package types

import (
	"encoding/hex"
	"fmt"
)

// Address represents the 20-byte address of an account.
type Address [20]byte

// HexToAddress converts a hex string to an Address.
func HexToAddress(s string) Address {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 20 {
		// In a production app, handle this error properly.
		panic(fmt.Sprintf("invalid address hex: %s", s))
	}
	var a Address
	copy(a[:], b)
	return a
}

// String returns the hex representation of the address.
func (a Address) String() string {
	return hex.EncodeToString(a[:])
}
