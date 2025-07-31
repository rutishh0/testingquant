package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rutishh0/testingquant/internal/config"
	"github.com/rutishh0/testingquant/internal/connector"
	"github.com/rutishh0/testingquant/internal/overledger"
	"github.com/rutishh0/testingquant/internal/tests"

	"github.com/gin-gonic/gin"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	connectorService connector.Service
	cfg              *config.Config
}

// NewHandlers creates a new handlers instance
func NewHandlers(connectorService connector.Service, cfg *config.Config) *Handlers {
	return &Handlers{
		connectorService: connectorService,
		cfg:              cfg,
	}
}

// Health handles health check requests
func (h *Handlers) Health(c *gin.Context) {
	health, err := h.connectorService.HealthCheck()
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "health_check_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	
	statusCode := http.StatusOK
	if health.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, health)
}

// Status handles status requests (legacy endpoint)
func (h *Handlers) Status(c *gin.Context) {
	response := connector.StatusResponse{
		Service:   "quant-connector",
		Status:    "running",
		Uptime:    "N/A", // In a real implementation, calculate actual uptime
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// Coinbase Handlers

// GetCoinbaseWallets handles GET /v1/coinbase/wallets
func (h *Handlers) GetCoinbaseWallets(c *gin.Context) {
	wallets, err := h.connectorService.GetCoinbaseWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_wallets_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, wallets)
}

// CreateCoinbaseWallet handles POST /v1/coinbase/wallets
func (h *Handlers) CreateCoinbaseWallet(c *gin.Context) {
	var req connector.CreateCoinbaseWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	wallet, err := h.connectorService.CreateCoinbaseWallet(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_create_wallet_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusCreated, wallet)
}

// GetCoinbaseWalletBalance handles GET /v1/coinbase/wallets/:walletId/balance
func (h *Handlers) GetCoinbaseWalletBalance(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_wallet_id",
			Message: "Wallet ID is required",
			Code:    400,
		})
		return
	}

	balance, err := h.connectorService.GetCoinbaseWalletBalance(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_balance_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, balance)
}

// CreateCoinbaseTransaction handles POST /v1/coinbase/wallets/:walletId/transactions
func (h *Handlers) CreateCoinbaseTransaction(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_wallet_id",
			Message: "Wallet ID is required",
			Code:    400,
		})
		return
	}

	var req connector.CreateCoinbaseTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	req.WalletID = walletID
	transaction, err := h.connectorService.CreateCoinbaseTransaction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_transaction_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusCreated, transaction)
}

// GetCoinbaseTransaction handles GET /v1/coinbase/transactions/:transactionId
func (h *Handlers) GetCoinbaseTransaction(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_transaction_id",
			Message: "Transaction ID is required",
			Code:    400,
		})
		return
	}

	transaction, err := h.connectorService.GetCoinbaseTransaction(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_get_transaction_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, transaction)
}

// GetCoinbaseWalletAddresses handles GET /v1/coinbase/wallets/:walletId/addresses
func (h *Handlers) GetCoinbaseWalletAddresses(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_wallet_id",
			Message: "Wallet ID is required",
			Code:    400,
		})
		return
	}

	addresses, err := h.connectorService.GetCoinbaseWalletAddresses(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_addresses_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, addresses)
}

// CreateCoinbaseWalletAddress handles POST /v1/coinbase/wallets/:walletId/addresses
func (h *Handlers) CreateCoinbaseWalletAddress(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_wallet_id",
			Message: "Wallet ID is required",
			Code:    400,
		})
		return
	}

	var req connector.CreateCoinbaseAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	address, err := h.connectorService.CreateCoinbaseWalletAddress(walletID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_create_address_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusCreated, address)
}

// GetCoinbaseTransactions handles GET /v1/coinbase/wallets/:walletId/transactions
func (h *Handlers) GetCoinbaseTransactions(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_wallet_id",
			Message: "Wallet ID is required",
			Code:    400,
		})
		return
	}

	// Parse query parameters
	limitStr := c.Query("limit")
	cursor := c.Query("cursor")
	
	limit := 25 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	transactions, err := h.connectorService.GetCoinbaseTransactions(walletID, limit, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_transactions_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

// GetCoinbaseAssets handles GET /v1/coinbase/assets
func (h *Handlers) GetCoinbaseAssets(c *gin.Context) {
	assets, err := h.connectorService.GetCoinbaseAssets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_assets_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// GetCoinbaseNetworks handles GET /v1/coinbase/networks
func (h *Handlers) GetCoinbaseNetworks(c *gin.Context) {
	networks, err := h.connectorService.GetCoinbaseNetworks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_networks_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, networks)
}

// GetCoinbaseExchangeRates handles GET /v1/coinbase/exchange-rates
func (h *Handlers) GetCoinbaseExchangeRates(c *gin.Context) {
	baseCurrency := c.Query("currency")
	
	rates, err := h.connectorService.GetCoinbaseExchangeRates(baseCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_exchange_rates_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, rates)
}

// EstimateCoinbaseTransactionFee handles POST /v1/coinbase/wallets/:walletId/transactions/estimate-fee
func (h *Handlers) EstimateCoinbaseTransactionFee(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_wallet_id",
			Message: "Wallet ID is required",
			Code:    400,
		})
		return
	}

	var req connector.EstimateFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	feeEstimate, err := h.connectorService.EstimateCoinbaseTransactionFee(walletID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "coinbase_fee_estimate_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, feeEstimate)
}

// RunTests handles GET /tests and returns automated test results
func (h *Handlers) RunTests(c *gin.Context) {
	results := tests.RunAll(h.connectorService, h.cfg)
	c.JSON(http.StatusOK, results)
}

// Overledger Handlers

// GetOverledgerNetworks handles GET /v1/overledger/networks
func (h *Handlers) GetOverledgerNetworks(c *gin.Context) {
	networks, err := h.connectorService.GetOverledgerNetworks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "overledger_networks_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, networks)
}

// GetOverledgerBalance handles GET /v1/overledger/networks/:networkId/addresses/:address/balance
func (h *Handlers) GetOverledgerBalance(c *gin.Context) {
	networkID := c.Param("networkId")
	address := c.Param("address")

	if networkID == "" || address == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_parameters",
			Message: "Network ID and address are required",
			Code:    400,
		})
		return
	}

	balance, err := h.connectorService.GetOverledgerBalance(networkID, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "overledger_balance_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, balance)
}

// CreateOverledgerTransaction handles POST /v1/overledger/transactions
func (h *Handlers) CreateOverledgerTransaction(c *gin.Context) {
	var req overledger.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	transaction, err := h.connectorService.CreateOverledgerTransaction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "overledger_transaction_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusCreated, transaction)
}

// GetOverledgerTransactionStatus handles GET /v1/overledger/networks/:networkId/transactions/:txHash/status
func (h *Handlers) GetOverledgerTransactionStatus(c *gin.Context) {
	networkID := c.Param("networkId")
	txHash := c.Param("txHash")
	
	if networkID == "" || txHash == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "missing_parameters",
			Message: "Network ID and transaction hash are required",
			Code:    400,
		})
		return
	}

	status, err := h.connectorService.GetOverledgerTransactionStatus(networkID, txHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "overledger_transaction_status_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}
	c.JSON(http.StatusOK, status)
}

// TestOverledgerConnection handles GET /v1/overledger/test
func (h *Handlers) TestOverledgerConnection(c *gin.Context) {
	err := h.connectorService.TestOverledgerConnection()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, connector.ErrorResponse{
			Error:   "overledger_connection_failed",
			Message: err.Error(),
			Code:    503,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "connected",
		"message": "Overledger API connection successful",
	})
}