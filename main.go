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
	addresses    map[string]bool
	transactions map[string][]*Transaction
}

// GetCurrentBlock returns the last parsed block
func (p *EthereumParser) GetCurrentBlock() int {
	return p.currentBlock
}

// Subscribe adds an address to the observer
func (p *EthereumParser) Subscribe(address string) bool {
	p.addresses[address] = true
	return true
}

// GetTransactions returns a list of inbound or outbound transactions for an address
func (p *EthereumParser) GetTransactions(address string) []*Transaction {
	transactions, exists := p.transactions[address]
	if exists {
		return transactions
	}
	return nil
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

	if response.StatusCode != http.StatusOK {
		fmt.Println("HTTP request failed with status:", response.Status)
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
		p.updateTransactions(&transaction)
	}

	p.currentBlock = blockNumber
}

// updateTransactions updates the internal data structure with a new transaction
func (p *EthereumParser) updateTransactions(transaction *Transaction) {
	if ok := p.addresses[transaction.From]; ok {
		p.transactions[transaction.From] = append(p.transactions[transaction.From], transaction)
	}

	if ok := p.addresses[transaction.To]; ok {
		p.transactions[transaction.To] = append(p.transactions[transaction.To], transaction)
	}
}

func main() {
	// Create an instance of EthereumParser
	parser := EthereumParser{
		currentBlock: 0,
		addresses:    make(map[string]bool),
		transactions: make(map[string][]*Transaction),
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
