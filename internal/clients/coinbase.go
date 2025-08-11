package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rutishh0/testingquant/internal/utils"
)

const coinbaseAPIPrefix = ""

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
		BaseURL: "https://api.coinbase.com",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DoRequest makes an authenticated request to the Coinbase API
func (c *CoinbaseClient) DoRequest(method, path string, body interface{}) (*http.Response, error) {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	// Only add /platform prefix if not already present
	fullPath := path
	if !strings.HasPrefix(path, coinbaseAPIPrefix) {
		fullPath = coinbaseAPIPrefix + path
	}
	
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

	// Generate auth headers with the exact path the HTTP request will use (incl. /platform)
	headers, err := utils.GenerateAuthHeaders(method, fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth headers: %v", err)
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	
    // Log request details in development mode
    if os.Getenv("LOG_LEVEL") == "debug" {
        log.Printf("[Coinbase Request] %s %s", method, c.BaseURL+fullPath)
        if kid := os.Getenv("COINBASE_API_KEY_ID"); kid != "" {
            log.Printf("[Coinbase Auth] Key ID configured")
        }
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
		
        // Log the error for debugging
        if os.Getenv("LOG_LEVEL") == "debug" {
            log.Printf("[Coinbase Error] Status: %d, Body: %s", resp.StatusCode, string(respBody))
        }
		
		var errorResp CoinbaseErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error.Message != "" {
			// Check for common auth errors and provide helpful messages
            if resp.StatusCode == 401 {
                return nil, fmt.Errorf("authentication failed (401): %s - %s", 
                    errorResp.Error.Code, errorResp.Error.Message)
            }
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
	// Simple connectivity check - we'll try to get networks which is a lightweight endpoint
	resp, err := c.Get("/v1/networks?limit=1")
	if err != nil {
		return fmt.Errorf("Coinbase API health check failed: %v", err)
	}
	defer resp.Body.Close()
	return nil
}

// GetAssets retrieves available assets for trading
func (c *CoinbaseClient) GetAssets() (*http.Response, error) {
    // Try v1 path first
    resp, err := c.Get("/v1/assets")
    if err == nil {
        return resp, nil
    }
    // If 404, retry without version prefix
    if strings.Contains(err.Error(), "HTTP error: 404") || strings.Contains(err.Error(), "no matching operation was found") {
        return c.Get("/assets")
    }
    return nil, err
}

// GetNetworks retrieves available networks
func (c *CoinbaseClient) GetNetworks() (*http.Response, error) {
    // Coinbase Mesh lists networks via POST /network/list
    return c.Post("/network/list", map[string]interface{}{})
}

// GetPortfolio retrieves portfolio information
func (c *CoinbaseClient) GetPortfolio() (*http.Response, error) {
	return c.Get("/v1/portfolios")
}

// GetWallets retrieves wallets
func (c *CoinbaseClient) GetWallets() (*http.Response, error) {
	return c.Get("/v1/wallets")
}

// GetBalances retrieves balances for a specific wallet
func (c *CoinbaseClient) GetBalances(walletID string) (*http.Response, error) {
	return c.Get(fmt.Sprintf("/v1/wallets/%s/balances", walletID))
}

// GetWalletAddresses retrieves addresses for a specific wallet
func (c *CoinbaseClient) GetWalletAddresses(walletID string) (*http.Response, error) {
	return c.Get(fmt.Sprintf("/v1/wallets/%s/addresses", walletID))
}

// CreateWalletAddress creates a new address for a wallet
func (c *CoinbaseClient) CreateWalletAddress(walletID, name string) (*http.Response, error) {
	return c.Post(fmt.Sprintf("/v1/wallets/%s/addresses", walletID), map[string]string{"name": name})
}

// CreateAddress is an alias for CreateWalletAddress to satisfy the Adapter interface
func (c *CoinbaseClient) CreateAddress(walletID, name string) (*http.Response, error) {
	return c.CreateWalletAddress(walletID, name)
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

// GetTransaction retrieves a single transaction by ID
func (c *CoinbaseClient) GetTransaction(transactionID string) (*http.Response, error) {
	return c.Get(fmt.Sprintf("/v1/transactions/%s", transactionID))
}

// CreateTransaction creates a new transaction
func (c *CoinbaseClient) CreateTransaction(to, currency string, amount float64) (*http.Response, error) {
	body := map[string]interface{}{
		"to":       to,
		"currency": currency,
		"amount":   amount,
	}
	return c.Post("/v1/transactions", body)
}



// EstimateFee estimates the fee for a transaction
func (c *CoinbaseClient) EstimateFee(walletID, to, currency string, amount float64) (*http.Response, error) {
	body := map[string]interface{}{
		"to":       to,
		"currency": currency,
		"amount":   amount,
	}
	return c.Post(fmt.Sprintf("/v1/wallets/%s/estimate-fee", walletID), body)
}

// CreateWallet creates a new wallet
func (c *CoinbaseClient) CreateWallet(name string) (*http.Response, error) {
	return c.Post("/v1/wallets", map[string]string{"name": name})
}
// GetExchangeRates retrieves current exchange rates
func (c *CoinbaseClient) GetExchangeRates(baseCurrency string) (*http.Response, error) {
    path1 := "/v1/exchange-rates"
    if baseCurrency != "" {
        path1 += "?currency=" + baseCurrency
    }
    resp, err := c.Get(path1)
    if err == nil {
        return resp, nil
    }
    path2 := "/exchange-rates"
    if baseCurrency != "" {
        path2 += "?currency=" + baseCurrency
    }
    return c.Get(path2)
}

// EstimateTransactionFee estimates the fee for a transaction
func (c *CoinbaseClient) EstimateTransactionFee(walletID string, req interface{}) (*http.Response, error) {
	return c.Post(fmt.Sprintf("/v1/wallets/%s/transactions/estimate-fee", walletID), req)
}

// BroadcastTransaction broadcasts a signed transaction
func (c *CoinbaseClient) BroadcastTransaction(walletID string, req interface{}) (*http.Response, error) {
	return c.Post(fmt.Sprintf("/v1/wallets/%s/transactions/broadcast", walletID), req)
}
