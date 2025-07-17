package mesh

// NetworkIdentifier uniquely identifies a network
type NetworkIdentifier struct {
	Blockchain string `json:"blockchain"`
	Network    string `json:"network"`
}

// AccountIdentifier uniquely identifies an account within a network
type AccountIdentifier struct {
	Address    string                 `json:"address"`
	SubAccount *SubAccountIdentifier `json:"sub_account,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SubAccountIdentifier identifies a sub-account
type SubAccountIdentifier struct {
	Address  string                 `json:"address"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Amount represents a monetary amount
type Amount struct {
	Value    string   `json:"value"`
	Currency Currency `json:"currency"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Currency represents a currency
type Currency struct {
	Symbol   string `json:"symbol"`
	Decimals int32  `json:"decimals"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Operation represents a state-changing action
type Operation struct {
	OperationIdentifier OperationIdentifier `json:"operation_identifier"`
	RelatedOperations   []OperationIdentifier `json:"related_operations,omitempty"`
	Type                string              `json:"type"`
	Status              *string             `json:"status,omitempty"`
	Account             *AccountIdentifier  `json:"account,omitempty"`
	Amount              *Amount             `json:"amount,omitempty"`
	CoinChange          *CoinChange         `json:"coin_change,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// OperationIdentifier uniquely identifies an operation
type OperationIdentifier struct {
	Index        int64  `json:"index"`
	NetworkIndex *int64 `json:"network_index,omitempty"`
}

// CoinChange represents a change in coin state
type CoinChange struct {
	CoinIdentifier CoinIdentifier `json:"coin_identifier"`
	CoinAction     string         `json:"coin_action"`
}

// CoinIdentifier uniquely identifies a coin
type CoinIdentifier struct {
	Identifier string `json:"identifier"`
}

// BlockIdentifier uniquely identifies a block
type BlockIdentifier struct {
	Index int64  `json:"index"`
	Hash  string `json:"hash"`
}

// PartialBlockIdentifier identifies a block by either index or hash
type PartialBlockIdentifier struct {
	Index *int64  `json:"index,omitempty"`
	Hash  *string `json:"hash,omitempty"`
}

// TransactionIdentifier uniquely identifies a transaction
type TransactionIdentifier struct {
	Hash string `json:"hash"`
}

// Transaction represents a transaction
type Transaction struct {
	TransactionIdentifier TransactionIdentifier `json:"transaction_identifier"`
	Operations            []Operation           `json:"operations"`
	RelatedTransactions   []RelatedTransaction  `json:"related_transactions,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// RelatedTransaction identifies a related transaction
type RelatedTransaction struct {
	NetworkIdentifier     *NetworkIdentifier      `json:"network_identifier,omitempty"`
	TransactionIdentifier TransactionIdentifier  `json:"transaction_identifier"`
	Direction             string                  `json:"direction"`
}

// Block represents a block
type Block struct {
	BlockIdentifier       BlockIdentifier `json:"block_identifier"`
	ParentBlockIdentifier BlockIdentifier `json:"parent_block_identifier"`
	Timestamp             int64           `json:"timestamp"`
	Transactions          []Transaction   `json:"transactions"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// Signature represents a cryptographic signature
type Signature struct {
	SigningPayload SigningPayload `json:"signing_payload"`
	PublicKey      PublicKey      `json:"public_key"`
	SignatureType  string         `json:"signature_type"`
	HexBytes       string         `json:"hex_bytes"`
}

// SigningPayload represents data to be signed
type SigningPayload struct {
	AccountIdentifier *AccountIdentifier `json:"account_identifier,omitempty"`
	HexBytes          string             `json:"hex_bytes"`
	SignatureType     *string            `json:"signature_type,omitempty"`
}

// PublicKey represents a public key
type PublicKey struct {
	HexBytes  string `json:"hex_bytes"`
	CurveType string `json:"curve_type"`
}

// Error represents an error response
type Error struct {
	Code        int32                  `json:"code"`
	Message     string                 `json:"message"`
	Description *string                `json:"description,omitempty"`
	Retriable   bool                   `json:"retriable"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Request/Response types for Construction API

// ConstructionPreprocessRequest represents a preprocess request
type ConstructionPreprocessRequest struct {
	NetworkIdentifier      NetworkIdentifier      `json:"network_identifier"`
	Operations             []Operation            `json:"operations"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
	MaxFee                 []Amount               `json:"max_fee,omitempty"`
	SuggestedFeeMultiplier *float64               `json:"suggested_fee_multiplier,omitempty"`
}

// ConstructionPreprocessResponse represents a preprocess response
type ConstructionPreprocessResponse struct {
	Options              map[string]interface{} `json:"options,omitempty"`
	RequiredPublicKeys   []AccountIdentifier    `json:"required_public_keys,omitempty"`
}

// ConstructionPayloadsRequest represents a payloads request
type ConstructionPayloadsRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	Operations        []Operation            `json:"operations"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	PublicKeys        []PublicKey            `json:"public_keys,omitempty"`
}

// ConstructionPayloadsResponse represents a payloads response
type ConstructionPayloadsResponse struct {
	UnsignedTransaction string           `json:"unsigned_transaction"`
	Payloads            []SigningPayload `json:"payloads"`
}

// ConstructionCombineRequest represents a combine request
type ConstructionCombineRequest struct {
	NetworkIdentifier   NetworkIdentifier `json:"network_identifier"`
	UnsignedTransaction string            `json:"unsigned_transaction"`
	Signatures          []Signature       `json:"signatures"`
}

// ConstructionCombineResponse represents a combine response
type ConstructionCombineResponse struct {
	SignedTransaction string `json:"signed_transaction"`
}

// ConstructionSubmitRequest represents a submit request
type ConstructionSubmitRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	SignedTransaction string            `json:"signed_transaction"`
}

// ConstructionSubmitResponse represents a submit response
type ConstructionSubmitResponse struct {
	TransactionIdentifier TransactionIdentifier `json:"transaction_identifier"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// AccountBalanceRequest represents a balance request
type AccountBalanceRequest struct {
	NetworkIdentifier NetworkIdentifier       `json:"network_identifier"`
	AccountIdentifier AccountIdentifier       `json:"account_identifier"`
	BlockIdentifier   *PartialBlockIdentifier `json:"block_identifier,omitempty"`
	Currencies        []Currency              `json:"currencies,omitempty"`
}

// AccountBalanceResponse represents a balance response
type AccountBalanceResponse struct {
	BlockIdentifier BlockIdentifier `json:"block_identifier"`
	Balances        []Amount        `json:"balances"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// BlockRequest represents a block request
type BlockRequest struct {
	NetworkIdentifier NetworkIdentifier       `json:"network_identifier"`
	BlockIdentifier   PartialBlockIdentifier `json:"block_identifier"`
}

// BlockResponse represents a block response
type BlockResponse struct {
	Block             *Block                 `json:"block,omitempty"`
	OtherTransactions []TransactionIdentifier `json:"other_transactions,omitempty"`
}

// NetworkStatusRequest represents a network status request
type NetworkStatusRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// NetworkStatusResponse represents a network status response
type NetworkStatusResponse struct {
	CurrentBlockIdentifier BlockIdentifier        `json:"current_block_identifier"`
	CurrentBlockTimestamp  int64                  `json:"current_block_timestamp"`
	GenesisBlockIdentifier BlockIdentifier        `json:"genesis_block_identifier"`
	OldestBlockIdentifier  *BlockIdentifier       `json:"oldest_block_identifier,omitempty"`
	SyncStatus             *SyncStatus            `json:"sync_status,omitempty"`
	Peers                  []Peer                 `json:"peers,omitempty"`
}

// SyncStatus represents synchronization status
type SyncStatus struct {
	CurrentIndex *int64 `json:"current_index,omitempty"`
	TargetIndex  *int64 `json:"target_index,omitempty"`
	Stage        *string `json:"stage,omitempty"`
	Synced       *bool   `json:"synced,omitempty"`
}

// Peer represents a network peer
type Peer struct {
	PeerID   string                 `json:"peer_id"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}