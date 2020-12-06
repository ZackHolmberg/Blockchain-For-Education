package blockchain

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ============================ Blockchain ============================

// Interface defines required methods of any Blockchain implementation
type Interface interface {
	NewBlockchain()
	Run()

	initialize()
	mine()
	getChain()
	createGenesisNode()
	terminate()
	initializeComponents()
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
	Initialize()
	Terminate()
}

// ConsensusComponent standardizes methods for any Blockchain consensus component
type ConsensusComponent interface {
	ConsensusMethod()
	Initialize()
	Terminate()
}

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	Initialize()
	InitializeWithPort(port int)
	GetPeerChains()
	RecieveFromNetwork()
	BroadcastToNetwork(msg Message)
	PingNetwork()
	GetMessageChannel() chan Message
	Terminate()
	PrunePeerNodes()
}

// NewBlockchain creates and returns a new Blockchain, with the Genesis Block and Components initialized
func NewBlockchain(com CommunicationComponent, p ProofComponent, con ConsensusComponent) Blockchain {

	// Define a new Blockchain with the passed componenet values
	newBlockchain := Blockchain{communicationComponent: com, proofComponent: p, consensusComponent: con}

	// Initialize the Blockchain
	newBlockchain.initialize()

	return newBlockchain
}

// Initialize initializes the Blockchain by initializing its components and serving itself on the network
func (b *Blockchain) initialize() {

	// Initialize the communication component
	b.communicationComponent.Initialize()

	// Initialize the consensus component
	b.consensusComponent.Initialize()

	// Initialize the proof component
	b.proofComponent.Initialize()

	// Run consensus to get latest copy of the chain from the network
	b.consensusComponent.ConsensusMethod()

	// If the chain is empty after consensus, then this peer is the first node on the network
	if len(b.blockchain) == 0 {
		// Initialize the chain by creating the genesis block
		b.createGenesisBlock()
	}
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

// Run utilizes the blockchain components to run this blockchain peer by sending/recieving
// requests and messages on the p2p network
func (b Blockchain) Run(wg *sync.WaitGroup) {

	//When Run() concludes, terminate() will be called to clean up the different Blockchain components
	defer b.terminate(wg)

	// Sets done equal to true if the user exits the program with ctrl+c, which will case the loop to finish and Run() to exit,
	// which will cause terminate() to run

	c := make(chan os.Signal)
	done := false
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		done = true
	}()

	fmt.Println("\nRunning Blockchain...")

	log.Println("Announcing self to peers...")

	var lastPing = time.Now()
	go b.communicationComponent.PingNetwork()

	fmt.Println()

	for !done {

		go b.communicationComponent.RecieveFromNetwork()

		select {
		case peerMsg := <-b.communicationComponent.GetMessageChannel():
			switch peerMsg.Command {
			case "PING":
				log.Printf("Recieved a ping from %s:%d\n", peerMsg.From.Address.IP.String(), peerMsg.From.Address.Port)
			case "MINE":
				// Start mining the new block
				// If the block is successfully mined by another node first, this node will receice
				// a message to quit mining the particular block and wait for the next
			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			//There was no message to read, thus do nothing
		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Seconds() >= 10 {
			log.Println("Blockchain peer sending pings...")
			go b.communicationComponent.PingNetwork()
			lastPing = time.Now()
		}

		// If this peer hasn't received a message from another peer for 75 seconds,
		// then remove that peer from the list of known nodes
		go b.communicationComponent.PrunePeerNodes()

		// Timeout for 1 millisecond to limit the number of iterations of the loop to 1 per ms
		time.Sleep(1 * time.Millisecond)

	}
}

// terminate calls all of the interface-defined component clean-up methods
func (b Blockchain) terminate(wg *sync.WaitGroup) {
	fmt.Println("\nTerminating Blockchain components...")

	b.communicationComponent.Terminate()
	b.proofComponent.Terminate()
	b.consensusComponent.Terminate()

	fmt.Println("Exiting Blockchain peer...")

	wg.Done()

}

func (b *Blockchain) initializeComponents() {
	b.communicationComponent.Initialize()
	b.consensusComponent.Initialize()
	b.proofComponent.Initialize()
}
