package coinbase

import (
	"encoding/json"
	"fmt"
	"github.com/rutishh0/testingquant/internal/clients"
	"github.com/rutishh0/testingquant/internal/models"
	"strings"
)

// CoinbaseAPIError wraps Coinbase API errors with additional context
type CoinbaseAPIError struct {
	Code       string
	Message    string
	Details    string
	StatusCode int
}

func (e *CoinbaseAPIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("Coinbase API error [%s]: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("Coinbase API error [%s]: %s", e.Code, e.Message)
}

// Adapter defines the interface for coinbase client operations
type Adapter interface {
	GetNetworks() ([]*models.CoinbaseNetwork, error)
	GetBalances(walletID string) (*models.CoinbaseBalance, error)
	GetWallets() (*models.CoinbaseWalletsResponse, error)
	CreateWallet(name string) (*models.CoinbaseWallet, error)
	GetAddresses(walletID string) ([]*models.CoinbaseAddress, error)
	CreateAddress(walletID, name string) (*models.CoinbaseAddress, error)
	GetTransactions(walletID string, limit int, cursor string) ([]*models.CoinbaseTransaction, error)
	GetTransactionsPaginated(walletID string, limit int, cursor string) (*models.CoinbaseTransactionsPaginatedResponse, error)
	GetTransaction(transactionID string) (*models.CoinbaseTransaction, error)
	CreateTransaction(to, currency string, amount float64) (*models.CoinbaseTransaction, error)
	GetAssets() ([]*models.CoinbaseAsset, error)
	GetExchangeRates(baseCurrency string) (*models.CoinbaseExchangeRates, error)
	EstimateFee(walletID, to, currency string, amount float64) (*models.CoinbaseFeeEstimate, error)
	Health() bool
}

type coinbaseAdapter struct {
    client *clients.CoinbaseClient
}

// NewAdapter creates a new coinbase adapter
func NewAdapter(client *clients.CoinbaseClient) Adapter {
    return &coinbaseAdapter{client: client}
}

// parseAPIError extracts Coinbase error details from API response errors
func (a *coinbaseAdapter) parseAPIError(err error) error {
	if err == nil {
		return nil
	}
	
	errStr := err.Error()
	
	// Check for Coinbase API error format
	if strings.Contains(errStr, "Coinbase API error") {
		// Already a properly formatted error from client
		return err
	}
	
	// Check for authentication errors
	if strings.Contains(errStr, "authentication failed") || strings.Contains(errStr, "401") {
		return &CoinbaseAPIError{
			Code:    "AUTHENTICATION_ERROR",
			Message: "Authentication failed - check API credentials",
			Details: errStr,
		}
	}
	
	// Check for permission errors
	if strings.Contains(errStr, "403") {
		return &CoinbaseAPIError{
			Code:    "PERMISSION_DENIED",
			Message: "Insufficient permissions for this operation",
			Details: errStr,
		}
	}
	
	// Check for not found errors
	if strings.Contains(errStr, "404") {
		return &CoinbaseAPIError{
			Code:    "NOT_FOUND",
			Message: "Requested resource not found",
			Details: errStr,
		}
	}
	
	// Default case - wrap the original error
	return &CoinbaseAPIError{
		Code:    "API_ERROR",
		Message: "An error occurred while calling Coinbase API",
		Details: errStr,
	}
}

func (a *coinbaseAdapter) GetNetworks() ([]*models.CoinbaseNetwork, error) {
	resp, err := a.client.GetNetworks()
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var networks []*models.CoinbaseNetwork
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode networks response: %v", err))
	}
	return networks, nil
}

func (a *coinbaseAdapter) GetBalances(walletID string) (*models.CoinbaseBalance, error) {
	resp, err := a.client.GetBalances(walletID)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var balance models.CoinbaseBalance
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode balance response: %v", err))
	}
	return &balance, nil
}

func (a *coinbaseAdapter) GetWallets() (*models.CoinbaseWalletsResponse, error) {
	resp, err := a.client.GetWallets()
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var wallets models.CoinbaseWalletsResponse
	if err := json.NewDecoder(resp.Body).Decode(&wallets); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode wallets response: %v", err))
	}
	return &wallets, nil
}

func (a *coinbaseAdapter) CreateWallet(name string) (*models.CoinbaseWallet, error) {
	resp, err := a.client.CreateWallet(name)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var wallet models.CoinbaseWallet
	if err := json.NewDecoder(resp.Body).Decode(&wallet); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode wallet response: %v", err))
	}
	return &wallet, nil
}
func (a *coinbaseAdapter) GetAddresses(walletID string) ([]*models.CoinbaseAddress, error) {
	resp, err := a.client.GetWalletAddresses(walletID)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var addrResp models.CoinbaseAddressesResponse
	if err := json.NewDecoder(resp.Body).Decode(&addrResp); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode addresses response: %v", err))
	}
	addresses := make([]*models.CoinbaseAddress, 0, len(addrResp.Addresses))
	for i := range addrResp.Addresses {
		addr := addrResp.Addresses[i]
		addresses = append(addresses, &addr)
	}
	return addresses, nil
}
func (a *coinbaseAdapter) CreateAddress(walletID, name string) (*models.CoinbaseAddress, error) {
	resp, err := a.client.CreateAddress(walletID, name)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var address models.CoinbaseAddress
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode address response: %v", err))
	}
	return &address, nil
}

// GetTransactions returns just the transactions (backward compatibility)
func (a *coinbaseAdapter) GetTransactions(walletID string, limit int, cursor string) ([]*models.CoinbaseTransaction, error) {
	paginatedResp, err := a.GetTransactionsPaginated(walletID, limit, cursor)
	if err != nil {
		return nil, err
	}
	return paginatedResp.Transactions, nil
}

// GetTransactionsPaginated returns transactions with pagination info
func (a *coinbaseAdapter) GetTransactionsPaginated(walletID string, limit int, cursor string) (*models.CoinbaseTransactionsPaginatedResponse, error) {
	resp, err := a.client.GetTransactions(walletID, limit, cursor)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var txResp models.CoinbaseTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&txResp); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode transactions response: %v", err))
	}
	
	transactions := make([]*models.CoinbaseTransaction, 0, len(txResp.Transactions))
	for i := range txResp.Transactions {
		tx := txResp.Transactions[i]
		transactions = append(transactions, &tx)
	}
	
	return &models.CoinbaseTransactionsPaginatedResponse{
		Transactions: transactions,
		NextCursor:   txResp.NextCursor,
		HasNext:      txResp.HasNext,
		TotalCount:   len(transactions),
	}, nil
}

func (a *coinbaseAdapter) GetTransaction(transactionID string) (*models.CoinbaseTransaction, error) {
	resp, err := a.client.GetTransaction(transactionID)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var tx models.CoinbaseTransaction
	if err := json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode transaction response: %v", err))
	}
	return &tx, nil
}
func (a *coinbaseAdapter) CreateTransaction(to, currency string, amount float64) (*models.CoinbaseTransaction, error) {
	resp, err := a.client.CreateTransaction(to, currency, amount)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var tx models.CoinbaseTransaction
	if err := json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode transaction response: %v", err))
	}
	return &tx, nil
}
func (a *coinbaseAdapter) GetAssets() ([]*models.CoinbaseAsset, error) {
	resp, err := a.client.GetAssets()
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var assetsResp models.CoinbaseAssetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&assetsResp); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode assets response: %v", err))
	}
	assets := make([]*models.CoinbaseAsset, 0, len(assetsResp.Assets))
	for i := range assetsResp.Assets {
		asset := assetsResp.Assets[i]
		assets = append(assets, &asset)
	}
	return assets, nil
}
func (a *coinbaseAdapter) GetExchangeRates(baseCurrency string) (*models.CoinbaseExchangeRates, error) {
	resp, err := a.client.GetExchangeRates(baseCurrency)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var rates models.CoinbaseExchangeRates
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode exchange rates response: %v", err))
	}
	return &rates, nil
}
func (a *coinbaseAdapter) EstimateFee(walletID, to, currency string, amount float64) (*models.CoinbaseFeeEstimate, error) {
	resp, err := a.client.EstimateFee(walletID, to, currency, amount)
	if err != nil {
		return nil, a.parseAPIError(err)
	}
	defer resp.Body.Close()

	var fee models.CoinbaseFeeEstimate
	if err := json.NewDecoder(resp.Body).Decode(&fee); err != nil {
		return nil, a.parseAPIError(fmt.Errorf("failed to decode fee estimate response: %v", err))
	}
	return &fee, nil
}

func (a *coinbaseAdapter) Health() bool {
	return a.client.Health() == nil
}