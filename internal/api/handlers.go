package api

import (
	"net/http"
	"time"

	"github.com/rutishh0/testingquant/internal/connector"
	"github.com/rutishh0/testingquant/internal/overledger"

	"github.com/gin-gonic/gin"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	connectorService connector.Service
}

// NewHandlers creates a new handlers instance
func NewHandlers(connectorService connector.Service) *Handlers {
	return &Handlers{
		connectorService: connectorService,
	}
}

// Health handles health check requests
func (h *Handlers) Health(c *gin.Context) {
	response := connector.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

// Status handles status requests
func (h *Handlers) Status(c *gin.Context) {
	response := connector.StatusResponse{
		Service:   "quant-mesh-connector",
		Status:    "running",
		Uptime:    "N/A", // In a real implementation, calculate actual uptime
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// Preprocess handles construction preprocess requests
func (h *Handlers) Preprocess(c *gin.Context) {
	var req connector.PreprocessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.Preprocess(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "preprocess_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Payloads handles construction payloads requests
func (h *Handlers) Payloads(c *gin.Context) {
	var req connector.PayloadsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.Payloads(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "payloads_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Combine handles construction combine requests
func (h *Handlers) Combine(c *gin.Context) {
	var req connector.CombineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.Combine(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "combine_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Submit handles construction submit requests
func (h *Handlers) Submit(c *gin.Context) {
	var req connector.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.Submit(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "submit_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetBalance handles balance requests
func (h *Handlers) GetBalance(c *gin.Context) {
	var req connector.BalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.GetBalance(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "balance_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetBlock handles block requests
func (h *Handlers) GetBlock(c *gin.Context) {
	var req connector.BlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.GetBlock(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "block_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTransaction handles transaction requests
func (h *Handlers) GetTransaction(c *gin.Context) {
	var req connector.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.GetTransaction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "transaction_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Overledger-specific handlers

// GetOverledgerNetworks handles Overledger networks requests
func (h *Handlers) GetOverledgerNetworks(c *gin.Context) {
	resp, err := h.connectorService.GetOverledgerNetworks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "overledger_networks_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetOverledgerBalance handles Overledger balance requests
func (h *Handlers) GetOverledgerBalance(c *gin.Context) {
	networkID := c.Param("networkId")
	address := c.Param("address")

	if networkID == "" || address == "" {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: "networkId and address are required",
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.GetOverledgerBalance(networkID, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "overledger_balance_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateOverledgerTransaction handles Overledger transaction creation
func (h *Handlers) CreateOverledgerTransaction(c *gin.Context) {
	var req struct {
		NetworkID   string                 `json:"networkId" binding:"required"`
		FromAddress string                 `json:"fromAddress" binding:"required"`
		ToAddress   string                 `json:"toAddress" binding:"required"`
		Amount      string                 `json:"amount" binding:"required"`
		TokenSymbol string                 `json:"tokenSymbol,omitempty"`
		Metadata    map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	// Map to Overledger transaction request
	txReq := &overledger.TransactionRequest{
		NetworkID:   req.NetworkID,
		FromAddress: req.FromAddress,
		ToAddress:   req.ToAddress,
		Amount:      req.Amount,
		TokenID:     req.TokenSymbol,
		Metadata:    req.Metadata,
	}

	resp, err := h.connectorService.CreateOverledgerTransaction(txReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "overledger_transaction_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// TestOverledgerConnection handles Overledger connection test
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