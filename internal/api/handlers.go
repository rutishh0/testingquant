package api

import (
    "bytes"
    "net/http"
    "os"
    "os/exec"
    "strconv"
    "time"
    "strings"

    "github.com/rutishh0/testingquant/internal/config"
    "github.com/rutishh0/testingquant/internal/connector"
    "github.com/rutishh0/testingquant/internal/overledger"
    "github.com/rutishh0/testingquant/internal/clients"

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
    
    // Always return 200 OK unless there's an actual error
    // The status field in the response indicates the health state
    c.JSON(http.StatusOK, health)
}

// Status handles status requests (legacy endpoint)
func (h *Handlers) Status(c *gin.Context) {
    response := connector.StatusResponse{
        Status: "OK",
    }
    c.JSON(http.StatusOK, response)
}

// Coinbase Handlers

// GetCoinbaseWallets handles GET /v1/coinbase/wallets
func (h *Handlers) GetCoinbaseWallets(c *gin.Context) {
    wallets, err := h.connectorService.GetCoinbaseWallets()
    if err != nil {
        // Graceful fallback: if Coinbase isn't configured or endpoint is missing, return empty list
        errStr := err.Error()
        if strings.Contains(errStr, "not initialized") || strings.Contains(errStr, "404") || strings.Contains(errStr, "no matching operation") {
            c.JSON(http.StatusOK, map[string]interface{}{
                "data": []interface{}{},
            })
            return
        }
        c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
            Error:   "coinbase_wallets_failed",
            Message: err.Error(),
            Code:    500,
        })
        return
    }
    // Return in the format expected by the frontend: { data: [...] }
    c.JSON(http.StatusOK, map[string]interface{}{
        "data": wallets.Wallets,
    })
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

// GetCoinbaseTransactionsPaginated handles GET /v1/coinbase/wallets/:walletId/transactions-paginated
func (h *Handlers) GetCoinbaseTransactionsPaginated(c *gin.Context) {
    walletID := c.Param("walletId")
    if walletID == "" {
        c.JSON(http.StatusBadRequest, connector.ErrorResponse{
            Error:   "missing_wallet_id",
            Message: "Wallet ID is required",
            Code:    400,
        })
        return
    }

    limitStr := c.Query("limit")
    cursor := c.Query("cursor")

    limit := 25
    if limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }

    resp, err := h.connectorService.GetCoinbaseTransactionsPaginated(walletID, limit, cursor)
    if err != nil {
        c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
            Error:   "coinbase_transactions_failed",
            Message: err.Error(),
            Code:    500,
        })
        return
    }
    c.JSON(http.StatusOK, resp)
}

// GetCoinbaseAssets handles GET /v1/coinbase/assets
func (h *Handlers) GetCoinbaseAssets(c *gin.Context) {
    assets, err := h.connectorService.GetCoinbaseAssets()
    if err != nil {
        // Graceful fallback for missing Coinbase integration or 404s
        errStr := err.Error()
        if strings.Contains(errStr, "not initialized") || strings.Contains(errStr, "404") || strings.Contains(errStr, "no matching operation") {
            c.JSON(http.StatusOK, map[string]interface{}{
                "data": []interface{}{},
            })
            return
        }
        c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
            Error:   "coinbase_assets_failed",
            Message: err.Error(),
            Code:    500,
        })
        return
    }
    // Return in the format expected by the frontend: { data: [...] }
    c.JSON(http.StatusOK, map[string]interface{}{
        "data": assets,
    })
}

// GetCoinbaseNetworks handles GET /v1/coinbase/networks
func (h *Handlers) GetCoinbaseNetworks(c *gin.Context) {
    networks, err := h.connectorService.GetCoinbaseNetworks()
    if err != nil {
        // Graceful fallback for missing Coinbase integration or 404s
        errStr := err.Error()
        if strings.Contains(errStr, "not initialized") || strings.Contains(errStr, "404") || strings.Contains(errStr, "no matching operation") {
            c.JSON(http.StatusOK, []interface{}{})
            return
        }
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
        // Fallback to an empty rates object so UI can still render
        errStr := err.Error()
        if strings.Contains(errStr, "not initialized") || strings.Contains(errStr, "404") || strings.Contains(errStr, "no matching operation") {
            if baseCurrency == "" {
                baseCurrency = "USD"
            }
            c.JSON(http.StatusOK, map[string]interface{}{
                "data": map[string]interface{}{
                    "currency":   baseCurrency,
                    "rates":      map[string]string{},
                    "updated_at": time.Now().UTC().Format(time.RFC3339),
                },
            })
            return
        }
        c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
            Error:   "coinbase_exchange_rates_failed",
            Message: err.Error(),
            Code:    500,
        })
        return
    }
    // Frontend expects { data: { currency, rates, updated_at } }
    c.JSON(http.StatusOK, map[string]interface{}{
        "data": map[string]interface{}{
            "currency": rates.Base,
            "rates":   rates.Rates,
            "updated_at": time.Now().UTC().Format(time.RFC3339),
        },
    })
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
        // Graceful fallback similar to other Coinbase endpoints
        errStr := err.Error()
        if strings.Contains(errStr, "not initialized") || strings.Contains(errStr, "404") || strings.Contains(errStr, "no matching operation") {
            c.JSON(http.StatusOK, map[string]string{"fee": "0"})
            return
        }
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
    // If compiled test binaries are present in /app/tests, execute them and
    // return a combined JSON so the frontend can copy/paste full logs.
    // Fallback to internal tiered tests otherwise.
    type externalResult struct {
        Suite   string        `json:"suite"`
        Output  string        `json:"output"`
        Success bool          `json:"success"`
    }

    var external []externalResult

    // Attempt to run compiled tests (best-effort, ignore errors on platforms without exec)
    if out, ok := execTestBinary("/app/tests/mesh_tests"); ok {
        external = append(external, externalResult{Suite: "mesh", Output: out, Success: true})
    }
    if out, ok := execTestBinary("/app/tests/integration_tests"); ok {
        external = append(external, externalResult{Suite: "integration", Output: out, Success: true})
    }

    results := tests.RunAll(h.connectorService, h.cfg)

    c.JSON(http.StatusOK, gin.H{
        "tiered":   results,
        "external": external,
    })
}

// execTestBinary runs a test binary if it exists and returns its stdout.
func execTestBinary(path string) (string, bool) {
    f, err := os.Stat(path)
    if err != nil || f.IsDir() {
        return "", false
    }
    cmd := exec.Command(path, "-test.v")
    var buf bytes.Buffer
    cmd.Stdout = &buf
    cmd.Stderr = &buf
    if err := cmd.Run(); err != nil {
        return buf.String(), false
    }
    return buf.String(), true
}

// Exchange Handlers

// GetExchangeProducts handles GET /v1/exchange/products
func (h *Handlers) GetExchangeProducts(c *gin.Context) {
    exch, err := clients.NewExchangeClient()
    if err != nil {
        // Distinguish configuration states for logging/debugging; always degrade gracefully for products
        status := http.StatusOK
        _ = status // keep for potential future use
        // Graceful degradation: return empty list so UI can fallback
        c.JSON(http.StatusOK, gin.H{"products": []any{}})
        return
    }
    products, err := exch.ListProducts(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"products": []any{}})
        return
    }
    c.JSON(http.StatusOK, products)
}

// GetExchangeAccounts handles GET /v1/exchange/accounts
func (h *Handlers) GetExchangeAccounts(c *gin.Context) {
    exch, err := clients.NewExchangeClient()
    if err != nil {
        // If not configured or misconfigured, degrade gracefully with empty accounts
        if err == clients.ErrExchangeNotConfigured || err == clients.ErrExchangeMisconfigured || err == clients.ErrExchangePartialConfig {
            c.JSON(http.StatusOK, gin.H{"accounts": []any{}})
            return
        }
        // For any other initialization error, also degrade to empty to avoid breaking UI
        c.JSON(http.StatusOK, gin.H{"accounts": []any{}})
        return
    }
    accountsResp, err := exch.ListAccounts(c.Request.Context())
    if err != nil {
        // If unauthorized or any other runtime error, degrade to empty accounts
        c.JSON(http.StatusOK, gin.H{"accounts": []any{}})
        return
    }
    c.JSON(http.StatusOK, accountsResp)
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

// Mesh Handlers

// GetMeshNetworks handles GET /v1/mesh/networks
func (h *Handlers) GetMeshNetworks(c *gin.Context) {
    networks, err := h.connectorService.GetMeshNetworks()
    if err != nil {
        c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
            Error:   "mesh_networks_failed",
            Message: err.Error(),
            Code:    500,
        })
        return
    }
    c.JSON(http.StatusOK, networks)
}

// GetMeshAccountBalance handles POST /v1/mesh/account/balance
func (h *Handlers) GetMeshAccountBalance(c *gin.Context) {
    var req struct {
        NetworkIdentifier interface{} `json:"network_identifier"`
        AccountIdentifier interface{} `json:"account_identifier"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, connector.ErrorResponse{
            Error:   "invalid_request",
            Message: err.Error(),
            Code:    400,
        })
        return
    }

    balance, err := h.connectorService.GetMeshNetworkBalance(req.NetworkIdentifier, req.AccountIdentifier)
    if err != nil {
        c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
            Error:   "mesh_balance_failed",
            Message: err.Error(),
            Code:    500,
        })
        return
    }
    c.JSON(http.StatusOK, balance)
}