package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite runs integration tests against the deployed service
type IntegrationTestSuite struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewIntegrationTestSuite() *IntegrationTestSuite {
	baseURL := os.Getenv("TEST_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	
	return &IntegrationTestSuite{
		baseURL: baseURL,
		apiKey:  os.Getenv("TEST_API_KEY"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// serverReachable quickly checks if the connector server is available; if not, tests depending on it will be skipped.
func serverReachable(baseURL string) bool {
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

// TestHealthEndpoint verifies the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	suite := NewIntegrationTestSuite()
	// Skip if no live connector server is available at baseURL
	if !serverReachable(suite.baseURL) {
		t.Skipf("Skipping live connector tests: server not running at %s", suite.baseURL)
	}
	
	resp, err := suite.get("/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)
	
	assert.Contains(t, health, "status")
	assert.Contains(t, health, "services")
	
	// Check individual service health
	services := health["services"].(map[string]interface{})
	
	// Log service statuses for debugging
	for service, status := range services {
		t.Logf("Service %s: %v", service, status)
	}
}

// TestCoinbaseIntegration tests Coinbase API endpoints
func TestCoinbaseIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite()
	
	// Skip if Coinbase credentials not configured
	if os.Getenv("COINBASE_API_KEY_ID") == "" {
		t.Skip("Skipping Coinbase tests - credentials not configured")
	}
	
	t.Run("GetAssets", func(t *testing.T) {
		resp, err := suite.get("/v1/coinbase/assets")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Contains(t, result, "data")
		}
	})
	
	t.Run("GetWallets", func(t *testing.T) {
		resp, err := suite.get("/v1/coinbase/wallets")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		// May return 404 if no wallets exist
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Contains(t, result, "data")
		}
	})
	
	t.Run("GetExchangeRates", func(t *testing.T) {
		resp, err := suite.get("/v1/coinbase/exchange-rates?currency=USD")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Contains(t, result, "data")
		}
	})
}

// TestOverledgerIntegration tests Overledger API endpoints
func TestOverledgerIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite()
	
	// Skip if Overledger credentials not configured
	if os.Getenv("OVERLEDGER_CLIENT_ID") == "" {
		t.Skip("Skipping Overledger tests - credentials not configured")
	}
	
	t.Run("TestConnection", func(t *testing.T) {
		resp, err := suite.get("/v1/overledger/test")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "connected", result["status"])
	})
	
	t.Run("GetNetworks", func(t *testing.T) {
		resp, err := suite.get("/v1/overledger/networks")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Contains(t, result, "networks")
		}
	})
}

// TestMeshProxyIntegration tests the Mesh API proxy
func TestMeshProxyIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite()
	// Skip if no live connector server is available at baseURL
	if !serverReachable(suite.baseURL) {
		t.Skipf("Skipping live connector tests: server not running at %s", suite.baseURL)
	}
	
	t.Run("NetworkList", func(t *testing.T) {
		payload := map[string]interface{}{}
		resp, err := suite.postJSON("/mesh/network/list", payload)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		// Log response for debugging
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Mesh network/list response: %s", string(body))
		
		// The proxy might return 502 if mesh server isn't running
		if resp.StatusCode != http.StatusBadGateway {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})
}

// TestExchangeIntegration tests Exchange API endpoints
func TestExchangeIntegration(t *testing.T) {
	suite := NewIntegrationTestSuite()
	
	// Skip if Exchange credentials not configured
	if os.Getenv("EXCHANGE_CREDENTIALS") == "" && os.Getenv("COINBASE_API_KEY") == "" {
		t.Skip("Skipping Exchange tests - credentials not configured")
	}
	
	t.Run("GetProducts", func(t *testing.T) {
		resp, err := suite.get("/v1/exchange/products")
		require.NoError(t, err)
		defer resp.Body.Close()
		
		// May return 500 if credentials are invalid
		if resp.StatusCode == http.StatusOK {
			var result interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
		}
	})
}

// TestAPIKeyAuthentication tests API key middleware
func TestAPIKeyAuthentication(t *testing.T) {
	suite := NewIntegrationTestSuite()
	
	// Only test if API key is configured
	if suite.apiKey == "" {
		t.Skip("Skipping API key tests - no key configured")
	}
	
	// Test without API key
	suite.apiKey = ""
	resp, err := suite.get("/v1/coinbase/assets")
	require.NoError(t, err)
	resp.Body.Close()
	
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	
	// Test with invalid API key
	suite.apiKey = "invalid-key"
	resp, err = suite.get("/v1/coinbase/assets")
	require.NoError(t, err)
	resp.Body.Close()
	
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	
	// Test with valid API key
	suite.apiKey = os.Getenv("TEST_API_KEY")
	resp, err = suite.get("/v1/coinbase/assets")
	require.NoError(t, err)
	resp.Body.Close()
	
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
}

// Helper methods

func (s *IntegrationTestSuite) get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	
	if s.apiKey != "" {
		req.Header.Set("X-API-Key", s.apiKey)
	}
	
	return s.httpClient.Do(req)
}

func (s *IntegrationTestSuite) postJSON(path string, payload interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", s.baseURL+path, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("X-API-Key", s.apiKey)
	}
	
	return s.httpClient.Do(req)
}

// TestFullIntegrationSuite runs all integration tests
func TestFullIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	t.Run("Health", TestHealthEndpoint)
	t.Run("Coinbase", TestCoinbaseIntegration)
	t.Run("Overledger", TestOverledgerIntegration)
	t.Run("Mesh", TestMeshProxyIntegration)
	t.Run("Exchange", TestExchangeIntegration)
	t.Run("Authentication", TestAPIKeyAuthentication)
}
