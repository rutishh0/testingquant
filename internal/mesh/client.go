package mesh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Mesh API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Mesh client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an HTTP request to the Mesh API
func (c *Client) makeRequest(method, endpoint string, payload interface{}, response interface{}) error {
	var body io.Reader
	
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

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
		var meshError Error
		if err := json.Unmarshal(respBody, &meshError); err == nil {
			return fmt.Errorf("mesh API error: %s (code: %d)", meshError.Message, meshError.Code)
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

// NetworkStatus gets the current network status
func (c *Client) NetworkStatus(req *NetworkStatusRequest) (*NetworkStatusResponse, error) {
	var resp NetworkStatusResponse
	err := c.makeRequest("POST", "/network/status", req, &resp)
	return &resp, err
}

// AccountBalance gets account balance
func (c *Client) AccountBalance(req *AccountBalanceRequest) (*AccountBalanceResponse, error) {
	var resp AccountBalanceResponse
	err := c.makeRequest("POST", "/account/balance", req, &resp)
	return &resp, err
}

// Block gets block information
func (c *Client) Block(req *BlockRequest) (*BlockResponse, error) {
	var resp BlockResponse
	err := c.makeRequest("POST", "/block", req, &resp)
	return &resp, err
}

// ConstructionPreprocess preprocesses a construction request
func (c *Client) ConstructionPreprocess(req *ConstructionPreprocessRequest) (*ConstructionPreprocessResponse, error) {
	var resp ConstructionPreprocessResponse
	err := c.makeRequest("POST", "/construction/preprocess", req, &resp)
	return &resp, err
}

// ConstructionPayloads creates payloads for signing
func (c *Client) ConstructionPayloads(req *ConstructionPayloadsRequest) (*ConstructionPayloadsResponse, error) {
	var resp ConstructionPayloadsResponse
	err := c.makeRequest("POST", "/construction/payloads", req, &resp)
	return &resp, err
}

// ConstructionCombine combines signatures with unsigned transaction
func (c *Client) ConstructionCombine(req *ConstructionCombineRequest) (*ConstructionCombineResponse, error) {
	var resp ConstructionCombineResponse
	err := c.makeRequest("POST", "/construction/combine", req, &resp)
	return &resp, err
}

// ConstructionSubmit submits a signed transaction
func (c *Client) ConstructionSubmit(req *ConstructionSubmitRequest) (*ConstructionSubmitResponse, error) {
	var resp ConstructionSubmitResponse
	err := c.makeRequest("POST", "/construction/submit", req, &resp)
	return &resp, err
}

// Health checks if the Mesh API is healthy
func (c *Client) Health() error {
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}