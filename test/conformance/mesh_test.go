package conformance

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Test against local server by default
	defaultTestURL = "http://localhost:8080"
	meshEndpoint   = "/mesh"
)

type MeshTestSuite struct {
	baseURL    string
	httpClient *http.Client
}

func NewMeshTestSuite(baseURL string) *MeshTestSuite {
	if baseURL == "" {
		baseURL = defaultTestURL
	}
	return &MeshTestSuite{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TestNetworkList tests the /network/list endpoint
func TestNetworkList(t *testing.T) {
	suite := NewMeshTestSuite("")

	resp, err := suite.post("/network/list", map[string]interface{}{})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Network list should return 200")

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify response structure
	networks, ok := result["network_identifiers"].([]interface{})
	assert.True(t, ok, "Response should contain network_identifiers array")
	assert.NotEmpty(t, networks, "Should have at least one network")

	// Check first network structure
	if len(networks) > 0 {
		network := networks[0].(map[string]interface{})
		assert.Contains(t, network, "blockchain")
		assert.Contains(t, network, "network")
	}
}

// TestNetworkStatus tests the /network/status endpoint
func TestNetworkStatus(t *testing.T) {
	suite := NewMeshTestSuite("")

	payload := map[string]interface{}{
		"network_identifier": map[string]string{
			"blockchain": "Coinbase",
			"network":    "Mainnet",
		},
	}

	resp, err := suite.post("/network/status", payload)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Network status should return 200")

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify required fields
	assert.Contains(t, result, "current_block_identifier")
	assert.Contains(t, result, "current_block_timestamp")
	assert.Contains(t, result, "genesis_block_identifier")
}

// TestNetworkOptions tests the /network/options endpoint
func TestNetworkOptions(t *testing.T) {
	suite := NewMeshTestSuite("")

	payload := map[string]interface{}{
		"network_identifier": map[string]string{
			"blockchain": "Coinbase",
			"network":    "Mainnet",
		},
	}

	resp, err := suite.post("/network/options", payload)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Network options should return 200")

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify required fields
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "allow")
}

// TestAccountBalance tests the /account/balance endpoint
func TestAccountBalance(t *testing.T) {
	suite := NewMeshTestSuite("")

	payload := map[string]interface{}{
		"network_identifier": map[string]string{
			"blockchain": "Coinbase",
			"network":    "Mainnet",
		},
		"account_identifier": map[string]string{
			"address": "0x1234567890abcdef1234567890abcdef12345678",
		},
	}

	resp, err := suite.post("/account/balance", payload)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Account balance may return 404 for non-existent accounts, which is valid
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result, "block_identifier")
		assert.Contains(t, result, "balances")
	}
}

// TestBlock tests the /block endpoint
func TestBlock(t *testing.T) {
	suite := NewMeshTestSuite("")

	payload := map[string]interface{}{
		"network_identifier": map[string]string{
			"blockchain": "Coinbase",
			"network":    "Mainnet",
		},
		"block_identifier": map[string]interface{}{
			"index": 1000000,
		},
	}

	resp, err := suite.post("/block", payload)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Verify block structure
		block, ok := result["block"].(map[string]interface{})
		assert.True(t, ok, "Response should contain block object")
		assert.Contains(t, block, "block_identifier")
		assert.Contains(t, block, "parent_block_identifier")
		assert.Contains(t, block, "timestamp")
	}
}

// TestMalformedRequests tests error handling for invalid requests
func TestMalformedRequests(t *testing.T) {
	suite := NewMeshTestSuite("")

	testCases := []struct {
		name     string
		endpoint string
		payload  interface{}
	}{
		{
			name:     "Empty network identifier",
			endpoint: "/network/status",
			payload:  map[string]interface{}{},
		},
		{
			name:     "Invalid JSON",
			endpoint: "/network/status",
			payload:  "not-json",
		},
		{
			name:     "Missing required fields",
			endpoint: "/account/balance",
			payload: map[string]interface{}{
				"network_identifier": map[string]string{
					"blockchain": "Coinbase",
				},
				// Missing account_identifier
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.post(tc.endpoint, tc.payload)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should return 4xx or 5xx for malformed requests
			assert.True(t, resp.StatusCode >= 400, "Should return error status for malformed request")
		})
	}
}

// Helper method to make POST requests
func (s *MeshTestSuite) post(endpoint string, payload interface{}) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}

	url := s.baseURL + meshEndpoint + endpoint
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return s.httpClient.Do(req)
}

// TestMeshAPIConformance runs all conformance tests
func TestMeshAPIConformance(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("NetworkList", TestNetworkList)
	t.Run("NetworkStatus", TestNetworkStatus)
	t.Run("NetworkOptions", TestNetworkOptions)
	t.Run("AccountBalance", TestAccountBalance)
	t.Run("Block", TestBlock)
	t.Run("MalformedRequests", TestMalformedRequests)
}
