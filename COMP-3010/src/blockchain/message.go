package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// Message is the struct that is marshalled/demarshalled between peers to communicate
type Message struct {
	From    Peer   `json:"from"`
	Command string `json:"command"`
	Data    Data   `json:"data,omitempty"`
}

// UnmarshalJSON is a custom JSON unmarshaller
func (m *Message) UnmarshalJSON(bytes []byte) error {

	var result map[string]interface{}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		log.Printf("Error unmarshalling message: %v\n", err)
		return err
	}
	// fmt.Printf("DEBUG - Unmarshal result is: %+v\n", result)

	// Unmarshal the data of the Peer who sent the message
	from := result["from"].(map[string]interface{})
	address := from["address"].(map[string]interface{})
	ip := net.ParseIP(address["IP"].(string))
	port := int(address["Port"].(float64))
	newAddress := net.UDPAddr{IP: ip, Port: port}
	lastMessageTime := from["lastMessageTime"].(string)
	parsedLastMessageTime, err := time.Parse(time.RFC3339, lastMessageTime)
	if err != nil {
		log.Println("Failed to parse time:", err.Error())
	}

	newPeer := Peer{Address: newAddress, LastMessageTime: parsedLastMessageTime}

	// Umarshall the command
	command := result["command"].(string)

	// Unmarshal the data
	data := result["data"]

	var dataStruct Data

	if data != nil {

		dataObject := data.(map[string]interface{})
		if val, ok := dataObject["from"]; ok {

			// Then the data is a transaction, so unmarshal into a Transaction struct
			from := val.(string)
			to := dataObject["to"].(string)
			amount := int(dataObject["amount"].(float64))
			dataStruct = Transaction{From: from, To: to, Amount: amount}

		} else if val, ok := dataObject["chainCopy"]; ok {
			// Then the data is a chain copy, so unmarshal into a ChainCopy struct
			list := val.([]interface{})
			newChain := Chain{ChainCopy: []Block{}}

			//Umarshal the chain
			for _, block := range list {
				blockMap := block.(map[string]interface{})
				dataMap := blockMap["Data"].(map[string]interface{})

				// We can assume that the Data will be of type Transaction, for now
				from := dataMap["from"].(string)
				to := dataMap["to"].(string)
				amount := int(dataMap["amount"].(float64))
				newTransaction := Transaction{From: from, To: to, Amount: amount}

				// Umarshal the rest of the block
				index := int(blockMap["Index"].(float64))
				timestamp := blockMap["Timestamp"].(string)
				prevHash := blockMap["PrevHash"].(string)
				hash := blockMap["Hash"].(string)
				newBlock := Block{Data: newTransaction, Index: index, Timestamp: timestamp, PrevHash: prevHash, Hash: hash}
				newChain.ChainCopy = append(newChain.ChainCopy, newBlock)
			}

			dataStruct = newChain

		} else if val, ok := dataObject["list"]; ok {
			// Then the data is a list of peer chains, so unmarshal into a PeerChains struct

			if val != nil {
				newList := PeerChains{List: [][]Block{}}
				list := val.([]interface{})

				for _, chain := range list {
					tempList := []Block{}

					//Umarshal the chain
					for _, block := range chain.([]interface{}) {
						blockMap := block.(map[string]interface{})

						dataMap := blockMap["Data"].(map[string]interface{})

						// We can assume that the Data will be of type Transaction, for now
						from := dataMap["from"].(string)
						to := dataMap["to"].(string)
						amount := int(dataMap["amount"].(float64))
						newTransaction := Transaction{From: from, To: to, Amount: amount}

						// Umarshal the rest of the block
						index := int(blockMap["Index"].(float64))
						timestamp := blockMap["Timestamp"].(string)
						prevHash := blockMap["PrevHash"].(string)
						hash := blockMap["Hash"].(string)
						newBlock := Block{Data: newTransaction, Index: index, Timestamp: timestamp, PrevHash: prevHash, Hash: hash}
						tempList = append(tempList, newBlock)
					}

					newList.List = append(newList.List, tempList)

				}

				dataStruct = newList

			} else {
				// Else if temp is nil, the Middlware sent an empty list, thus there are no existing peer chains, return
				// an empty PeerChains struct
				fmt.Println("No existing peer chains.")
				dataStruct = PeerChains{}
			}
		} else {
			// Else the data is of an unsupported type
			err := errors.New("error unmarshalling Data object: unsupported type")
			return err
		}
	}

	m.Command = command
	m.From = newPeer
	m.Data = dataStruct

	return nil
}

// TODO: Create helper functions that can be called and resued in different data cases. Including unmashalling a transaction, a block, and a chain
