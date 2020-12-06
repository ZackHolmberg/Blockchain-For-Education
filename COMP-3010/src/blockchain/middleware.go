package blockchain

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Middleware is the Middleware object
type Middleware struct {
	communicationComponent CommunicationComponent
	transactionQueue       *list.List
	newTransaction         chan Transaction
}

func handleNewTransaction(w http.ResponseWriter, req *http.Request) {

	// Receive transaction data from client

	// Transform into a Transaction struct

	// Create a new Message struct, containing the new Transactio as data

	// Set the command to MINE

	// Add it the the queue of transactions to be sent out
	fmt.Fprintf(w, "endpoint hit!\n")
	fmt.Fprintf(w, "your request: %+v\n", *req)

}

// Need a function that will facilitate block mining.
// Essentially, this middleware will keep a queue of transactions.
// The middleware will pop a transaction off the queue and broadcast it to all
// the nodes on the network. It then waits for one of the peers to mine the block,
// which it will know because a message will be sent to this middleware. The middleware
// will then broadcast a msg to all nodes telling them to stop mining, and will
// ignore any other solutions that came after the first one. Then, the process repeats.

// NewMiddleware creates and returns a new Middleware, with the Genesis Block initialized
func NewMiddleware(com CommunicationComponent) Middleware {

	// Define a new Middleware with the passed component value
	newMiddleware := Middleware{communicationComponent: com}

	// Initialize the Middleware
	newMiddleware.Initialize()

	return newMiddleware
}

// Initialize initializes the middleware by starting the request handlers,
// initializing its components and serving itself on the network
func (m *Middleware) Initialize() {

	// Intialize communication component
	m.communicationComponent.InitializeWithPort(8080)

	// Initialize Transaction queue
	m.transactionQueue = list.New()

	// Initialize Transaction channel
	m.newTransaction = make(chan Transaction)

	// Initialize newTransaction request handler
	http.HandleFunc("/newTransaction", handleNewTransaction)

	// Serve the http server
	go http.ListenAndServe(":8090", nil)

}

// Run utilizes its component and the http package to receive requests
// from clients and send/recieve messages to blockchain peers on the p2p network
func (m Middleware) Run(wg *sync.WaitGroup) {

	defer m.terminate(wg)

	// Calls the cleanup function terminate() if the user exits the program with ctrl+c
	c := make(chan os.Signal)
	done := false
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		done = true
	}()

	fmt.Println("\nRunning Middleware...")

	log.Println("Announcing self to peers...")

	var lastPing = time.Now()
	go m.communicationComponent.PingNetwork()

	fmt.Println()

	for !done {

		// Handle http request from client

		// Handle message from peers
		go m.communicationComponent.RecieveFromNetwork()

		select {
		case peerMsg := <-m.communicationComponent.GetMessageChannel():
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
func (m Middleware) terminate(wg *sync.WaitGroup) {
	fmt.Println("\nTerminating Middleware components...")

	m.communicationComponent.Terminate()

	fmt.Println("Exiting Middleware...")

	wg.Done()

}
