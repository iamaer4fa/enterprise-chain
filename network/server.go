package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/iamaer4a/enterprise-chain/core"
	"github.com/iamaer4a/enterprise-chain/core/mempool"
)

// Server handles P2P connections and message routing.
type Server struct {
	ListenAddr string
	Listener   net.Listener
	Mempool    *mempool.Mempool
	Blockchain *core.Blockchain

	mu    sync.RWMutex
	peers map[net.Conn]bool
}

// NewServer creates a new P2P network node.
func NewServer(addr string, pool *mempool.Mempool, bc *core.Blockchain) *Server {
	return &Server{
		ListenAddr: addr,
		Mempool:    pool,
		Blockchain: bc,
		peers:      make(map[net.Conn]bool),
	}
}

// Start boots up the TCP server and begins accepting connections.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.Listener = ln
	log.Printf("P2P Server listening on %s\n", s.ListenAddr)

	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v\n", err)
			continue
		}

		s.mu.Lock()
		s.peers[conn] = true
		s.mu.Unlock()

		log.Printf("New peer connected: %s\n", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.peers, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		var env Envelope
		decoder := gob.NewDecoder(conn)
		if err := decoder.Decode(&env); err != nil {
			log.Printf("Connection closed or decode error: %v\n", err)
			break
		}

		s.routeMessage(env)
	}
}

func (s *Server) routeMessage(env Envelope) {
	decoder := gob.NewDecoder(bytes.NewReader(env.Data))

	switch env.Type {
	case MsgTx:
		var txMsg TxMessage
		if err := decoder.Decode(&txMsg); err == nil {
			log.Println("Received new transaction via gossip protocol")
			s.Mempool.Add(txMsg.Transaction)
		}
	case MsgBlock:
		var blockMsg BlockMessage
		if err := decoder.Decode(&blockMsg); err == nil {
			log.Printf("Received new block: %d via gossip protocol\n", blockMsg.Block.Header.Number)

			// --- NEW: Pass the block to the state machine! ---
			if err := s.Blockchain.ProcessBlock(blockMsg.Block); err != nil {
				log.Printf("Gossiped block rejected: %v\n", err)
			}
			// -------------------------------------------------
		}
	default:
		log.Println("Received unknown message type")
	}
}

// Broadcast sends an Envelope to all connected peers (The Gossip Protocol).
func (s *Server) Broadcast(env Envelope) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for peer := range s.peers {
		encoder := gob.NewEncoder(peer)
		if err := encoder.Encode(env); err != nil {
			fmt.Printf("Failed to broadcast to peer %s: %v\n", peer.RemoteAddr(), err)
		}
	}
}

// ConnectToPeer dials another node and establishes a two-way gossip channel.
func (s *Server) ConnectToPeer(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.peers[conn] = true
	s.mu.Unlock()

	log.Printf("Successfully connected to peer: %s\n", addr)

	// Start listening to this peer in the background
	go s.handleConnection(conn)
	return nil
}
