package connector

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/rutishh0/testingquant/internal/clients"
	"github.com/rutishh0/testingquant/internal/overledger"
)

// Service defines the connector service interface
type Service interface {
	// Coinbase operations
	GetCoinbaseWallets() (*CoinbaseWalletsResponse, error)
	CreateCoinbaseWallet(req *CreateCoinbaseWalletRequest) (*CoinbaseWalletResponse, error)
	GetCoinbaseWalletBalance(walletID string) (*CoinbaseBalanceResponse, error)
	GetCoinbaseWalletAddresses(walletID string) (*CoinbaseAddressesResponse, error)
	CreateCoinbaseWalletAddress(walletID string, req *CreateCoinbaseAddressRequest) (*CoinbaseAddressResponse, error)
	CreateCoinbaseTransaction(req *CreateCoinbaseTransactionRequest) (*CoinbaseTransactionResponse, error)
	GetCoinbaseTransaction(transactionID string) (*CoinbaseTransactionResponse, error)
	GetCoinbaseTransactions(walletID string, limit int, cursor string) (*CoinbaseTransactionsResponse, error)
	GetCoinbaseAssets() (*CoinbaseAssetsResponse, error)
	GetCoinbaseNetworks() (*CoinbaseNetworksResponse, error)
	GetCoinbaseExchangeRates(baseCurrency string) (*CoinbaseExchangeRatesResponse, error)
	EstimateCoinbaseTransactionFee(walletID string, req *EstimateFeeRequest) (*EstimateFeeResponse, error)
	
	// Overledger operations
	GetOverledgerNetworks() (*overledger.NetworksResponse, error)
	GetOverledgerBalance(networkID, address string) (*overledger.BalanceResponse, error)
	CreateOverledgerTransaction(req *overledger.TransactionRequest) (*overledger.TransactionResponse, error)
	GetOverledgerTransactionStatus(networkID, txHash string) (*overledger.TransactionStatusResponse, error)
	TestOverledgerConnection() error
	
	// Health and status
	HealthCheck() (*HealthResponse, error)
}

// service implements the Service interface
type service struct {
	coinbaseClient   *clients.CoinbaseClient
	overledgerClient *overledger.Client
}

// NewService creates a new connector service
func NewService(coinbaseClient *clients.CoinbaseClient, overledgerClient *overledger.Client) Service {
	return &service{
		coinbaseClient:   coinbaseClient,
		overledgerClient: overledgerClient,
	}
}

// Coinbase operations

func (s *service) GetCoinbaseWallets() (*CoinbaseWalletsResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.Get("/v1/wallets")
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var walletsResp CoinbaseWalletsResponse
	if err := json.Unmarshal(body, &walletsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &walletsResp, nil
}

func (s *service) CreateCoinbaseWallet(req *CreateCoinbaseWalletRequest) (*CoinbaseWalletResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.Post("/v1/wallets", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var walletResp CoinbaseWalletResponse
	if err := json.Unmarshal(body, &walletResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &walletResp, nil
}

func (s *service) GetCoinbaseWalletBalance(walletID string) (*CoinbaseBalanceResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.Get(fmt.Sprintf("/v1/wallets/%s/balances", walletID))
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var balanceResp CoinbaseBalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &balanceResp, nil
}

func (s *service) CreateCoinbaseTransaction(req *CreateCoinbaseTransactionRequest) (*CoinbaseTransactionResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.Post(fmt.Sprintf("/v1/wallets/%s/transactions", req.WalletID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var txResp CoinbaseTransactionResponse
	if err := json.Unmarshal(body, &txResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &txResp, nil
}

func (s *service) GetCoinbaseTransaction(transactionID string) (*CoinbaseTransactionResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.Get(fmt.Sprintf("/v1/transactions/%s", transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var txResp CoinbaseTransactionResponse
	if err := json.Unmarshal(body, &txResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &txResp, nil
}

func (s *service) GetCoinbaseWalletAddresses(walletID string) (*CoinbaseAddressesResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.GetWalletAddresses(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet addresses: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var addressesResp CoinbaseAddressesResponse
	if err := json.Unmarshal(body, &addressesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &addressesResp, nil
}

func (s *service) CreateCoinbaseWalletAddress(walletID string, req *CreateCoinbaseAddressRequest) (*CoinbaseAddressResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.CreateWalletAddress(walletID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet address: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
}

	var addressResp CoinbaseAddressResponse
	if err := json.Unmarshal(body, &addressResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &addressResp, nil
}

func (s *service) GetCoinbaseTransactions(walletID string, limit int, cursor string) (*CoinbaseTransactionsResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.GetTransactions(walletID, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var transactionsResp CoinbaseTransactionsResponse
	if err := json.Unmarshal(body, &transactionsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &transactionsResp, nil
}

func (s *service) GetCoinbaseAssets() (*CoinbaseAssetsResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.GetAssets()
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var assetsResp CoinbaseAssetsResponse
	if err := json.Unmarshal(body, &assetsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &assetsResp, nil
}

func (s *service) GetCoinbaseNetworks() (*CoinbaseNetworksResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.GetNetworks()
	if err != nil {
		return nil, fmt.Errorf("failed to get networks: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var networksResp CoinbaseNetworksResponse
	if err := json.Unmarshal(body, &networksResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &networksResp, nil
}

func (s *service) GetCoinbaseExchangeRates(baseCurrency string) (*CoinbaseExchangeRatesResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.GetExchangeRates(baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
		}
		
	var ratesResp CoinbaseExchangeRatesResponse
	if err := json.Unmarshal(body, &ratesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &ratesResp, nil
}

func (s *service) EstimateCoinbaseTransactionFee(walletID string, req *EstimateFeeRequest) (*EstimateFeeResponse, error) {
	if s.coinbaseClient == nil {
		return nil, errors.New("coinbase client not initialized")
	}

	resp, err := s.coinbaseClient.EstimateTransactionFee(walletID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate transaction fee: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var feeResp EstimateFeeResponse
	if err := json.Unmarshal(body, &feeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &feeResp, nil
}

// Overledger operations

func (s *service) GetOverledgerNetworks() (*overledger.NetworksResponse, error) {
	if s.overledgerClient == nil {
		return nil, errors.New("overledger client not initialized")
	}
	return s.overledgerClient.GetNetworks()
}

func (s *service) GetOverledgerBalance(networkID, address string) (*overledger.BalanceResponse, error) {
	if s.overledgerClient == nil {
		return nil, errors.New("overledger client not initialized")
	}
	return s.overledgerClient.GetAccountBalance(networkID, address)
}

func (s *service) CreateOverledgerTransaction(req *overledger.TransactionRequest) (*overledger.TransactionResponse, error) {
	if s.overledgerClient == nil {
		return nil, errors.New("overledger client not initialized")
	}
	return s.overledgerClient.CreateTransaction(req)
}

func (s *service) GetOverledgerTransactionStatus(networkID, txHash string) (*overledger.TransactionStatusResponse, error) {
	if s.overledgerClient == nil {
		return nil, errors.New("overledger client not initialized")
	}
	return s.overledgerClient.GetTransactionStatus(networkID, txHash)
}

func (s *service) TestOverledgerConnection() error {
	if s.overledgerClient == nil {
		return errors.New("overledger client not initialized")
	}
	return s.overledgerClient.TestConnection()
}

// Health and status

func (s *service) HealthCheck() (*HealthResponse, error) {
	health := &HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Services:  make(map[string]ServiceHealth),
	}

	// Check Coinbase connection
	if s.coinbaseClient != nil {
		err := s.coinbaseClient.Health()
		if err != nil {
			health.Services["coinbase"] = ServiceHealth{
				Status: "unhealthy",
				Error:  err.Error(),
			}
			health.Status = "degraded"
		} else {
			health.Services["coinbase"] = ServiceHealth{
				Status: "healthy",
			}
		}
	} else {
		health.Services["coinbase"] = ServiceHealth{
			Status: "not_configured",
		}
	}

	// Check Overledger connection
	if s.overledgerClient != nil {
		err := s.overledgerClient.TestConnection()
		if err != nil {
			health.Services["overledger"] = ServiceHealth{
				Status: "unhealthy",
				Error:  err.Error(),
			}
			health.Status = "degraded"
		} else {
			health.Services["overledger"] = ServiceHealth{
				Status: "healthy",
			}
		}
	} else {
		health.Services["overledger"] = ServiceHealth{
			Status: "not_configured",
		}
	}

	return health, nil
}