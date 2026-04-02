package network

import (
	"github.com/iamaer4a/enterprise-chain/core/types"
)

// MessageType defines the kind of data being sent over TCP.
type MessageType byte

const (
	MsgTx MessageType = iota
	MsgBlock
)

// Envelope wraps our network data so the receiver knows how to decode it.
type Envelope struct {
	Type MessageType
	Data []byte
}

// TxMessage is sent when a node wants to gossip a transaction.
type TxMessage struct {
	Transaction types.Transaction
}

// BlockMessage is sent when a validator mints a new block.
type BlockMessage struct {
	Block types.Block
}
