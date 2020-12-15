package blockchain

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Middleware is the Middleware object
type Middleware struct {
	communicationComponent CommunicationComponent
	transactionQueue       *list.List
	newTransaction         chan Transaction
	peerChains             PeerChains
}

func (m *Middleware) handleNewTransaction(w http.ResponseWriter, r *http.Request) {

	// Receive transaction data from client
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	from := r.FormValue("from")
	to := r.FormValue("to")
	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Printf("Error converting int to string: %v", err)
		fmt.Fprintf(w, "Something bad occured, please try your request again!")
		return
	}

	// Transform into a Transaction struct
	newTransaction := Transaction{From: from, To: to, Amount: amount}

	// Add it the the queue of transactions to be sent out
	m.transactionQueue.PushBack(newTransaction)

	// Notify the client that the transaction was successfully processed
	fmt.Fprintf(w, "Transaction processed succesfully!\n\n")

}

// NewMiddleware creates and returns a new Middleware, with the Genesis Block initialized
func NewMiddleware(com CommunicationComponent, udpPort int, serverPort int) (Middleware, error) {

	// Define a new Middleware with the passed component value
	newMiddleware := Middleware{communicationComponent: com}

	// Initialize the Middleware
	err := newMiddleware.Initialize(udpPort, serverPort)
	if err != nil {
		fmt.Printf("Error initializing Middleware: %+v\n", err)
		return Middleware{}, err
	}

	return newMiddleware, nil
}

// Initialize initializes the middleware by starting the request handlers,
// initializing its components and serving itself on the network
func (m *Middleware) Initialize(udpPort int, serverPort int) error {

	// Intialize communication component
	err := m.communicationComponent.InitializeWithPort(udpPort)
	if err != nil {
		fmt.Printf("Error initializing communication component: %+v\n", err)
		return err
	}

	// Initialize Transaction queue
	m.transactionQueue = list.New()

	// Initialize Transaction channel
	m.newTransaction = make(chan Transaction)

	// Intialize PeerChains struct
	m.peerChains = PeerChains{}

	// Initialize newTransaction request handler
	http.HandleFunc("/newTransaction", m.handleNewTransaction)

	// Serve the http server
	go http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)

	return nil
}

// Run utilizes its component and the http package to receive requests
// from clients and send/recieve messages to blockchain peers on the p2p network
func (m *Middleware) Run() {

	//When Run() concludes, terminate() will be called to clean up the different Blockchain components
	defer m.terminate()

	// Sets done equal to true if the user exits the program with ctrl+c, which will case the loop to finish and Run() to exit,
	// which will cause terminate() to run
	c := make(chan os.Signal)
	done := false
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		done = true
	}()

	fmt.Println("\nRunning Middleware...")

	var lastPing = time.Now()
	go func() {
		err := m.communicationComponent.PingNetwork()
		if err != nil {
			log.Printf("Error pinging network: %+v\n", err)
		}
	}()
	fmt.Println()

	// Session variables
	peersMining, proofFound, running := false, false, false
	newPeerChains := [][]Block{}

	for !done {

		// Get message from peers
		// go func() {
		err := m.communicationComponent.RecieveFromNetwork(false)
		if err != nil {
			log.Printf("Fatal Error recieving from network: %+v\n", err)
			done = true
		}
		// }()

		select {
		case peerMsg := <-m.communicationComponent.GetMessageChannel():
			switch peerMsg.Command {
			case "PING":
				log.Printf("Recieved a ping from %s:%d\n", peerMsg.From.Address.IP.String(), peerMsg.From.Address.Port)

			case "PROOF":

				//Dont run this in a goroutine as not to cause possible concurrency issues

				// A blockchain peer is returning a proof for the current mining session, thus set proofFound to true
				// so that the mining session is concluded
				if !proofFound {
					proofFound = true
					log.Println("Peer found proof. Ending current mining session...")

					// Send a reward to the successful miner, which will tell them to add the mined block to their chain,
					// which will become the new global chain once consensus is run
					newData := Transaction{From: "", To: "", Amount: 5}

					err := m.communicationComponent.SendMsgToPeer("REWARD", newData, peerMsg.From)
					if err != nil {
						// This would be a fatal error because if a peer doesn't recieve the reward,
						// it wont add the new block to its chain, and the block will get lost if the next session begins
						log.Printf("Fatal Error sending reward to peer: %v", err)
						done = true
					}

				}

				// Else this was not the first proof found, so ignore this message

			case "PEER_CHAIN":

				go func() {

					// Get the chain copy that was sent by the peer
					chain := peerMsg.Data.(Chain)

					// Append it to the new list of peer chains
					newPeerChains = append(newPeerChains, chain.ChainCopy)

					// If we're initializing the chain for the first time,
					// immediately set the Middleware's chain list to the new chain
					if len(m.peerChains.List) == 0 {
						m.peerChains.List = newPeerChains
						log.Println("Chain initialized")
					}

					log.Printf("\n\nNew peer chain list: %+v\n\n\n", newPeerChains)

				}()

			case "GET_CHAINS":

				go func() {

					err := m.communicationComponent.SendMsgToPeer("CONSENSUS", m.peerChains, peerMsg.From)
					if err != nil {
						// Not a fatal error since consensus can still occur amongst other nodes, and the
						// peer who didn't recieve a chain copy will deal with it but the Middleware should keep running
						log.Printf("Fatal Error sending reward to peer: %v", err)
					}

				}()

			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			//There was no message to read, thus do nothing
		}

		// If !peersMining and there is at least one transaction to be mind, pop a transaction
		// from the queue and broacast to network, starting a new mining session
		if !peersMining && m.transactionQueue.Len() > 0 {
			log.Println("Beginning a new mining session...")

			// Pop a Message from the transactionQueue
			toMine := m.pop()

			// Broadcast the new transaction to the peers on the network
			err := m.communicationComponent.BroadcastMsgToNetwork("MINE", toMine)
			if err != nil {
				// This would be a fatal error
				log.Printf("Error broadcasting message: %v", err)
				done = true
			}
			// Set peersMining to true and proofFound to false since the above function call will cause the blockchain
			// peers to begin mining the new transaction, starting a new mining session
			peersMining = true
			proofFound = false

			// Reset the list of peer chains
			newPeerChains = nil

		} else if peersMining && proofFound && !running {
			// End the mining session. Running to be set before the goroutine starts or else it will
			// get called multiple times
			running = true

			go func() {

				// Else if proofFound, conclude the current mining session by broadcasting a halt message to the peers
				// to halt mining and cause consensus to be run so every peer gets a copy of the new longest chain
				log.Println("Broadcasting halt message to peers...")

				// Broadcast the halt message to the peers on the network. All peers will send their chain copies when they
				// recieve a halt
				err := m.communicationComponent.BroadcastMsgToNetwork("HALT", nil)
				if err != nil {
					// This would be a fatal error
					log.Printf("Error broadcasting message: %v", err)
					done = true
				}

				// Wait until all peers have sent their chain copies or the timeout length has pssed
				start := time.Now()
				for len(newPeerChains) != len(m.communicationComponent.GetPeerNodes()) || time.Since(start).Seconds() >= 5 {
				}

				// Now consensus should occur, so trigger it with a broadcast
				newData := PeerChains{List: newPeerChains}

				err = m.communicationComponent.BroadcastMsgToNetwork("CONSENSUS", newData)

				if err != nil {
					// This would be a fatal error because if consensus doesn't run,
					// the newly mined block will be lost
					log.Printf("Fatal Error triggering consensus: %v", err)
					done = true
				}

				// Timeout for 3 seconds to give peers time to conclude their mining sessions ans
				// send their copies of the chain to the middleware to distribute for consensus
				log.Println("Timing out for 3 seconds while peers run consensus...")
				time.Sleep(3 * time.Second)

				// Set the middleware's peer chains list to the new list that was used for consensus
				m.peerChains = newData

				// Reset state
				newPeerChains = nil
				running = false
				peersMining = false

				log.Println("Mining session concluded.")

			}()

		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Minutes() >= 1 {
			err := m.communicationComponent.PingNetwork()
			if err != nil {
				log.Printf("Error pinging network: %+v\n", err)
			}
			lastPing = time.Now()
		}

		// If this peer hasn't received a message from another peer for 75 seconds,
		// then remove that peer from the list of known nodes
		m.communicationComponent.PrunePeerNodes()

		// Timeout for 5 milliseconds to limit the number of iterations of the loop to 20 per s
		time.Sleep(5 * time.Millisecond)

	}
}

// terminate calls all of the interface-defined component clean-up methods
func (m Middleware) terminate() {

	fmt.Println("\nTerminating Middleware components...")

	m.communicationComponent.Terminate()

	fmt.Println("Exiting Middleware...")
}

// Pops a message of the Middleware's transactionQueue and returns it
func (m *Middleware) pop() Transaction {

	// Get element from the front of the list
	poppedElement := m.transactionQueue.Front()

	// Remove the element, essentially "popping" it
	m.transactionQueue.Remove(poppedElement)

	//Convert the popped element, which is of type *Element, to *Message
	toMine := poppedElement.Value.(Transaction)

	return toMine
}
