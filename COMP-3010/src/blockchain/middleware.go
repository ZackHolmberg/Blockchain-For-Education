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
}

// example request: curl -X POST -d 'from=jerry&to=bob&amount=10' localhost:8090/newTransaction

func (m *Middleware) handleNewTransaction(w http.ResponseWriter, r *http.Request) {

	// Receive transaction data from client
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	from := r.FormValue("from")
	to := r.FormValue("to")

	fmt.Fprintf(w, "%+v\n", r.Form)
	fmt.Fprintf(w, "from: %s, to: %s\n", from, to)

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Printf("Error converting int to string: %v\n", err)
		fmt.Fprintf(w, "Something bad occured, please try your request again!\n")
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

// Initializes the middleware by starting the request handlers,
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

// Utilizes its component and the http package to receive requests
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

	for !done {

		// Get message from peers
		go func() {
			err := m.communicationComponent.RecieveFromNetwork(true)
			if err != nil {
				log.Printf("Fatal Error recieving from network: %+v\n", err)
				done = true
			}
		}()

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

					toSend, err := m.communicationComponent.GenerateMessage("REWARD", newData)
					if err != nil {
						log.Printf("Fatal error generating message: %v\n", err)
						done = true
					}

					err = m.communicationComponent.SendMsgToPeer(toSend, peerMsg.From)
					if err != nil {
						// This would be a fatal error because if a peer doesn't recieve the reward,
						// it wont add the new block to its chain, and the block will get lost if the next session begins
						log.Printf("Fatal Error sending reward to peer: %v", err)
						done = true
					}

				}

			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			//There was no message to read, thus do nothing
		}

		// If we aren't already in a mining session and there is at least one transaction to be mined, pop a transaction
		// from the queue and broacast to network, starting a new mining session
		if !peersMining && m.transactionQueue.Len() > 0 && len(m.communicationComponent.GetPeerNodes()) > 0 {

			log.Println("Beginning a new mining session...")

			// Pop a Message from the transactionQueue
			toMine := m.pop()

			toSend, err := m.communicationComponent.GenerateMessage("MINE", toMine)
			if err != nil {
				log.Printf("Fatal error generating message: %v\n", err)
				done = true
			}

			// Broadcast the new transaction to the peers on the network
			err = m.communicationComponent.BroadcastMsgToNetwork(toSend)
			if err != nil {
				// This would be a fatal error
				log.Printf("Fatal error broadcasting message: %v", err)
				done = true
			}
			// Set peersMining to true and proofFound to false since the above function call will cause the blockchain
			// peers to begin mining the new transaction, starting a new mining session
			peersMining = true
			proofFound = false

		} else if peersMining && proofFound && !running {
			// End the mining session. Running to be set before the goroutine starts or else it will
			// get called multiple times
			running = true

			go func() {

				// Else if proofFound, conclude the current mining session by broadcasting a halt message to the peers
				// to halt mining and cause consensus to be run so every peer gets a copy of the new longest chain
				log.Println("Broadcasting halt message to peers...")

				toSend, err := m.communicationComponent.GenerateMessage("HALT", nil)
				if err != nil {
					// This would be a fatal error
					log.Printf("Fatal error generating message: %v\n", err)
					done = true
				}

				// Broadcast the halt message to the peers on the network. All peers will send their chain copies when they
				// recieve a halt
				err = m.communicationComponent.BroadcastMsgToNetwork(toSend)
				if err != nil {
					// This would be a fatal error
					log.Printf("Error broadcasting message: %v", err)
					done = true
				}

				// Timeout for 5 seconds to give peers time to conclude their mining sessions and
				// send their copies of the chain to the middleware to distribute for consensus
				log.Println("Timing out for 5 seconds while peers distribute new chain...")
				time.Sleep(5 * time.Second)

				// Reset state
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

		// Timeout for 5 milliseconds to limit the number of iterations of the loop to 20 per second
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
