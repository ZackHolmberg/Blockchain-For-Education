package blockchain

import (
	"encoding/json"
	"fmt"
)

// Data is an interface used to standardize methods for any type of Block data
type Data interface {
	GetData() Data
	ToString() string
}

// =========== Transaction ===========

// Transaction is a type of Data
type Transaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

// GetData is the interface method that is required to retrieve Data object
func (t Transaction) GetData() Data {
	return t
}

// ToString is the interface method that is required to transform the Data object into a string for communication
func (t Transaction) ToString() string {
	b, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

// =========== Chain ===========

// Chain contains a slice, or chain, of blocks, representing a blockchain
type Chain struct {
	ChainCopy []Block `json:"chainCopy"`
}

// GetData is the interface method that is required to retrieve Data object
func (c Chain) GetData() Data {
	return c
}

// ToString is the interface method that is required to transform the Data object into a string for communication
func (c Chain) ToString() string {
	b, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

// =========== PeerChains ===========

// PeerChains is a list of all the copies of the blockchain on the network
type PeerChains struct {
	List [][]Block `json:"list"`
}

// GetData is the interface method that is required to retrieve Data object
func (p PeerChains) GetData() Data {
	return p
}

// ToString is the interface method that is required to transform the Data object into a string for communication
func (p PeerChains) ToString() string {
	b, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

// =========== LotteryEntry ===========

// LotteryEntry represents one entry in the proof of stake lottery
type LotteryEntry struct {
	Stake int         `json:"stake"`
	Peer  PeerAddress `json:"peer"`
}

// GetData is the interface method that is required to retrieve Data object
func (l LotteryEntry) GetData() Data {
	return l
}

// ToString is the interface method that is required to transform the Data object into a string for communication
func (p LotteryEntry) ToString() string {
	b, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}
