package types

// BlockHeader contains the metadata for the block.
type BlockHeader struct {
	Version   uint32
	PrevHash  [32]byte
	StateRoot [32]byte // Hash of the BadgerDB state after these txs are applied
	TxRoot    [32]byte // Merkle Root of the transactions
	Timestamp uint64
	Number    uint64  // Block height
	Signer    Address // The PoA Validator who created this block
	Signature []byte  // Validator's signature of the header
}

// Block is a complete block containing the header and all transactions.
type Block struct {
	Header       BlockHeader
	Transactions []Transaction
}
