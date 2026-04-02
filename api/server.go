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

type Server struct {
	listenAddr string
	bc         *core.Blockchain
	pool       *mempool.Mempool
}

func NewServer(addr string, bc *core.Blockchain, pool *mempool.Mempool) *Server {
	return &Server{
		listenAddr: addr,
		bc:         bc,
		pool:       pool,
	}
}

// --- NEW: Universal CORS Helper ---
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) Start() error {
	// Ensure the routes are mapped to the new endpoints
	http.HandleFunc("/account", s.handleGetAccount)
	http.HandleFunc("/tx", s.handleSubmitTx)
	http.HandleFunc("/status", s.handleStatus)

	log.Printf("REST API listening on http://%s\n", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, nil)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	// Handle browser preflight checks
	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{
		"height":      s.bc.GetTipHeight(),
		"pending_txs": len(s.pool.GetPending()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

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

	// Fetch the FULL account from the database
	acc, err := s.bc.StateDB().GetAccount(address)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	// Convert the byte storage to readable strings
	readableStorage := make(map[string]string)
	for k, v := range acc.Storage {
		readableStorage[k] = string(v)
	}

	response := map[string]interface{}{
		"address": addrHex,
		"balance": acc.Balance.String(),
		"nonce":   acc.Nonce,
		"storage": readableStorage,
		"history": acc.TxHistory,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSubmitTx(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
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

	if err := s.pool.Add(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Transaction %x accepted into mempool", tx.Hash[:4])
}
