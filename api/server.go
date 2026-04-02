package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/iamaer4a/enterprise-chain/core"
	"github.com/iamaer4a/enterprise-chain/core/mempool"
	"github.com/iamaer4a/enterprise-chain/core/types"
)

// Server exposes the blockchain data via HTTP.
type Server struct {
	listenAddr string
	bc         *core.Blockchain
	pool       *mempool.Mempool
}

// NewServer initializes the API server.
func NewServer(addr string, bc *core.Blockchain, pool *mempool.Mempool) *Server {
	return &Server{
		listenAddr: addr,
		bc:         bc,
		pool:       pool,
	}
}

// Start boots up the HTTP server.
func (s *Server) Start() error {
	http.HandleFunc("/balance", s.handleGetBalance)
	http.HandleFunc("/tx", s.handleSubmitTx)
	http.HandleFunc("/status", s.handleStatus)

	log.Printf("REST API listening on http://%s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, nil)
}

// NEW: handleStatus returns the current chain height and mempool count
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Allow cross-origin requests from our HTML file
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"height":      s.bc.GetTipHeight(),
		"pending_txs": len(s.pool.GetPending()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetBalance returns the current balance of an address.
func (s *Server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	// Allow cross-origin requests from our HTML file
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addrHex := r.URL.Query().Get("address")
	if addrHex == "" {
		http.Error(w, "Missing address parameter", http.StatusBadRequest)
		return
	}

	address := types.HexToAddress(addrHex)

	// Call the actual database via the Blockchain helper method
	balance, err := s.bc.GetBalance(address)
	if err != nil {
		http.Error(w, "Failed to read balance", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"address": addrHex,
		"balance": balance.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSubmitTx accepts a signed transaction and adds it to the mempool.
// Example: POST /tx with JSON body
func (s *Server) handleSubmitTx(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tx types.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid transaction payload", http.StatusBadRequest)
		return
	}

	// 1. Verify the ECDSA signature here using core/crypto
	// 2. Add to mempool
	if err := s.pool.Add(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Transaction %x accepted into mempool", tx.Hash[:4])

}
