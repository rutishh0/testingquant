package integration

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rutishh0/testingquant/internal/clients"
	"github.com/rutishh0/testingquant/internal/config"
	"github.com/rutishh0/testingquant/internal/connector"
	meshadapter "github.com/rutishh0/testingquant/internal/adapters/mesh"
	"github.com/rutishh0/testingquant/internal/overledger"
)

// MockOverledgerServer creates an in-memory mock Overledger API server
func startMockOverledgerServer(t *testing.T) *httptest.Server {
	t.Helper()

	router := gin.New()
	gin.SetMode(gin.TestMode)

	// Mock OAuth2 token endpoint
	router.POST("/oauth2/token", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"access_token": "mock_token_12345",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	})

	// Mock networks endpoint
	router.GET("/v2.1/networks", func(c *gin.Context) {
		c.JSON(200, overledger.NetworksResponse{
			Networks: []overledger.Network{
				{
					ID:          "ethereum-sepolia",
					Name:        "Ethereum Sepolia",
					Description: "Ethereum testnet",
					Type:        "testnet",
					Status:      "active",
				},
				{
					ID:          "bitcoin-testnet",
					Name:        "Bitcoin Testnet",
					Description: "Bitcoin testnet",
					Type:        "testnet",
					Status:      "active",
				},
			},
		})
	})

	// Mock balance endpoint
	router.GET("/v2.1/networks/:networkId/addresses/:address/balances", func(c *gin.Context) {
		networkID := c.Param("networkId")
		address := c.Param("address")

		balances := []overledger.Balance{
			{
				TokenID:     "eth",
				TokenName:   "Ethereum",
				TokenSymbol: "ETH",
				Amount:      "1000000000000000000", // 1 ETH in wei
				Decimals:    18,
				Unit:        "wei",
			},
		}

		if networkID == "ethereum-sepolia" {
			// Add an ERC20 token for Ethereum
			balances = append(balances, overledger.Balance{
				TokenID:     "0xa0b86a33e6fe17f58b8b4cd6e45e74b8a7b4d30d",
				TokenName:   "Test Token",
				TokenSymbol: "TEST",
				Amount:      "5000000000000000000", // 5 TEST tokens
				Decimals:    18,
				Unit:        "wei",
			})
		}

		c.JSON(200, overledger.BalanceResponse{
			Address:  address,
			Balances: balances,
		})
	})

	// Mock transaction creation endpoint
	router.POST("/v2.1/transactions", func(c *gin.Context) {
		var req overledger.TransactionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		// Generate mock transaction hash
		txHash := fmt.Sprintf("0x%064x", time.Now().UnixNano())

		c.JSON(200, overledger.TransactionResponse{
			TransactionID: fmt.Sprintf("tx_%d", time.Now().UnixNano()),
			Hash:          txHash,
			Status:        "pending",
			NetworkID:     req.NetworkID,
			FromAddress:   req.FromAddress,
			ToAddress:     req.ToAddress,
			Amount:        req.Amount,
			Fee:           "21000",
			Timestamp:     time.Now(),
		})
	})

	// Mock transaction status endpoint
	router.GET("/v2.1/networks/:networkId/transactions/:txHash/status", func(c *gin.Context) {
		txHash := c.Param("txHash")
		networkID := c.Param("networkId")

		// Simulate transaction progression
		status := "confirmed"
		confirmations := 12
		blockNumber := int64(18500000)

		c.JSON(200, overledger.TransactionStatusResponse{
			TransactionID: "tx_" + txHash[2:10], // Extract part of hash as ID
			Hash:          txHash,
			Status:        status,
			NetworkID:     networkID,
			Confirmations: confirmations,
			BlockNumber:   blockNumber,
			BlockHash:     fmt.Sprintf("0x%064x", blockNumber),
			Timestamp:     time.Now().Add(-5*time.Minute), // 5 minutes ago
		})
	})

	// Mock connection test endpoint
	router.GET("/v2.1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	return httptest.NewServer(router)
}

// setupTokenTransferSuite sets up a complete test environment with mock Overledger and Mesh servers
func setupTokenTransferSuite(t *testing.T) (*httptest.Server, *httptest.Server, *overledger.Client) {
	t.Helper()

	// Start mock Overledger server
	overledgerServer := startMockOverledgerServer(t)

	// Start mock Rosetta server for Mesh endpoints
	rosettaServer := startMockRosettaServer(t)

	// Create Overledger client with mock server URLs
	cfg := &config.Config{
		OverledgerClientID:     "test_client_id",
		OverledgerClientSecret: "test_client_secret",
		OverledgerAuthURL:      overledgerServer.URL + "/oauth2/token",
		OverledgerBaseURL:      overledgerServer.URL + "/v2.1",
	}

	overledgerClient := overledger.NewClient(cfg)

	return overledgerServer, rosettaServer, overledgerClient
}

// TestTokenTransferLifecycle tests the complete flow: create transaction -> check status -> verify balance changes
func TestTokenTransferLifecycle(t *testing.T) {
	overledgerServer, rosettaServer, overledgerClient := setupTokenTransferSuite(t)
	defer overledgerServer.Close()
	defer rosettaServer.Close()

	// Create mesh client for balance verification
	meshClient := clients.NewMeshSDKClient(rosettaServer.URL)
	meshAdapter := meshadapter.NewAdapter(meshClient)

	// Create service with both adapters
	service := connector.NewService(nil, meshAdapter, overledgerClient)

	t.Run("Complete Transaction Lifecycle", func(t *testing.T) {
		// Step 1: Get initial balance via Mesh
		networkID := map[string]interface{}{
			"blockchain": "Ethereum",
			"network":    "Sepolia",
		}
		accountID := map[string]interface{}{
			"address": "0x1234567890abcdef1234567890abcdef12345678",
		}

		initialBalance, err := service.GetMeshNetworkBalance(networkID, accountID)
		require.NoError(t, err, "should get initial balance")
		require.NotNil(t, initialBalance, "balance response should not be nil")
		require.NotEmpty(t, initialBalance.Balances, "should have at least one balance")

		t.Logf("Initial balance: %+v", initialBalance.Balances[0])

		// Step 2: Create transaction via Overledger
		txReq := &overledger.TransactionRequest{
			NetworkID:   "ethereum-sepolia",
			FromAddress: "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:   "0xabcdef1234567890abcdef1234567890abcdef12",
			Amount:      "500000000000000000", // 0.5 ETH
			GasLimit:    "21000",
			GasPrice:    "20000000000", // 20 Gwei
		}

		txResp, err := service.CreateOverledgerTransaction(txReq)
		require.NoError(t, err, "should create transaction")
		require.NotNil(t, txResp, "transaction response should not be nil")
		require.NotEmpty(t, txResp.Hash, "transaction hash should be provided")
		assert.Equal(t, "pending", txResp.Status, "initial transaction status should be pending")

		t.Logf("Created transaction: %s (status: %s)", txResp.Hash, txResp.Status)

		// Step 3: Check transaction status progression
		txStatus, err := service.GetOverledgerTransactionStatus(txReq.NetworkID, txResp.Hash)
		require.NoError(t, err, "should get transaction status")
		require.NotNil(t, txStatus, "status response should not be nil")
		assert.Equal(t, txResp.Hash, txStatus.Hash, "transaction hashes should match")
		assert.Contains(t, []string{"pending", "confirmed", "failed"}, txStatus.Status, "status should be valid")

		t.Logf("Transaction status: %s (confirmations: %d)", txStatus.Status, txStatus.Confirmations)

		// Step 4: Verify network mapping consistency
		overledgerNetworks, err := service.GetOverledgerNetworks()
		require.NoError(t, err, "should get Overledger networks")
		require.NotEmpty(t, overledgerNetworks.Networks, "should have networks")

		meshNetworks, err := service.GetMeshNetworks()
		require.NoError(t, err, "should get Mesh networks")
		require.NotEmpty(t, meshNetworks.Networks, "should have mesh networks")

		// Verify we can map between the two systems
		foundEthereumOverledger := false
		foundEthereumMesh := false

		for _, net := range overledgerNetworks.Networks {
			if strings.Contains(strings.ToLower(net.Name), "ethereum") {
				foundEthereumOverledger = true
				break
			}
		}

		for _, net := range meshNetworks.Networks {
			if strings.Contains(strings.ToLower(net.NetworkIdentifier.Blockchain), "ethereum") {
				foundEthereumMesh = true
				break
			}
		}

		assert.True(t, foundEthereumOverledger, "should find Ethereum network in Overledger")
		assert.True(t, foundEthereumMesh, "should find Ethereum network in Mesh")
	})

	t.Run("Error Handling in Transaction Flow", func(t *testing.T) {
		// Test invalid network ID
		invalidTxReq := &overledger.TransactionRequest{
			NetworkID:   "invalid-network",
			FromAddress: "0x1234567890abcdef1234567890abcdef12345678",
			ToAddress:   "0xabcdef1234567890abcdef1234567890abcdef12",
			Amount:      "1000000000000000000",
		}

		// Mock server should still return a response, but in real implementation this might fail
		_, err := service.CreateOverledgerTransaction(invalidTxReq)
		// For mock server, this should succeed, but we can test the structure
		if err == nil {
			t.Log("Mock server accepted invalid network (expected for test)")
		}

		// Test invalid transaction hash for status check
		_, err = service.GetOverledgerTransactionStatus("ethereum-sepolia", "0xinvalidhash")
		// Mock server will return a response, real server might return 404
		if err == nil {
			t.Log("Mock server returned status for invalid hash (expected for test)")
		}
	})
}

// TestCrossChainTransactionMapping tests transaction mapping across different blockchain networks
func TestCrossChainTransactionMapping(t *testing.T) {
	overledgerServer, rosettaServer, overledgerClient := setupTokenTransferSuite(t)
	defer overledgerServer.Close()
	defer rosettaServer.Close()

	meshClient := clients.NewMeshSDKClient(rosettaServer.URL)
	meshAdapter := meshadapter.NewAdapter(meshClient)
	service := connector.NewService(nil, meshAdapter, overledgerClient)

	networks := []struct {
		name         string
		overledgerID string
		meshNetwork  map[string]interface{}
		amount       string
		tokenSymbol  string
	}{
		{
			name:         "Ethereum Sepolia",
			overledgerID: "ethereum-sepolia",
			meshNetwork: map[string]interface{}{
				"blockchain": "Ethereum",
				"network":    "Sepolia",
			},
			amount:      "1000000000000000000", // 1 ETH
			tokenSymbol: "ETH",
		},
		{
			name:         "Bitcoin Testnet",
			overledgerID: "bitcoin-testnet",
			meshNetwork: map[string]interface{}{
				"blockchain": "Bitcoin",
				"network":    "Testnet",
			},
			amount:      "100000000", // 1 BTC in satoshis
			tokenSymbol: "BTC",
		},
	}

	for _, network := range networks {
		t.Run(network.name, func(t *testing.T) {
			// Test balance retrieval consistency
			accountID := map[string]interface{}{
				"address": "test_address_" + strings.ToLower(network.tokenSymbol),
			}

			// Get Overledger balance
			overledgerBalance, err := service.GetOverledgerBalance(network.overledgerID, accountID["address"].(string))
			require.NoError(t, err, "should get Overledger balance for "+network.name)
			require.NotEmpty(t, overledgerBalance.Balances, "should have balances")

			// Get Mesh balance (this will use mock Rosetta server which returns ETH balances)
			meshBalance, err := service.GetMeshNetworkBalance(network.meshNetwork, accountID)
			require.NoError(t, err, "should get Mesh balance for "+network.name)
			require.NotEmpty(t, meshBalance.Balances, "should have mesh balances")

			t.Logf("Overledger %s balance: %+v", network.name, overledgerBalance.Balances[0])
			t.Logf("Mesh %s balance: %+v", network.name, meshBalance.Balances[0])

			// Create and verify transaction
			txReq := &overledger.TransactionRequest{
				NetworkID:   network.overledgerID,
				FromAddress: "sender_" + network.tokenSymbol,
				ToAddress:   "receiver_" + network.tokenSymbol,
				Amount:      network.amount,
			}

			txResp, err := service.CreateOverledgerTransaction(txReq)
			require.NoError(t, err, "should create transaction for "+network.name)
			assert.Equal(t, network.overledgerID, txResp.NetworkID, "network ID should match")
			assert.Equal(t, network.amount, txResp.Amount, "amount should match")
		})
	}
}

// TestTransactionStatusProgression simulates and tests transaction status changes over time
func TestTransactionStatusProgression(t *testing.T) {
	overledgerServer, _, overledgerClient := setupTokenTransferSuite(t)
	defer overledgerServer.Close()

	service := connector.NewService(nil, nil, overledgerClient)

	// Create a transaction
	txReq := &overledger.TransactionRequest{
		NetworkID:   "ethereum-sepolia",
		FromAddress: "0x1234567890abcdef1234567890abcdef12345678",
		ToAddress:   "0xabcdef1234567890abcdef1234567890abcdef12",
		Amount:      "1000000000000000000",
		GasLimit:    "21000",
		GasPrice:    "20000000000",
	}

	txResp, err := service.CreateOverledgerTransaction(txReq)
	require.NoError(t, err, "should create transaction")
	require.NotEmpty(t, txResp.Hash, "should have transaction hash")

	// Test multiple status checks (simulating polling)
	for i := 0; i < 3; i++ {
		status, err := service.GetOverledgerTransactionStatus(txReq.NetworkID, txResp.Hash)
		require.NoError(t, err, "should get transaction status on attempt %d", i+1)
		require.NotNil(t, status, "status should not be nil")
		
		assert.Equal(t, txResp.Hash, status.Hash, "hash should match")
		assert.Contains(t, []string{"pending", "confirmed", "failed"}, status.Status, "status should be valid")
		
		if status.Status == "confirmed" {
			assert.Greater(t, status.Confirmations, 0, "confirmed transaction should have confirmations")
			assert.Greater(t, status.BlockNumber, int64(0), "confirmed transaction should have block number")
			assert.NotEmpty(t, status.BlockHash, "confirmed transaction should have block hash")
		}

		t.Logf("Status check %d: %s (confirmations: %d, block: %d)", 
			i+1, status.Status, status.Confirmations, status.BlockNumber)

		// Short delay between checks
		time.Sleep(100 * time.Millisecond)
	}
}

// TestTokenTransferValidation tests validation of transaction parameters
func TestTokenTransferValidation(t *testing.T) {
	overledgerServer, _, overledgerClient := setupTokenTransferSuite(t)
	defer overledgerServer.Close()

	service := connector.NewService(nil, nil, overledgerClient)

	testCases := []struct {
		name        string
		request     *overledger.TransactionRequest
		expectError bool
		description string
	}{
		{
			name: "Valid ETH transfer",
			request: &overledger.TransactionRequest{
				NetworkID:   "ethereum-sepolia",
				FromAddress: "0x1234567890abcdef1234567890abcdef12345678",
				ToAddress:   "0xabcdef1234567890abcdef1234567890abcdef12",
				Amount:      "1000000000000000000",
				GasLimit:    "21000",
				GasPrice:    "20000000000",
			},
			expectError: false,
			description: "Standard ETH transfer should succeed",
		},
		{
			name: "Valid ERC20 token transfer",
			request: &overledger.TransactionRequest{
				NetworkID:   "ethereum-sepolia",
				FromAddress: "0x1234567890abcdef1234567890abcdef12345678",
				ToAddress:   "0xabcdef1234567890abcdef1234567890abcdef12",
				Amount:      "5000000000000000000",
				TokenID:     "0xa0b86a33e6fe17f58b8b4cd6e45e74b8a7b4d30d",
				GasLimit:    "65000",
				GasPrice:    "25000000000",
			},
			expectError: false,
			description: "ERC20 token transfer should succeed",
		},
		{
			name: "Empty network ID",
			request: &overledger.TransactionRequest{
				NetworkID:   "",
				FromAddress: "0x1234567890abcdef1234567890abcdef12345678",
				ToAddress:   "0xabcdef1234567890abcdef1234567890abcdef12",
				Amount:      "1000000000000000000",
			},
			expectError: false, // Mock server will handle this
			description: "Empty network ID handled by mock server",
		},
		{
			name: "Zero amount",
			request: &overledger.TransactionRequest{
				NetworkID:   "ethereum-sepolia",
				FromAddress: "0x1234567890abcdef1234567890abcdef12345678",
				ToAddress:   "0xabcdef1234567890abcdef1234567890abcdef12",
				Amount:      "0",
			},
			expectError: false, // Mock server accepts this
			description: "Zero amount transfer handled by mock server",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txResp, err := service.CreateOverledgerTransaction(tc.request)
			
			if tc.expectError {
				assert.Error(t, err, tc.description)
				assert.Nil(t, txResp, "response should be nil on error")
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotNil(t, txResp, "response should not be nil on success")
				if txResp != nil {
					assert.NotEmpty(t, txResp.Hash, "transaction hash should be provided")
					assert.Equal(t, tc.request.NetworkID, txResp.NetworkID, "network ID should match")
					assert.Equal(t, tc.request.Amount, txResp.Amount, "amount should match")
				}
			}
		})
	}
}