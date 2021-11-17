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
	middleware  Peer
}

// Peer represents a peer on the network and contains metadata about that peer
type Peer struct {
	Address         net.UDPAddr `json:"address"`
	LastMessageTime time.Time   `json:"lastMessageTime"`
}

// GetMessageChannel is the interface retriever method that returns the channel that a message from a peer is put into upon read
func (c Communicator) GetMessageChannel() chan Message {
	return c.peerMessage
}

// GetPeerNodes is the interface retriever method that returns this node's list of peers
func (c Communicator) GetPeerNodes() []Peer {
	return c.peerNodes
}

// GetMiddlewarePeer is the interface retriever method that returns a pointer to the Middlware peer
func (c Communicator) GetMiddlewarePeer() Peer {
	return c.middleware
}

// RecieveFromNetwork is the interface method that
// returns a Message that it reads from this peer's UDP socket
func (c *Communicator) RecieveFromNetwork(withTimeout bool) error {

	buf := make([]byte, 65535)
	if withTimeout {
		c.socket.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
	}

	// Read from the socket
	len, _, err := c.socket.ReadFromUDP(buf)
	if err != nil {
		if er, ok := err.(net.Error); ok && er.Timeout() {
			// This was a timeout error, so just return as there was nothing to be read
			return nil
		} else if err != nil {
			// This was an error, but not a timeout, so print it out
			log.Printf("Error reading from socket: %v\n", err)
			return err
		}
	}
	// fmt.Printf("DEBUG - Read a message from socket: %+v\n", string(buf))

	message := new(Message)

	err = message.UnmarshalJSON(buf[:len])
	if err != nil {
		log.Printf("Error unmarshalling message: %v\n", err)
		return err
	}

	// fmt.Printf("DEBUG - Unmarshalled message from socket: %+v\n", message)

	// If the peer that sent the message is not a known peer, add it to the peerNodes list
	if !knownPeer(c.peerNodes, message.From) {
		// fmt.Println("DEBUG - Peer is not known")
		c.peerNodes = append(c.peerNodes, message.From)
	}
	// fmt.Println("DEBUG - Checked if known peer")

	// Update the known peer's LastMessage value
	c.peerNodes = updateLastMessage(c.peerNodes, message.From)
	// fmt.Println("DEBUG - Updated last message")

	c.peerMessage <- *message
	// fmt.Println("DEBUG - Successfully received message")

	return nil
}

// PingNetwork is the interface method that sends a ping to all known peer nodes
func (c Communicator) PingNetwork() error {

	if len(c.peerNodes) > 0 {

		log.Println("Broadcasting pings...")

		// call BroadcastToNetwork to broadcast ping message
		err := c.BroadcastMsgToNetwork("PING", nil)
		if err != nil {
			log.Printf("Error broadcasting pings: %v", err)
			return err
		}

	} else {
		log.Println("No known peer nodes, not sending pings")
	}

	return nil
}

// Initialize initializes a new communicator by initializing
// a socket and ZeroConf service and discovering other services
func (c *Communicator) Initialize() error {

	// Initialize the socket that this peer will communicate through
	c.socket = initializeSocket()

	// Get the UDPAddr of this peer's socket
	addr, err := getUDPAddr(c.socket)
	if err != nil {
		log.Printf("Failed to get UDP address: %+v\n", err)
		return err
	}

	// Initialize the service that this peer will join the p2p network through
	c.self = initializeService(addr.Port)

	// Discover peers on the p2p network
	c.discoverPeers(addr)

	// Initialize peer Message channel. Need to have a 1 message buffer for peer startup.
	c.peerMessage = make(chan Message, 1)

	return nil

}

// InitializeWithPort "overloads" thr Initialize method, and initializes
// the communicator with the passed well-defined port, instead of dynamically assgning a port
func (c *Communicator) InitializeWithPort(port int) error {

	// Initialize the socket that this peer will communicate through
	c.socket = initializeSocketWithPort(port)

	// Get the UDPAddr of this peer's socket
	addr, err := getUDPAddr(c.socket)
	if err != nil {
		log.Printf("Failed to get UDP address: %+v\n", err)
		return err
	}

	// Initialize the service that this peer will join the p2p network through
	c.self = initializeService(port)

	// Discover peers on the p2p network
	err = c.discoverPeers(addr)
	if err != nil {
		log.Printf("Error discovering peers: %+v\n", err)
		return err
	}

	// Initialize peer Message channel
	c.peerMessage = make(chan Message)

	return nil
}

// Terminate is the interface method that calls this component's cleanup method
func (c Communicator) Terminate() {
	c.terminateCommunicator()
}

// BroadcastMsgToNetwork is the interface method that uses
// helper methods to broadcast messages
func (c Communicator) BroadcastMsgToNetwork(cmd string, d Data) error {

	toSend, err := c.GenerateMessage(cmd, d)
	if err != nil {
		log.Printf("Error generating message: %v\n", err)
		return err
	}

	//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
	temp, _ := net.LookupIP("localhost")
	toSend.From.Address.IP = temp[0]
	// =====================================================

	err = c.broadcastToNetwork(toSend)
	if err != nil {
		log.Printf("Error broadcasting message: %v\n", err)
		return err
	}

	return nil
}

// SendMsgToPeer is the interface method that uses
// helper methods to send a message to a peer
func (c Communicator) SendMsgToPeer(cmd string, d Data, p Peer) error {

	toSend, err := c.GenerateMessage(cmd, d)
	if err != nil {
		log.Printf("Error generating message: %v\n", err)
		return err
	}

	//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
	temp, _ := net.LookupIP("localhost")
	toSend.From.Address.IP = temp[0]
	// =====================================================

	err = c.sendToPeer(toSend, p)
	if err != nil {
		log.Printf("Error sending message to Peer: %v\n", err)
		return err
	}

	return nil
}

// PrunePeerNodes is the interface method that removes nodes from the peerNodes list which  have
// not sent a message within the previous 75 seconds, as we assume that node to have gone down
// in this case
func (c *Communicator) PrunePeerNodes() {

	for i, peer := range c.peerNodes {
		if time.Since(peer.LastMessageTime).Seconds() >= 75 {
			log.Printf("Pruning peer node: %+v\n", peer)
			c.peerNodes = removeFromList(c.peerNodes, peer, i)
		}
	}

}

// GenerateMessage uses the passed values to generate a new Message
func (c *Communicator) GenerateMessage(cmd string, data Data) (Message, error) {

	addr, err := getUDPAddr(c.socket)
	if err != nil {
		log.Fatalf("Failed to get UDP address: %+v\n", err)
		return Message{}, err
	}

	newMessage := Message{Command: cmd, Data: data, From: Peer{Address: addr, LastMessageTime: time.Now()}}

	return newMessage, nil
}

// ==================== Non-interface, helper methods ========================

// terminateCommunicator cleans up and terminates this peer's socket and service
func (c Communicator) terminateCommunicator() {

	log.Println("Terminating communicator...")

	//Shutdown self service instance
	c.self.Shutdown()

	//Close the socket
	err := c.socket.Close()

	if err != nil {
		log.Printf("Error closing socket: %+v\n", err)
	}
}

func initializeSocket() *net.UDPConn {

	//Dynamically get an unused port assigned by opening a socket with port set to 0
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %+v\n", err)
	}

	s, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to create new socket: %+v\n", err)
	}

	return s
}

// initializeSocketWithPort "overloads" the initializeSocket method, and initializes
// the socket with the passed well-defined port, instead of dynamically assgning a port
func initializeSocketWithPort(port int) *net.UDPConn {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v\n", err)
	}

	s, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to create new socket: %v\n", err)
	}

	return s
}

// Errors that occur within this function and similar ones do not need to be passed up to the caller
// because the program just exits if an error occurs
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

func (c *Communicator) discoverPeers(addr net.UDPAddr) error {
	// Discover all services on the '_blockchain-P2P-Network._udp' blockchain network
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Printf("Failed to initialize resolver: %+v\n", err)
		return err
	}

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
				continue
			}

			newPeer := Peer{Address: *newAddr, LastMessageTime: time.Now()}

			//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
			temp, err := net.LookupIP("localhost")
			if err != nil {
				log.Println("Failed to LookupIP:", err.Error())
				continue
			}
			newPeer.Address.IP = temp[0]
			// =====================================================

			if newAddr.Port == 8080 {
				// If this peer is the Middleware, set it as the communicators Middleware node value
				c.middleware = newPeer
			}

			// Append the new peer to the Communicator's list of known peer nodes
			c.peerNodes = append(c.peerNodes, newPeer)

		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(3))
	defer cancel()
	err = resolver.Browse(ctx, "_blockchain-P2P-Network._udp", "local.", entries)
	if err != nil {
		log.Printf("Failed to browse: %+v\n", err)
		return err
	}

	<-ctx.Done()

	// Wait some additional time to see debug messages on go routine shutdown.
	time.Sleep(1 * time.Second)

	fmt.Printf("Discovered peer nodes: %+v\n", c.peerNodes)

	return nil

}

// broadcastToNetwork is the helper method that uses
// UDP to broadcast a message to all the peers on the network
func (c Communicator) broadcastToNetwork(msg Message) error {

	// Marshal the Message into JSON
	endcodedMessage, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling message: %v\n", err)
		return err
	}

	// fmt.Printf("DEBUG - Broadcasting to peer list: %+v\n", c.peerNodes)

	// Send the message to each known peer node
	for _, peer := range c.peerNodes {

		//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
		temp, _ := net.LookupIP("localhost")
		peer.Address.IP = temp[0]
		// =====================================================

		// fmt.Printf("DEBUG - Broadcasting a message to peer: %+v\n", peer.Address)

		_, err := c.socket.WriteToUDP(endcodedMessage, &peer.Address)
		if err != nil {
			log.Printf("Couldn't send message to peer during broadcast: %v\n", err)
			return err
		}

	}

	return nil
}

// sendToPeer is the helper method that sends a
// UDP message to one peer on the network
func (c Communicator) sendToPeer(msg Message, p Peer) error {

	// Marshal the Message into JSON
	endcodedMessage, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling message: %v\n", err)
		return err
	}

	//=========== TODO: REMOVE AFTER DEVELOPMENT ===========
	temp, _ := net.LookupIP("localhost")
	p.Address.IP = temp[0]
	// =====================================================

	// Send the message to the peer
	// fmt.Printf("DEBUG - Sending a message to peer: %+v\n", p.Address)
	_, err = c.socket.WriteToUDP(endcodedMessage, &p.Address)
	if err != nil {
		log.Printf("Couldn't send message to peer: %v\n", err)
		return err
	}

	return nil
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

func getUDPAddr(c *net.UDPConn) (net.UDPAddr, error) {

	port := c.LocalAddr().(*net.UDPAddr).Port

	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Error resolving hostname: %+v", err)
		return net.UDPAddr{}, nil
	}

	ip, err := net.LookupIP(hostname)
	if err != nil {
		log.Printf("Error resolving hostname IP: %+v", err)
		return net.UDPAddr{}, nil
	}

	addr := net.UDPAddr{IP: ip[0], Port: port, Zone: ""}

	return addr, nil
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
			newList[i].LastMessageTime = time.Now()
		}
	}

	return newList
}

func equalPeers(p1 net.UDPAddr, p2 net.UDPAddr) bool {
	// fmt.Printf("DEBUG - Comparing p1: %#v ane p2 %#v", p1, p2)
	return p1.IP.String() == p2.IP.String() && p1.Port == p2.Port
}
