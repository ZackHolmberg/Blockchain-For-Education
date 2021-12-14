package blockchain

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const MIDDLEWARE_URL = "http://localhost:8090/newTransaction"

var commandDescriptions = [][]string{
	{"help", "Lists all valid commands with their descriptions."},
	{"transaction", "Prompts user for recipient and amount values to send a new transaction. Expected input is of the form 'index of recipient in peers list,amount'. For example, '1,5' excluding the apostrophes."},
	{"peers", "Lists all of the peers on the network that the user can send currency to.\nExample output:\n'index=1, ip=::1, port=55514'"},
	{"bal", "Prints out the user's current wallet balance"},
}

// ============================ Client ============================

type Client struct {
	// publicKey           string
	// privateKey          string
	wallet              *int
	commandDescriptions [][]string
	communicator        CommunicationComponent
}

// Initialize is the interface method that calls this component's initialize method
func (c *Client) Initialize(com CommunicationComponent, wallet *int) error {

	c.commandDescriptions = commandDescriptions
	c.communicator = com
	c.wallet = wallet
	// TODO: Generate public and private keys for digital signing

	// Start a new thread running the sendTransaction method after the Peer has had some time to initialize
	time.AfterFunc(3*time.Second, c.sendTransaction)
	return nil
}

// Terminate is the interface method that calls this component's cleanup method
func (c Client) Terminate() {
	// No clean-up needed for this implementation
}

// SendTransaction is the interface method that calls this component's cleanup method
func (c Client) sendTransaction() {

	// We run indefinitely
	for {
		consoleReader := bufio.NewReader(os.Stdin)
		log.Println("Enter a command or enter 'help' for a list of commands: ")
		input, _ := consoleReader.ReadString('\n')

		input = strings.ToLower(input)
		input = strings.TrimRight(input, "\n")

	CommandSwitch:
		switch input {
		case "help":
			c.printCommands()
		case "peers":
			c.listPeers()
		case "transaction":
			// There must be at least one other node on the network to send currency to
			if len(c.communicator.GetPeerNodes()) > 1 {
				fmt.Println("Enter transaction data or 'cancel' to cancel.")
				//Get transaction input
				input, _ = consoleReader.ReadString('\n')
				input = strings.TrimRight(input, "\n")
				if input == "cancel" {
					break CommandSwitch
				} else {
					s := strings.Split(input, ",")

					recipientIndex, err := strconv.Atoi(s[0])
					if err != nil {
						fmt.Println("Incorrect input, please enter 'help' to see expected transaction input and try again")
						break CommandSwitch
					}

					amount, err := strconv.Atoi(s[1])
					if err != nil {
						fmt.Println("Incorrect input, please enter 'help' to see expected transaction input and try again")
						break CommandSwitch
					}

					err = c.createNewTransaction(recipientIndex, amount)
					if err != nil {
						fmt.Printf("Error creating new transaction: %+v\n", err)
						break CommandSwitch
					}
				}
			} else {
				fmt.Println("Warning: Cannot create new transaction, no other Peers on the network")
			}
		case "bal":
			fmt.Printf("Current wallet balance: %d\n", *c.wallet)
		default:
			fmt.Printf("Error: Invalid command '%s', Please try again.\n", input)
		}
	}
}

func (c Client) printCommands() {
	fmt.Println("===== Commands =====")
	for _, desc := range c.commandDescriptions {
		fmt.Printf("%s - %s\n", desc[0], desc[1])
	}
	fmt.Println("====================")

}

func (c Client) listPeers() {

	peersList := c.communicator.GetPeerNodes()
	if len(peersList) > 1 {
		fmt.Println("===== Known Peers =====")
		for i, peer := range peersList {
			// Exclude the middleware from the list of peers to send currency to
			if peer.Address.Port != 8080 {
				fmt.Printf("index=%d, ip=%v, port=%v\n", i, peer.Address.IP, peer.Address.Port)
			} else {
				fmt.Printf("index=%d, ip=%v, port=%v [Middleware Peer]\n", i, peer.Address.IP, peer.Address.Port)
			}
		}
		fmt.Println("====================")
	} else {
		fmt.Println("No other Peers on the network.")
	}

}

func (c Client) createNewTransaction(index int, amount int) error {

	// First, check if the user has the amount of currency they are wanting to send
	if amount <= *c.wallet {

		peersList := c.communicator.GetPeerNodes()
		recipient := peersList[index]

		// Prevent the user from sending to the middleware
		if recipient.Address.Port != c.communicator.GetMiddlewarePeer().Address.Port {

			//Send currency to the intended transaction recipient
			toStr := recipient.String()
			fromStr := c.communicator.GetSelfAddress().String()
			data := Transaction{From: fromStr, To: toStr, Amount: amount}

			toSend, err := c.communicator.GenerateMessage("TRANSACTION", data)
			if err != nil {
				return err
			}

			err = c.communicator.SendMsgToPeer(toSend, recipient)
			if err != nil {
				return err
			}

			// If sending the transaction is succesful, hit the Middleware endpoint
			// to create an entry in the blockchain for the transaction
			amountStr := fmt.Sprint(amount)
			values := url.Values{"to": {toStr}, "from": {fromStr}, "amount": {amountStr}}

			// Hit the Middleware's create transaction endpoint
			resp, err := http.PostForm(MIDDLEWARE_URL, values)

			if err != nil {
				return err
			}

			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			// Return the error from the server if the request is not successful
			if resp.StatusCode != 200 {
				return errors.New(string(body))
			} else {
				fmt.Println(string(body))
			}

		} else {
			return errors.New("can't transfer money to Middleware")
		}
	} else {
		return errors.New("amount entered to send is greater than balance")
	}

	return nil
}
