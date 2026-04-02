package crypto

import (
	"crypto/sha256"

	"github.com/iamaer4a/enterprise-chain/core/types"
)

// ComputeMerkleRoot generates the root hash for a list of transactions.
func ComputeMerkleRoot(txs []types.Transaction) [32]byte {
	if len(txs) == 0 {
		return [32]byte{}
	}

	// Extract the hashes of all transactions
	var nodes [][32]byte
	for _, tx := range txs {
		var hash [32]byte
		copy(hash[:], tx.Hash)
		nodes = append(nodes, hash)
	}

	// Recursively hash pairs until 1 root remains
	for len(nodes) > 1 {
		if len(nodes)%2 != 0 {
			// If odd number of nodes, duplicate the last one
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		var nextLevel [][32]byte
		for i := 0; i < len(nodes); i += 2 {
			combined := append(nodes[i][:], nodes[i+1][:]...)
			hash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, hash)
		}
		nodes = nextLevel
	}

	return nodes[0]
}
