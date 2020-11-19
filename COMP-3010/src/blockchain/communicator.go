package blockchain

// ============================ Communication ============================

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	GetPeerChains()
	PingNetwork()
	RecieveFromClient()
	SendToClient()
	RecieveFromNetwork()
	BroadcastToNetwork()
}

// Communicator implements CommunicationsComponent and facilities Blockchain communication
type Communicator struct {
}

func (c Communicator) GetPeerChains() {
}

func (c Communicator) PingNetwork() {
}

func (c Communicator) RecieveFromClient() {
}

func (c Communicator) SendToClient() {
}

func (c Communicator) RecieveFromNetwork() {
}

func (c Communicator) BroadcastToNetwork() {
}
