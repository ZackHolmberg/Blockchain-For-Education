package blockchain

// ============================ Communication ============================

// Communicator implements CommunicationsComponent and facilities Blockchain communication
type Communicator struct {
}

// GetPeerChains is the interface method that retrieves the copy of the blockchain from every peer on the network
func (c Communicator) GetPeerChains() {
}

// RecieveFromClient is the interface method that
func (c Communicator) RecieveFromClient() {
}

// SendToClient is the interface method that
func (c Communicator) SendToClient() {
}

// RecieveFromNetwork is the interface method that
func (c Communicator) RecieveFromNetwork() {
}

// BroadcastToNetwork is the interface method that
func (c Communicator) BroadcastToNetwork() {
}

//Non-interface methods

// PingNetwork is the interface method that
func (c Communicator) PingNetwork() {
	//Will use SendToFromNetwork
}

// HandlePingFromNetwork is the interface method that
func (c Communicator) HandlePingFromNetwork() {
	//Will use RecieveFromNetwork
}
