package connector

// Overledger-compatible request/response models

// PreprocessRequest represents an Overledger preprocess request
type PreprocessRequest struct {
	DLT       string                 `json:"dlt" binding:"required"`
	Network   string                 `json:"network" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	Transfers []Transfer             `json:"transfers,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PreprocessResponse represents an Overledger preprocess response
type PreprocessResponse struct {
	Options             map[string]interface{} `json:"options,omitempty"`
	RequiredSigners     []string               `json:"requiredSigners,omitempty"`
	TransactionFee      string                 `json:"transactionFee"`
	GatewayFee          string                 `json:"gatewayFee"`
	PreparedTransaction map[string]interface{} `json:"preparedTransaction"`
}

// PayloadsRequest represents an Overledger payloads request
type PayloadsRequest struct {
	DLT        string                 `json:"dlt" binding:"required"`
	Network    string                 `json:"network" binding:"required"`
	Type       string                 `json:"type" binding:"required"`
	Transfers  []Transfer             `json:"transfers,omitempty"`
	PublicKeys []PublicKey            `json:"publicKeys,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PayloadsResponse represents an Overledger payloads response
type PayloadsResponse struct {
	UnsignedTransaction string           `json:"unsignedTransaction"`
	Payloads            []SigningPayload `json:"payloads"`
}

// CombineRequest represents an Overledger combine request
type CombineRequest struct {
	DLT                 string                 `json:"dlt" binding:"required"`
	Network             string                 `json:"network" binding:"required"`
	UnsignedTransaction string                 `json:"unsignedTransaction" binding:"required"`
	Signatures          []Signature            `json:"signatures" binding:"required"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// CombineResponse represents an Overledger combine response
type CombineResponse struct {
	SignedTransaction string `json:"signedTransaction"`
}

// SubmitRequest represents an Overledger submit request
type SubmitRequest struct {
	DLT               string                 `json:"dlt" binding:"required"`
	Network           string                 `json:"network" binding:"required"`
	SignedTransaction string                 `json:"signedTransaction" binding:"required"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// SubmitResponse represents an Overledger submit response
type SubmitResponse struct {
	TransactionID string                 `json:"transactionId"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BalanceRequest represents an Overledger balance request
type BalanceRequest struct {
	DLT      string                 `json:"dlt" binding:"required"`
	Network  string                 `json:"network" binding:"required"`
	Address  string                 `json:"address" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BalanceResponse represents an Overledger balance response
type BalanceResponse struct {
	Address  string                 `json:"address"`
	Balances []Balance              `json:"balances"`
	Block    BlockInfo              `json:"block"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BlockRequest represents an Overledger block request
type BlockRequest struct {
	DLT         string                 `json:"dlt" binding:"required"`
	Network     string                 `json:"network" binding:"required"`
	BlockNumber *uint64                `json:"blockNumber,omitempty"`
	BlockHash   string                 `json:"blockHash,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BlockResponse represents an Overledger block response
type BlockResponse struct {
	BlockID      string                 `json:"blockId"`
	Number       int64                  `json:"number"`
	Transactions []TransactionInfo      `json:"transactions"`
	Timestamp    int64                  `json:"timestamp"`
	ParentHash   string                 `json:"parentHash"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionRequest represents an Overledger transaction request
type TransactionRequest struct {
	DLT           string                 `json:"dlt" binding:"required"`
	Network       string                 `json:"network" binding:"required"`
	TransactionID string                 `json:"transactionId" binding:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionResponse represents an Overledger transaction response
type TransactionResponse struct {
	TxID      string                 `json:"txId"`
	Status    string                 `json:"status"`
	Block     BlockInfo              `json:"block"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Transfer represents a token transfer
type Transfer struct {
	From        string                 `json:"from" binding:"required"`
	To          string                 `json:"to" binding:"required"`
	Amount      string                 `json:"amount" binding:"required"`
	TokenSymbol string                 `json:"tokenSymbol" binding:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PublicKey represents a public key
type PublicKey struct {
	HexBytes  string `json:"hexBytes" binding:"required"`
	CurveType string `json:"curveType" binding:"required"`
}

// SigningPayload represents data to be signed
type SigningPayload struct {
	Address       string `json:"address,omitempty"`
	HexBytes      string `json:"hexBytes" binding:"required"`
	SignatureType string `json:"signatureType,omitempty"`
}

// Signature represents a cryptographic signature
type Signature struct {
	PublicKey      PublicKey `json:"publicKey" binding:"required"`
	SignatureType  string    `json:"signatureType" binding:"required"`
	SignatureBytes string    `json:"signatureBytes" binding:"required"`
	HexBytes       string    `json:"hexBytes" binding:"required"`
}

// Balance represents an account balance
type Balance struct {
	Amount      string                 `json:"amount"`
	TokenSymbol string                 `json:"tokenSymbol"`
	Decimals    int32                  `json:"decimals"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BlockInfo represents basic block information
type BlockInfo struct {
	Number int64  `json:"number"`
	Hash   string `json:"hash"`
}

// TransactionInfo represents basic transaction information
type TransactionInfo struct {
	TxID     string                 `json:"txId"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version"`
}

// StatusResponse represents a status response
type StatusResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	Timestamp int64  `json:"timestamp"`
}