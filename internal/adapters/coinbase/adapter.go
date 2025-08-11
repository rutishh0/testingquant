package coinbase

import (
	"encoding/json"
	"github.com/rutishh0/testingquant/internal/clients"
	"github.com/rutishh0/testingquant/internal/models"
)

// Adapter defines the interface for coinbase client operations
type Adapter interface {
	GetNetworks() ([]*models.CoinbaseNetwork, error)
	GetBalances(walletID string) (*models.CoinbaseBalance, error)
	GetWallets() (*models.CoinbaseWalletsResponse, error)
	CreateWallet(name string) (*models.CoinbaseWallet, error)
	GetAddresses(walletID string) ([]*models.CoinbaseAddress, error)
	CreateAddress(walletID, name string) (*models.CoinbaseAddress, error)
	GetTransactions(walletID string, limit int, cursor string) ([]*models.CoinbaseTransaction, error)
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

func (a *coinbaseAdapter) GetNetworks() ([]*models.CoinbaseNetwork, error) {
	resp, err := a.client.GetNetworks()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var networks []*models.CoinbaseNetwork
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, err
	}
	return networks, nil
}

func (a *coinbaseAdapter) GetBalances(walletID string) (*models.CoinbaseBalance, error) {
	resp, err := a.client.GetBalances(walletID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var balance models.CoinbaseBalance
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

func (a *coinbaseAdapter) GetWallets() (*models.CoinbaseWalletsResponse, error) {
	resp, err := a.client.GetWallets()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wallets models.CoinbaseWalletsResponse
	if err := json.NewDecoder(resp.Body).Decode(&wallets); err != nil {
		return nil, err
	}
	return &wallets, nil
}

func (a *coinbaseAdapter) CreateWallet(name string) (*models.CoinbaseWallet, error) {
	panic("not implemented")
}
func (a *coinbaseAdapter) GetAddresses(walletID string) ([]*models.CoinbaseAddress, error) {
	panic("not implemented")
}
func (a *coinbaseAdapter) CreateAddress(walletID, name string) (*models.CoinbaseAddress, error) {
	resp, err := a.client.CreateAddress(walletID, name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var address models.CoinbaseAddress
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		return nil, err
	}
	return &address, nil
}
func (a *coinbaseAdapter) GetTransactions(walletID string, limit int, cursor string) ([]*models.CoinbaseTransaction, error) {
	panic("not implemented")
}
func (a *coinbaseAdapter) GetTransaction(transactionID string) (*models.CoinbaseTransaction, error) {
	panic("not implemented")
}
func (a *coinbaseAdapter) CreateTransaction(to, currency string, amount float64) (*models.CoinbaseTransaction, error) {
	panic("not implemented")
}
func (a *coinbaseAdapter) GetAssets() ([]*models.CoinbaseAsset, error) {
	panic("not implemented")
}
func (a *coinbaseAdapter) GetExchangeRates(baseCurrency string) (*models.CoinbaseExchangeRates, error) {
	panic("not implemented")
}
func (a *coinbaseAdapter) EstimateFee(walletID, to, currency string, amount float64) (*models.CoinbaseFeeEstimate, error) {
	panic("not implemented")
}

func (a *coinbaseAdapter) Health() bool {
	return a.client.Health() == nil
}