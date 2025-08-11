package connector

import (
	"errors"
	"time"

	"github.com/rutishh0/testingquant/internal/adapters/coinbase"
	"github.com/rutishh0/testingquant/internal/adapters/mesh"
	"github.com/rutishh0/testingquant/internal/models"
	"github.com/rutishh0/testingquant/internal/overledger"
)

// Request/Response types
type CreateCoinbaseWalletRequest struct {
	Name string `json:"name"`
}

type CreateCoinbaseAddressRequest struct {
	Name string `json:"name"`
}

type CreateCoinbaseTransactionRequest struct {
	WalletID string  `json:"walletId"`
	To       string  `json:"to"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type EstimateFeeRequest struct {
	To       string  `json:"to"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthResponse struct {
	Status    string                   `json:"status"`
	Message   string                   `json:"message,omitempty"`
	Timestamp string                   `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services"`
}

// Service defines the connector service interface
type Service interface {
	// Coinbase operations
	GetCoinbaseWallets() (*models.CoinbaseWalletsResponse, error)
	CreateCoinbaseWallet(req *CreateCoinbaseWalletRequest) (*models.CoinbaseWallet, error)
	GetCoinbaseWalletBalance(walletID string) (*models.CoinbaseBalance, error)
	GetCoinbaseWalletAddresses(walletID string) ([]*models.CoinbaseAddress, error)
	CreateCoinbaseWalletAddress(walletID string, req *CreateCoinbaseAddressRequest) (*models.CoinbaseAddress, error)
	CreateCoinbaseTransaction(req *CreateCoinbaseTransactionRequest) (*models.CoinbaseTransaction, error)
	GetCoinbaseTransaction(transactionID string) (*models.CoinbaseTransaction, error)
	GetCoinbaseTransactions(walletID string, limit int, cursor string) ([]*models.CoinbaseTransaction, error)
	GetCoinbaseAssets() ([]*models.CoinbaseAsset, error)
	GetCoinbaseNetworks() ([]*models.CoinbaseNetwork, error)
	// Mesh operations
	GetMeshNetworks() (*models.MeshNetworksResponse, error)
	GetMeshNetworkBalance(networkIdentifier, accountIdentifier interface{}) (*models.MeshBalanceResponse, error)
	GetCoinbaseExchangeRates(baseCurrency string) (*models.CoinbaseExchangeRates, error)
	EstimateCoinbaseTransactionFee(walletID string, req *EstimateFeeRequest) (*models.CoinbaseFeeEstimate, error)
	
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
	coinbaseAdapter coinbase.Adapter
	meshAdapter     mesh.Adapter
	overledgerClient *overledger.Client
}

// NewService creates a new connector service
func NewService(coinbaseAdapter coinbase.Adapter, meshAdapter mesh.Adapter, overledgerClient *overledger.Client) Service {
	return &service{
		coinbaseAdapter: coinbaseAdapter,
		meshAdapter:     meshAdapter,
		overledgerClient: overledgerClient,
	}
}

// Coinbase operations

func (s *service) GetCoinbaseWallets() (*models.CoinbaseWalletsResponse, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetWallets()
}

func (s *service) CreateCoinbaseWallet(req *CreateCoinbaseWalletRequest) (*models.CoinbaseWallet, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.CreateWallet(req.Name)
}

func (s *service) GetCoinbaseWalletBalance(walletID string) (*models.CoinbaseBalance, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetBalances(walletID)
}

func (s *service) CreateCoinbaseTransaction(req *CreateCoinbaseTransactionRequest) (*models.CoinbaseTransaction, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.CreateTransaction(req.To, req.Currency, req.Amount)
}

func (s *service) GetCoinbaseTransaction(transactionID string) (*models.CoinbaseTransaction, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetTransaction(transactionID)
}

func (s *service) GetCoinbaseWalletAddresses(walletID string) ([]*models.CoinbaseAddress, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetAddresses(walletID)
}

func (s *service) CreateCoinbaseWalletAddress(walletID string, req *CreateCoinbaseAddressRequest) (*models.CoinbaseAddress, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.CreateAddress(walletID, req.Name)
}

func (s *service) GetCoinbaseTransactions(walletID string, limit int, cursor string) ([]*models.CoinbaseTransaction, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetTransactions(walletID, limit, cursor)
}

func (s *service) GetCoinbaseAssets() ([]*models.CoinbaseAsset, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetAssets()
}

func (s *service) GetCoinbaseNetworks() ([]*models.CoinbaseNetwork, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetNetworks()
}

func (s *service) GetCoinbaseExchangeRates(baseCurrency string) (*models.CoinbaseExchangeRates, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.GetExchangeRates(baseCurrency)
}

func (s *service) EstimateCoinbaseTransactionFee(walletID string, req *EstimateFeeRequest) (*models.CoinbaseFeeEstimate, error) {
	if s.coinbaseAdapter == nil {
		return nil, errors.New("coinbase adapter not initialized")
	}
	return s.coinbaseAdapter.EstimateFee(walletID, req.To, req.Currency, req.Amount)
}

// Overledger operations

func (s *service) GetOverledgerNetworks() (*overledger.NetworksResponse, error) {
	if s.overledgerClient == nil {
		return nil, errors.New("overledger client not initialized")
	}
	return s.overledgerClient.GetNetworks()
}

// Mesh operations
func (s *service) GetMeshNetworks() (*models.MeshNetworksResponse, error) {
	if s.meshAdapter == nil {
		return nil, errors.New("mesh adapter not initialized")
	}
	return s.meshAdapter.ListNetworks()
}

// GetMeshNetworkBalance retrieves the balance of an account on a specified network
func (s *service) GetMeshNetworkBalance(networkIdentifier, accountIdentifier interface{}) (*models.MeshBalanceResponse, error) {
	if s.meshAdapter == nil {
		return nil, errors.New("mesh adapter not initialized")
	}
	return s.meshAdapter.AccountBalance(networkIdentifier, accountIdentifier)
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
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]ServiceHealth),
	}

	// Check Coinbase health
	if s.coinbaseAdapter != nil {
		if !s.coinbaseAdapter.Health() {
			health.Services["coinbase"] = ServiceHealth{
				Status:  "unhealthy",
				Message: "Coinbase API is not accessible",
			}
			health.Status = "degraded"
		} else {
			health.Services["coinbase"] = ServiceHealth{
				Status:  "healthy",
				Message: "Coinbase API is accessible",
			}
		}
	} else {
		health.Services["coinbase"] = ServiceHealth{
			Status:  "disabled",
			Message: "Coinbase client not configured (missing credentials)",
		}
	}

	// Check Overledger health
	if s.overledgerClient != nil {
		if err := s.overledgerClient.TestConnection(); err != nil {
			health.Services["overledger"] = ServiceHealth{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			health.Status = "degraded"
		} else {
			health.Services["overledger"] = ServiceHealth{
				Status:  "healthy",
				Message: "Overledger API is accessible",
			}
		}
	} else {
		health.Services["overledger"] = ServiceHealth{
			Status:  "disabled",
			Message: "Overledger client not configured (missing credentials)",
		}
	}

	// If both services are disabled or unhealthy, mark overall status appropriately
	allDisabled := true
	anyHealthy := false
	for _, service := range health.Services {
		if service.Status != "disabled" {
			allDisabled = false
		}
		if service.Status == "healthy" {
			anyHealthy = true
		}
	}

	if allDisabled {
		health.Status = "degraded"
		health.Message = "No services configured - please set API credentials"
	} else if !anyHealthy && health.Status == "degraded" {
		health.Message = "Some services are experiencing issues"
	} else if anyHealthy {
		health.Status = "healthy"
		health.Message = "Service is operational"
	}

	return health, nil
}