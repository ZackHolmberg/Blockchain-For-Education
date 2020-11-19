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

func (b Block) Mine() {

}

// Data is an interface used to standardize methods for any type of Block data
type Data interface {
	GetData()
	ToString()
}

// Transaction is a type of Block data
type Transaction struct {
	From   string
	To     string
	Amount int
}

func (t Transaction) GetData() {
}

func (t Transaction) ToString() {
}
