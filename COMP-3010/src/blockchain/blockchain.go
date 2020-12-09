package blockchain

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ============================ Blockchain ============================

// Node defines required methods of any Blockchain Node implementation
type Node interface {
	NewBlockchain(com CommunicationComponent, p ProofComponent, con ConsensusComponent) *Node
	Run()

	initialize() error
	mine()
	getChain()
	createGenesisNode()
	terminate()
	initializeComponents() error
}

// Blockchain is the Blockchain object
type Blockchain struct {
	communicationComponent CommunicationComponent
	proofComponent         ProofComponent
	consensusComponent     ConsensusComponent
	genesisBlock           *Block
	blockchain             []Block
	mining                 bool
	wallet                 int
}

//ProofComponent standardizes methods for any Blockchain proof component
type ProofComponent interface {
	CalculateHash(nonce int, block Block) string
	ProofMethod(b Block, m bool) string
	ValidateProof(s string) bool
	Initialize() error
	Terminate()
}

// ConsensusComponent standardizes methods for any Blockchain consensus component
type ConsensusComponent interface {
	ConsensusMethod(c CommunicationComponent) ([]Block, error)
	Initialize() error
	Terminate()
}

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	Initialize() error
	InitializeWithPort(port int) error
	GetPeerChains() ([][]Block, error)
	RecieveFromNetwork() error
	BroadcastToNetwork(m Message) error
	SendToPeer(m Message, p Peer) error
	PingNetwork() error
	GenerateMessage(cmd string, d Data) (Message, error)
	GetMessageChannel() chan Message
	Terminate()
	PrunePeerNodes()
	GetMiddlewarePeer() Peer
}

// NewBlockchain creates and returns a new Blockchain, with the Genesis Block and Components initialized
func NewBlockchain(com CommunicationComponent, p ProofComponent, con ConsensusComponent) (Blockchain, error) {

	// Define a new Blockchain with the passed componenet values
	newBlockchain := Blockchain{communicationComponent: com, proofComponent: p, consensusComponent: con}

	// Initialize the Blockchain
	err := newBlockchain.initialize()

	// If there was an error initializing the Blockchain peer
	if err != nil {
		fmt.Printf("Error initializing Blockchain peer: %+v\n", err)
		newBlockchain.terminate()
		return Blockchain{}, nil
	}

	return newBlockchain, nil
}

// Initialize initializes the Blockchain by initializing its components and serving itself on the network
func (b *Blockchain) initialize() error {

	// Initialize Blockchain peer components
	err := b.initializeComponents()

	// If there was an error initializing one of the components
	if err != nil {
		fmt.Printf("Error initializing Blockchain peer: %+v\n", err)
		return err
	}

	// Run consensus to get latest copy of the chain from the network
	newChain, err := b.consensusComponent.ConsensusMethod(b.communicationComponent)
	// If there was an error during consensus
	if err != nil {
		fmt.Printf("Error running consensus during initialization: %+v\n", err)
		return err
	}

	b.blockchain = newChain

	// If the chain is empty after consensus, then this peer is the first node on the network
	if len(b.blockchain) == 0 {
		// Initialize the chain by creating the genesis block
		log.Println("No existing blockchain discovered, initializing new blockchain...")
		b.createGenesisBlock()
	}

	return nil
}

// createGenesisBlock initializes and adds a genesis block to the blockchain
func (b *Blockchain) createGenesisBlock() {

	genesisBlock := Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().String()
	genesisBlock.Data = Transaction{}
	genesisBlock.PrevHash = ""
	genesisBlock.Hash = b.proofComponent.ProofMethod(genesisBlock, true)

	b.genesisBlock = &genesisBlock
	b.blockchain = append(b.blockchain, genesisBlock)
}

// getChain returns this Blockchains current chain
func (b *Blockchain) getChain() []Block {
	return b.blockchain
}

// mine implements functionality to mine a new block to the chain
func (b *Blockchain) mine(data Data) Block {

	//Create a new block
	newBlock := Block{
		Index:     len(b.blockchain),
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  b.blockchain[len(b.blockchain)-1].Hash,
		Hash:      ""}

	//Calculate this block's proof
	newBlock.Hash = b.proofComponent.ProofMethod(newBlock, b.mining)

	return newBlock
}

// Run utilizes the blockchain components to run this blockchain peer by sending/recieving
// requests and messages on the p2p network
func (b *Blockchain) Run() {

	//When Run() concludes, terminate() will be called to clean up the different Blockchain components
	defer b.terminate()

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
	go func() {
		err := b.communicationComponent.PingNetwork()
		if err != nil {
			log.Printf("Error pinging network: %+v\n", err)
		}
	}()
	fmt.Println()

	var candidateBlock Block
	var newBlock Block
	b.mining = false

	for !done {

		go func() {
			err := b.communicationComponent.RecieveFromNetwork()
			if err != nil {
				log.Printf("Fatal Error recieving from network: %+v\n", err)
				done = true
			}
		}()

		select {
		case peerMsg := <-b.communicationComponent.GetMessageChannel():
			switch peerMsg.Command {

			case "PING":
				log.Printf("Recieved a ping from %s:%d\n", peerMsg.From.Address.IP.String(), peerMsg.From.Address.Port)

			case "MINE":

				// Start a new mining session
				newTransaction := peerMsg.Data
				b.mining = true
				candidateBlock = Block{}
				log.Println("Recieved a new transaction, beginning new mining session...")

				go func() {

					// Mine the new block
					newBlock = b.mine(newTransaction)
					b.mining = false

					// If the new block's hash isnt empty, then this peer succesffuly mined the block
					if newBlock.Hash != "" {
						log.Println("Block mined successfully")
						candidateBlock = newBlock

						toSend, err := b.communicationComponent.GenerateMessage("PROOF", nil)
						if err != nil {
							log.Printf("Error generating message: %v\n", err)
							return
						}

						log.Println("Sending proof to Middleware...")
						err = b.communicationComponent.SendToPeer(toSend, b.communicationComponent.GetMiddlewarePeer())
						if err != nil {
							log.Printf("Error sending message to peer: %v\n", err)
							return
						}
					}

				}()

			case "HALT":
				if b.mining {

					// If mining is true, then this peer was halted whilst still mining the block
					log.Println("Recieved a Halt, ending current mining session...")

					// Set mining to false which will end the mining session if its still in process,
					// as another peer has already successfully mined the block
					b.mining = false
				} else {

					log.Println("Recieved a Halt, but this peer's mining session already ended")

				}

				go func() {

					fmt.Printf("DEBUG - Chain before consensus: %+v\n", b.blockchain)

					// Run consensus to get the updated chain, in any case
					newChain, err := b.consensusComponent.ConsensusMethod(b.communicationComponent)
					if err != nil {
						log.Printf("Fatal Error running consensus method: %+v\n", err)
						done = true
					}

					b.blockchain = newChain

					fmt.Printf("DEBUG - Chain after consensus: %+v\n", b.blockchain)
				}()

			case "REWARD":

				// If this peer was the first peer to successfully mine the block, append the candidate block to this peer's blockchain
				// so that other nodes will get the block when consensus occurs
				log.Println("Appending new mined block to local chain")
				b.blockchain = append(b.blockchain, candidateBlock)

				// Add the reward that was sent to this peer for succesfully
				// mining the new block to this peer's wallet
				log.Println("Recieved a reward, adding amount to wallet... ")
				b.wallet += peerMsg.Data.(*Transaction).Amount
				log.Printf("Updated balance: %d\n", b.wallet)

			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")

			}

		default:
			//There was no message to read, thus do nothing
		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Minutes() >= 1 {
			log.Println("Blockchain peer sending pings...")
			go func() {
				err := b.communicationComponent.PingNetwork()
				if err != nil {
					log.Printf("Error pinging network: %+v\n", err)
				}
			}()
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
func (b Blockchain) terminate() {
	fmt.Println("\nTerminating Blockchain components...")

	b.communicationComponent.Terminate()
	b.proofComponent.Terminate()
	b.consensusComponent.Terminate()

	fmt.Println("Exiting Blockchain peer...")
}

func (b *Blockchain) initializeComponents() error {
	// Initialize the communication component
	err := b.communicationComponent.Initialize()

	// Initialize the consensus component
	err = b.consensusComponent.Initialize()

	// Initialize the proof component
	err = b.proofComponent.Initialize()

	// If there was an error initializing one of the components
	if err != nil {
		fmt.Printf("Error initializing Blockchain components: %+v", err)
		return err
	}

	return nil
}
