package blockchain

import (
	"fmt"
	"log"
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
	RecieveFromNetwork()
	BroadcastToNetwork(msg Message)
	PingNetwork()
	GetMessageChannel() chan Message
	TerminateCommunicationComponent()
}

// NewBlockchain creates and returns a new Blockchain, with the Genesis Block initialized
func NewBlockchain(com CommunicationComponent, p ProofComponent, con ConsensusComponent) Blockchain {

	// Initialize a new Blockchain with the passed componenet values
	newBlockchain := Blockchain{communicationComponent: com, proofComponent: p, consensusComponent: con}

	// Initialize the communication component
	com.InitializeCommunicator()

	// Run consensus to get latest copy of the chain from the network
	con.ConsensusMethod()

	// If the chain is empty after consensus, then this peer is the first node on the network
	if len(newBlockchain.blockchain) == 0 {
		// Initialize the chain by creating the genesis block
		newBlockchain.createGenesisBlock()
	}

	return newBlockchain
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

	var lastPing = time.Now()

	fmt.Println("\nRunning Blockchain...")
	fmt.Println()

	for {

		go b.communicationComponent.RecieveFromNetwork()

		select {
		case peerMsg := <-b.communicationComponent.GetMessageChannel():
			switch peerMsg.Command {
			case "PING":
				go log.Printf("Recieved a ping from %s:%d\n", peerMsg.From.Address.IP.String(), peerMsg.From.Address.Port)
			default:
				go log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			// Ping all peer nodes on the network once every minute
			// if time.Since(lastPing).Minutes() >= 1 {
			if time.Since(lastPing).Seconds() >= 5 {
				go b.communicationComponent.PingNetwork()
				lastPing = time.Now()
			}
		}

		time.Sleep(1 * time.Millisecond)

		// TODO: If this peer hasn't received a message from another peer for 75 seconds,
		// then remove that peer from the list of known nodes

		// TODO: Need to set a handler/listener in here to catch ctrl+c quitting, which calls exits this loop
		// so that terminate can run

	}
}

// terminate calls all of the interface-defined component clean-up methods
func (b Blockchain) terminate() {
	fmt.Println("Terminating Blockchain components...")

	b.communicationComponent.TerminateCommunicationComponent()
	b.proofComponent.TerminateProofComponent()
	b.consensusComponent.TerminateConsensusComponent()

	fmt.Println("Exiting...")

}
