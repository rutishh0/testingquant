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

// TransactionPrepareRequest represents a transaction preparation request for Overledger
type TransactionPrepareRequest struct {
	Location Location `json:"location"`
	Type     string   `json:"type"`
	Urgency  string   `json:"urgency"`
	RequestDetails RequestDetails `json:"requestDetails"`
}

// Location represents the blockchain network location
type Location struct {
	Technology string `json:"technology"`
	Network    string `json:"network"`
}

// RequestDetails represents the transaction details
type RequestDetails struct {
	Destination []DestinationAccount `json:"destination"`
	Message     string               `json:"message"`
	OverledgerSigningType string   `json:"overledgerSigningType"`
	Origin      []OriginAccount      `json:"origin"`
	Overrides   *TransactionOverrides `json:"overrides,omitempty"`
}

// DestinationAccount represents the destination account for payment
type DestinationAccount struct {
	DestinationID string  `json:"destinationId"`
	Payment       Payment `json:"payment"`
}

// Payment represents the payment details
type Payment struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

// OriginAccount represents the origin account
type OriginAccount struct {
	OriginID string `json:"originId"`
}

// OriginatorAccount represents the account sending the transaction
type OriginatorAccount struct {
	AccountID string `json:"accountId"`
	Unit      string `json:"unit"`
	Amount    string `json:"amount"`
}

// RecipientDetails represents the transaction recipient
type RecipientDetails struct {
	ReceivingAccounts []ReceivingAccount `json:"receivingAccounts"`
}

// ReceivingAccount represents an account receiving the transaction
type ReceivingAccount struct {
	AccountID string `json:"accountId"`
	Amount    string `json:"amount"`
	Unit      string `json:"unit"`
}

// TransactionOverrides represents optional transaction parameters
type TransactionOverrides struct {
	GasLimit string `json:"gasLimit,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	MaxFeePerGas string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
}

// TransactionPrepareResponse represents the response from transaction preparation
type TransactionPrepareResponse struct {
	PreparationTransactionSearchResponse PreparationTransactionSearchResponse `json:"preparationTransactionSearchResponse"`
	GatewayFee                          GatewayFee                          `json:"gatewayFee,omitempty"`
	DltData                             DltData                             `json:"dltData"`
}

// PreparationTransactionSearchResponse contains preparation details
type PreparationTransactionSearchResponse struct {
	RequestID  string     `json:"requestId"`
	GatewayFee GatewayFee `json:"gatewayFee"`
}

// GatewayFee represents the gateway fee
type GatewayFee struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

// DltData contains transaction data for signing
type DltData struct {
	Data        TransactionData `json:"data"`
	SigningData SigningData     `json:"signingData"`
	NativeData  *NativeData     `json:"nativeData,omitempty"`
}

// TransactionData contains the transaction details
type TransactionData struct {
	TransactionHash string `json:"transactionHash,omitempty"`
	RawTransaction  string `json:"rawTransaction,omitempty"`
	NativeData      *NativeData `json:"nativeData,omitempty"`
}

// SigningData contains data needed for transaction signing
type SigningData struct {
	Signature string `json:"signature,omitempty"`
}

// TransactionExecuteRequest represents a transaction execution request
type TransactionExecuteRequest struct {
	Signed    string `json:"signed"`
	RequestID string `json:"requestId"`
}

// TransactionExecuteResponse represents the response from transaction execution
type TransactionExecuteResponse struct {
	PreparationTransactionSearchResponse PreparationTransactionSearchResponse `json:"preparationTransactionSearchResponse"`
	ExecutionTransactionSearchResponse   ExecutionTransactionSearchResponse   `json:"executionTransactionSearchResponse"`
}

// ExecutionTransactionSearchResponse contains execution details
type ExecutionTransactionSearchResponse struct {
	TransactionID string          `json:"transactionId"`
	Status        ExecutionStatus `json:"status"`
	Location      Location       `json:"location"`
	Timestamp     string         `json:"timestamp"`
	Message       string         `json:"message,omitempty"`
}

// ExecutionStatus mirrors Overledger nested status object
type ExecutionStatus struct {
	Value       string `json:"value"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// SandboxSigningRequest represents a request to the sandbox signing endpoint
type SandboxSigningRequest struct {
	KeyID     string     `json:"keyId"`
	GatewayFee GatewayFee `json:"gatewayFee"`
	RequestID string     `json:"requestId"`
	DltFee    GatewayFee `json:"dltFee"`
	TransactionSigningResponderName string `json:"transactionSigningResponderName"`
	NativeData NativeData `json:"nativeData"`
}

// NativeData represents the native transaction data for signing
type NativeData struct {
	Chain              string `json:"chain"`
	Data               string `json:"data"`
	ChainID            int    `json:"chainId"`
	Gas                string `json:"gas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	To                 string `json:"to"`
	MaxFeePerGas       string `json:"maxFeePerGas"`
	Nonce              int    `json:"nonce"`
	Hardfork           string `json:"hardfork"`
	Value              string `json:"value"`
}

// SandboxSigningResponse represents the response from sandbox signing
type SandboxSigningResponse struct {
	SignedTransaction string `json:"signed,omitempty"`
	Signature         string `json:"signature,omitempty"`
}

// Legacy TransactionRequest for backward compatibility
type TransactionRequest struct {
	NetworkID   string                 `json:"networkId"`
	FromAddress string                 `json:"fromAddress"`
	ToAddress   string                 `json:"toAddress"`
	Amount      string                 `json:"amount"`
	TokenID     string                 `json:"tokenId,omitempty"`
	GasLimit    string                 `json:"gasLimit,omitempty"`
	GasPrice    string                 `json:"gasPrice,omitempty"`
	MaxFeePerGas string                `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string        `json:"maxPriorityFeePerGas,omitempty"`
	Nonce       *int                   `json:"nonce,omitempty"`
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