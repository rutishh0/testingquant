package models

// CoinbaseWallet represents a single wallet
type CoinbaseWallet struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// CoinbaseWalletsResponse represents the response for a list of wallets
type CoinbaseWalletsResponse struct {
	Wallets []CoinbaseWallet `json:"wallets"`
}

// CoinbaseNetwork represents a single network
type CoinbaseNetwork struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CoinbaseNetworksResponse represents the response for a list of networks
type CoinbaseNetworksResponse struct {
	Networks []CoinbaseNetwork `json:"networks"`
}

// CoinbaseBalance represents the balance of a single asset in a wallet
type CoinbaseBalance struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// CoinbaseBalanceResponse represents the response for a wallet's balances
type CoinbaseBalanceResponse struct {
	Balances []CoinbaseBalance `json:"balances"`
}

// CoinbaseAsset represents a single asset
type CoinbaseAsset struct {
	AssetID         string `json:"asset_id"`
	ID              string `json:"id"` // Backward compatibility
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Decimals        int    `json:"decimals"`
	DisplayName     string `json:"display_name"`
	AddressFormat   string `json:"address_format"`
	ExplorerURL     string `json:"explorer_url,omitempty"`
	ContractAddress string `json:"contract_address,omitempty"`
	ImageURL        string `json:"image_url,omitempty"`
}

// CoinbaseAssetsResponse represents the response for a list of assets
type CoinbaseAssetsResponse struct {
	Assets []CoinbaseAsset `json:"assets"`
}

// CoinbaseTransaction represents a single transaction
type CoinbaseTransaction struct {
	ID     string `json:"id"`
	Amount struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Status string `json:"status"`
}

// CoinbaseTransactionsResponse represents the response for a list of transactions
type CoinbaseTransactionsResponse struct {
	Transactions []CoinbaseTransaction `json:"transactions"`
	NextCursor   string                `json:"next_cursor,omitempty"`
	HasNext      bool                  `json:"has_next"`
}

// CoinbaseAddress represents a single address
type CoinbaseAddress struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

// CoinbaseAddressesResponse represents the response for a list of addresses
type CoinbaseAddressesResponse struct {
	Addresses []CoinbaseAddress `json:"addresses"`
}

// CoinbaseAssetInfo represents information about a single asset
type CoinbaseAssetInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CoinbaseNetworkInfo represents information about a single network
type CoinbaseNetworkInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CoinbaseExchangeRates represents the exchange rates for a base currency
type CoinbaseExchangeRates struct {
	Base  string            `json:"base"`
	Rates map[string]string `json:"rates"`
}

// CoinbaseFeeEstimate represents the estimated fee for a transaction
type CoinbaseFeeEstimate struct {
	Fee string `json:"fee"`
}

// CoinbaseError represents Coinbase API error details
type CoinbaseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// CoinbaseTransactionsPaginatedResponse includes pagination info
type CoinbaseTransactionsPaginatedResponse struct {
	Transactions []*CoinbaseTransaction `json:"transactions"`
	NextCursor   string                 `json:"next_cursor,omitempty"`
	HasNext      bool                   `json:"has_next"`
	TotalCount   int                    `json:"total_count,omitempty"`
}
