package state

import (
	"math/big"

	"github.com/iamaer4a/enterprise-chain/core/types"
)

// Account represents a single participant's state in the network.
type Account struct {
	Address   types.Address     `json:"address"`
	Nonce     uint64            `json:"nonce"`
	Balance   *big.Int          `json:"balance"`
	Storage   map[string][]byte `json:"storage"` // For smart contract state
	TxHistory []string          `json:"tx_history"`
}

// NewAccount initializes a new account with a zero balance.
func NewAccount(address types.Address) *Account {
	return &Account{
		Address: address,
		Nonce:   0,
		Balance: big.NewInt(0),
	}
}

// AddBalance securely adds funds to the account.
func (a *Account) AddBalance(amount *big.Int) {
	a.Balance.Add(a.Balance, amount)
}

// SubBalance subtracts funds, returning false if insufficient funds.
func (a *Account) SubBalance(amount *big.Int) bool {
	if a.Balance.Cmp(amount) < 0 {
		return false // Insufficient funds
	}
	a.Balance.Sub(a.Balance, amount)
	return true
}
