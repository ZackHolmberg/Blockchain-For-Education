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
	clientComponent        ClientComponent
	chain                  []Block
	wallet                 int
}

//ConsensusComponent standardizes methods for any Peer consensus component
type ConsensusComponent interface {
	ValidateBlock(b Block) bool
	HandleCommand(msg Message, p *Peer) error
	GetCandidateBlock() Block
	Initialize() error
	Terminate()
}

// CommunicationComponent standardizes methods for any Peer communcation component
type CommunicationComponent interface {
	Initialize() error
	InitializeWithPort(port int) error
	GetPeerNodes() []PeerAddress
	GetMiddlewarePeer() PeerAddress
	GetSelfAddress() PeerAddress
	GetMessageChannel() chan Message
	RecieveFromNetwork(withTimeout bool) error
	GenerateMessage(cmd string, data Data) (Message, error)
	BroadcastMsgToNetwork(m Message) error
	SendMsgToPeer(m Message, p PeerAddress) error
	PingNetwork() error
	Terminate()
	PrunePeerNodes()
}

// ClientComponent standardizes methods for any Peer transaction component
type ClientComponent interface {
	Initialize(com CommunicationComponent, wallet *int) error
	Terminate()
}

// NewPeer creates and returns a new Peer, with the Genesis Block and Components initialized
func NewPeer(c CommunicationComponent, p ConsensusComponent, t ClientComponent) (Peer, error) {

	// Define a new Peer with the passed componenet values
	newPeer := Peer{communicationComponent: c, consensusComponent: p, clientComponent: t}

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
func (p *Peer) initialize() error {

	// Initialize Peer peer components
	err := p.initializeComponents()

	// If there was an error initializing one of the components
	if err != nil {
		fmt.Printf("Error initializing Peer peer: %+v\n", err)
		return err
	}

	// Initialize the chain
	err = p.initializeChain()

	// If there was an error initializing the chain
	if err != nil {
		fmt.Printf("Error initializing Peer: %+v\n", err)
		return err
	}

	p.wallet = 10

	return nil
}

// createGenesisBlock initializes and adds a genesis block to the Peer
func (p *Peer) createGenesisBlock() {

	genesisBlock := Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().String()
	genesisBlock.Data = Transaction{}
	genesisBlock.PrevHash = ""
	genesisBlock.Nonce = 0
	genesisBlock.Hash = "0"

	p.chain = append(p.chain, genesisBlock)
}

// Run utilizes the Peer components to run this Peer peer by sending/recieving
// requests and messages on the p2p network
func (p *Peer) Run() {

	//When Run() concludes, terminate() will be called to clean up the different Peer components
	defer p.terminate()

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
		err := p.communicationComponent.PingNetwork()
		if err != nil {
			log.Printf("Error pinging network: %+v\n", err)
		}
	}()
	fmt.Println()

	for !done {

		// Get message from peers
		go func() {
			err := p.communicationComponent.RecieveFromNetwork(true)
			if err != nil {
				log.Printf("Fatal Error recieving from network: %+v\n", err)
				done = true
			}
		}()

		select {
		case peerMsg := <-p.communicationComponent.GetMessageChannel():
			switch peerMsg.Command {
			case "PING":
				log.Printf("Recieved a ping from %s\n", peerMsg.From.String())

			case "PEER_CHAIN":
				go func() {
					peerChain := peerMsg.Data.(Chain).ChainCopy
					// fmt.Printf("\n\nDEBUG - Chain before consensus: %+v\n\n\n", p.chain)
					// If the received chain is longer than the current chain, use the received chain as this peer's new chain copy, and broadcast our copy again
					if len(peerChain) > len(p.chain) {
						p.chain = peerChain
						log.Println("Recieved a longer copy of the chain, setting it as new local copy")

						p.broadcastChainCopy()
					}
					// fmt.Printf("\n\nDEBUG - Chain after consensus: %+v\n\n\n", p.chain)
				}()
			case "GET_CHAIN":
				go p.broadcastChainCopy()

			case "TRANSACTION":
				go func() {

					amount := peerMsg.Data.(Transaction).Amount

					// Determine whether this is a transaction from a fellow Peer or a reward from the Middleware
					if peerMsg.From.String() == p.communicationComponent.GetMiddlewarePeer().String() {
						// If this peer was the first peer to successfully mine the block, append the candidate block to this peer's Peer
						// so that other nodes will get the block when consensus occurs
						p.chain = append(p.chain, p.consensusComponent.GetCandidateBlock())

						log.Println("Recieved a reward, adding amount to wallet and appending new mined block to local chain")
					} else {
						sender := peerMsg.Data.(Transaction).From
						log.Printf("Recieved a transaction from: %v with amount: %d\n", sender, amount)

					}

					// Whether it's a reward or not, add the amount to this Peer's wallet
					p.wallet += amount

					log.Printf("Updated balance: %d\n", p.wallet)

				}()

			case "VALIDATE":
				go func() {
					// Tell the middleware if the received block is valid or not
					log.Println("Received candidate block from Middleware, validating...")
					candidateBlock := peerMsg.Data.(CandidateBlock).Block
					valid := p.consensusComponent.ValidateBlock(candidateBlock)
					if valid {
						log.Println("Verified received candidate block is valid")
						toSend, err := p.communicationComponent.GenerateMessage("BLOCK_VALID", nil)
						if err != nil {
							log.Printf("Error generating message: %v\n", err)
						}

						err = p.communicationComponent.SendMsgToPeer(toSend, p.communicationComponent.GetMiddlewarePeer())
						if err != nil {
							log.Printf("Error sending message to Middleware: %v\n", err)
						}
					}

				}()

			// If the the received command isn't supported by the peer, then it must be a component-specific
			// command, and thus we hand it off for the component to handle. Currently, only the consensus component
			// so we can assume the command is meant for it and hand it off
			default:
				err := p.consensusComponent.HandleCommand(peerMsg, p)
				if err != nil {
					log.Printf("Consensus component had error when handling message: %+v\n", err)
				}
			}
		default:
			//There was no message to read, thus do nothing
		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Minutes() >= 1 {
			err := p.communicationComponent.PingNetwork()
			if err != nil {
				log.Printf("Error pinging network: %+v\n", err)
			} else {
				lastPing = time.Now()
			}
		}

		// If this peer hasn't received a message from another peer for 75 seconds,
		// then remove that peer from the list of known nodes
		p.communicationComponent.PrunePeerNodes()

		// Timeout for 5 milliseconds to limit the number of iterations of the loop to 20 per s
		time.Sleep(5 * time.Millisecond)

	}
}

// terminate calls all of the interface-defined component clean-up methods
func (p *Peer) terminate() {
	fmt.Println("\nTerminating Peer components...")

	p.communicationComponent.Terminate()
	p.consensusComponent.Terminate()
	p.clientComponent.Terminate()
	fmt.Println("Exiting Blockchain Peer...")
}

func (p *Peer) initializeComponents() error {

	// Initialize the communication component
	err := p.communicationComponent.Initialize()

	if err != nil {
		fmt.Printf("Error initializing Peer communication component: %+v", err)
		return err
	}

	// Initialize the consensus component
	err = p.consensusComponent.Initialize()

	// If there was an error initializing the consensus component
	if err != nil {
		fmt.Printf("Error initializing Peer consensus component: %+v", err)
		return err
	}

	// Initialize the transactions component
	err = p.clientComponent.Initialize(p.communicationComponent, &p.wallet)

	// If there was an error initializing the transactions component
	if err != nil {
		fmt.Printf("Error initializing Peer transactions component: %+v", err)
		return err
	}

	return nil
}

func (p *Peer) initializeChain() error {

	// Create the genesis block in case we are the first peer on the network
	p.createGenesisBlock()

	// Get peer chains in case we are not the first peer on the network
	toSend, err := p.communicationComponent.GenerateMessage("GET_CHAIN", nil)
	if err != nil {
		log.Printf("Error generating message: %v\n", err)
	}

	err = p.communicationComponent.BroadcastMsgToNetwork(toSend)
	if err != nil {
		log.Printf("Error broadcasting message: %v\n", err)
	}

	return nil
}

func (p *Peer) broadcastChainCopy() {
	// Broadcast this peer's copy of the chain so that the new chain can be distributed
	data := Chain{ChainCopy: p.chain}

	toSend, err := p.communicationComponent.GenerateMessage("PEER_CHAIN", data)
	if err != nil {
		log.Printf("Error generating message: %v\n", err)
	}

	err = p.communicationComponent.BroadcastMsgToNetwork(toSend)
	if err != nil {
		log.Printf("Error broadcasting message: %v\n", err)
	}
}
