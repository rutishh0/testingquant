package overledger

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rutishh0/testingquant/internal/config"
)

// Client represents an Overledger API client
type Client struct {
	config      *config.Config
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
	mutex       sync.RWMutex
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// NewClient creates a new Overledger API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// authenticate obtains an access token using client credentials
func (c *Client) authenticate() error {
	c.mutex.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-1*time.Minute)) {
		c.mutex.RUnlock()
		return nil
	}
	c.mutex.RUnlock()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Re-check after acquiring write lock
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-1*time.Minute)) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", c.config.OverledgerAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	credentials := base64.StdEncoding.EncodeToString([]byte(c.config.OverledgerClientID + ":" + c.config.OverledgerClientSecret))
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// makeRequest makes an authenticated HTTP request to the Overledger API
func (c *Client) makeRequest(method, endpoint string, payload interface{}, response interface{}) error {
	// Ensure we have a valid access token
	if err := c.authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.config.OverledgerBaseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	c.mutex.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mutex.RUnlock()
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to parse structured error response
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error.Message != "" {
			return fmt.Errorf("Overledger API error: %s (code: %s, details: %s, status: %d)",
				errorResp.Error.Message, errorResp.Error.Code, errorResp.Error.Details, resp.StatusCode)
		}
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
	}

	if response != nil {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// baseHasVersion returns true if OverledgerBaseURL already contains a version segment like /v2 or /v2.1
func (c *Client) baseHasVersion() bool {
	u, err := url.Parse(c.config.OverledgerBaseURL)
	if err != nil {
		return strings.Contains(c.config.OverledgerBaseURL, "/v")
	}
	return strings.Contains(u.Path, "/v")
}

// GetNetworks retrieves available networks from Overledger
func (c *Client) GetNetworks() (*NetworksResponse, error) {
	var resp NetworksResponse
	if c.baseHasVersion() {
		// Base URL already includes version like /v2.1
		if err := c.makeRequest("GET", "/networks", nil, &resp); err == nil {
			return &resp, nil
		}
		return nil, fmt.Errorf("failed to fetch networks from known endpoints")
	}
	// Base URL without version: try v2.1, then v2, then legacy
	if err := c.makeRequest("GET", "/v2.1/networks", nil, &resp); err == nil {
		return &resp, nil
	} else if !strings.Contains(err.Error(), "HTTP error: 404") {
		return nil, err
	}
	if err := c.makeRequest("GET", "/v2/networks", nil, &resp); err == nil {
		return &resp, nil
	} else if !strings.Contains(err.Error(), "HTTP error: 404") {
		return nil, err
	}
	if err := c.makeRequest("GET", "/v2/mdapi/networks", nil, &resp); err == nil {
		return &resp, nil
	}
	return nil, fmt.Errorf("failed to fetch networks from known endpoints")
}

// GetAccountBalance retrieves account balance for a specific network and address
func (c *Client) GetAccountBalance(networkID, address string) (*BalanceResponse, error) {
	var resp BalanceResponse
	if c.baseHasVersion() {
		endpoint := fmt.Sprintf("/networks/%s/addresses/%s/balances", networkID, address)
		if err := c.makeRequest("GET", endpoint, nil, &resp); err == nil {
			return &resp, nil
		}
		return nil, fmt.Errorf("failed to fetch balance from known endpoints")
	}
	endpointV21 := fmt.Sprintf("/v2.1/networks/%s/addresses/%s/balances", networkID, address)
	if err := c.makeRequest("GET", endpointV21, nil, &resp); err == nil {
		return &resp, nil
	} else if !strings.Contains(err.Error(), "HTTP error: 404") {
		return nil, err
	}
	endpointV2 := fmt.Sprintf("/v2/networks/%s/addresses/%s/balances", networkID, address)
	if err := c.makeRequest("GET", endpointV2, nil, &resp); err == nil {
		return &resp, nil
	}
	return nil, fmt.Errorf("failed to fetch balance from known endpoints")
}

// CreateTransaction creates a transaction on the specified network
func (c *Client) CreateTransaction(req *TransactionRequest) (*TransactionResponse, error) {
	var resp TransactionResponse
	if c.baseHasVersion() {
		// With versioned base, unified endpoint is "/transactions"
		if err := c.makeRequest("POST", "/transactions", req, &resp); err == nil {
			return &resp, nil
		}
		// Some deployments may still use network-scoped path
		endpoint := fmt.Sprintf("/networks/%s/transactions", req.NetworkID)
		if err := c.makeRequest("POST", endpoint, req, &resp); err == nil {
			return &resp, nil
		}
		return nil, fmt.Errorf("failed to create transaction on known endpoints")
	}
	// Base URL without version: try v2.1 unified, then v2 network-scoped
	if err := c.makeRequest("POST", "/v2.1/transactions", req, &resp); err == nil {
		return &resp, nil
	} else if !strings.Contains(err.Error(), "HTTP error: 404") {
		return nil, err
	}
	endpoint := fmt.Sprintf("/v2/networks/%s/transactions", req.NetworkID)
	if err := c.makeRequest("POST", endpoint, req, &resp); err == nil {
		return &resp, nil
	}
	return nil, fmt.Errorf("failed to create transaction on known endpoints")
}

// GetTransactionStatus retrieves the status of a transaction
func (c *Client) GetTransactionStatus(networkID, txHash string) (*TransactionStatusResponse, error) {
	var resp TransactionStatusResponse
	if c.baseHasVersion() {
		endpoint := fmt.Sprintf("/networks/%s/transactions/%s/status", networkID, txHash)
		if err := c.makeRequest("GET", endpoint, nil, &resp); err == nil {
			return &resp, nil
		}
		return nil, fmt.Errorf("failed to fetch transaction status from known endpoints")
	}
	endpointV21 := fmt.Sprintf("/v2.1/networks/%s/transactions/%s/status", networkID, txHash)
	if err := c.makeRequest("GET", endpointV21, nil, &resp); err == nil {
		return &resp, nil
	} else if !strings.Contains(err.Error(), "HTTP error: 404") {
		return nil, err
	}
	endpointV2 := fmt.Sprintf("/v2/networks/%s/transactions/%s/status", networkID, txHash)
	if err := c.makeRequest("GET", endpointV2, nil, &resp); err == nil {
		return &resp, nil
	}
	return nil, fmt.Errorf("failed to fetch transaction status from known endpoints")
}

// TestConnection tests the connection to Overledger API
func (c *Client) TestConnection() error {
	return c.authenticate()
}