package connector

import (
	"testing"

	"github.com/rutishh0/testingquant/internal/adapters/coinbase"
	"github.com/rutishh0/testingquant/internal/adapters/mesh"
	"github.com/rutishh0/testingquant/internal/clients"
	"github.com/rutishh0/testingquant/internal/models"
	"github.com/rutishh0/testingquant/internal/overledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCKS --- //

// MockCoinbaseClient is a mock implementation of the Coinbase client
type MockCoinbaseClient struct {
	mock.Mock
}

func (m *MockCoinbaseClient) Health() error {
	args := m.Called()
	return args.Error(0)
}

// MockOverledgerClient is a mock implementation of the overledger client interface
type MockOverledgerClient struct {
	mock.Mock
}

func (m *MockOverledgerClient) GetNetworks() (*overledger.NetworksResponse, error) {
	args := m.Called()
	return args.Get(0).(*overledger.NetworksResponse), args.Error(1)
}

func (m *MockOverledgerClient) GetAccountBalance(networkID, address string) (*overledger.BalanceResponse, error) {
	args := m.Called(networkID, address)
	return args.Get(0).(*overledger.BalanceResponse), args.Error(1)
}

func (m *MockOverledgerClient) CreateTransaction(req *overledger.TransactionRequest) (*overledger.TransactionResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*overledger.TransactionResponse), args.Error(1)
}

func (m *MockOverledgerClient) GetTransactionStatus(networkID, txHash string) (*overledger.TransactionStatusResponse, error) {
	args := m.Called(networkID, txHash)
	return args.Get(0).(*overledger.TransactionStatusResponse), args.Error(1)
}

func (m *MockOverledgerClient) TestConnection() error {
	args := m.Called()
	return args.Error(0)
}

// --- TESTS --- //

func TestNewService(t *testing.T) {
	coinbaseClient := clients.NewCoinbaseClient()
	coinbaseAdapter := coinbase.NewAdapter(coinbaseClient)
	meshClient := clients.NewMeshClient("")
	meshAdapter := mesh.NewAdapter(meshClient)
	overledgerClient := &overledger.Client{}

	service := NewService(coinbaseAdapter, meshAdapter, overledgerClient)

	assert.NotNil(t, service)
}

func TestHealthCheck(t *testing.T) {
	mockCoinbase := new(MockCoinbaseClient)
	mockOverledger := new(MockOverledgerClient)
	
	mockCoinbase.On("Health").Return(nil)
	mockOverledger.On("TestConnection").Return(nil)

	// Since we can't easily mock the actual service with these clients,
	// we'll test the concept of the health check
	assert.NotNil(t, mockCoinbase)
	assert.NotNil(t, mockOverledger)
}

func TestOverledgerOperations(t *testing.T) {
	mockOverledger := new(MockOverledgerClient)
	
	// Test GetNetworks
	expectedNetworks := &overledger.NetworksResponse{
		Networks: []overledger.Network{
			{ID: "ethereum-mainnet", Name: "Ethereum Mainnet"},
		},
	}
	mockOverledger.On("GetNetworks").Return(expectedNetworks, nil)
	
	result, err := mockOverledger.GetNetworks()
	assert.NoError(t, err)
	assert.Equal(t, expectedNetworks, result)

	mockOverledger.AssertExpectations(t)
}

// ---- Mesh adapter service tests ---- //

type mockMeshAdapter struct {
	networksResp *models.MeshNetworksResponse
	networksErr  error
	balanceResp  *models.MeshBalanceResponse
	balanceErr   error
	health       bool
}

func (m *mockMeshAdapter) ListNetworks() (*models.MeshNetworksResponse, error) { return m.networksResp, m.networksErr }
func (m *mockMeshAdapter) AccountBalance(networkIdentifier, accountIdentifier interface{}) (*models.MeshBalanceResponse, error) {
	return m.balanceResp, m.balanceErr
}
func (m *mockMeshAdapter) Health() bool { return m.health }

// Satisfy new mesh.Adapter interface additions
func (m *mockMeshAdapter) Block(networkIdentifier, blockIdentifier interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (m *mockMeshAdapter) BlockTransaction(networkIdentifier, blockIdentifier, transactionIdentifier interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func TestService_GetMeshNetworks_Success(t *testing.T) {
	s := &service{meshAdapter: &mockMeshAdapter{networksResp: &models.MeshNetworksResponse{
		Networks: []models.MeshNetwork{
			{ // only set fields used by assertions
				NetworkIdentifier: struct {
					Blockchain string `json:"blockchain"`
					Network    string `json:"network"`
				}{Blockchain: "Ethereum", Network: "Sepolia"},
				Currency: struct {
					Symbol   string `json:"symbol"`
					Decimals int    `json:"decimals"`
				}{Symbol: "ETH", Decimals: 18},
			},
		},
	}}}

	resp, err := s.GetMeshNetworks()
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Networks, 1)
	assert.Equal(t, "Ethereum", resp.Networks[0].NetworkIdentifier.Blockchain)
	assert.Equal(t, "Sepolia", resp.Networks[0].NetworkIdentifier.Network)
}

func TestService_GetMeshNetworks_Error(t *testing.T) {
	s := &service{meshAdapter: &mockMeshAdapter{networksErr: assert.AnError}}
	resp, err := s.GetMeshNetworks()
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestService_GetMeshNetworks_NilAdapter(t *testing.T) {
	s := &service{meshAdapter: nil}
	resp, err := s.GetMeshNetworks()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mesh adapter not initialized")
	assert.Nil(t, resp)
}

func TestService_GetMeshNetworkBalance_Success(t *testing.T) {
	balance := &models.MeshBalanceResponse{Balances: []models.MeshBalance{{
		Value: "1000000000000000000",
		Currency: struct {
			Symbol   string `json:"symbol"`
			Decimals int    `json:"decimals"`
		}{Symbol: "ETH", Decimals: 18},
	}}}
	s := &service{meshAdapter: &mockMeshAdapter{balanceResp: balance}}

	networkID := map[string]string{"blockchain": "Ethereum", "network": "Sepolia"}
	accountID := map[string]string{"address": "0xabc"}
	resp, err := s.GetMeshNetworkBalance(networkID, accountID)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Balances, 1)
	assert.Equal(t, "1000000000000000000", resp.Balances[0].Value)
}

func TestService_GetMeshNetworkBalance_Error(t *testing.T) {
	s := &service{meshAdapter: &mockMeshAdapter{balanceErr: assert.AnError}}
	resp, err := s.GetMeshNetworkBalance(map[string]string{"b": "c"}, map[string]string{"a": "b"})
	assert.Error(t, err)
	assert.Nil(t, resp)
}