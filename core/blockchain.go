package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/iamaer4a/enterprise-chain/core/consensus"
	"github.com/iamaer4a/enterprise-chain/core/crypto"
	"github.com/iamaer4a/enterprise-chain/core/state"
	"github.com/iamaer4a/enterprise-chain/core/types"
	"github.com/iamaer4a/enterprise-chain/database"
)

// Blockchain represents the active ledger.
type Blockchain struct {
	db        *database.Store
	stateDB   *state.StateDB
	poa       *consensus.PoAEngine
	tipHash   [32]byte
	tipHeight uint64
}

// NewBlockchain initializes the chain.
func NewBlockchain(db *database.Store, poa *consensus.PoAEngine) *Blockchain {
	return &Blockchain{
		db:      db,
		stateDB: state.NewStateDB(db),
		poa:     poa,
	}
}

// ProcessBlock is the main entry point for blocks received from the network.
func (bc *Blockchain) ProcessBlock(block types.Block) error {
	// 1. Validate the Block Header & PoA Consensus
	if err := bc.validateHeader(block.Header); err != nil {
		return fmt.Errorf("block header validation failed: %v", err)
	}

	// 2. Validate the Merkle Root
	computedTxRoot := crypto.ComputeMerkleRoot(block.Transactions)
	if !bytes.Equal(computedTxRoot[:], block.Header.TxRoot[:]) {
		return errors.New("merkle root mismatch: transaction data is corrupted")
	}

	// 3. Execute Transactions and Transition State
	if err := bc.executeTransactions(block.Transactions); err != nil {
		return fmt.Errorf("transaction execution failed: %v", err)
	}

	// 4. Save Block to Database (Simplified: In production, serialize block first)
	// Example: bc.db.Put(block.Header.Hash(), serializedBlock)

	// 5. Update Chain Tip
	bc.tipHeight = block.Header.Number
	// bc.tipHash = block.Header.Hash() // Assuming a Hash() method exists on Header

	for _, tx := range block.Transactions {
		// ... signature validation ...
		// ... deduct sender balance ...
		// ... update sender nonce ...

		// --- CHECK YOUR FILE FOR THIS BLOCK ---
		// If the transaction contains data, process it as a Smart Contract
		if len(tx.Data) > 0 {
			type ContractCall struct {
				Method string `json:"method"`
				Key    string `json:"key"`
				Value  string `json:"value"`
			}

			var call ContractCall
			if err := json.Unmarshal(tx.Data, &call); err == nil {
				if call.Method == "SetRecord" {
					contractAcc, _ := bc.stateDB.GetAccount(tx.Recipient)
					if contractAcc.Storage == nil {
						contractAcc.Storage = make(map[string][]byte)
					}

					contractAcc.Storage[call.Key] = []byte(call.Value)
					bc.stateDB.SaveAccount(contractAcc)

					log.Printf("Contract Executed: Stored [%s] = [%s] at address %x\n", call.Key, call.Value, tx.Recipient[:4])
				}
			}
		}
		// --------------------------------------
	}

	log.Printf("Block %d processed and appended successfully by validator %x\n", block.Header.Number, block.Header.Signer[:4])
	return nil
}

// GetBalance safely queries the state database for an account's balance.
func (bc *Blockchain) GetBalance(addr types.Address) (*big.Int, error) {
	// StateDB's GetAccount handles returning a fresh account (0 balance)
	// if the address doesn't exist in BadgerDB yet.
	account, err := bc.stateDB.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	return account.Balance, nil
}

// StateDB exposes the underlying state manager so external packages
// (like main) can access the accounts directly if needed.
func (bc *Blockchain) StateDB() *state.StateDB {
	return bc.stateDB
}

// GetTipHeight returns the height of the most recently accepted block.
func (bc *Blockchain) GetTipHeight() uint64 {
	return bc.tipHeight
}

// validateHeader ensures the block was created by an authorized validator.
func (bc *Blockchain) validateHeader(header types.BlockHeader) error {
	// Is the signer an authorized Authority?
	if !bc.poa.IsAuthorized(header.Signer) {
		return errors.New("block signer is not an authorized validator")
	}

	// In a complete implementation, you would verify the ECDSA signature of the header here:
	// if !crypto.VerifySignature(header.SignerPubKey, header.HashBytes(), header.Signature) { ... }

	// Check if this block strictly follows our current tip (Simple Fork Resolution)
	if header.Number != bc.tipHeight+1 {
		return fmt.Errorf("block height %d does not match expected height %d", header.Number, bc.tipHeight+1)
	}

	return nil
}

// executeTransactions sequentially applies each transaction to the state database.
func (bc *Blockchain) executeTransactions(txs []types.Transaction) error {
	for _, tx := range txs {
		// Load sender and recipient
		sender, err := bc.stateDB.GetAccount(tx.Sender)
		if err != nil {
			return err
		}
		recipient, err := bc.stateDB.GetAccount(tx.Recipient)
		if err != nil {
			return err
		}

		// Verify Nonce (Replay Protection)
		if tx.Nonce != sender.Nonce {
			return fmt.Errorf("invalid nonce for sender %x: expected %d, got %d", tx.Sender[:4], sender.Nonce, tx.Nonce)
		}

		// Verify Balance
		if !sender.SubBalance(tx.Amount) {
			return fmt.Errorf("insufficient funds for sender %x", tx.Sender[:4])
		}

		// Update State
		recipient.AddBalance(tx.Amount)
		sender.Nonce++ // Increment nonce after successful execution

		// Save updated accounts back to BadgerDB
		bc.stateDB.SaveAccount(sender)
		bc.stateDB.SaveAccount(recipient)
	}

	return nil
}
