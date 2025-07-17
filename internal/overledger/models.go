package overledger

import "time"

// NetworksResponse represents the response from the networks endpoint
type NetworksResponse struct {
	Networks []Network `json:"networks"`
}

// Network represents a blockchain network in Overledger
type Network struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
}

// BalanceResponse represents the response from the balance endpoint
type BalanceResponse struct {
	Address  string    `json:"address"`
	Balances []Balance `json:"balances"`
}

// Balance represents a token balance
type Balance struct {
	TokenID     string `json:"tokenId"`
	TokenName   string `json:"tokenName"`
	TokenSymbol string `json:"tokenSymbol"`
	Amount      string `json:"amount"`
	Decimals    int    `json:"decimals"`
	Unit        string `json:"unit"`
}

// TransactionRequest represents a transaction creation request
type TransactionRequest struct {
	NetworkID   string                 `json:"networkId"`
	FromAddress string                 `json:"fromAddress"`
	ToAddress   string                 `json:"toAddress"`
	Amount      string                 `json:"amount"`
	TokenID     string                 `json:"tokenId,omitempty"`
	GasLimit    string                 `json:"gasLimit,omitempty"`
	GasPrice    string                 `json:"gasPrice,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionResponse represents the response from transaction creation
type TransactionResponse struct {
	TransactionID string                 `json:"transactionId"`
	Hash          string                 `json:"hash"`
	Status        string                 `json:"status"`
	NetworkID     string                 `json:"networkId"`
	FromAddress   string                 `json:"fromAddress"`
	ToAddress     string                 `json:"toAddress"`
	Amount        string                 `json:"amount"`
	Fee           string                 `json:"fee,omitempty"`
	BlockNumber   int64                  `json:"blockNumber,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionStatusResponse represents the response from transaction status endpoint
type TransactionStatusResponse struct {
	TransactionID string                 `json:"transactionId"`
	Hash          string                 `json:"hash"`
	Status        string                 `json:"status"`
	NetworkID     string                 `json:"networkId"`
	Confirmations int                    `json:"confirmations"`
	BlockNumber   int64                  `json:"blockNumber,omitempty"`
	BlockHash     string                 `json:"blockHash,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse represents an error response from Overledger API
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details,omitempty"`
	} `json:"error"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// NetworkStatus represents the status of a network
type NetworkStatus struct {
	NetworkID     string    `json:"networkId"`
	Status        string    `json:"status"`
	BlockHeight   int64     `json:"blockHeight"`
	LastUpdated   time.Time `json:"lastUpdated"`
	SyncProgress  float64   `json:"syncProgress"`
	PeerCount     int       `json:"peerCount,omitempty"`
}

// AddressInfo represents information about an address
type AddressInfo struct {
	Address     string    `json:"address"`
	NetworkID   string    `json:"networkId"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"createdAt"`
	LastActive  time.Time `json:"lastActive,omitempty"`
	TxCount     int       `json:"transactionCount"`
}

// TokenInfo represents information about a token
type TokenInfo struct {
	TokenID     string `json:"tokenId"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	TotalSupply string `json:"totalSupply,omitempty"`
	ContractAddress string `json:"contractAddress,omitempty"`
	NetworkID   string `json:"networkId"`
	Type        string `json:"type"`
}

// WebhookRequest represents a webhook subscription request
type WebhookRequest struct {
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	NetworkID string   `json:"networkId,omitempty"`
	Address   string   `json:"address,omitempty"`
	Active    bool     `json:"active"`
}

// WebhookResponse represents a webhook subscription response
type WebhookResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	NetworkID string    `json:"networkId,omitempty"`
	Address   string    `json:"address,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}