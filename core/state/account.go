package state

import (
	"math/big"

	"github.com/iamaer4a/enterprise-chain/core/types"
)

// Account represents a single participant's state in the network.
type Account struct {
	Address types.Address
	Nonce   uint64
	Balance *big.Int
	Storage map[string][]byte // For smart contract state
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
