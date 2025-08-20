package clients

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rotypes "github.com/coinbase/rosetta-sdk-go/types"
)

// TestNewMeshSDKClient tests the MeshSDKClient constructor
func TestNewMeshSDKClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "with custom base URL",
			baseURL:     "http://example.com:9090/mesh",
			expectedURL: "http://example.com:9090/mesh",
		},
		{
			name:        "with trailing slash",
			baseURL:     "http://example.com:9090/mesh/",
			expectedURL: "http://example.com:9090/mesh",
		},
		{
			name:        "empty base URL uses default",
			baseURL:     "",
			expectedURL: "http://localhost:8080/mesh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMeshSDKClient(tt.baseURL)
			assert.NotNil(t, client, "client should not be nil")
			assert.Equal(t, tt.expectedURL, client.baseURL, "base URL should match expected")
			assert.NotNil(t, client.apiClient, "API client should be initialized")
		})
	}
}

// mockRosettaServer creates a test server that mocks Rosetta endpoints
func mockRosettaServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.URL.Path {
		case "/network/list":
			response := &rotypes.NetworkListResponse{
				NetworkIdentifiers: []*rotypes.NetworkIdentifier{
					{Blockchain: "Ethereum", Network: "Sepolia"},
					{Blockchain: "Bitcoin", Network: "Testnet3"},
				},
			}
			json.NewEncoder(w).Encode(response)
			
		case "/network/status":
			var req rotypes.NetworkRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			response := &rotypes.NetworkStatusResponse{
				CurrentBlockIdentifier: &rotypes.BlockIdentifier{
					Index: 123456,
					Hash:  "0xabcdef1234567890",
				},
				CurrentBlockTimestamp: 1640000000000,
				GenesisBlockIdentifier: &rotypes.BlockIdentifier{
					Index: 0,
					Hash:  "0x0000000000000000",
				},
			}
			json.NewEncoder(w).Encode(response)
			
		case "/network/options":
			var req rotypes.NetworkRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			response := &rotypes.NetworkOptionsResponse{
				Version: &rotypes.Version{
					RosettaVersion: "1.4.0",
					NodeVersion:    "1.0.0",
				},
				Allow: &rotypes.Allow{
					OperationStatuses: []*rotypes.OperationStatus{
						{Status: "SUCCESS", Successful: true},
						{Status: "FAILURE", Successful: false},
					},
					OperationTypes: []string{"TRANSFER"},
				},
			}
			json.NewEncoder(w).Encode(response)
			
		case "/account/balance":
			var req rotypes.AccountBalanceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			response := &rotypes.AccountBalanceResponse{
				BlockIdentifier: &rotypes.BlockIdentifier{
					Index: 123456,
					Hash:  "0xabcdef1234567890",
				},
				Balances: []*rotypes.Amount{
					{
						Value: "1000000000000000000",
						Currency: &rotypes.Currency{
							Symbol:   "ETH",
							Decimals: 18,
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/block":
			var req rotypes.BlockRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			block := &rotypes.Block{
				BlockIdentifier: &rotypes.BlockIdentifier{Index: 123456, Hash: "0xblockhash"},
				ParentBlockIdentifier: &rotypes.BlockIdentifier{Index: 123455, Hash: "0xparenthash"},
				Timestamp: 1640000000000,
				Transactions: []*rotypes.Transaction{},
			}
			json.NewEncoder(w).Encode(&rotypes.BlockResponse{Block: block})

		case "/block/transaction":
			var req rotypes.BlockTransactionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			tx := &rotypes.Transaction{
				TransactionIdentifier: &rotypes.TransactionIdentifier{Hash: req.TransactionIdentifier.Hash},
				Operations:           []*rotypes.Operation{},
			}
			json.NewEncoder(w).Encode(&rotypes.BlockTransactionResponse{Transaction: tx})
			
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// TestMeshSDKClient_ListNetworks tests the ListNetworks method
func TestMeshSDKClient_ListNetworks(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)
	
	resp, err := client.ListNetworks()
	require.NoError(t, err, "ListNetworks should not return error")
	require.NotNil(t, resp, "response should not be nil")
	
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status code should be 200")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "content type should be JSON")
	
	// Decode and validate response structure
	var networkList rotypes.NetworkListResponse
	err = json.NewDecoder(resp.Body).Decode(&networkList)
	require.NoError(t, err, "should decode response body")
	defer resp.Body.Close()
	
	assert.Len(t, networkList.NetworkIdentifiers, 2, "should have 2 networks")
	assert.Equal(t, "Ethereum", networkList.NetworkIdentifiers[0].Blockchain)
	assert.Equal(t, "Sepolia", networkList.NetworkIdentifiers[0].Network)
	assert.Equal(t, "Bitcoin", networkList.NetworkIdentifiers[1].Blockchain)
	assert.Equal(t, "Testnet3", networkList.NetworkIdentifiers[1].Network)
}

// TestMeshSDKClient_NetworkStatus tests the NetworkStatus method
func TestMeshSDKClient_NetworkStatus(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)
	
	// Test with map input
	networkID := map[string]interface{}{
		"blockchain": "Ethereum",
		"network":    "Sepolia",
	}
	
	resp, err := client.NetworkStatus(networkID, nil)
	require.NoError(t, err, "NetworkStatus should not return error")
	require.NotNil(t, resp, "response should not be nil")
	
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status code should be 200")
	
	// Decode and validate response
	var status rotypes.NetworkStatusResponse
	err = json.NewDecoder(resp.Body).Decode(&status)
	require.NoError(t, err, "should decode response body")
	defer resp.Body.Close()
	
	assert.NotNil(t, status.CurrentBlockIdentifier, "should have current block identifier")
	assert.Equal(t, int64(123456), status.CurrentBlockIdentifier.Index)
	assert.Equal(t, "0xabcdef1234567890", status.CurrentBlockIdentifier.Hash)
}

// TestMeshSDKClient_NetworkOptions tests the NetworkOptions method
func TestMeshSDKClient_NetworkOptions(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)
	
	// Test with NetworkIdentifier struct
	networkID := &rotypes.NetworkIdentifier{
		Blockchain: "Ethereum",
		Network:    "Sepolia",
	}
	
	resp, err := client.NetworkOptions(networkID)
	require.NoError(t, err, "NetworkOptions should not return error")
	require.NotNil(t, resp, "response should not be nil")
	
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status code should be 200")
	
	// Decode and validate response
	var options rotypes.NetworkOptionsResponse
	err = json.NewDecoder(resp.Body).Decode(&options)
	require.NoError(t, err, "should decode response body")
	defer resp.Body.Close()
	
	assert.NotNil(t, options.Version, "should have version info")
	assert.Equal(t, "1.4.0", options.Version.RosettaVersion)
	assert.NotNil(t, options.Allow, "should have allow info")
	assert.Contains(t, options.Allow.OperationTypes, "TRANSFER")
}

// TestMeshSDKClient_AccountBalance tests the AccountBalance method
func TestMeshSDKClient_AccountBalance(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)
	
	networkID := map[string]interface{}{
		"blockchain": "Ethereum",
		"network":    "Sepolia",
	}
	accountID := map[string]interface{}{
		"address": "0x1234567890abcdef1234567890abcdef12345678",
	}
	
	resp, err := client.AccountBalance(networkID, accountID)
	require.NoError(t, err, "AccountBalance should not return error")
	require.NotNil(t, resp, "response should not be nil")
	
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status code should be 200")
	
	// Decode and validate response
	var balance rotypes.AccountBalanceResponse
	err = json.NewDecoder(resp.Body).Decode(&balance)
	require.NoError(t, err, "should decode response body")
	defer resp.Body.Close()
	
	assert.NotNil(t, balance.BlockIdentifier, "should have block identifier")
	assert.Len(t, balance.Balances, 1, "should have 1 balance")
	assert.Equal(t, "1000000000000000000", balance.Balances[0].Value)
	assert.Equal(t, "ETH", balance.Balances[0].Currency.Symbol)
	assert.Equal(t, int32(18), balance.Balances[0].Currency.Decimals)
}

// New: TestMeshSDKClient_Block tests the Block method
func TestMeshSDKClient_Block(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)

	networkID := map[string]interface{}{
		"blockchain": "Ethereum",
		"network":    "Sepolia",
	}
	blockID := map[string]interface{}{
		"index": 123456,
	}

	resp, err := client.Block(networkID, blockID)
	require.NoError(t, err, "Block should not return error")
	require.NotNil(t, resp, "response should not be nil")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var blockResp rotypes.BlockResponse
	err = json.NewDecoder(resp.Body).Decode(&blockResp)
	require.NoError(t, err)
	defer resp.Body.Close()

	if assert.NotNil(t, blockResp.Block) {
		assert.Equal(t, int64(123456), blockResp.Block.BlockIdentifier.Index)
		assert.Equal(t, "0xblockhash", blockResp.Block.BlockIdentifier.Hash)
		assert.Equal(t, int64(1640000000000), blockResp.Block.Timestamp)
	}
}

// New: TestMeshSDKClient_BlockTransaction tests the BlockTransaction method
func TestMeshSDKClient_BlockTransaction(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)

	networkID := map[string]interface{}{
		"blockchain": "Ethereum",
		"network":    "Sepolia",
	}
	blockID := map[string]interface{}{
		"index": 123456,
	}
	transactionID := map[string]interface{}{
		"hash": "0xtxhash",
	}

	resp, err := client.BlockTransaction(networkID, blockID, transactionID)
	require.NoError(t, err, "BlockTransaction should not return error")
	require.NotNil(t, resp, "response should not be nil")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var txResp rotypes.BlockTransactionResponse
	err = json.NewDecoder(resp.Body).Decode(&txResp)
	require.NoError(t, err)
	defer resp.Body.Close()

	if assert.NotNil(t, txResp.Transaction) {
		assert.Equal(t, "0xtxhash", txResp.Transaction.TransactionIdentifier.Hash)
	}
}

// TestMeshSDKClient_Health tests the Health method
func TestMeshSDKClient_Health(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)
	
	health := client.Health()
	assert.True(t, health, "health check should pass when ListNetworks succeeds")
}

// TestMeshSDKClient_Health_Failure tests the Health method when server is down
func TestMeshSDKClient_Health_Failure(t *testing.T) {
	// Use a non-existent server URL
	client := NewMeshSDKClient("http://localhost:99999")
	
	health := client.Health()
	assert.False(t, health, "health check should fail when server is unreachable")
}

// TestToNetworkIdentifier tests the toNetworkIdentifier helper function
func TestToNetworkIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		expected    *rotypes.NetworkIdentifier
	}{
		{
			name: "map input",
			input: map[string]interface{}{
				"blockchain": "Ethereum",
				"network":    "Sepolia",
			},
			expectError: false,
			expected: &rotypes.NetworkIdentifier{
				Blockchain: "Ethereum",
				Network:    "Sepolia",
			},
		},
		{
			name: "NetworkIdentifier pointer",
			input: &rotypes.NetworkIdentifier{
				Blockchain: "Bitcoin",
				Network:    "Mainnet",
			},
			expectError: false,
			expected: &rotypes.NetworkIdentifier{
				Blockchain: "Bitcoin",
				Network:    "Mainnet",
			},
		},
		{
			name: "NetworkIdentifier struct",
			input: rotypes.NetworkIdentifier{
				Blockchain: "Ethereum",
				Network:    "Mainnet",
			},
			expectError: false,
			expected: &rotypes.NetworkIdentifier{
				Blockchain: "Ethereum",
				Network:    "Mainnet",
			},
		},
		{
			name: "invalid map - missing blockchain",
			input: map[string]interface{}{
				"network": "Sepolia",
			},
			expectError: true,
		},
		{
			name: "invalid map - missing network",
			input: map[string]interface{}{
				"blockchain": "Ethereum",
			},
			expectError: true,
		},
		{
			name:        "unsupported type",
			input:       "invalid string",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toNetworkIdentifier(tt.input)
			
			if tt.expectError {
				assert.Error(t, err, "should return error for invalid input")
				assert.Nil(t, result, "result should be nil on error")
			} else {
				assert.NoError(t, err, "should not return error for valid input")
				assert.Equal(t, tt.expected, result, "result should match expected")
			}
		})
	}
}

// TestToAccountIdentifier tests the toAccountIdentifier helper function
func TestToAccountIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		expected    *rotypes.AccountIdentifier
	}{
		{
			name: "map input",
			input: map[string]interface{}{
				"address": "0x1234567890abcdef1234567890abcdef12345678",
			},
			expectError: false,
			expected: &rotypes.AccountIdentifier{
				Address: "0x1234567890abcdef1234567890abcdef12345678",
			},
		},
		{
			name: "AccountIdentifier pointer",
			input: &rotypes.AccountIdentifier{
				Address: "0xabcdef1234567890abcdef1234567890abcdef12",
			},
			expectError: false,
			expected: &rotypes.AccountIdentifier{
				Address: "0xabcdef1234567890abcdef1234567890abcdef12",
			},
		},
		{
			name: "AccountIdentifier struct",
			input: rotypes.AccountIdentifier{
				Address: "0xfedcba0987654321fedcba0987654321fedcba09",
			},
			expectError: false,
			expected: &rotypes.AccountIdentifier{
				Address: "0xfedcba0987654321fedcba0987654321fedcba09",
			},
		},
		{
			name: "invalid map - missing address",
			input: map[string]interface{}{
				"name": "test account",
			},
			expectError: true,
		},
		{
			name:        "unsupported type",
			input:       123,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toAccountIdentifier(tt.input)
			
			if tt.expectError {
				assert.Error(t, err, "should return error for invalid input")
				assert.Nil(t, result, "result should be nil on error")
			} else {
				assert.NoError(t, err, "should not return error for valid input")
				assert.Equal(t, tt.expected, result, "result should match expected")
			}
		})
	}
}

// TestWrapJSONResponse tests the wrapJSONResponse helper function
func TestWrapJSONResponse(t *testing.T) {
	testData := map[string]interface{}{
		"message": "test response",
		"status":  "success",
		"count":   42,
	}
	
	resp, err := wrapJSONResponse(testData)
	require.NoError(t, err, "should not return error for valid data")
	require.NotNil(t, resp, "response should not be nil")
	
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status code should be 200")
	assert.Equal(t, "200 OK", resp.Status, "status should be 200 OK")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "content type should be JSON")
	
	// Read and parse body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "should read body without error")
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "should unmarshal JSON body")
	
	// JSON numbers decode to float64 by default when unmarshaling into map[string]interface{}
	// so compare fields individually instead of whole-map equality.
	assert.Equal(t, testData["message"], result["message"], "message should match")
	assert.Equal(t, testData["status"], result["status"], "status should match")
	if cv, ok := result["count"]; ok {
	    switch v := cv.(type) {
	    case float64:
	        assert.Equal(t, float64(42), v, "count should match (float64 JSON number)")
	    case json.Number:
	        n, convErr := v.Int64()
	        require.NoError(t, convErr, "count should be convertible to int64")
	        assert.Equal(t, int64(42), n, "count should match (json.Number)")
	    default:
	        t.Fatalf("unexpected type for count: %T (%v)", v, v)
	    }
	} else {
	    t.Fatalf("count field missing in response")
	}
}

// TestWrapJSONResponse_Error tests wrapJSONResponse with invalid data
func TestWrapJSONResponse_Error(t *testing.T) {
	// Use a channel which cannot be marshaled to JSON
	invalidData := make(chan int)
	
	resp, err := wrapJSONResponse(invalidData)
	assert.Error(t, err, "should return error for invalid data")
	assert.Nil(t, resp, "response should be nil on error")
	assert.Contains(t, err.Error(), "failed to marshal response", "error should mention marshaling failure")
}

// TestMeshSDKClient_Integration tests the SDK client against server errors
func TestMeshSDKClient_Integration_Errors(t *testing.T) {
	// Create a server that returns errors
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer errorServer.Close()

	client := NewMeshSDKClient(errorServer.URL)
	
	// Test that methods handle server errors gracefully
	t.Run("ListNetworks error handling", func(t *testing.T) {
		resp, err := client.ListNetworks()
		assert.Error(t, err, "should return error when server returns 500")
		assert.Nil(t, resp, "response should be nil on error")
	})
	
	t.Run("NetworkStatus error handling", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.NetworkStatus(networkID, nil)
		assert.Error(t, err, "should return error when server returns 500")
		assert.Nil(t, resp, "response should be nil on error")
	})
	
	t.Run("NetworkOptions error handling", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.NetworkOptions(networkID)
		assert.Error(t, err, "should return error when server returns 500")
		assert.Nil(t, resp, "response should be nil on error")
	})
	
	t.Run("AccountBalance error handling", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		accountID := map[string]interface{}{"address": "0x1234567890abcdef1234567890abcdef12345678"}
		resp, err := client.AccountBalance(networkID, accountID)
		assert.Error(t, err, "should return error when server returns 500")
		assert.Nil(t, resp, "response should be nil on error")
	})

	// New error handling tests
	t.Run("Block error handling", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		blockID := map[string]interface{}{
			"index": 1,
		}
		resp, err := client.Block(networkID, blockID)
		assert.Error(t, err, "should return error when server returns 500")
		assert.Nil(t, resp, "response should be nil on error")
	})

	t.Run("BlockTransaction error handling", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		blockID := map[string]interface{}{
			"index": 1,
		}
		txID := map[string]interface{}{
			"hash": "0xabc",
		}
		resp, err := client.BlockTransaction(networkID, blockID, txID)
		assert.Error(t, err, "should return error when server returns 500")
		assert.Nil(t, resp, "response should be nil on error")
	})
	
	t.Run("Health check with server errors", func(t *testing.T) {
		health := client.Health()
		assert.False(t, health, "health check should fail when server returns errors")
	})
}

// TestMeshSDKClient_TypeConversionErrors tests error handling in type conversion
func TestMeshSDKClient_TypeConversionErrors(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)
	
	t.Run("NetworkStatus with invalid network identifier", func(t *testing.T) {
		resp, err := client.NetworkStatus("invalid", nil)
		assert.Error(t, err, "should return error for invalid network identifier")
		assert.Nil(t, resp, "response should be nil on error")
		assert.Contains(t, strings.ToLower(err.Error()), "unsupported network_identifier type")
	})
	
	t.Run("NetworkOptions with invalid network identifier", func(t *testing.T) {
		resp, err := client.NetworkOptions(123)
		assert.Error(t, err, "should return error for invalid network identifier")
		assert.Nil(t, resp, "response should be nil on error")
	})
	
	t.Run("AccountBalance with invalid network identifier", func(t *testing.T) {
		accountID := map[string]interface{}{
			"address": "0x1234567890abcdef1234567890abcdef12345678",
		}
		resp, err := client.AccountBalance([]int{1, 2, 3}, accountID)
		assert.Error(t, err, "should return error for invalid network identifier")
		assert.Nil(t, resp, "response should be nil on error")
	})
	
	t.Run("AccountBalance with invalid account identifier", func(t *testing.T) {
		networkID := map[string]interface{}{
			"blockchain": "Ethereum",
			"network":    "Sepolia",
		}
		resp, err := client.AccountBalance(networkID, "invalid")
		assert.Error(t, err, "should return error for invalid account identifier")
		assert.Nil(t, resp, "response should be nil on error")
		assert.Contains(t, strings.ToLower(err.Error()), "unsupported account_identifier type")
	})

	// New: Block invalid conversions
	t.Run("Block with invalid network identifier", func(t *testing.T) {
		resp, err := client.Block("invalid", map[string]interface{}{"index": 1})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Block with invalid block identifier type", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.Block(networkID, true)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, strings.ToLower(err.Error()), "unsupported block_identifier type")
	})

	// New: BlockTransaction invalid conversions
	t.Run("BlockTransaction with invalid network identifier", func(t *testing.T) {
		resp, err := client.BlockTransaction("invalid", map[string]interface{}{"index": 1}, map[string]interface{}{"hash": "0xabc"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("BlockTransaction with invalid block identifier type", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.BlockTransaction(networkID, []byte("invalid"), map[string]interface{}{"hash": "0xabc"})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, strings.ToLower(err.Error()), "unsupported block_identifier type")
	})

	t.Run("BlockTransaction with invalid transaction identifier type", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.BlockTransaction(networkID, map[string]interface{}{"index": 1}, 123)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, strings.ToLower(err.Error()), "unsupported transaction_identifier type")
	})

	t.Run("BlockTransaction with empty block identifier map", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.BlockTransaction(networkID, map[string]interface{}{}, map[string]interface{}{"hash": "0xabc"})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid block_identifier map")
	})

	t.Run("BlockTransaction with empty transaction identifier map", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.BlockTransaction(networkID, map[string]interface{}{"index": 1}, map[string]interface{}{})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid transaction_identifier map")
	})

	// New: BlockTransaction with invalid block identifier map value types (string index, numeric hash)
	t.Run("BlockTransaction with invalid block identifier map types", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		invalidBlockID := map[string]interface{}{"index": "1", "hash": 123}
		txID := map[string]interface{}{"hash": "0xabc"}
		resp, err := client.BlockTransaction(networkID, invalidBlockID, txID)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, strings.ToLower(err.Error()), "invalid block_identifier map")
	})
}

// TestToPartialBlockIdentifier tests the toPartialBlockIdentifier helper function
func TestToPartialBlockIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		expected    *rotypes.PartialBlockIdentifier
	}{
		{
			name:        "nil input",
			input:       nil,
			expectError: false,
			expected:    &rotypes.PartialBlockIdentifier{},
		},
		{
			name: "PartialBlockIdentifier pointer",
			input: &rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(100); return &v }(),
				Hash:  func() *string { v := "0xhash123"; return &v }(),
			},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(100); return &v }(),
				Hash:  func() *string { v := "0xhash123"; return &v }(),
			},
		},
		{
			name: "PartialBlockIdentifier struct",
			input: rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(200); return &v }(),
			},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(200); return &v }(),
			},
		},
		{
			name: "map with index (int)",
			input: map[string]interface{}{
				"index": 42,
			},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(42); return &v }(),
			},
		},
		{
			name: "map with index (int64)",
			input: map[string]interface{}{
				"index": int64(999),
			},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(999); return &v }(),
			},
		},
		{
			name: "map with index (float64)",
			input: map[string]interface{}{
				"index": 123.0,
			},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(123); return &v }(),
			},
		},
		{
			name: "map with hash",
			input: map[string]interface{}{
				"hash": "0xabcdef",
			},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{
				Hash: func() *string { v := "0xabcdef"; return &v }(),
			},
		},
		{
			name: "map with both index and hash",
			input: map[string]interface{}{
				"index": 55,
				"hash":  "0x123456",
			},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{
				Index: func() *int64 { v := int64(55); return &v }(),
				Hash:  func() *string { v := "0x123456"; return &v }(),
			},
		},
		{
			name: "empty map",
			input: map[string]interface{}{},
			expectError: false,
			expected: &rotypes.PartialBlockIdentifier{},
		},
		{
			name:        "unsupported type",
			input:       "invalid string",
			expectError: true,
		},
		{
			name:        "slice type",
			input:       []int{1, 2, 3},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toPartialBlockIdentifier(tt.input)

			if tt.expectError {
				assert.Error(t, err, "should return error for invalid input")
				assert.Nil(t, result, "result should be nil on error")
			} else {
				assert.NoError(t, err, "should not return error for valid input")
				if tt.expected.Index != nil && result.Index != nil {
					assert.Equal(t, *tt.expected.Index, *result.Index, "index should match")
				} else {
					assert.Equal(t, tt.expected.Index, result.Index, "index pointer should match")
				}
				if tt.expected.Hash != nil && result.Hash != nil {
					assert.Equal(t, *tt.expected.Hash, *result.Hash, "hash should match")
				} else {
					assert.Equal(t, tt.expected.Hash, result.Hash, "hash pointer should match")
				}
			}
		})
	}
}

// TestToFullBlockIdentifier tests the toFullBlockIdentifier helper function
func TestToFullBlockIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		expected    *rotypes.BlockIdentifier
	}{
		{
			name: "BlockIdentifier pointer",
			input: &rotypes.BlockIdentifier{
				Index: 123,
				Hash:  "0xhash456",
			},
			expectError: false,
			expected: &rotypes.BlockIdentifier{
				Index: 123,
				Hash:  "0xhash456",
			},
		},
		{
			name: "BlockIdentifier struct",
			input: rotypes.BlockIdentifier{
				Index: 456,
				Hash:  "0xhash789",
			},
			expectError: false,
			expected: &rotypes.BlockIdentifier{
				Index: 456,
				Hash:  "0xhash789",
			},
		},
		{
			name: "map with index (int)",
			input: map[string]interface{}{
				"index": 789,
			},
			expectError: false,
			expected: &rotypes.BlockIdentifier{
				Index: 789,
				Hash:  "",
			},
		},
		{
			name: "map with index (int64)",
			input: map[string]interface{}{
				"index": int64(1000),
			},
			expectError: false,
			expected: &rotypes.BlockIdentifier{
				Index: 1000,
				Hash:  "",
			},
		},
		{
			name: "map with index (float64)",
			input: map[string]interface{}{
				"index": 2000.0,
			},
			expectError: false,
			expected: &rotypes.BlockIdentifier{
				Index: 2000,
				Hash:  "",
			},
		},
		{
			name: "map with hash only",
			input: map[string]interface{}{
				"hash": "0xonlyhash",
			},
			expectError: false,
			expected: &rotypes.BlockIdentifier{
				Index: 0,
				Hash:  "0xonlyhash",
			},
		},
		{
			name: "map with both index and hash",
			input: map[string]interface{}{
				"index": 333,
				"hash":  "0xbothhash",
			},
			expectError: false,
			expected: &rotypes.BlockIdentifier{
				Index: 333,
				Hash:  "0xbothhash",
			},
		},
		{
			name: "empty map - should error",
			input: map[string]interface{}{},
			expectError: true,
		},
		{
			name:        "unsupported type",
			input:       12.5,
			expectError: true,
		},
		{
			name:        "channel type",
			input:       make(chan int),
			expectError: true,
		},
		// New error-path cases
		{
			name:        "map with invalid index type (string) - should error",
			input:       map[string]interface{}{"index": "100"},
			expectError: true,
		},
		{
			name:        "map with invalid hash type (number) - should error",
			input:       map[string]interface{}{"hash": 12345},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toFullBlockIdentifier(tt.input)

			if tt.expectError {
				assert.Error(t, err, "should return error for invalid input")
				assert.Nil(t, result, "result should be nil on error")
			} else {
				assert.NoError(t, err, "should not return error for valid input")
				assert.Equal(t, tt.expected, result, "result should match expected")
			}
		})
	}
}

// TestToTransactionIdentifier tests the toTransactionIdentifier helper function
func TestToTransactionIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		expected    *rotypes.TransactionIdentifier
	}{
		{
			name: "TransactionIdentifier pointer",
			input: &rotypes.TransactionIdentifier{
				Hash: "0xtxhash123",
			},
			expectError: false,
			expected: &rotypes.TransactionIdentifier{
				Hash: "0xtxhash123",
			},
		},
		{
			name: "TransactionIdentifier struct",
			input: rotypes.TransactionIdentifier{
				Hash: "0xtxhash456",
			},
			expectError: false,
			expected: &rotypes.TransactionIdentifier{
				Hash: "0xtxhash456",
			},
		},
		{
			name: "map with hash",
			input: map[string]interface{}{
				"hash": "0xmaphash",
			},
			expectError: false,
			expected: &rotypes.TransactionIdentifier{
				Hash: "0xmaphash",
			},
		},
		{
			name: "invalid map - missing hash",
			input: map[string]interface{}{
				"id": "not_hash",
			},
			expectError: true,
		},
		{
			name: "invalid map - empty hash",
			input: map[string]interface{}{
				"hash": "",
			},
			expectError: true,
		},
		{
			name:        "unsupported type",
			input:       []string{"hash1", "hash2"},
			expectError: true,
		},
		{
			name:        "nil input",
			input:       nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toTransactionIdentifier(tt.input)

			if tt.expectError {
				assert.Error(t, err, "should return error for invalid input")
				assert.Nil(t, result, "result should be nil on error")
			} else {
				assert.NoError(t, err, "should not return error for valid input")
				assert.Equal(t, tt.expected, result, "result should match expected")
			}
		})
	}
}

// TestMeshSDKClient_SDKErrors tests error handling when SDK calls fail
func TestMeshSDKClient_SDKErrors(t *testing.T) {
	// Use a client with an invalid base URL to force SDK errors
	client := NewMeshSDKClient("http://invalid-host:99999/mesh")

	t.Run("ListNetworks SDK error", func(t *testing.T) {
		resp, err := client.ListNetworks()
		assert.Error(t, err, "should return error when SDK fails")
		assert.Nil(t, resp, "response should be nil on error")
	})

	t.Run("NetworkStatus SDK error", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.NetworkStatus(networkID, nil)
		assert.Error(t, err, "should return error when SDK fails")
		assert.Nil(t, resp, "response should be nil on error")
	})

	t.Run("NetworkOptions SDK error", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.NetworkOptions(networkID)
		assert.Error(t, err, "should return error when SDK fails")
		assert.Nil(t, resp, "response should be nil on error")
	})

	t.Run("AccountBalance SDK error", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		accountID := map[string]interface{}{"address": "0x1234567890abcdef1234567890abcdef12345678"}
		resp, err := client.AccountBalance(networkID, accountID)
		assert.Error(t, err, "should return error when SDK fails")
		assert.Nil(t, resp, "response should be nil on error")
	})

	t.Run("Block SDK error", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		blockID := map[string]interface{}{"index": 1}
		resp, err := client.Block(networkID, blockID)
		assert.Error(t, err, "should return error when SDK fails")
		assert.Nil(t, resp, "response should be nil on error")
	})

	t.Run("BlockTransaction SDK error", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		blockID := map[string]interface{}{"index": 1}
		txID := map[string]interface{}{"hash": "0xabc"}
		resp, err := client.BlockTransaction(networkID, blockID, txID)
		assert.Error(t, err, "should return error when SDK fails")
		assert.Nil(t, resp, "response should be nil on error")
	})
}

// TestMeshSDKClient_EdgeCases tests additional edge cases
func TestMeshSDKClient_EdgeCases(t *testing.T) {
	server := mockRosettaServer()
	defer server.Close()

	client := NewMeshSDKClient(server.URL)

	t.Run("Block with nil block identifier", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		resp, err := client.Block(networkID, nil)
		require.NoError(t, err, "Block should handle nil block identifier")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AccountBalance with SubAccount in AccountIdentifier", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		accountID := &rotypes.AccountIdentifier{
			Address: "0x1234567890abcdef1234567890abcdef12345678",
			SubAccount: &rotypes.SubAccountIdentifier{
				Address: "sub-account",
			},
		}
		resp, err := client.AccountBalance(networkID, accountID)
		require.NoError(t, err, "AccountBalance should handle SubAccount")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NetworkStatus with blockIdentifier parameter ignored", func(t *testing.T) {
		networkID := map[string]interface{}{"blockchain": "Ethereum", "network": "Sepolia"}
		// blockIdentifier is ignored in NetworkStatus per comment in method
		blockID := map[string]interface{}{"index": 123}
		resp, err := client.NetworkStatus(networkID, blockID)
		require.NoError(t, err, "NetworkStatus should ignore blockIdentifier")
		require.NotNil(t, resp, "response should not be nil")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}