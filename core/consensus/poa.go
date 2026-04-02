package consensus

import (
	"sync"

	"github.com/iamaer4a/enterprise-chain/core/types"
)

// PoAEngine manages the list of authorized signers for the private network.
type PoAEngine struct {
	mu         sync.RWMutex
	validators map[types.Address]bool
}

// NewPoAEngine initializes the consensus engine with the genesis validators.
func NewPoAEngine(initialValidators []types.Address) *PoAEngine {
	engine := &PoAEngine{
		validators: make(map[types.Address]bool),
	}
	for _, v := range initialValidators {
		engine.validators[v] = true
	}
	return engine
}

// IsAuthorized checks if a given address is allowed to sign and mint blocks.
func (poa *PoAEngine) IsAuthorized(address types.Address) bool {
	poa.mu.RLock()
	defer poa.mu.RUnlock()
	return poa.validators[address]
}

// AddValidator allows the network to vote in a new authority (simplified).
func (poa *PoAEngine) AddValidator(address types.Address) {
	poa.mu.Lock()
	defer poa.mu.Unlock()
	poa.validators[address] = true
}

// RemoveValidator revokes signing privileges from an address.
func (poa *PoAEngine) RemoveValidator(address types.Address) {
	poa.mu.Lock()
	defer poa.mu.Unlock()
	delete(poa.validators, address)
}
