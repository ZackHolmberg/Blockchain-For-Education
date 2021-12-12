package blockchain

// ============================ Block ============================

// Block is the Block object
type Block struct {
	Index     int
	Timestamp string
	Data      Data
	PrevHash  string
	Hash      string
	Nonce     int
}
