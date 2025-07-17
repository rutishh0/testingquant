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

	"github.com/your-username/quant-mesh-connector/internal/config"
)

// Client represents an Overledger API client with OAuth2 authentication
type Client struct {
	config      *config.Config
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
	mutex       sync.RWMutex
}

// TokenResponse represents the OAuth2 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// NewClient creates a new Overledger client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// authenticate obtains an OAuth2 access token using client credentials
func (c *Client) authenticate() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we already have a valid token
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	// Prepare Basic Auth header
	auth := base64.StdEncoding.EncodeToString(
		[]byte(c.config.OverledgerClientID + ":" + c.config.OverledgerClientSecret),
	)

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", c.config.OverledgerAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make auth request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: %d - %s", resp.StatusCode, string(respBody))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	// Set expiry with a 5-minute buffer
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

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
		// Try to parse error response
		var errorResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			return fmt.Errorf("Overledger API error: %v (status: %d)", errorResp, resp.StatusCode)
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

// GetNetworks retrieves available networks from Overledger
func (c *Client) GetNetworks() (*NetworksResponse, error) {
	var resp NetworksResponse
	err := c.makeRequest("GET", "/v2/networks", nil, &resp)
	return &resp, err
}

// GetAccountBalance retrieves account balance for a specific network and address
func (c *Client) GetAccountBalance(networkID, address string) (*BalanceResponse, error) {
	var resp BalanceResponse
	endpoint := fmt.Sprintf("/v2/networks/%s/addresses/%s/balances", networkID, address)
	err := c.makeRequest("GET", endpoint, nil, &resp)
	return &resp, err
}

// CreateTransaction creates a transaction on the specified network
func (c *Client) CreateTransaction(req *TransactionRequest) (*TransactionResponse, error) {
	var resp TransactionResponse
	endpoint := fmt.Sprintf("/v2/networks/%s/transactions", req.NetworkID)
	err := c.makeRequest("POST", endpoint, req, &resp)
	return &resp, err
}

// GetTransactionStatus retrieves the status of a transaction
func (c *Client) GetTransactionStatus(networkID, txHash string) (*TransactionStatusResponse, error) {
	var resp TransactionStatusResponse
	endpoint := fmt.Sprintf("/v2/networks/%s/transactions/%s/status", networkID, txHash)
	err := c.makeRequest("GET", endpoint, nil, &resp)
	return &resp, err
}

// TestConnection tests the connection to Overledger API
func (c *Client) TestConnection() error {
	return c.authenticate()
}