package connector

import (
	"testing"

	"github.com/rutishh0/testingquant/internal/mesh"
	"github.com/rutishh0/testingquant/internal/overledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCKS --- //

// MockMeshClient is a mock implementation of the mesh client interface.
type MockMeshClient struct {
	mock.Mock
}

func (m *MockMeshClient) ConstructionPreprocess(req *mesh.ConstructionPreprocessRequest) (*mesh.ConstructionPreprocessResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.ConstructionPreprocessResponse), args.Error(1)
}

func (m *MockMeshClient) ConstructionPayloads(req *mesh.ConstructionPayloadsRequest) (*mesh.ConstructionPayloadsResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.ConstructionPayloadsResponse), args.Error(1)
}

func (m *MockMeshClient) ConstructionCombine(req *mesh.ConstructionCombineRequest) (*mesh.ConstructionCombineResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.ConstructionCombineResponse), args.Error(1)
}

func (m *MockMeshClient) ConstructionSubmit(req *mesh.ConstructionSubmitRequest) (*mesh.ConstructionSubmitResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.ConstructionSubmitResponse), args.Error(1)
}

func (m *MockMeshClient) AccountBalance(req *mesh.AccountBalanceRequest) (*mesh.AccountBalanceResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.AccountBalanceResponse), args.Error(1)
}

func (m *MockMeshClient) Block(req *mesh.BlockRequest) (*mesh.BlockResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.BlockResponse), args.Error(1)
}

// MockOverledgerClient is a mock implementation of the overledger client interface.
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

func (m *MockOverledgerClient) TestConnection() error {
	args := m.Called()
	return args.Error(0)
}


// --- TESTS --- //

func TestNewService(t *testing.T) {
	mockMeshClient := &MockMeshClient{}
	mockOverledgerClient := &MockOverledgerClient{}
	// We pass the mocks to the original NewService function which now accepts interfaces
	service := NewService(mockMeshClient, mockOverledgerClient)

	assert.NotNil(t, service)
	assert.Implements(t, (*Service)(nil), service)
}

func TestService_Preprocess(t *testing.T) {
	mockMeshClient := &MockMeshClient{}
	mockOverledgerClient := &MockOverledgerClient{}
	service := NewService(mockMeshClient, mockOverledgerClient)

	// Test data
	req := &PreprocessRequest{
		DLT:     "ethereum",
		Network: "mainnet",
		Type:    "TRANSFER",
		Transfers: []Transfer{
			{
				From:        "0xfromAddress",
				To:          "0xtoAddress",
				Amount:      "100000",
				TokenSymbol: "ETH",
			},
		},
	}

	expectedMeshResp := &mesh.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			"gas_limit": "21000",
			"gas_price": "20000000000",
		},
		RequiredPublicKeys: []mesh.AccountIdentifier{
			{Address: "0xfromAddress"},
		},
	}

	mockMeshClient.On("ConstructionPreprocess", mock.AnythingOfType("*mesh.ConstructionPreprocessRequest")).Return(expectedMeshResp, nil)

	// Execute
	resp, err := service.Preprocess(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Options)
	assert.Equal(t, "21000", resp.Options["gas_limit"])
	assert.Equal(t, "20000000000", resp.Options["gas_price"])
	assert.Equal(t, []string{"0xfromAddress"}, resp.RequiredSigners)

	mockMeshClient.AssertExpectations(t)
}