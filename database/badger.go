package database

import (
	"log"

	"github.com/dgraph-io/badger/v4"
)

// Store wraps the BadgerDB instance.
type Store struct {
	db *badger.DB
}

// NewStore initializes the BadgerDB at the given path.
func NewStore(dbPath string) (*Store, error) {
	opts := badger.DefaultOptions(dbPath)
	// Disable logging for cleaner CLI output, or customize it
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// Put saves a key-value pair to the database.
func (s *Store) Put(key, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get retrieves a value by its key.
func (s *Store) Get(key []byte) ([]byte, error) {
	var valCopy []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return err
	})
	return valCopy, err
}

// Close safely shuts down the database.
func (s *Store) Close() {
	if err := s.db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}
