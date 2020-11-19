package blockchain

// ============================ Communication ============================

// Communicator implements CommunicationsComponent and facilities Blockchain communication
type Communicator struct {
	peerNodes []string
}

// GetPeerChains is the interface method that retrieves the copy of the blockchain from every peer on the network
func (c Communicator) GetPeerChains() {
	//TODO: Implement
}

// RecieveFromClient is the interface method that
func (c Communicator) RecieveFromClient() {
	//TODO: Implement
}

// SendToClient is the interface method that
func (c Communicator) SendToClient() {
	//TODO: Implement
}

// RecieveFromNetwork is the interface method that
func (c Communicator) RecieveFromNetwork() {
	//TODO: Implement
}

// BroadcastToNetwork is the interface method that
func (c Communicator) BroadcastToNetwork() {
	//TODO: Implement
}

// == Non-interface methods ==

// PingNetwork is the interface method that
func (c Communicator) PingNetwork() {
	//Will use SendToFromNetwork
	//TODO: Implement
}

// HandlePingFromNetwork is the interface method that
func (c Communicator) HandlePingFromNetwork() {
	//Will use RecieveFromNetwork
	//TODO: Implement
}
