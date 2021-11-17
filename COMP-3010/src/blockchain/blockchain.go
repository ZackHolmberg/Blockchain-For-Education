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
	createGenesisNode()
	terminate()
	initializeComponents() error
}

// Blockchain is the Blockchain object
type Blockchain struct {
	communicationComponent CommunicationComponent
	proofComponent         ProofComponent
	consensusComponent     ConsensusComponent
	chain                  []Block
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
	ConsensusMethod(peerChains [][]Block, currChain []Block) ([]Block, error)
	Initialize() error
	Terminate()
}

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	Initialize() error
	InitializeWithPort(port int) error
	GetPeerNodes() []Peer
	GetMiddlewarePeer() Peer
	GetMessageChannel() chan Message
	RecieveFromNetwork(withTimeout bool) error
	GenerateMessage(cmd string, data Data) (Message, error)
	BroadcastMsgToNetwork(cmd string, d Data) error
	SendMsgToPeer(cmd string, d Data, p Peer) error
	PingNetwork() error
	Terminate()
	PrunePeerNodes()
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
		return Blockchain{}, err
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

	// Initialize the chain
	err = b.initializeChain()

	// If there was an error initializing the chain
	if err != nil {
		fmt.Printf("Error initializing blockchain: %+v\n", err)
		return err
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

	b.chain = append(b.chain, genesisBlock)
}

// mine implements functionality to mine a new block to the chain
func (b *Blockchain) mine(data Data) Block {

	//Create a new block
	newBlock := Block{
		Index:     len(b.chain),
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  b.chain[len(b.chain)-1].Hash,
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

	// Session variables
	var candidateBlock Block
	var newBlock Block
	b.mining = false

	for !done {

		// Get message from peers
		go func() {
			err := b.communicationComponent.RecieveFromNetwork(true)
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

				go func() {

					// Start a new mining session
					newTransaction := peerMsg.Data
					b.mining = true
					candidateBlock = Block{}
					log.Println("Recieved a new transaction, beginning new mining session...")

					// Mine the new block
					newBlock = b.mine(newTransaction)
					b.mining = false

					// If the new block's hash isnt empty, then this peer succesffuly mined the block
					if newBlock.Hash != "" {
						log.Println("Block mined successfully")
						candidateBlock = newBlock
						log.Println("Sending proof to Middleware...")

						err := b.communicationComponent.SendMsgToPeer("PROOF", nil, b.communicationComponent.GetMiddlewarePeer())
						if err != nil {
							log.Printf("Error sending message to Middleware: %v\n", err)
							return
						}
					}

				}()

			case "HALT":
				go func() {

					if b.mining {

						// If mining is true, then this peer was halted whilst still mining the block
						log.Println("Recieved a Halt, ending current mining session...")

						// Set mining to false which will end the mining session if its still in process,
						// as another peer has already successfully mined the block
						b.mining = false
					} else {

						log.Println("Recieved a Halt, but this peer's mining session already ended")

					}

					// Send this peer's copy of the chain so it can broadcast a list of peer chains to all nodes for consensus
					data := Chain{ChainCopy: b.chain}

					err := b.communicationComponent.SendMsgToPeer("PEER_CHAIN", data, b.communicationComponent.GetMiddlewarePeer())
					if err != nil {
						log.Printf("Error sending message to Middleware: %v\n", err)
						return
					}

				}()

			case "REWARD":

				go func() {

					// If this peer was the first peer to successfully mine the block, append the candidate block to this peer's blockchain
					// so that other nodes will get the block when consensus occurs
					log.Println("Appending new mined block to local chain")
					b.chain = append(b.chain, candidateBlock)

					// Add the reward that was sent to this peer for succesfully
					// mining the new block to this peer's wallet
					log.Println("Recieved a reward, adding amount to wallet... ")
					b.wallet += peerMsg.Data.(Transaction).Amount
					log.Printf("Updated balance: %d\n", b.wallet)
				}()

			case "CONSENSUS":

				go func() {

					peerChains := peerMsg.Data.(PeerChains).List

					fmt.Printf("\n\nDEBUG - Chain before consensus: %+v\n\n\n", b.chain)

					// Run consensus to get the updated chain, in any case
					newChain, err := b.consensusComponent.ConsensusMethod(peerChains, b.chain)
					if err != nil {
						log.Printf("Fatal Error running consensus method: %+v\n", err)
						done = true
					}

					b.chain = newChain
					fmt.Printf("\n\nDEBUG - Chain after consensus: %+v\n\n\n", b.chain)

					log.Println("Mining session concluded.")

				}()

			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			//There was no message to read, thus do nothing
		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Seconds() >= 10 {
			err := b.communicationComponent.PingNetwork()
			if err != nil {
				log.Printf("Error pinging network: %+v\n", err)
			} else {
				lastPing = time.Now()
			}
		}

		// If this peer hasn't received a message from another peer for 75 seconds,
		// then remove that peer from the list of known nodes
		b.communicationComponent.PrunePeerNodes()

		// Timeout for 5 milliseconds to limit the number of iterations of the loop to 20 per s
		time.Sleep(5 * time.Millisecond)

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

	if err != nil {
		fmt.Printf("Error initializing Blockchain communication component: %+v", err)
		return err
	}

	// Initialize the consensus component
	err = b.consensusComponent.Initialize()

	if err != nil {
		fmt.Printf("Error initializing Blockchain consensus component: %+v", err)
		return err
	}

	// Initialize the proof component
	err = b.proofComponent.Initialize()

	// If there was an error initializing one of the components
	if err != nil {
		fmt.Printf("Error initializing Blockchain proof component: %+v", err)
		return err
	}

	return nil
}

func (b *Blockchain) initializeChain() error {

	// Get the current list of peer node chains from the Middlware
	err := b.communicationComponent.SendMsgToPeer("GET_CHAINS", nil, b.communicationComponent.GetMiddlewarePeer())
	if err != nil {
		log.Printf("Error sending message to Middleware: %v\n", err)
		return err
	}

	err = b.communicationComponent.RecieveFromNetwork(false)
	if err != nil {
		log.Printf("Error recieving from network: %+v\n", err)
		return err
	}
	peerMsg := <-b.communicationComponent.GetMessageChannel()
	peerChains := peerMsg.Data.(PeerChains).List

	// If the peerChain list is empty, then this peer is the first node on the network
	if len(peerChains) == 0 {
		// Initialize the chain by creating the genesis block
		log.Println("No existing blockchain discovered, initializing new blockchain...")
		b.createGenesisBlock()

		// Tell the Middleware that a chain has been initialized
		data := Chain{ChainCopy: b.chain}

		err := b.communicationComponent.SendMsgToPeer("PEER_CHAIN", data, b.communicationComponent.GetMiddlewarePeer())
		if err != nil {
			log.Printf("Error sending message to Middleware: %v\n", err)
			return err
		}

	} else {
		// Else there are existing peer chains, so run consensus to get the right copy of the chain

		// Run consensus
		log.Println("Received peer chains, running consensus...")
		newChain, err := b.consensusComponent.ConsensusMethod(peerChains, b.chain)

		// If there was an error during consensus
		if err != nil {
			fmt.Printf("Error running consensus during initialization: %+v\n", err)
			return err
		}

		b.chain = newChain

		log.Printf("Chain copy after consensus: %+v\n", b.chain)

	}

	return nil
}
