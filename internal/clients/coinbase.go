package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rutishh0/testingquant/internal/utils"
)

const coinbaseAPIPrefix = "/platform"

type CoinbaseClient struct {
	BaseURL string
	Client  *http.Client
}

// CoinbaseError represents a Coinbase API error response
type CoinbaseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// CoinbaseErrorResponse represents the full error response structure
type CoinbaseErrorResponse struct {
	Error CoinbaseError `json:"error"`
}

func NewCoinbaseClient() *CoinbaseClient {
	return &CoinbaseClient{
		BaseURL: "https://api.cdp.coinbase.com",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DoRequest makes an authenticated request to the Coinbase API
func (c *CoinbaseClient) DoRequest(method, path string, body interface{}) (*http.Response, error) {
	// ensure prefix
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	fullPath := coinbaseAPIPrefix + path
	var reqBody io.Reader = nil
	
	// Marshal request body if provided
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Create the request
	req, err := http.NewRequest(method, c.BaseURL+fullPath, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Generate auth headers (must include the same path the request will use)
	headers, err := utils.GenerateAuthHeaders(method, fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth headers: %v", err)
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Execute the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	// Handle API errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		
		var errorResp CoinbaseErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error.Message != "" {
			return nil, fmt.Errorf("Coinbase API error (%d): %s - %s", 
				resp.StatusCode, errorResp.Error.Code, errorResp.Error.Message)
		}
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
	}

	return resp, nil
}

// Get makes a GET request to the Coinbase API
func (c *CoinbaseClient) Get(path string) (*http.Response, error) {
	return c.DoRequest(http.MethodGet, path, nil)
}

// Post makes a POST request to the Coinbase API
func (c *CoinbaseClient) Post(path string, body interface{}) (*http.Response, error) {
	return c.DoRequest(http.MethodPost, path, body)
}

// Put makes a PUT request to the Coinbase API
func (c *CoinbaseClient) Put(path string, body interface{}) (*http.Response, error) {
	return c.DoRequest(http.MethodPut, path, body)
}

// Delete makes a DELETE request to the Coinbase API
func (c *CoinbaseClient) Delete(path string) (*http.Response, error) {
	return c.DoRequest(http.MethodDelete, path, nil)
}

// Health checks the Coinbase API availability
func (c *CoinbaseClient) Health() error {
	// Use a simple GET request to check connectivity
	resp, err := c.Get("/v1/wallets?limit=1")
	if err != nil {
		return fmt.Errorf("Coinbase API health check failed: %v", err)
	}
	defer resp.Body.Close()
	return nil
}

// GetAssets retrieves available assets for trading
func (c *CoinbaseClient) GetAssets() (*http.Response, error) {
	return c.Get("/v1/assets")
}

// GetNetworks retrieves available networks
func (c *CoinbaseClient) GetNetworks() (*http.Response, error) {
	return c.Get("/v1/networks")
}

// GetPortfolio retrieves portfolio information
func (c *CoinbaseClient) GetPortfolio() (*http.Response, error) {
	return c.Get("/v1/portfolios")
}

// GetWalletAddresses retrieves addresses for a specific wallet
func (c *CoinbaseClient) GetWalletAddresses(walletID string) (*http.Response, error) {
	return c.Get(fmt.Sprintf("/v1/wallets/%s/addresses", walletID))
}

// CreateWalletAddress creates a new address for a wallet
func (c *CoinbaseClient) CreateWalletAddress(walletID string, req interface{}) (*http.Response, error) {
	return c.Post(fmt.Sprintf("/v1/wallets/%s/addresses", walletID), req)
}

// GetTransactions retrieves transactions for a wallet
func (c *CoinbaseClient) GetTransactions(walletID string, limit int, cursor string) (*http.Response, error) {
	path := fmt.Sprintf("/v1/wallets/%s/transactions", walletID)
	if limit > 0 || cursor != "" {
		path += "?"
		if limit > 0 {
			path += fmt.Sprintf("limit=%d", limit)
		}
		if cursor != "" {
			if limit > 0 {
				path += "&"
			}
			path += fmt.Sprintf("cursor=%s", cursor)
		}
	}
	return c.Get(path)
}

// GetExchangeRates retrieves current exchange rates
func (c *CoinbaseClient) GetExchangeRates(baseCurrency string) (*http.Response, error) {
	path := "/v1/exchange-rates"
	if baseCurrency != "" {
		path += "?currency=" + baseCurrency
	}
	return c.Get(path)
}

// EstimateTransactionFee estimates the fee for a transaction
func (c *CoinbaseClient) EstimateTransactionFee(walletID string, req interface{}) (*http.Response, error) {
	return c.Post(fmt.Sprintf("/v1/wallets/%s/transactions/estimate-fee", walletID), req)
}

// BroadcastTransaction broadcasts a signed transaction
func (c *CoinbaseClient) BroadcastTransaction(walletID string, req interface{}) (*http.Response, error) {
	return c.Post(fmt.Sprintf("/v1/wallets/%s/transactions/broadcast", walletID), req)
}
