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

//ProofComponent standardizes methods for any Blockchain proof component
type ProofComponent interface {
	ProofMethod()
	ValidateProof()
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
