package blockchain

import (
	"time"
)

// ============================ Blockchain ============================

// Blockchain is the Blockchain object
type Blockchain struct {
	CommunicationComponent CommunicationComponent
	ProofComponent         ProofComponent
	ConsensusComponent     ConsensusComponent
	GenesisBlock           *Block
	Blockchain             []Block
}

//ProofComponent standardizes methods for any Blockchain proof component
type ProofComponent interface {
	CalculateHash(nonce int, block Block) string
	ProofMethod()
	ValidateProof() bool
}

// ConsensusComponent standardizes methods for any Blockchain consensus component
type ConsensusComponent interface {
	ConsensusMethod()
}

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	GetPeerChains()
	RecieveFromClient()
	SendToClient()
	RecieveFromNetwork()
	BroadcastToNetwork()
}

// NewBlockchain creates and returns a new Blockchain
func NewBlockchain() Blockchain {
	newBlockcain := Blockchain{}
	//TODO: Implement
	return newBlockcain
}

// CreateGenesisBlock add the genesis block to the blockchain
func (b *Blockchain) CreateGenesisBlock() {

	genesisBlock := Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().String()
	genesisBlock.Data = Transaction{}
	genesisBlock.PrevHash = "0"
	genesisBlock.Hash = b.ProofComponent.CalculateHash(0, genesisBlock)

	b.GenesisBlock = &genesisBlock
	b.Blockchain = append(b.Blockchain, genesisBlock)
}

// Mine implements functionality to mine a new block to the chain
func (b *Blockchain) Mine(block Block) {
	//TODO: Implement
}
