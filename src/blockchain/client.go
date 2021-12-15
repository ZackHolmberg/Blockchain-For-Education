package blockchain

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
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
	{"bal", "Prints out the user's current wallet balance."},
}

// ============================ Client ============================

type Client struct {
	publicKey           ecdsa.PublicKey
	peerPublicKeys      map[int]*ecdsa.PublicKey
	privateKey          *ecdsa.PrivateKey
	peer                *Peer
	commandDescriptions [][]string
	communicator        CommunicationComponent
}

// Initialize is the interface method that calls this component's initialize method
func (c *Client) Initialize(com CommunicationComponent, p *Peer) error {

	c.commandDescriptions = commandDescriptions
	c.communicator = com
	c.peer = p

	// Generate public and private keys for digital signing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	c.privateKey = privateKey
	c.publicKey = c.privateKey.PublicKey

	// Initialize peerPublicKeys map
	c.peerPublicKeys = make(map[int]*ecdsa.PublicKey)

	// Broadcast public key and get Peers' public keys on startup
	data := c.ecdsaPublicKeytoPublicKey()
	toSend, err := com.GenerateMessage("PUBLIC_KEYS", data)
	if err != nil {
		log.Printf("Error generating message: %v\n", err)
	}

	err = com.BroadcastMsgToNetwork(toSend)
	if err != nil {
		log.Printf("Error sending Public Key to Peer: %v\n", err)
	}

	// Start a new thread running the sendTransaction method after the Peer has had some time to initialize
	time.AfterFunc(3*time.Second, c.sendTransaction)

	return nil
}

// Terminate is the interface method that calls this component's cleanup method
func (c Client) Terminate() {
	// No clean-up needed for this implementation
}

func (c Client) addPublicKey(pk PublicKey, peer PeerAddress) {

	publicKey := hexToPublicKey(pk.X, pk.Y)

	for port := range c.peerPublicKeys {
		// If we already have a key registered for the peer, just return
		if peer.Address.Port == port {
			return
		}
	}
	// fmt.Println("DEBUG - Adding new public key")
	c.peerPublicKeys[peer.Address.Port] = publicKey
}

func (c Client) Sign(t Transaction) (Transaction, error) {
	hash := t.ToString()
	signature, err := ecdsa.SignASN1(rand.Reader, c.privateKey, []byte(hash))
	if err != nil {
		return t, err
	}
	t.Signature = hex.EncodeToString(signature)
	return t, nil
}

func (c Client) Verify(t Transaction) bool {

	mapKey, _ := strconv.Atoi(strings.Split(t.From, ":")[1])
	key := c.peerPublicKeys[mapKey]
	temp := Transaction{To: t.To, From: t.From, Amount: t.Amount}
	hash := temp.ToString()
	sig, _ := hex.DecodeString(t.Signature)
	return ecdsa.VerifyASN1(key, []byte(hash), sig)
}

// HandleCommand is the interface method that handles the passed message
func (c *Client) HandleCommand(msg Message, com CommunicationComponent) (err error) {

	switch msg.Command {
	case "PUBLIC_KEY":
		go func() {
			log.Println("Received Peer's Public Key")
			publicKey := msg.Data.(PublicKey)
			c.addPublicKey(publicKey, msg.From)
		}()

	case "PUBLIC_KEYS":
		go func() {
			log.Println("Received request for Peer's Public Key, sending...")

			publicKey := msg.Data.(PublicKey)
			c.addPublicKey(publicKey, msg.From)
			data := c.ecdsaPublicKeytoPublicKey()
			toSend, err := com.GenerateMessage("PUBLIC_KEY", data)
			if err != nil {
				log.Printf("Error generating message: %v\n", err)
			}

			err = com.SendMsgToPeer(toSend, msg.From)
			if err != nil {
				log.Printf("Error sending Public Key to Peer: %v\n", err)
			}
		}()

	default:
		err = errors.New("command not supported")
	}

	return err

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
			// There must be at least one other node on the network (other than the Middleware) to send currency to
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
			fmt.Printf("Current wallet balance: %d\n", c.peer.wallet)
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
	if amount <= c.peer.wallet {

		peersList := c.communicator.GetPeerNodes()
		recipient := peersList[index]

		// Prevent the user from sending to the middleware
		if recipient.Address.Port != c.communicator.GetMiddlewarePeer().Address.Port {

			toStr := recipient.String()
			fromStr := c.communicator.GetSelfAddress().String()

			// Hit the Middleware endpoint
			// to create an entry in the blockchain for the transaction
			data := Transaction{From: fromStr, To: toStr, Amount: amount}

			data, err := c.Sign(data)
			if err != nil {
				return err
			}

			values := url.Values{"to": {data.To}, "from": {data.From}, "amount": {fmt.Sprint(amount)}, "signature": {data.Signature}}

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

			// If the Middleware accepts the transaction, send the currency to
			// the intended recipient

			toSend, err := c.communicator.GenerateMessage("TRANSACTION", data)
			if err != nil {
				return err
			}

			err = c.communicator.SendMsgToPeer(toSend, recipient)
			if err != nil {
				return err
			}

			// Finally, subtract the amount from this Peer's wallet
			c.peer.wallet -= amount

		} else {
			return errors.New("can't transfer money to Middleware")
		}
	} else {
		return errors.New("amount entered to send is greater than balance")
	}

	return nil
}

// hex.EncodeToString(hashed)
func hexToPublicKey(xHex string, yHex string) *ecdsa.PublicKey {
	xBytes, _ := hex.DecodeString(xHex)
	x := new(big.Int)
	x.SetBytes(xBytes)

	yBytes, _ := hex.DecodeString(yHex)
	y := new(big.Int)
	y.SetBytes(yBytes)

	pub := new(ecdsa.PublicKey)
	pub.X = x
	pub.Y = y

	pub.Curve = elliptic.P256()

	return pub
}

func (c Client) ecdsaPublicKeytoPublicKey() PublicKey {
	return PublicKey{X: hex.EncodeToString(c.publicKey.X.Bytes()), Y: hex.EncodeToString(c.publicKey.Y.Bytes())}
}
