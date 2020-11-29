package blockchain

import (
	"time"
)

// ============================ Blockchain ============================

//TODO: Could have a Blockchain interface with following methods:
// - Mine()
// - GetChain()
// - CreateGenesisNode()
// - Run()

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
	ProofMethod(b Block) string
	ValidateProof(s string) bool
}

// ConsensusComponent standardizes methods for any Blockchain consensus component
type ConsensusComponent interface {
	ConsensusMethod()
}

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	InitializeCommunicator()
	GetPeerChains()
	RecieveFromNetwork() (Message, error)
	BroadcastToNetwork(msg Message)
	PingNetwork()
	HandlePingFromNetwork()
}

// NewBlockchain creates and returns a new Blockchain, with the Genesis Block initialized
func NewBlockchain(com CommunicationComponent, p ProofComponent, con ConsensusComponent) Blockchain {

	// Initialize a new Blockchain with the passed componenet values
	newBlockcain := Blockchain{CommunicationComponent: com, ProofComponent: p, ConsensusComponent: con}

	// Initialize the communication component
	com.InitializeCommunicator()

	// Ping the network so that this new peer is discovered by all existing peers
	com.PingNetwork()

	// Run consensus to get latest copy of the chain from the network
	con.ConsensusMethod()

	// If the chain is empty after consensus, then this peer is the first node on the network
	if len(newBlockcain.Blockchain) == 0 {
		// Initialize the chain by creating the genesis block
		newBlockcain.CreateGenesisBlock()
	}

	return newBlockcain
}

// CreateGenesisBlock initializes and adds a genesis block to the blockchain
func (b *Blockchain) CreateGenesisBlock() {

	genesisBlock := Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().String()
	genesisBlock.Data = Transaction{}
	genesisBlock.PrevHash = ""
	genesisBlock.Hash = b.ProofComponent.ProofMethod(genesisBlock)

	b.GenesisBlock = &genesisBlock
	b.Blockchain = append(b.Blockchain, genesisBlock)
}

// GetChain returns this Blockchains current chain
func (b *Blockchain) GetChain() []Block {
	return b.Blockchain
}

// Mine implements functionality to mine a new block to the chain
func (b *Blockchain) Mine(data Data) {

	//Create a new block
	newBlock := Block{
		Index:     len(b.Blockchain),
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  b.Blockchain[len(b.Blockchain)-1].Hash,
		Hash:      ""}

	//Calculate this block's proof
	newBlock.Hash = b.ProofComponent.ProofMethod(newBlock)

	//Add the new block to the chain
	b.Blockchain = append(b.Blockchain, newBlock)
}

// Run uses the 3 blockchain components to run this blockchain peer by sending/recieving
// requests and messages on the p2p network
func (b Blockchain) Run() {

}

// TODO: Consider a blockchain clean up function when program ends.
// 	     Would call the each components clean up and exit function,
//       if there is a need for such functions. These would be
//       interface-defined functions.
