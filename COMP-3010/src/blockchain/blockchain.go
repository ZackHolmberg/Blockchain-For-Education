package blockchain

import (
	"fmt"
	"time"
)

// ============================ Blockchain ============================

// Interface defines required methods of any Blockchain implementation
type Interface interface {
	NewBlockchain()
	Run()

	mine()
	getChain()
	createGenesisNode()
	terminate()
}

// Blockchain is the Blockchain object
type Blockchain struct {
	communicationComponent CommunicationComponent
	proofComponent         ProofComponent
	consensusComponent     ConsensusComponent
	genesisBlock           *Block
	blockchain             []Block
}

//ProofComponent standardizes methods for any Blockchain proof component
type ProofComponent interface {
	CalculateHash(nonce int, block Block) string
	ProofMethod(b Block) string
	ValidateProof(s string) bool
	TerminateProofComponent()
}

// ConsensusComponent standardizes methods for any Blockchain consensus component
type ConsensusComponent interface {
	ConsensusMethod()
	TerminateConsensusComponent()
}

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	InitializeCommunicator()
	GetPeerChains()
	RecieveFromNetwork() (Message, error)
	BroadcastToNetwork(msg Message)
	PingNetwork()
	TerminateCommunicationComponent()
}

// NewBlockchain creates and returns a new Blockchain, with the Genesis Block initialized
func NewBlockchain(com CommunicationComponent, p ProofComponent, con ConsensusComponent) Blockchain {

	// Initialize a new Blockchain with the passed componenet values
	newBlockcain := Blockchain{communicationComponent: com, proofComponent: p, consensusComponent: con}

	// Initialize the communication component
	com.InitializeCommunicator()

	// Ping the network so that this new peer is discovered by all existing peers
	com.PingNetwork()

	// Run consensus to get latest copy of the chain from the network
	con.ConsensusMethod()

	// If the chain is empty after consensus, then this peer is the first node on the network
	if len(newBlockcain.blockchain) == 0 {
		// Initialize the chain by creating the genesis block
		newBlockcain.createGenesisBlock()
	}

	return newBlockcain
}

// createGenesisBlock initializes and adds a genesis block to the blockchain
func (b *Blockchain) createGenesisBlock() {

	genesisBlock := Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().String()
	genesisBlock.Data = Transaction{}
	genesisBlock.PrevHash = ""
	genesisBlock.Hash = b.proofComponent.ProofMethod(genesisBlock)

	b.genesisBlock = &genesisBlock
	b.blockchain = append(b.blockchain, genesisBlock)
}

// getChain returns this Blockchains current chain
func (b *Blockchain) getChain() []Block {
	return b.blockchain
}

// mine implements functionality to mine a new block to the chain
func (b *Blockchain) mine(data Data) {

	//Create a new block
	newBlock := Block{
		Index:     len(b.blockchain),
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  b.blockchain[len(b.blockchain)-1].Hash,
		Hash:      ""}

	//Calculate this block's proof
	newBlock.Hash = b.proofComponent.ProofMethod(newBlock)

	//Add the new block to the chain
	b.blockchain = append(b.blockchain, newBlock)
}

// Run uses the 3 blockchain components to run this blockchain peer by sending/recieving
// requests and messages on the p2p network
func (b Blockchain) Run() {

	//When Run() concludes, terminate() will be called to clean up the different Blockchain components
	defer b.terminate()

	b.communicationComponent.PingNetwork()
	var lastPing = time.Now()

	for {
		msg, err := b.communicationComponent.RecieveFromNetwork()
		if err != nil {

			fmt.Printf("Error receiving from network: %v", err)

			// continue forces the program to enter the next iteration of
			// the loop, skipping all code in the remainder of the loop
			continue
		}

		switch msg.command {
		case "PING":
			go fmt.Printf("Recieved a ping from %#v\n", msg.from)
		default:
			go fmt.Println("Warning - Command \"" + msg.command + "\" not supported")
		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Minutes() >= 1 {
			go b.communicationComponent.PingNetwork()
			lastPing = time.Now()
		}

		// If this peer hasn't received a message from another peer for 75 seconds,
		// then remove that peer from the list of known nodes

	}
}

// terminate calls all of the interface-defined component clean-up methods
func (b Blockchain) terminate() {
	b.communicationComponent.TerminateCommunicationComponent()
	b.proofComponent.TerminateProofComponent()
	b.consensusComponent.TerminateConsensusComponent()
}
