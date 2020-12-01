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
	peerMessage chan Message
}

// Message is the struct that is marshalled/demarshalled between peers to communicate
type Message struct {
	From    Peer
	Command string
	Data    Data
}

// Peer represents a peer on the network and contains metadata about that peer
type Peer struct {
	Address     net.UDPAddr
	LastMessage time.Time
}

// GetMessageChannel is the interface gett method that returns the channel that a message from a peer is put into upon read
func (c Communicator) GetMessageChannel() chan Message {
	return c.peerMessage
}

// GetPeerChains is the interface method that retrieves the copy of the blockchain from every peer on the network
func (c Communicator) GetPeerChains() {
	// TODO: Implement
	// Wait - how actually am I going to do this?
	// Maybe Middleware will query every node and return a list perhaps??
}

// RecieveFromNetwork is the interface method that
// returns a Message that it reads from this peer's UDP socket
func (c Communicator) RecieveFromNetwork() {

	buf := make([]byte, 2048)
	c.socket.SetReadDeadline(time.Now().Add(1 * time.Millisecond))

	// Read from the socket
	len, _, err := c.socket.ReadFromUDP(buf)
	if err != nil {
		if er, ok := err.(net.Error); ok && er.Timeout() {
			// This was a timeout error, so just return as there was nothing to be read
			return
		} else if err != nil {
			// This was an error, but not a timeout, so print it out
			log.Printf("Error reading from socket: %v\n", err)
			return
		}
	}
	fmt.Println("DEBUG - Read a message from socket", string(buf))

	var message Message

	// Unmarshal the JSON into a Message
	err = json.Unmarshal(buf[:len], &message)
	if err != nil {
		log.Printf("Error unmarshalling message: %v\n", err)
		return
	}

	// fmt.Printf("DEBUG - Unmarshalled message from socket: %+v\n", message)

	// If the peer that sent the message is not a known peer, add it to the peerNodes list
	if !knownPeer(c.peerNodes, message.From) {
		c.peerNodes = append(c.peerNodes, message.From)
	} else {
		//TODO: Update the existing peer's LastMessage value

	}

	c.peerMessage <- message
}

// BroadcastToNetwork is the interface method that uses
// UDP to broadcast a message to all the peers on the network
func (c Communicator) BroadcastToNetwork(msg Message) {

	// fmt.Printf("DEBUG - Unmarshalled Message to send: %+v\n", msg)

	// Marshal the Message into JSON
	endcodedMessage, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling message: %v\n", err)
		return
	}
	// fmt.Printf("DEBUG - Marshalled Message to send: %+v\n", string(endcodedMessage))

	// Send the message to each known peer node
	for _, peer := range c.peerNodes {
		fmt.Printf("DEBUG - Broadcasting a message to peer: %+v\n", peer.Address)
		_, err := c.socket.WriteToUDP(endcodedMessage, &peer.Address)
		if err != nil {
			log.Printf("Couldn't send message to peer: %v\n", err)
		}
		fmt.Println("DEBUG - Sent a message from socket")

	}
}

// PingNetwork is the interface method that sends a ping to all known peer nodes
func (c Communicator) PingNetwork() {

	log.Println("Sending pings...")

	addr := getUDPAddr(c.socket)

	// Initialize the ping message to broadcast to peers on the network
	pingMessage := Message{Command: "PING", Data: nil, From: Peer{Address: addr, LastMessage: time.Now()}}

	// call BroadcastToNetwork to broadcast ping message
	c.BroadcastToNetwork(pingMessage)
}

// InitializeCommunicator initializes a new communicator by initializing
// a socket and ZeroConf service and discovering other services
func (c *Communicator) InitializeCommunicator() {

	// Initialize the socket that this peer will communicate through
	c.socket = initializeSocket()

	// Initialize the service that this peer will join the p2p network through
	c.self = initializeService(getUDPAddr(c.socket).Port)

	// Discover peers on the p2p network
	c.peerNodes = discoverPeers(getUDPAddr(c.socket))

	// Initialize peer Message channel
	c.peerMessage = make(chan Message)

}

// TerminateCommunicationComponent is the interface method that calls this component's cleanup method
func (c Communicator) TerminateCommunicationComponent() {
	c.terminateCommunicator()
}

// == Non-interface methods ==

// terminateCommunicator cleans up and terminates this peer's socket and service
func (c Communicator) terminateCommunicator() {

	log.Println("Terminating communicator...")

	//Shutdown self service instance
	c.self.Shutdown()

	//Close the socket
	c.socket.Close()
}

func initializeSocket() *net.UDPConn {

	//Dynamically get an unused port assigned by opening a socket with port set to 0
	addr, err := net.ResolveUDPAddr("udp", "0")

	s, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("Failed to create new socket: %v\n", err)
		return nil
	}

	return s
}

func initializeService(port int) *zeroconf.Server {

	//Get name of host to use in the peerName
	se, err := os.Hostname()
	if err != nil {
		log.Fatalln("Error getting hostname: ", err)
	}

	//Generate a UUID to ensure peerName is unique
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatalln("Error generating UUID: ", err)
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
		log.Fatalln("Error registering service: ", err)
	}

	// Print out a summary of the new service
	fmt.Println("\nPublished service:")
	fmt.Println("- Name:", peerName)
	fmt.Println("- Type:", serviceName)
	fmt.Println("- Domain:", domain)
	fmt.Println("- Port:", port)

	return self
}

func discoverPeers(addr net.UDPAddr) []Peer {
	// Discover all services on the '_blockchain-P2P-Network._udp' blockchain network
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	newPeerNodes := []Peer{}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		//For all peers found on the network
		for entry := range results {
			newAddr, err := net.ResolveUDPAddr(addr.Network(), fmt.Sprintf("%v:%d", entry.AddrIPv4[0], entry.Port))
			// if newAddr.IP.String() == addr.IP.String() && newAddr.Port == addr.Port {
			// 	//If the entry is the service corresponding to this peer, ignore it
			// 	continue
			// }
			fmt.Printf("newAddr - %+v\n", newAddr)
			if err != nil {
				log.Println("Failed to resolve UDP address:", err.Error())
				continue
			}
			newPeer := Peer{Address: *newAddr, LastMessage: time.Now()}
			fmt.Printf("newPeer address - %+v\n", newPeer.Address)
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
		if c.Address.IP.String() == p.Address.IP.String() && c.Address.Port == p.Address.Port {
			return true
		}
	}
	return false
}

func getUDPAddr(c *net.UDPConn) net.UDPAddr {

	port := c.LocalAddr().(*net.UDPAddr).Port

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("Error getting hostname: ", err)
	}

	ip, err := net.LookupIP(hostname)
	if err != nil {
		log.Fatalln("Error getting hostname: ", err)
	}

	return net.UDPAddr{IP: ip[0], Port: port, Zone: ""}
}
