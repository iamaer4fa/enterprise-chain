package types

import (
	"encoding/binary"
	"math/big"
)

// Transaction represents a transfer of value or state change.
type Transaction struct {
	Sender    Address  `json:"sender"`
	Recipient Address  `json:"recipient"`
	Amount    *big.Int `json:"amount"`
	Nonce     uint64   `json:"nonce"`
	Data      []byte   `json:"data"` // Smart Contract Payload
	Signature []byte   `json:"signature"`
	Hash      []byte   `json:"hash"`
}

// Payload returns the data that needs to be signed (excluding the signature and hash itself).
// In a real build, you would use an RLP (Recursive Length Prefix) or Protocol Buffers encoder here.
func (tx *Transaction) Payload() []byte {
	// Placeholder: Concatenate fields to byte slice
	var data []byte
	data = append(data, tx.Sender[:]...)
	data = append(data, tx.Recipient[:]...)
	if tx.Amount != nil {
		data = append(data, tx.Amount.Bytes()...)
	}
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, tx.Nonce)
	data = append(data, nonceBytes...)
	if len(tx.Data) > 0 {
		data = append(data, tx.Data...)
	}
	return data
}
