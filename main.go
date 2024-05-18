package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ethereumAPIURL = "https://cloudflare-eth.com"
)

// Transaction represents a transaction on the Ethereum blockchain
type Transaction struct {
	Hash     string `json:"hash"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	GasPrice string `json:"gasPrice"`
}

// EthereumParser represents the Ethereum blockchain parser
type EthereumParser struct {
	currentBlock int
	addresses    []string
	transactions map[string][]Transaction
}

// GetCurrentBlock returns the last parsed block
func (p *EthereumParser) GetCurrentBlock() int {
	return p.currentBlock
}

// Subscribe adds an address to the observer
func (p *EthereumParser) Subscribe(address string) bool {
	if !p.isAddressSubscribed(address) {
		p.addresses = append(p.addresses, address)
		return true
	}
	return false
}

// GetTransactions returns a list of inbound or outbound transactions for an address
func (p *EthereumParser) GetTransactions(address string) []Transaction {
	transactions, exists := p.transactions[address]
	if exists {
		return transactions
	}
	return []Transaction{}
}

// isAddressSubscribed checks if an address is already subscribed
func (p *EthereumParser) isAddressSubscribed(address string) bool {
	for _, addr := range p.addresses {
		if addr == address {
			return true
		}
	}
	return false
}

// parseBlock parses the transactions in a block and updates the internal data structure
func (p *EthereumParser) parseBlock(blockNumber int) {
	requestBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":1}`, blockNumber)

	response, err := http.Post(ethereumAPIURL, "application/json", strings.NewReader(requestBody))
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response:", err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error parsing JSON response:", err)
		return
	}

	if result["result"] == nil {
		return
	}

	blockData := result["result"].(map[string]interface{})
	blockTransactions := blockData["transactions"].([]interface{})

	for _, tx := range blockTransactions {
		transactionData := tx.(map[string]interface{})
		transaction := Transaction{
			Hash:     transactionData["hash"].(string),
			From:     transactionData["from"].(string),
			To:       transactionData["to"].(string),
			Value:    transactionData["value"].(string),
			GasPrice: transactionData["gasPrice"].(string),
		}
		p.updateTransactions(transaction)
	}

	p.currentBlock = blockNumber
}

// updateTransactions updates the internal data structure with a new transaction
func (p *EthereumParser) updateTransactions(transaction Transaction) {
	for _, addr := range p.addresses {
		if transaction.From == addr || transaction.To == addr {
			p.transactions[addr] = append(p.transactions[addr], transaction)
		}
	}
}

func main() {
	// Create an instance of EthereumParser
	parser := EthereumParser{
		currentBlock: 0,
		addresses:    []string{},
		transactions: make(map[string][]Transaction),
	}

	// Subscribe to an address
	parser.Subscribe("0x1234567890abcdef")
	parser.Subscribe("0x9876543210fedcba")

	// Parse new blocks
	parser.parseBlock(1207)
	parser.parseBlock(1208)

	// Get transactions for an address
	transactions := parser.GetTransactions("0x1234567890abcdef")
	fmt.Println("Transactions for address 0x1234567890abcdef:")
	for _, tx := range transactions {
		fmt.Println("Hash:", tx.Hash)
		fmt.Println("From:", tx.From)
		fmt.Println("To:", tx.To)
		fmt.Println("Value:", tx.Value)
		fmt.Println("Gas Price:", tx.GasPrice)
		fmt.Println()
	}
}
