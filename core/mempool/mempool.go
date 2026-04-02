package mempool

import (
	"fmt"
	"sync"

	"github.com/iamaer4a/enterprise-chain/core/types"
)

// Mempool stores unconfirmed transactions.
type Mempool struct {
	mu           sync.RWMutex
	transactions map[string]types.Transaction
}

// NewMempool initializes an empty mempool.
func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[string]types.Transaction),
	}
}

// Add inserts a new transaction into the pool.
// In a production system, you would validate the ECDSA signature and account balance here first.
func (mp *Mempool) Add(tx types.Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	hashHex := fmt.Sprintf("%x", tx.Hash)
	if _, exists := mp.transactions[hashHex]; exists {
		return fmt.Errorf("transaction %s already in mempool", hashHex)
	}

	mp.transactions[hashHex] = tx
	return nil
}

// GetPending returns all transactions currently in the pool.
func (mp *Mempool) GetPending() []types.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var txs []types.Transaction
	for _, tx := range mp.transactions {
		txs = append(txs, tx)
	}
	return txs
}

// Clear removes specific transactions from the pool (usually called after a block is minted).
func (mp *Mempool) Clear(txs []types.Transaction) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, tx := range txs {
		hashHex := fmt.Sprintf("%x", tx.Hash)
		delete(mp.transactions, hashHex)
	}
}
