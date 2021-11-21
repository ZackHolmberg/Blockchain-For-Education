package blockchain

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ============================ Peer ============================

// Node defines required methods of any Peer Node implementation
type Node interface {
	NewPeer(com CommunicationComponent, p ConsensusComponent) *Node
	Run()

	initialize() error
	mine()
	createGenesisNode()
	terminate()
	initializeComponents() error
}

// Peer is the Peer object
type Peer struct {
	communicationComponent CommunicationComponent
	consensusComponent     ConsensusComponent
	chain                  []Block
	mining                 bool
	wallet                 int
}

//ConsensusComponent standardizes methods for any Peer consensus component
type ConsensusComponent interface {
	CalculateHash(nonce int, block Block) string
	ProofMethod(b Block, m bool) string
	ValidateProof(s string) bool
	Initialize() error
	Terminate()
}

// CommunicationComponent standardizes methods for any Peer communcation component
type CommunicationComponent interface {
	Initialize() error
	InitializeWithPort(port int) error
	GetPeerNodes() []PeerAddress
	GetMiddlewarePeer() PeerAddress
	GetMessageChannel() chan Message
	RecieveFromNetwork(withTimeout bool) error
	GenerateMessage(cmd string, data Data) (Message, error)
	BroadcastMsgToNetwork(m Message) error
	SendMsgToPeer(m Message, p PeerAddress) error
	PingNetwork() error
	Terminate()
	PrunePeerNodes()
}

// NewPeer creates and returns a new Peer, with the Genesis Block and Components initialized
func NewPeer(com CommunicationComponent, p ConsensusComponent) (Peer, error) {

	// Define a new Peer with the passed componenet values
	newPeer := Peer{communicationComponent: com, consensusComponent: p}

	// Initialize the Peer
	err := newPeer.initialize()

	// If there was an error initializing the Peer peer
	if err != nil {
		fmt.Printf("Error initializing Peer peer: %+v\n", err)
		newPeer.terminate()
		return Peer{}, err
	}

	return newPeer, nil
}

// Initialize initializes the Peer by initializing its components and serving itself on the network
func (b *Peer) initialize() error {

	// Initialize Peer peer components
	err := b.initializeComponents()

	// If there was an error initializing one of the components
	if err != nil {
		fmt.Printf("Error initializing Peer peer: %+v\n", err)
		return err
	}

	// Initialize the chain
	err = b.initializeChain()

	// If there was an error initializing the chain
	if err != nil {
		fmt.Printf("Error initializing Peer: %+v\n", err)
		return err
	}

	return nil
}

// createGenesisBlock initializes and adds a genesis block to the Peer
func (b *Peer) createGenesisBlock() {

	genesisBlock := Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().String()
	genesisBlock.Data = Transaction{}
	genesisBlock.PrevHash = ""
	genesisBlock.Hash = b.consensusComponent.ProofMethod(genesisBlock, true)

	b.chain = append(b.chain, genesisBlock)
}

// mine implements functionality to mine a new block to the chain
func (b *Peer) mine(data Data) Block {

	//Create a new block
	newBlock := Block{
		Index:     len(b.chain),
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  b.chain[len(b.chain)-1].Hash,
		Hash:      ""}

	//Calculate this block's proof
	newBlock.Hash = b.consensusComponent.ProofMethod(newBlock, b.mining)

	return newBlock
}

// Run utilizes the Peer components to run this Peer peer by sending/recieving
// requests and messages on the p2p network
func (b *Peer) Run() {

	//When Run() concludes, terminate() will be called to clean up the different Peer components
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

	fmt.Println("\nRunning Peer...")

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
					newTransaction := peerMsg.Data.(Transaction)
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

						toSend, err := b.communicationComponent.GenerateMessage("PROOF", nil)
						if err != nil {
							log.Printf("Fatal error generating message: %v\n", err)
							done = true
						}

						err = b.communicationComponent.SendMsgToPeer(toSend, b.communicationComponent.GetMiddlewarePeer())
						if err != nil {
							log.Printf("Error sending message to Middleware: %v\n", err)
							return
						}
					}
				}()

			case "HALT":
				go func() {
					if b.mining {
						// Set mining to false which will end the mining session
						// as another peer has already successfully mined the block
						b.mining = false
					}

					// Broadcast this peer's copy of the chain so that the new chain can be distributed
					data := Chain{ChainCopy: b.chain}

					toSend, err := b.communicationComponent.GenerateMessage("PEER_CHAIN", data)
					if err != nil {
						log.Printf("Error generating message: %v\n", err)
					}

					err = b.communicationComponent.BroadcastMsgToNetwork(toSend)
					if err != nil {
						log.Printf("Error broadcasting message: %v\n", err)
					}

				}()

			case "REWARD":
				go func() {
					// If this peer was the first peer to successfully mine the block, append the candidate block to this peer's Peer
					// so that other nodes will get the block when consensus occurs
					log.Println("Appending new mined block to local chain")
					b.chain = append(b.chain, candidateBlock)

					// Add the reward that was sent to this peer for succesfully
					// mining the new block to this peer's wallet
					log.Println("Recieved a reward, adding amount to wallet... ")
					b.wallet += peerMsg.Data.(Transaction).Amount
					log.Printf("Updated balance: %d\n", b.wallet)
				}()

			case "PEER_CHAIN":
				go func() {
					peerChain := peerMsg.Data.(Chain).ChainCopy
					// fmt.Printf("\n\nDEBUG - Chain before consensus: %+v\n\n\n", b.chain)

					// If the received chain is longer than the current chain, use the received chain as this peer's new chain copy, and broadcast our copy again
					if len(peerChain) > len(b.chain) {
						b.chain = peerChain
						log.Println("Recieved a longer copy of the chain, setting it as new local copy")

						b.broadcastChainCopy()
					}
					// fmt.Printf("\n\nDEBUG - Chain after consensus: %+v\n\n\n", b.chain)
				}()
			case "GET_CHAIN":
				go b.broadcastChainCopy()

			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			//There was no message to read, thus do nothing
		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Minutes() >= 1 {
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
func (b *Peer) terminate() {
	fmt.Println("\nTerminating Peer components...")

	b.communicationComponent.Terminate()
	b.consensusComponent.Terminate()

	fmt.Println("Exiting Blockchain Peer...")
}

func (b *Peer) initializeComponents() error {
	// Initialize the communication component
	err := b.communicationComponent.Initialize()

	if err != nil {
		fmt.Printf("Error initializing Peer communication component: %+v", err)
		return err
	}

	// Initialize the proof component
	err = b.consensusComponent.Initialize()

	// If there was an error initializing one of the components
	if err != nil {
		fmt.Printf("Error initializing Peer proof component: %+v", err)
		return err
	}

	return nil
}

func (b *Peer) initializeChain() error {

	// Create the genesis block in case we are the first peer on the network
	b.createGenesisBlock()

	// Get peer chains in case we are not the first peer on the network
	toSend, err := b.communicationComponent.GenerateMessage("GET_CHAIN", nil)
	if err != nil {
		log.Printf("Error generating message: %v\n", err)
	}

	err = b.communicationComponent.BroadcastMsgToNetwork(toSend)
	if err != nil {
		log.Printf("Error broadcasting message: %v\n", err)
	}

	return nil
}

func (b *Peer) broadcastChainCopy() {
	// Broadcast this peer's copy of the chain so that the new chain can be distributed
	data := Chain{ChainCopy: b.chain}

	toSend, err := b.communicationComponent.GenerateMessage("PEER_CHAIN", data)
	if err != nil {
		log.Printf("Error generating message: %v\n", err)
	}

	err = b.communicationComponent.BroadcastMsgToNetwork(toSend)
	if err != nil {
		log.Printf("Error broadcasting message: %v\n", err)
	}
}
