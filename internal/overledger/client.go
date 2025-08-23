package overledger

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"math/big"
	"os"
	"strconv"
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

// mapNetworkToLocation maps a networkID to Overledger "location" fields based on official documentation
func (c *Client) mapNetworkToLocation(networkID string) (technology, network string) {
	networkIDLower := strings.ToLower(networkID)
	
	// Handle Ethereum networks according to Overledger documentation (lowercase as per screenshots)
	if strings.Contains(networkIDLower, "ethereum") {
		technology = "ethereum"  // Use lowercase as shown in screenshots
		if strings.Contains(networkIDLower, "sepolia") {
			// Exact network name from screenshots
			network = "ethereum sepolia testnet"
			return
		}
		if strings.Contains(networkIDLower, "mainnet") {
			network = "ethereum mainnet"
			return
		}
		// Default to mainnet if not specified
		network = "ethereum mainnet"
		return
	}
	
	// Handle Polygon networks
	if strings.Contains(networkIDLower, "polygon") {
		technology = "ethereum"  // Polygon uses Ethereum technology
		if strings.Contains(networkIDLower, "amoy") || strings.Contains(networkIDLower, "test") {
			network = "polygon amoy testnet"
		} else {
			network = "polygon mainnet"
		}
		return
	}
	
	// Handle Bitcoin networks
	if strings.Contains(networkIDLower, "bitcoin") {
		technology = "bitcoin"
		if strings.Contains(networkIDLower, "test") {
			network = "testnet"
		} else {
			network = "mainnet"
		}
		return
	}
	
	// Handle XRP networks
	if strings.Contains(networkIDLower, "xrp") || strings.Contains(networkIDLower, "ripple") {
		technology = "xrp ledger"
		if strings.Contains(networkIDLower, "test") {
			network = "xrp ledger testnet"
		} else {
			network = "xrp ledger mainnet"
		}
		return
	}
	
	// Fallback: attempt to map unknown networks
	parts := strings.Split(networkIDLower, "-")
	if len(parts) > 0 {
		technology = parts[0]  // Keep lowercase
	}
	// Best-effort mapping for unknown testnets
	if strings.Contains(networkIDLower, "test") || strings.Contains(networkIDLower, "sepolia") || strings.Contains(networkIDLower, "goerli") {
		network = technology + " testnet"
	} else {
		network = technology + " mainnet"
	}
	return
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
	// Base URL without version: try versioned endpoints
	if err := c.makeRequest("GET", "/v2.1/networks", nil, &resp); err == nil {
		return &resp, nil
	} else if !strings.Contains(err.Error(), "HTTP error: 404") {
		return nil, err
	}
	if err := c.makeRequest("GET", "/v2/networks", nil, &resp); err == nil {
		return &resp, nil
	}
	return nil, fmt.Errorf("failed to fetch networks from known endpoints")
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

// PrepareTransaction prepares a transaction for signing using Overledger's preparation endpoint
func (c *Client) PrepareTransaction(req *TransactionPrepareRequest) (*TransactionPrepareResponse, error) {
	var resp TransactionPrepareResponse
	errs := []string{}

	// Primary endpoint: /v2/preparation/transaction (as shown in screenshots)
	if err := c.makeRequest("POST", "/v2/preparation/transaction", req, &resp); err == nil {
		return &resp, nil
	} else {
		errs = append(errs, fmt.Sprintf("/v2/preparation/transaction: %v", err))
		// Continue to try other endpoints only if this is a 404
		if !strings.Contains(err.Error(), "HTTP error: 404") {
			return nil, fmt.Errorf("failed to prepare transaction: %s", strings.Join(errs, " | "))
		}
	}

	// Fallback: Try native transaction preparation endpoint
	if err := c.makeRequest("POST", "/v2/preparation/nativetransaction", req, &resp); err == nil {
		return &resp, nil
	} else {
		errs = append(errs, fmt.Sprintf("/v2/preparation/nativetransaction: %v", err))
		if !strings.Contains(err.Error(), "HTTP error: 404") {
			return nil, fmt.Errorf("failed to prepare transaction: %s", strings.Join(errs, " | "))
		}
	}

	// Last resort: Try autoexecution preparation endpoint
	if err := c.makeRequest("POST", "/v2/autoexecution/preparation/transaction", req, &resp); err == nil {
		return &resp, nil
	} else {
		errs = append(errs, fmt.Sprintf("/v2/autoexecution/preparation/transaction: %v", err))
	}

	return nil, fmt.Errorf("failed to prepare transaction on all known endpoints: %s", strings.Join(errs, " | "))
}

// ExecuteTransaction executes a signed transaction using Overledger's execution endpoint
func (c *Client) ExecuteTransaction(req *TransactionExecuteRequest) (*TransactionExecuteResponse, error) {
	var resp TransactionExecuteResponse
	errs := []string{}

	// Primary endpoint: v2 execution transaction (from API documentation)
	if err := c.makeRequest("POST", "/v2/execution/transaction", req, &resp); err == nil {
		return &resp, nil
	} else {
		errs = append(errs, fmt.Sprintf("/v2/execution/transaction: %v", err))
		// Continue to try other endpoints only if this is a 404
		if !strings.Contains(err.Error(), "HTTP error: 404") {
			return nil, fmt.Errorf("failed to execute transaction: %s", strings.Join(errs, " | "))
		}
	}

	// Fallback: Try native transaction execution endpoint
	if err := c.makeRequest("POST", "/v2/execution/nativetransaction", req, &resp); err == nil {
		return &resp, nil
	} else {
		errs = append(errs, fmt.Sprintf("/v2/execution/nativetransaction: %v", err))
		if !strings.Contains(err.Error(), "HTTP error: 404") {
			return nil, fmt.Errorf("failed to execute transaction: %s", strings.Join(errs, " | "))
		}
	}

	// Fallback: Try autoexecution endpoint
	if err := c.makeRequest("POST", "/v2/autoexecution/execution", req, &resp); err == nil {
		return &resp, nil
	} else {
		errs = append(errs, fmt.Sprintf("/v2/autoexecution/execution: %v", err))
		if !strings.Contains(err.Error(), "HTTP error: 404") {
			return nil, fmt.Errorf("failed to execute transaction: %s", strings.Join(errs, " | "))
		}
	}

	return nil, fmt.Errorf("failed to execute transaction on all known endpoints: %s", strings.Join(errs, " | "))
}

// CreateTransaction implements the exact 3-step Overledger workflow from the curl examples
func (c *Client) CreateTransaction(req *TransactionRequest) (*TransactionResponse, error) {
	// Step 1: Get Bearer Token (hardcoded format)
	bearerToken, err := c.getBearerTokenHardcoded()
	if err != nil {
		return nil, fmt.Errorf("failed to get bearer token: %w", err)
	}

	// Step 2: Prepare Transaction (hardcoded format)
	prepareResp, err := c.prepareTransactionHardcoded(bearerToken, req)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare transaction: %w", err)
	}

	// Step 3: Sign Transaction (hardcoded format)
	signedData, err := c.signTransactionHardcoded(bearerToken, prepareResp, req)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Step 4: Execute Transaction (hardcoded format)
	execResp, err := c.executeTransactionHardcoded(bearerToken, signedData, prepareResp.PreparationTransactionSearchResponse.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	// Convert to legacy response format
	result := &TransactionResponse{
		TransactionID: execResp.ExecutionTransactionSearchResponse.TransactionID,
		Hash:          execResp.ExecutionTransactionSearchResponse.TransactionID,
		Status:        execResp.ExecutionTransactionSearchResponse.Status.Value,
		NetworkID:     req.NetworkID,
		FromAddress:   req.FromAddress,
		ToAddress:     req.ToAddress,
		Amount:        req.Amount,
		Timestamp:     time.Now(),
	}
	// Surface any execution message so the UI can display it
	if execResp.ExecutionTransactionSearchResponse.Message != "" {
		if result.Metadata == nil {
			result.Metadata = map[string]interface{}{}
		}
		result.Metadata["message"] = execResp.ExecutionTransactionSearchResponse.Message
	}
	// Also surface execution status code/description for frontend display
	if result.Metadata == nil {
		result.Metadata = map[string]interface{}{}
	}
	result.Metadata["execution"] = map[string]string{
		"value":       execResp.ExecutionTransactionSearchResponse.Status.Value,
		"code":        execResp.ExecutionTransactionSearchResponse.Status.Code,
		"description": execResp.ExecutionTransactionSearchResponse.Status.Description,
	}
	// Provide transactionId in metadata for resilient frontend linking
	result.Metadata["transactionId"] = execResp.ExecutionTransactionSearchResponse.TransactionID

	// Increment nonce after successful creation so the next tx uses a new nonce.
	bumpNonceFile()
	return result, nil
}

// getBearerTokenHardcoded follows the exact curl format for getting bearer token
func (c *Client) getBearerTokenHardcoded() (string, error) {
	// Exact curl format: 
	// curl --request POST \
	// --url https://auth.overledger.dev/oauth2/token \
	// --header 'accept: application/json' \
	// --header 'authorization: Basic M25ocXBzdDkzNXYwa3F1bWMzczc2amNxNDY6MTAybDBlYWJycXVtNHBhbGRvbXY3bDBvNmUyNWk3M3V0dm45aHRkdjByamprdXNyYmxqZQ==' \
	// --header 'content-type: application/x-www-form-urlencoded' \
	// --data grant_type=client_credentials

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://auth.overledger.dev/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	// Exact headers from curl
	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", "Basic M25ocXBzdDkzNXYwa3F1bWMzczc2amNxNDY6MTAybDBlYWJycXVtNHBhbGRvbXY3bDBvNmUyNWk3M3V0dm45aHRkdjByamprdXNyYmxqZQ==")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// prepareTransactionHardcoded follows the exact curl format for transaction preparation
func (c *Client) prepareTransactionHardcoded(bearerToken string, req *TransactionRequest) (*TransactionPrepareResponse, error) {
	// Exact curl format:
	// curl --request POST \
	// --url https://api.overledger.dev/v2/preparation/transaction \
	// --header 'accept: application/json' \
	// --header 'authorization: Bearer <bearer token here>' \
	// --header 'content-type: application/json' \
	// --data '{ "location": { "technology": "ethereum", "network": "ethereum sepolia testnet" }, "type": "PAYMENT", "urgency": "normal", "requestDetails": { "destination": [{ "destinationId": "0x6E32dA6eDbea6b4c794CD50c830753F9b134DEf0", "payment": { "amount": "0.001", "unit": "ETH" } }], "message": "OVL Transaction Message", "overledgerSigningType": "overledger-javascript-library", "origin": [{ "originId": "0x6E32dA6eDbea6b4c794CD50c830753F9b134DEf0" }] } }'

	// Hardcode the exact JSON structure from your curl example
	prepareBody := map[string]interface{}{
		"location": map[string]interface{}{
			"technology": "ethereum",
			"network":    "ethereum sepolia testnet",
		},
		"type":    "PAYMENT",
		"urgency": "normal",
		"requestDetails": map[string]interface{}{
			"destination": []map[string]interface{}{
				{
					"destinationId": req.ToAddress,
					"payment": map[string]interface{}{
						"amount": req.Amount,
						"unit":   "ETH",
					},
				},
			},
			"message":                "OVL Transaction Message",
			"overledgerSigningType": "overledger-javascript-library",
			"origin": []map[string]interface{}{
				{
					"originId": req.FromAddress,
				},
			},
		},
	}

	jsonData, err := json.Marshal(prepareBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prepare request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.overledger.dev/v2/preparation/transaction", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create prepare request: %w", err)
	}

	// Exact headers from curl
	httpReq.Header.Set("accept", "application/json")
	httpReq.Header.Set("authorization", "Bearer "+bearerToken)
	httpReq.Header.Set("content-type", "application/json")
	// IMPORTANT: Do NOT include API-Version on v2 endpoints to avoid ambiguity
	httpReq.Header.Del("API-Version")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("prepare request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read prepare response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("prepare failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var prepareResp TransactionPrepareResponse
	if err := json.Unmarshal(respBody, &prepareResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prepare response: %w", err)
	}

	// Attempt to normalize requestId in case the key casing/placement differs
	if strings.TrimSpace(prepareResp.PreparationTransactionSearchResponse.RequestID) == "" {
		var generic map[string]interface{}
		if err := json.Unmarshal(respBody, &generic); err == nil {
			if v, ok := generic["preparationTransactionSearchResponse"]; ok {
				if m, ok := v.(map[string]interface{}); ok {
					if id, ok := m["requestId"].(string); ok && id != "" {
						prepareResp.PreparationTransactionSearchResponse.RequestID = strings.Trim(id, "{} \"\n\t")
					}
				}
			}
			if strings.TrimSpace(prepareResp.PreparationTransactionSearchResponse.RequestID) == "" {
				if id, ok := generic["requestId"].(string); ok && id != "" {
					prepareResp.PreparationTransactionSearchResponse.RequestID = strings.Trim(id, "{} \"\n\t")
				}
			}
		}
	}

	// Debug: Print the preparation response to understand structure
	fmt.Printf("[DEBUG] Preparation Response: %s\n", string(respBody))

	return &prepareResp, nil
}

// signTransactionHardcoded follows the exact curl format for transaction signing
func (c *Client) signTransactionHardcoded(bearerToken string, prepareResp *TransactionPrepareResponse, req *TransactionRequest) (string, error) {
	// Exact curl format with ALL PLACEHOLDERS REPLACED:
	// curl --request POST \
	// --url https://api.overledger.dev/api/transaction-signing-sandbox \
	// --header 'accept: application/json' \
	// --header 'authorization: Bearer <bearer token here>' \
	// --header 'content-type: application/json' \
	// --data '{ "keyId": "<Sender Wallet Address>", "gatewayFee": { "amount": "<amount>", "unit": "ETH" }, "requestId": "<Request ID>", "dltFee": { "amount": "0.000019897764079968", "unit": "ETH" }, "transactionSigningResponderName": "CTA", "nativeData": { "chain": "testnet", "data": "000000004f564c2054657374204d657373616765", "chainId": 11155111, "gas": "21438", "maxPriorityFeePerGas": "347639493", "to": "<Destination Wallet Address>", "maxFeePerGas": "928153936", "nonce": 2, "hardfork": "london", "value": "20000000000000000" } }'

	// HARDCODED structure with placeholder replacement as per your specification:
	// Normalize requestId to match strict UUID format (remove braces/quotes/whitespace)
	normalizedRequestID := strings.Trim(prepareResp.PreparationTransactionSearchResponse.RequestID, "{} \"\n\t")

	// Compute value in wei from ETH input for EVM networks
	weiValue := "20000000000000000"
	if strings.TrimSpace(req.Amount) != "" {
		if isEvmNetworkID(req.NetworkID) {
			if v, err := ethToWeiString(req.Amount); err == nil {
				weiValue = v
			}
		} else {
			weiValue = req.Amount
		}
	}

	signBody := map[string]interface{}{
		"keyId": req.FromAddress, // <Sender Wallet Address>
		// gatewayFee in QNT per Flow; default 0
		"gatewayFee": map[string]interface{}{
			"amount": "0",
			"unit":   "QNT",
		},
		"requestId": normalizedRequestID, // <Request ID>
		"dltFee": map[string]interface{}{
			"amount": "0.000019897764079968", // HARDCODED as per your spec
			"unit":   "ETH",
		},
		"transactionSigningResponderName": "CTA", // HARDCODED as per your spec
		"nativeData": map[string]interface{}{
			"chain":                "testnet",                        // HARDCODED as per your spec
			"data":                 "000000004f564c2054657374204d657373616765", // HARDCODED as per your spec
			"chainId":              11155111,                         // HARDCODED as per your spec
			"gas":                  chooseOrDefault(req.GasLimit, "22086"),
			"maxPriorityFeePerGas": chooseOrDefault(req.MaxPriorityFeePerGas, "1500000"),
			"to":                   req.ToAddress,                    // <Destination Wallet Address>
			"maxFeePerGas":         chooseOrDefault(req.MaxFeePerGas, "9618390"),
			"nonce":                chooseNonce(req.Nonce, readAndMaybeInitNonce()),
			"hardfork":             "london",                        // HARDCODED as per your spec
			"value":                weiValue,
		},
	}

	jsonData, err := json.Marshal(signBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sign request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.overledger.dev/api/transaction-signing-sandbox", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create sign request: %w", err)
	}

	// Exact headers from curl
	httpReq.Header.Set("accept", "application/json")
	// Use a single Authorization header; Go canonicalizes keys, duplicates can conflict
	httpReq.Header.Set("Authorization", "Bearer "+bearerToken)
	httpReq.Header.Set("content-type", "application/json")
	// Required API version header ONLY for signing endpoint
	httpReq.Header.Set("API-Version", "3.0.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("sign request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read sign response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("sign failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var signResp SandboxSigningResponse
	if err := json.Unmarshal(respBody, &signResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal sign response: %w", err)
	}

	// Debug: Print the signing response
	fmt.Printf("[DEBUG] Signing Response: %s\n", string(respBody))

	if signResp.SignedTransaction != "" {
		return signResp.SignedTransaction, nil
	}
	if signResp.Signature != "" {
		return signResp.Signature, nil
	}

	return "", fmt.Errorf("signing response did not contain signed transaction or signature")
}

// chooseOrDefault returns provided value if non-empty, otherwise the default
func chooseOrDefault(value string, def string) string {
    if strings.TrimSpace(value) != "" {
        return value
    }
    return def
}

// chooseNonce returns pointed nonce value if provided, otherwise fallback
func chooseNonce(nonce *int, fallback int) int {
    if nonce != nil {
        return *nonce
    }
    return fallback
}

// nonce.txt management (very simple, project-root local file)
func nonceFilePath() string {
    return "nonce.txt"
}

func readAndMaybeInitNonce() int {
    // Try read
    b, err := os.ReadFile(nonceFilePath())
    if err == nil {
        s := strings.TrimSpace(string(b))
        if n, err := strconv.Atoi(s); err == nil && n >= 0 {
            return n
        }
    }
    // default to 4 and write it so next read works
    _ = os.WriteFile(nonceFilePath(), []byte("4"), 0644)
    return 4
}

func bumpNonceFile() {
    curr := readAndMaybeInitNonce()
    next := curr + 1
    _ = os.WriteFile(nonceFilePath(), []byte(strconv.Itoa(next)), 0644)
}

// ethToWeiString converts a decimal ETH string (e.g., "0.02") to a wei string using big.Int
func ethToWeiString(eth string) (string, error) {
    // Split on decimal point
    parts := strings.SplitN(strings.TrimSpace(eth), ".", 2)
    whole := parts[0]
    frac := ""
    if len(parts) == 2 {
        frac = parts[1]
    }
    // Pad fractional to 18 digits
    if len(frac) > 18 {
        frac = frac[:18]
    } else {
        frac = frac + strings.Repeat("0", 18-len(frac))
    }
    // Build big.Int from whole and frac
    wholeInt := new(big.Int)
    if whole == "" || whole == "+" || whole == "-" {
        whole = "0"
    }
    if _, ok := wholeInt.SetString(whole, 10); !ok {
        return "", fmt.Errorf("invalid whole part")
    }
    weiPerEth := new(big.Int)
    weiPerEth.Exp(big.NewInt(10), big.NewInt(18), nil)
    wholeWei := new(big.Int).Mul(wholeInt, weiPerEth)

    fracInt := new(big.Int)
    if frac != "" {
        if _, ok := fracInt.SetString(frac, 10); !ok {
            return "", fmt.Errorf("invalid fractional part")
        }
    }
    total := new(big.Int).Add(wholeWei, fracInt)
    return total.String(), nil
}

// isEvmNetworkID checks if the network looks like an EVM chain
func isEvmNetworkID(networkID string) bool {
    s := strings.ToLower(networkID)
    keys := []string{"eth", "ethereum", "sepolia", "holesky", "goerli", "polygon", "matic", "bsc", "binance", "arbitrum", "optimism", "base", "avalanche", "celo", "fantom", "linea", "zksync", "scroll"}
    for _, k := range keys {
        if strings.Contains(s, k) {
            return true
        }
    }
    return false
}

// executeTransactionHardcoded follows the exact curl format for transaction execution
func (c *Client) executeTransactionHardcoded(bearerToken, signedData, requestID string) (*TransactionExecuteResponse, error) {
	// Exact curl format:
	// curl --request POST \
	// --url https://api.overledger.dev/v2/execution/transaction \
	// --header 'accept: application/json' \
	// --header 'authorization: Bearer <bearer token here>' \
	// --header 'content-type: application/json' \
	// --data '{ "signed": "<The raw data after transaction signing>", "requestId": "<The ID assigned to a preparation>" }'

	execBody := map[string]interface{}{
		"signed":    signedData,
		"requestId": requestID,
	}

	jsonData, err := json.Marshal(execBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execute request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.overledger.dev/v2/execution/transaction", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create execute request: %w", err)
	}

	// Exact headers from curl
	httpReq.Header.Set("accept", "application/json")
	httpReq.Header.Set("authorization", "Bearer "+bearerToken)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read execute response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("execute failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var execResp TransactionExecuteResponse
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execute response: %w", err)
	}

	// Debug: Print the execution response
	fmt.Printf("[DEBUG] Execution Response: %s\n", string(respBody))

	return &execResp, nil
}

// convertToOverledgerFormat converts legacy TransactionRequest to Overledger format
func (c *Client) convertToOverledgerFormat(req *TransactionRequest) *TransactionPrepareRequest {
	technology, network := c.mapNetworkToLocation(req.NetworkID)

	// Set unit to native currency based on technology (proper case)
	unit := "ETH" // Default to ETH
	if strings.Contains(strings.ToLower(technology), "bitcoin") {
		unit = "BTC"
	} else if strings.Contains(strings.ToLower(technology), "xrp") {
		unit = "XRP"
	} else if strings.Contains(strings.ToLower(technology), "polygon") {
		unit = "MATIC"
	} else if strings.Contains(strings.ToLower(technology), "avalanche") {
		unit = "AVAX"
	}

	return &TransactionPrepareRequest{
		Location: Location{
			Technology: technology,
			Network:    network,
		},
		Type:    "PAYMENT",
		Urgency: "normal",
		RequestDetails: RequestDetails{
			Destination: []DestinationAccount{
				{
					DestinationID: req.ToAddress,
					Payment: Payment{
						Amount: req.Amount,
						Unit:   unit,
					},
				},
			},
			Message: "OVL Transaction Message",
			OverledgerSigningType: "overledger-javascript-library",
			Origin: []OriginAccount{
				{
					OriginID: req.FromAddress,
				},
			},
			// Remove nested message field
			Overrides: c.createOverrides(req),
		},
	}
}

// createOverrides creates transaction overrides from legacy request
func (c *Client) createOverrides(req *TransactionRequest) *TransactionOverrides {
	if req.GasLimit == "" && req.GasPrice == "" {
		return nil
	}

	return &TransactionOverrides{
		GasLimit: req.GasLimit,
		GasPrice: req.GasPrice,
	}
}

// GetAccountBalance retrieves the account balance for a given address on a specific network
func (c *Client) GetAccountBalance(networkID, address string) (*BalanceResponse, error) {
	var resp BalanceResponse
	if c.baseHasVersion() {
		endpoint := fmt.Sprintf("/networks/%s/accounts/%s/balance", networkID, address)
		if err := c.makeRequest("GET", endpoint, nil, &resp); err == nil {
			return &resp, nil
		}
		return nil, fmt.Errorf("failed to fetch account balance from known endpoints")
	}
	endpointV21 := fmt.Sprintf("/v2.1/networks/%s/accounts/%s/balance", networkID, address)
	if err := c.makeRequest("GET", endpointV21, nil, &resp); err == nil {
		return &resp, nil
	} else if !strings.Contains(err.Error(), "HTTP error: 404") {
		return nil, err
	}
	endpointV2 := fmt.Sprintf("/v2/networks/%s/accounts/%s/balance", networkID, address)
	if err := c.makeRequest("GET", endpointV2, nil, &resp); err == nil {
		return &resp, nil
	}
	return nil, fmt.Errorf("failed to fetch account balance from known endpoints")
}