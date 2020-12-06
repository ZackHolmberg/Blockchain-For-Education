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
func (c *Communicator) RecieveFromNetwork() {

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
	// fmt.Println("DEBUG - Read a message from socket", string(buf))

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
	}

	// Update the known peer's LastMessage value
	c.peerNodes = updateLastMessage(c.peerNodes, message.From)

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
		// fmt.Println("DEBUG - Sent a message from socket")

	}
}

// PingNetwork is the interface method that sends a ping to all known peer nodes
func (c Communicator) PingNetwork() {

	if len(c.peerNodes) > 0 {

		log.Println("Sending pings...")
		fmt.Printf("DEBUG - Peer list: %+v\n", c.peerNodes)

		addr := getUDPAddr(c.socket)

		// Initialize the ping message to broadcast to peers on the network
		pingMessage := Message{Command: "PING", Data: nil, From: Peer{Address: addr, LastMessage: time.Now()}}

		//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
		temp, _ := net.LookupIP("localhost")
		pingMessage.From.Address.IP = temp[0]
		// =====================================================

		// call BroadcastToNetwork to broadcast ping message
		c.BroadcastToNetwork(pingMessage)
	} else {
		log.Println("No known peer nodes, not sending pings")
	}
}

// Initialize initializes a new communicator by initializing
// a socket and ZeroConf service and discovering other services
func (c *Communicator) Initialize() {

	// Initialize the socket that this peer will communicate through
	c.socket = initializeSocket()

	// Initialize the service that this peer will join the p2p network through
	c.self = initializeService(getUDPAddr(c.socket).Port)

	// Discover peers on the p2p network
	c.peerNodes = discoverPeers(getUDPAddr(c.socket))

	// Initialize peer Message channel
	c.peerMessage = make(chan Message)

}

// InitializeWithPort "overloads" thr Initialize method, and initializes
// the communicator with the passed well-defined port, instead of dynamically assgning a port
func (c *Communicator) InitializeWithPort(port int) {

	// Initialize the socket that this peer will communicate through
	c.socket = initializeSocketWithPort(port)

	// Initialize the service that this peer will join the p2p network through
	c.self = initializeService(port)

	// Discover peers on the p2p network
	c.peerNodes = discoverPeers(getUDPAddr(c.socket))

	// Initialize peer Message channel
	c.peerMessage = make(chan Message)
}

// Terminate is the interface method that calls this component's cleanup method
func (c Communicator) Terminate() {
	c.terminateCommunicator()
}

// PrunePeerNodes is the interface method that removes nodes from the peerNodes list which  have
// not sent a message within the previous 75 seconds, as we assume that node to have gone down
// in this case
func (c *Communicator) PrunePeerNodes() {

	for i, peer := range c.peerNodes {
		if time.Since(peer.LastMessage).Seconds() >= 20 {
			log.Printf("Pruning peer node: %+v\n", peer)
			fmt.Printf("DEBUG - List before removing: %+v\n", c.peerNodes)
			c.peerNodes = removeFromList(c.peerNodes, peer, i)
			fmt.Printf("DEBUG - List after removing: %+v\n", c.peerNodes)

		}
	}

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
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v\n", err)
		return nil
	}

	s, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("Failed to create new socket: %v\n", err)
		return nil
	}

	return s
}

// initializeSocketWithPort "overloads" the initializeSocket method, and initializes
// the socket with the passed well-defined port, instead of dynamically assgning a port
func initializeSocketWithPort(port int) *net.UDPConn {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v\n", err)
		return nil
	}

	s, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to create new socket: %v\n", err)
		return nil
	}

	fmt.Printf("DEBUG - Intialized socket: %+v\n", addr)

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
			if err != nil {
				log.Println("Failed to resolve UDP address:", err.Error())
				continue
			}

			//If the entry is the service corresponding to this peer, ignore it
			if equalPeers(*newAddr, addr) {
				// fmt.Println("DEBUG - Peers equal but not skipping")

				continue
			}

			newPeer := Peer{Address: *newAddr, LastMessage: time.Now()}

			//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
			temp, err := net.LookupIP("localhost")
			newPeer.Address.IP = temp[0]
			// =====================================================

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

	addr := net.UDPAddr{IP: ip[0], Port: port, Zone: ""}

	return addr
}

func removeFromList(peers []Peer, p Peer, i int) []Peer {

	peers[i] = peers[len(peers)-1]
	peers[len(peers)-1] = Peer{}
	return peers[:len(peers)-1]
}

func updateLastMessage(p []Peer, t Peer) []Peer {

	newList := p
	for i := range newList {
		if equalPeers(newList[i].Address, t.Address) {
			newList[i].LastMessage = time.Now()
		}
	}

	return newList
}

func equalPeers(p1 net.UDPAddr, p2 net.UDPAddr) bool {
	fmt.Printf("DEBUG - Comparing peers: %+v and %+v\n", p1, p2)
	return p1.IP.String() == p2.IP.String() && p1.Port == p2.Port
}
