package blockchain

import (
	"container/list"
	"fmt"
	"log"
	"net"
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

	// Create a new Message struct, containing the new Transaction as data and setting command to mine
	newMessage, err := m.communicationComponent.GenerateMessage("MINE", newTransaction)
	if err != nil {
		log.Printf("Error generating message: %v", err)
		fmt.Fprintf(w, "Something bad occured, please try your request again!")
		return
	}

	// Add it the the queue of transactions to be sent out
	m.transactionQueue.PushBack(newMessage)

	// Notify the client that the transaction was successfully processed
	fmt.Fprintf(w, "Transaction processed succesfully!")

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

	// Initialize newTransaction request handler
	http.HandleFunc("/newTransaction", m.handleNewTransaction)

	// Serve the http server
	go http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)

	return nil
}

// Run utilizes its component and the http package to receive requests
// from clients and send/recieve messages to blockchain peers on the p2p network
func (m Middleware) Run() {

	defer m.terminate()

	// Calls the cleanup function terminate() if the user exits the program with ctrl+c
	c := make(chan os.Signal)
	done := false
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		done = true
	}()

	fmt.Println("\nRunning Middleware...")

	var lastPing = time.Now()
	go m.communicationComponent.PingNetwork()

	fmt.Println()

	peersMining, proofFound := false, false

	for !done {

		// If !peersMining and there is at least one transaction to be mind, pop a transaction
		// from the queue and broacast to network, starting a new mining session
		if !peersMining && m.transactionQueue.Len() > 0 {
			log.Println("Beginning a new mining session...")

			// Pop a Message from the transactionQueue
			toMine := m.pop()

			// Broadcast the new transaction to the peers on the network
			m.communicationComponent.BroadcastToNetwork(toMine)

			// Set peersMining to true and proofFound to false since the above function call will cause the blockchain
			// peers to begin mining the new transaction, starting a new mining session
			peersMining = true
			proofFound = false

		} else if peersMining && proofFound {

			// Else if proofFound, conclude the current mining session by broadcasting a halt message to the peers
			// to halt mining and cause consensus to be run so every peer gets a copy of the new longest chain
			log.Println("Broadcasting halt message to peers...")

			toSend, err := m.communicationComponent.GenerateMessage("HALT", nil)
			if err != nil {
				log.Printf("Error generating message: %v", err)
				// If an error occurs, skip to the next iteration of the loop and try again
				continue
			}

			// Broadcast the halt message to the peers on the network
			m.communicationComponent.BroadcastToNetwork(toSend)

			// Timeout for 5 seconds to give peers time to conduct consensus
			log.Println("Sleeping for 3 seconds while peers run mining session concludes...")
			time.Sleep(3 * time.Second)

			peersMining = false
		}

		// Handle message from peers
		go m.communicationComponent.RecieveFromNetwork()

		select {
		case peerMsg := <-m.communicationComponent.GetMessageChannel():
			switch peerMsg.Command {
			case "PING":
				log.Printf("Recieved a ping from %s:%d\n", peerMsg.From.Address.IP.String(), peerMsg.From.Address.Port)
			case "PROOF":
				// A blockchain peer is returning a proof for the current mining session, thus set proofFound to true
				// so that the mining session is concluded
				if !proofFound {
					proofFound = true
					log.Println("Peer found proof. Ending current mining session...")

					// Send a reward to the successful miner, which will tell them to add the mined block to their chain,
					// which will become the new global chain once consensus is run
					newData := Transaction{From: "", To: "", Amount: 5}
					toSend, err := m.communicationComponent.GenerateMessage("REWARD", newData)
					if err != nil {
						log.Printf("Error generating message: %v", err)
					}

					//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
					temp, _ := net.LookupIP("localhost")
					peerMsg.From.Address.IP = temp[0]
					// =====================================================

					log.Println("Sending reward to successful peer")
					err = m.communicationComponent.SendToPeer(toSend, peerMsg.From)
					if err != nil {
						// This would be a fatal error because if a peer doesn't recieve the reward,
						// it wont add the new block to its chain, and the block will get lost if the next session begins
						log.Printf("Fatal Error sending reward to peer: %v", err)
						done = true
					}

				}

				// Else this was not the first proof found, so ignore this message

			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			//There was no message to read, thus do nothing
		}

		// Ping all peer nodes on the network once every minute
		if time.Since(lastPing).Minutes() >= 1 {
			log.Println("Middleware peer sending pings...")
			go m.communicationComponent.PingNetwork()
			lastPing = time.Now()
		}

		// If this peer hasn't received a message from another peer for 75 seconds,
		// then remove that peer from the list of known nodes
		go m.communicationComponent.PrunePeerNodes()

		// Timeout for 1 millisecond to limit the number of iterations of the loop to 1 per ms
		time.Sleep(1 * time.Millisecond)

	}
}

// terminate calls all of the interface-defined component clean-up methods
func (m Middleware) terminate() {

	fmt.Println("\nTerminating Middleware components...")

	m.communicationComponent.Terminate()

	fmt.Println("Exiting Middleware...")
}

// Pops a message of the Middleware's transactionQueue and returns it
func (m *Middleware) pop() Message {

	// Get element from the front of the list
	poppedElement := m.transactionQueue.Front()

	// Remove the element, essentially "popping" it
	m.transactionQueue.Remove(poppedElement)

	//Convert the popped element, which is of type *Element, to *Message
	toMine := poppedElement.Value.(Message)

	return toMine
}
