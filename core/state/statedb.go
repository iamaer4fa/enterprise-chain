package state

import (
	"encoding/json"

	"github.com/iamaer4a/enterprise-chain/core/types"
	"github.com/iamaer4a/enterprise-chain/database"
)

// StateDB manages account states using BadgerDB.
type StateDB struct {
	db *database.Store
}

// NewStateDB initializes the state manager.
func NewStateDB(db *database.Store) *StateDB {
	return &StateDB{db: db}
}

// GetAccount retrieves an account from the database, or returns a new one if it doesn't exist.
func (s *StateDB) GetAccount(addr types.Address) (*Account, error) {
	data, err := s.db.Get(addr[:])
	if err != nil {
		// If not found, return a fresh account
		return NewAccount(addr), nil
	}

	var account Account
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

// SaveAccount writes the updated account state to the database.
func (s *StateDB) SaveAccount(account *Account) error {
	data, err := json.Marshal(account)
	if err != nil {
		return err
	}
	return s.db.Put(account.Address[:], data)
}
