package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// ============================ Communication ============================

// Communicator implements CommunicationsComponent and facilities Blockchain communication
type Communicator struct {
	self        *zeroconf.Server
	socket      *net.UDPConn
	peerNodes   []Peer
	peerMessage chan string
}

// Message is the struct that is marshalled/demarshalled between peers to communicate
type Message struct {
	from    Peer
	command string
	data    Data
}

// Peer represents a peer on the network and contains metadata about that peer
type Peer struct {
	address     net.UDPAddr
	lastMessage time.Time
}

// GetPeerChains is the interface method that retrieves the copy of the blockchain from every peer on the network
func (c Communicator) GetPeerChains() {
	//TODO: Implement
	// Wait - how actually am I going to do this?
	// Maybe Middleware will query every node and return a list perhaps??
}

// RecieveFromNetwork is the interface method that
// returns a Message that it reads from this peer's UDP socket
func (c Communicator) RecieveFromNetwork() (Message, error) {

	buf := make([]byte, 2048)

	// Read from the socket
	_, _, err := c.socket.ReadFromUDP(buf)
	if err != nil {
		fmt.Printf("Error reading from socket: %v", err)
		return Message{}, err
	}

	var message Message

	// Unmarshal the JSON into a Message
	err = json.Unmarshal(buf, &message)
	if err != nil {
		fmt.Printf("Error unmarshalling message: %v", err)
		return Message{}, err
	}

	// This removes the need to have a HandlePingFromNetwork function
	// Resets the peer's lastMessage timer
	message.from.lastMessage = time.Now()

	// If the peer that sent the message is not a known peer, add it to the peerNodes list
	if !knownPeer(c.peerNodes, message.from) {
		c.peerNodes = append(c.peerNodes, message.from)
	}

	return message, nil
}

// BroadcastToNetwork is the interface method that uses
// UDP to broadcast a message to all the peers on the network
func (c Communicator) BroadcastToNetwork(msg Message) {

	// Marshal the Message into JSON
	endcodedMessage, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message into json:", err)
		return
	}

	// Send the message to each known peer node
	for _, peer := range c.peerNodes {
		_, err := c.socket.WriteToUDP(endcodedMessage, &peer.address)
		if err != nil {
			fmt.Printf("Couldn't send message to peer %v", err)
		}
	}
}

// PingNetwork is the interface method that sends a ping to all known peer nodes
func (c Communicator) PingNetwork() {

	// Initialize the ping message to broadcast to peers on the network
	pingMessage := Message{command: "PING", data: nil}

	// call BroadcastToNetwork to broadcast ping message
	c.BroadcastToNetwork(pingMessage)
}

// InitializeCommunicator initializes a new communicator by initializing
// a socket and ZeroConf service and discovering other services
func (c *Communicator) InitializeCommunicator() {

	// Initialize the socket that this peer will communicate through
	c.socket = initializeSocket()

	// Initialize the service that this peer will join the p2p network through
	c.self = initializeService(c.socket.LocalAddr().(*net.UDPAddr).Port)

	// Discover peers on the p2p network
	c.peerNodes = discoverPeers()

}

// TerminateCommunicationComponent is the interface method that calls this component's cleanup method
func (c Communicator) TerminateCommunicationComponent() {
	c.terminateCommunicator()
}

// == Non-interface methods ==

// terminateCommunicator cleans up and terminates this peer's socket and service
func (c Communicator) terminateCommunicator() {

	fmt.Println("Terminating communicator...")

	//Shutdown self service instance
	c.self.Shutdown()

	//Close the socket
	c.socket.Close()
}

func initializeSocket() *net.UDPConn {
	//Dynamically get an unused port assigned by opening a socket with port set to 0
	addr := net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("127.0.0.1"),
	}
	s, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Failed to create new socket: %v\n", err)
		return nil
	}

	return s
}

func initializeService(port int) *zeroconf.Server {

	//Get name of host to use in the peerName
	se, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//Generate a UUID to ensure peerName is unique
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}

	//The name of the peer (must be unique on the network)
	peerName := fmt.Sprintf("%s-%s", se, out)
	peerName = strings.TrimSuffix(peerName, "\n")

	//The name of the service
	serviceName := "_blockchain-P2P-Network._udp"

	//The service's domain
	domain := "local."

	fmt.Println("Initiliazing service...")

	//Register the service
	self, err := zeroconf.Register(peerName, serviceName, domain, port, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		panic(err)
	}

	// Print out a summary of the new service
	fmt.Println("\nPublished service:")
	fmt.Println("- Name:", peerName)
	fmt.Println("- Type:", serviceName)
	fmt.Println("- Domain:", domain)
	fmt.Println("- Port:", port)

	return self
}

func discoverPeers() []Peer {
	// Discover all services on the '_blockchain-P2P-Network._udp' blockchain network
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
		return nil
	}

	newPeerNodes := []Peer{}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		//For all peers found on the network
		for entry := range results {

			newAddr := net.UDPAddr{IP: entry.AddrIPv4[0], Port: entry.Port, Zone: ""}
			newPeer := Peer{address: newAddr, lastMessage: time.Now()}
			// Append the new peer to the Communicator's list of known peer nodes
			newPeerNodes = append(newPeerNodes, newPeer)
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()
	err = resolver.Browse(ctx, "_blockchain-P2P-Network._udp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()

	// Wait some additional time to see debug messages on go routine shutdown.
	time.Sleep(1 * time.Second)

	fmt.Printf("Discovered peer nodes: %+v\n", newPeerNodes)

	return newPeerNodes
}

// knownPeer takes a slice of peers and looks for a particular peer in it. If found it will
// return true, otherwise it will return false.
func knownPeer(slice []Peer, p Peer) bool {
	for _, c := range slice {
		if c.address.IP.String() == p.address.IP.String() && c.address.Port == p.address.Port {
			return true
		}
	}
	return false
}
