package blockchain

// ============================ Blockchain ============================

// Blockchain is the Blockchain object
type Blockchain struct {
	CommunicationComponent CommunicationComponent
	ProofComponent         ProofComponent
	ConsensusComponent     ConsensusComponent
	GenesisBlock           *Block
}

// CreateGenesisBlock add the genesis block to the blockchain
func (b Blockchain) CreateGenesisBlock() {
	block := &Block{}
	b.GenesisBlock = block
}
