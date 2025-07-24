package connector

import "time"

// Coinbase API Models

type CoinbaseWalletsResponse struct {
	Data    []CoinbaseWallet `json:"data"`
	HasMore bool             `json:"has_more"`
	Cursor  string           `json:"cursor,omitempty"`
}

type CoinbaseWallet struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	PrimaryAddress   string            `json:"primary_address"`
	DefaultNetwork   CoinbaseNetwork   `json:"default_network"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	Features         []string          `json:"features"`
}

type CoinbaseNetwork struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	ChainID     int    `json:"chain_id"`
	IsTestnet   bool   `json:"is_testnet"`
}

type CreateCoinbaseWalletRequest struct {
	Name           string `json:"name" binding:"required"`
	UseServerSigner bool   `json:"use_server_signer,omitempty"`
}

type CoinbaseWalletResponse struct {
	Data CoinbaseWallet `json:"data"`
}

type CoinbaseBalanceResponse struct {
	Data    []CoinbaseBalance `json:"data"`
	HasMore bool              `json:"has_more"`
	Cursor  string            `json:"cursor,omitempty"`
}

type CoinbaseBalance struct {
	Amount   string         `json:"amount"`
	Asset    CoinbaseAsset  `json:"asset"`
	Network  CoinbaseNetwork `json:"network"`
}

type CoinbaseAsset struct {
	AssetID         string `json:"asset_id"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Decimals        int    `json:"decimals"`
	DisplayName     string `json:"display_name"`
	AddressFormat   string `json:"address_format"`
	ContractAddress string `json:"contract_address,omitempty"`
}

type CreateCoinbaseTransactionRequest struct {
	WalletID          string                     `json:"-"`
	Amount            string                     `json:"amount" binding:"required"`
	AssetID           string                     `json:"asset_id" binding:"required"`
	Destination       string                     `json:"destination" binding:"required"`
	GaslessSend       bool                       `json:"gasless_send,omitempty"`
	Network           string                     `json:"network,omitempty"`
	Speed             string                     `json:"speed,omitempty"`
	FeeRate           string                     `json:"fee_rate,omitempty"`
	DestinationTag    string                     `json:"destination_tag,omitempty"`
}

type CoinbaseTransactionResponse struct {
	Data CoinbaseTransaction `json:"data"`
}

type CoinbaseTransaction struct {
	TransactionID     string                   `json:"transaction_id"`
	Status            string                       `json:"status"`
	UnsignedPayload   string                   `json:"unsigned_payload,omitempty"`
	SignedPayload     string                   `json:"signed_payload,omitempty"`
	TransactionHash   string                   `json:"transaction_hash,omitempty"`
	TransactionLink   string                   `json:"transaction_link,omitempty"`
	FromAddress       string                   `json:"from_address"`
	ToAddress         string                   `json:"to_address"`
	Amount            string                   `json:"amount"`
	NetworkFee        string                   `json:"network_fee"`
	Asset             CoinbaseAsset            `json:"asset"`
	Network           CoinbaseNetwork          `json:"network"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

// Additional Coinbase API Models

type CoinbaseAddressesResponse struct {
	Data    []CoinbaseAddress `json:"data"`
	HasMore bool              `json:"has_more"`
	Cursor  string            `json:"cursor,omitempty"`
}

type CoinbaseAddress struct {
	ID          string          `json:"id"`
	Address     string          `json:"address"`
	Network     CoinbaseNetwork `json:"network"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	PublicKey   string          `json:"public_key,omitempty"`
	AddressInfo CoinbaseAddressInfo `json:"address_info,omitempty"`
}

type CoinbaseAddressInfo struct {
	Balance   string `json:"balance,omitempty"`
	Received  string `json:"received,omitempty"`
	Sent      string `json:"sent,omitempty"`
	TxCount   int    `json:"tx_count,omitempty"`
}

type CreateCoinbaseAddressRequest struct {
	Name       string `json:"name,omitempty"`
	NetworkID  string `json:"network_id" binding:"required"`
}

type CoinbaseAddressResponse struct {
	Data CoinbaseAddress `json:"data"`
}

type CoinbaseTransactionsResponse struct {
	Data    []CoinbaseTransaction `json:"data"`
	HasMore bool                  `json:"has_more"`
	Cursor  string                `json:"cursor,omitempty"`
	Total   int                   `json:"total,omitempty"`
}

type CoinbaseAssetsResponse struct {
	Data    []CoinbaseAssetInfo `json:"data"`
	HasMore bool                `json:"has_more"`
	Cursor  string              `json:"cursor,omitempty"`
}

type CoinbaseAssetInfo struct {
	AssetID         string `json:"asset_id"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Decimals        int    `json:"decimals"`
	DisplayName     string `json:"display_name"`
	AddressFormat   string `json:"address_format"`
	ExplorerURL     string `json:"explorer_url,omitempty"`
	ContractAddress string `json:"contract_address,omitempty"`
	ImageURL        string `json:"image_url,omitempty"`
	SupportedNetworks []CoinbaseNetwork `json:"supported_networks,omitempty"`
}

type CoinbaseNetworksResponse struct {
	Data    []CoinbaseNetworkInfo `json:"data"`
	HasMore bool                  `json:"has_more"`
	Cursor  string                `json:"cursor,omitempty"`
}

type CoinbaseNetworkInfo struct {
	ID                 string   `json:"id"`
	DisplayName        string   `json:"display_name"`
	ChainID            int      `json:"chain_id"`
	IsTestnet          bool     `json:"is_testnet"`
	BlockExplorerURL   string   `json:"block_explorer_url,omitempty"`
	NativeCurrency     CoinbaseAsset `json:"native_currency"`
	SupportedAssets    []string `json:"supported_assets,omitempty"`
	FeaturesToSupport  []string `json:"features_to_support,omitempty"`
}

type CoinbaseExchangeRatesResponse struct {
	Data CoinbaseExchangeRates `json:"data"`
}

type CoinbaseExchangeRates struct {
	Currency string                    `json:"currency"`
	Rates    map[string]string         `json:"rates"`
	UpdatedAt time.Time                `json:"updated_at"`
}

type EstimateFeeRequest struct {
	Amount       string `json:"amount" binding:"required"`
	AssetID      string `json:"asset_id" binding:"required"`
	Destination  string `json:"destination" binding:"required"`
	NetworkID    string `json:"network_id,omitempty"`
	Speed        string `json:"speed,omitempty"` // "slow", "standard", "fast"
}

type EstimateFeeResponse struct {
	Data CoinbaseFeeEstimate `json:"data"`
}

type CoinbaseFeeEstimate struct {
	EstimatedFee  string        `json:"estimated_fee"`
	AssetID       string        `json:"asset_id"`
	NetworkID     string        `json:"network_id"`
	Speed         string        `json:"speed"`
	EstimatedTime string        `json:"estimated_time,omitempty"`
	FeeBreakdown  []CoinbaseFeeBreakdown `json:"fee_breakdown,omitempty"`
}

type CoinbaseFeeBreakdown struct {
	Type   string `json:"type"`
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

// Health Check Models

type HealthResponse struct {
	Status    string                    `json:"status"`
	Timestamp int64                     `json:"timestamp"`
	Services  map[string]ServiceHealth  `json:"services"`
}

type ServiceHealth struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Legacy status response for backward compatibility
type StatusResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	Timestamp int64  `json:"timestamp"`
}

// Error response model
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}