package blockchain

// ============================ Block ============================

// Block is the Block object
type Block struct {
	Index     int
	Timestamp string
	Data      Data
	PrevHash  string
	Hash      string
}

// Data is an interface used to standardize methods for any type of Block data
type Data interface {
	GetData() Data
	ToString() string
}

// Transaction is a type of Block data
type Transaction struct {
	From   string
	To     string
	Amount int
}

// GetData is the interface method that is required to retrieve Data object
func (t Transaction) GetData() Data {
	return t
}

// ToString is the interface method that is required to transform the Data object into a string for communication
func (t Transaction) ToString() string {
	//TODO: Implement
	return "temp"
}
