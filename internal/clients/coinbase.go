package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"your-project/internal/utils"
)

type CoinbaseClient struct {
	BaseURL string
	Client  *http.Client
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
	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Generate auth headers
	headers, err := utils.GenerateAuthHeaders(method, path)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth headers: %v", err)
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Execute the request
	return c.Client.Do(req)
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
