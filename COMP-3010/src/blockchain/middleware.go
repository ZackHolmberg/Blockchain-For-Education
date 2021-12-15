package blockchain

import (
	"container/list"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"github.com/mroth/weightedrand"
)

// Middleware is the Middleware object
type Middleware struct {
	communicationComponent CommunicationComponent
	transactionQueue       *list.List
	newTransaction         chan Transaction
	lotteryPool            []LotteryEntry
	candidateBlockQueue    *list.List
	blockValidators        int
	blockValid             bool
	proofFound             bool
}

const REWARD_AMOUNT = 5

// example request: curl -X POST -d 'from=jerry&to=bob&amount=10' localhost:8090/newTransaction

func (m *Middleware) handleNewTransaction(w http.ResponseWriter, r *http.Request) {

	// Receive transaction data from client
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	from := r.FormValue("from")
	to := r.FormValue("to")
	signature := r.FormValue("signature")

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Printf("Error converting int to string: %v\n", err)
		fmt.Fprintf(w, "Something bad occured, please try your request again!\n")
		return
	}

	// Transform into a Transaction struct
	newTransaction := Transaction{From: from, To: to, Amount: amount, Signature: signature}

	// Add it the the queue of transactions to be sent out
	m.transactionQueue.PushBack(newTransaction)

	// Notify the client that the transaction was successfully processed
	fmt.Fprintf(w, "Transaction processed succesfully!\n\n")

}

// NewMiddleware creates and returns a new Middleware
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

	// Initialize Candidate Block queue
	m.candidateBlockQueue = list.New()

	// Initialize validator count to 0
	m.blockValidators = 0

	// Initialize proof found boolean to false
	m.proofFound = false

	// Initialize valid block boolean to false
	m.blockValid = false

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
	peersMining, running := false, false

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
				log.Printf("Recieved a ping from %s\n", peerMsg.From.String())
			case "GET_CHAIN", "PEER_CHAIN", "PUBLIC_KEYS":
				// Ignore, this is only a peer-relevant command but since Middleware is a part of the network it will get the messages
			case "PROOF":
				go func() {
					candidateBlock := peerMsg.Data.(CandidateBlock)
					// We push every candidate block we receive on the queue
					// in case the initial proof fails validation
					m.candidateBlockQueue.PushBack(candidateBlock)
					if !m.proofFound {
						m.proofFound = true

						// Once the peer receives a candidate block with a proof, run network validation to
						// ensure the block is valid
						log.Println("First proof received. Running validation...")

						err := m.runValidation()
						if err != nil {
							// This would be a fatal error because if the network isn't able to validate
							// the block then malicious nodes could succeed
							log.Printf("Fatal error validating candidate block: %v", err)
							done = true
						}
					}
				}()

			case "STAKE":
				go func() {

					newLotteryEntry := peerMsg.Data.(LotteryEntry)

					if len(m.lotteryPool) == 0 {
						log.Println("Running lottery in 10 seconds...")
						time.AfterFunc(10*time.Second, func() {
							err := m.runLottery()
							if err != nil {
								// This would be a fatal error because if we cannot run the lottery
								// then no new transactions are going to be mined
								log.Printf("Fatal error running lottery: %v\n", err)
								done = true
							}

						})
					}
					log.Printf("Received lottery entry: %+v\n", newLotteryEntry)

					m.lotteryPool = append(m.lotteryPool, newLotteryEntry)
				}()

			case "BLOCK_VALID":
				go func() {

					m.blockValidators++
					log.Printf("Received block validation, support count is now: %+v\n", m.blockValidators)
				}()
			default:
				log.Println("Warning: Command \"" + peerMsg.Command + "\" not supported")
			}
		default:
			// Timeout for 5 milliseconds to limit the number of iterations of the loop to 20 per second
			time.Sleep(5 * time.Millisecond)
		}

		// If we aren't already in a mining session and there is at least one transaction to be mined, pop a transaction
		// from the queue and broacast to network, starting a new mining session
		if !peersMining && m.transactionQueue.Len() > 0 && len(m.communicationComponent.GetPeerNodes()) > 0 {

			log.Println("Beginning a new mining session...")

			// Pop a Message from the transactionQueue
			toMine := m.popTransaction()

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
			m.proofFound = false

		} else if peersMining && m.proofFound && !running && m.blockValid {
			// End the mining session. Running to be set before the goroutine starts or else it will
			// get called multiple times
			running = true

			go func() {

				// If the lottery pool isn't empty, that means the network is using Proof of Stake,
				// so return the peers' stakes that they sent for the lottery
				if len(m.lotteryPool) > 0 {
					for _, entry := range m.lotteryPool {

						toSend, err := m.communicationComponent.GenerateMessage("STAKE", entry)
						if err != nil {
							// This would be a fatal error
							log.Printf("Fatal error generating message: %v\n", err)
							done = true
						}

						err = m.communicationComponent.SendMsgToPeer(toSend, entry.Peer)
						if err != nil {
							// This would be a fatal error
							log.Printf("Error sending message to peer: %v", err)
							done = true
						}

					}
					// Reset the state
					m.lotteryPool = nil
				}

				// Else if proofFound, conclude the current mining session by broadcasting a CONSENSUS message to the peers
				// to halt mining and cause consensus to be run so every peer gets a copy of the new longest chain
				log.Println("Broadcasting CONSENSUS message to peers...")

				toSend, err := m.communicationComponent.GenerateMessage("CONSENSUS", nil)
				if err != nil {
					// This would be a fatal error
					log.Printf("Fatal error generating message: %v\n", err)
					done = true
				}

				// Broadcast the CONSENSUS message to the peers on the network. All peers will send their chain copies when they
				// recieve a CONSENSUS
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
				m.blockValid = false
				m.candidateBlockQueue.Init()

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

	}
}

// terminate calls all of the interface-defined component clean-up methods
func (m Middleware) terminate() {

	fmt.Println("\nTerminating Middleware components...")

	m.communicationComponent.Terminate()

	fmt.Println("Exiting Middleware...")
}

// Pops a message of the Middleware's transactionQueue and returns it
func (m *Middleware) popTransaction() Transaction {

	// Get element from the front of the list
	poppedElement := m.transactionQueue.Front()

	// Remove the element, essentially "popping" it
	m.transactionQueue.Remove(poppedElement)

	//Convert the popped element, which is of type *Element, to *Message
	toMine := poppedElement.Value.(Transaction)

	return toMine
}

// Pops a message of the Middleware's transactionQueue and returns it
func (m *Middleware) popCandidateBlock() CandidateBlock {

	// Get element from the front of the list
	poppedElement := m.candidateBlockQueue.Front()

	// Remove the element, essentially "popping" it
	m.candidateBlockQueue.Remove(poppedElement)

	//Convert the popped element, which is of type *Element, to *Message
	toMine := poppedElement.Value.(CandidateBlock)

	return toMine
}

// Picks the winner of the proof of stake lottery and sends them the transaction data to mine
func (m *Middleware) runLottery() error {

	log.Println("Running lottery...")

	entries := []weightedrand.Choice{}

	for _, entry := range m.lotteryPool {
		newChoice := weightedrand.NewChoice(entry.Peer, uint(entry.Stake))
		entries = append(entries, newChoice)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	c, err := weightedrand.NewChooser(entries...)
	if err != nil {
		log.Printf("Error choosing lottery winner: %+v\n", err)
		return err
	}

	lotteryWinner := c.Pick().(PeerAddress)
	log.Printf("Lottery winner: %+v\n", lotteryWinner)

	toSend, err := m.communicationComponent.GenerateMessage("WINNER", nil)
	if err != nil {
		log.Printf("Eror generating message: %v\n", err)
		return err
	}

	err = m.communicationComponent.SendMsgToPeer(toSend, lotteryWinner)
	if err != nil {
		log.Printf("Error sending message to peer: %v", err)
		return err
	}

	return nil

}

// Checks if the candidate block got enough validation from the network. If not,
// applies penalty to the malicious peer who submitted the block and restarts the
// process with the next candidate block in the queue
func (m *Middleware) runValidation() (err error) {

	m.blockValidators = 0

	candidateBlock := m.popCandidateBlock()

	// Else if we're doing proof of stake, run the lottery again to choose the next winner

	toSend, msg_err := m.communicationComponent.GenerateMessage("VALIDATE", candidateBlock)
	if msg_err != nil {
		log.Printf("Fatal error generating message: %v\n", err)
		return msg_err
	}

	msg_err = m.communicationComponent.BroadcastMsgToNetwork(toSend)
	if msg_err != nil {
		// This would be a fatal error because if the network isn't able to validate
		// the block then malicious nodes could succeed
		log.Printf("Fatal error broadasting message to network: %v", err)
		return msg_err
	}

	// Give some time for peers to process the candidate block and respond with their validation
	time.AfterFunc(5*time.Second, func() {

		//If the block got enough validation (At least 50% of the network)
		if m.blockValidators >= int(math.Round(float64(len(m.communicationComponent.GetPeerNodes()))/2.0)) {
			log.Println("Validation successful. Ending current mining session...")
			// Send a reward to the successful miner, which will tell them to add the mined block to their chain,
			// which will become the new global chain once consensus is run
			fmt.Printf("DEBUG - from: '%+v', to: '%+v'\n", m.communicationComponent.GetSelfAddress().String(), candidateBlock.Miner.String())

			newData := Transaction{From: m.communicationComponent.GetSelfAddress().String(), To: candidateBlock.Miner.String(), Amount: REWARD_AMOUNT}

			toSend, msg_err := m.communicationComponent.GenerateMessage("TRANSACTION", newData)
			if msg_err != nil {
				log.Printf("Fatal error generating message: %v\n", err)
				err = msg_err
			}

			msg_err = m.communicationComponent.SendMsgToPeer(toSend, candidateBlock.Miner)
			if msg_err != nil {
				// This would be a fatal error because if a peer doesn't recieve the reward,
				// it wont add the new block to its chain, and the block will get lost if the next session begins
				log.Printf("Fatal Error sending reward to peer: %v", err)
				err = msg_err
			}

			m.blockValid = true

		} else {
			// If the validation is unsuccessful and we're doing PoS (lotteryPool len > 0), then remove the
			// lottery winner from the lottery pool so they don't receive  a stake refund
			if len(m.lotteryPool) > 0 {
				entryIndex := indexOf(candidateBlock.Miner, m.lotteryPool)
				m.lotteryPool[entryIndex] = m.lotteryPool[len(m.lotteryPool)-1]
				m.lotteryPool = m.lotteryPool[:len(m.lotteryPool)-1]
			}

			// We do not call this method again for proof of stake because when we run the lottery again, a new user is gonna send
			// the proof which is gonna cause this function to be run again anyway. We dont need to re-run ourself,
			// unlike for proof of work
			if len(m.lotteryPool) == 0 {
				// Reattempt validation
				m.runValidation()
			} else {
				m.proofFound = false
				m.runLottery()
			}
		}

	})

	return err
}

func indexOf(entry PeerAddress, entries []LotteryEntry) int {
	for i, e := range entries {
		if reflect.DeepEqual(e.Peer, entry) {
			return i
		}
	}
	// Return -1 if not find
	return -1
}
