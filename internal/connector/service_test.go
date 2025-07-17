package connector

import (
	"testing"

	"github.com/quant-mesh-connector/internal/mesh"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMeshClient is a mock implementation of mesh.Client
type MockMeshClient struct {
	mock.Mock
}

func (m *MockMeshClient) NetworkStatus(req *mesh.NetworkStatusRequest) (*mesh.NetworkStatusResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.NetworkStatusResponse), args.Error(1)
}

func (m *MockMeshClient) AccountBalance(req *mesh.AccountBalanceRequest) (*mesh.AccountBalanceResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.AccountBalanceResponse), args.Error(1)
}

func (m *MockMeshClient) Block(req *mesh.BlockRequest) (*mesh.BlockResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*mesh.BlockResponse), args.Error(1)
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

func (m *MockMeshClient) Health() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	mockClient := &MockMeshClient{}
	service := NewService(mockClient)

	assert.NotNil(t, service)
	assert.IsType(t, &service{}, service)
}

func TestService_GetBalance(t *testing.T) {
	mockClient := &MockMeshClient{}
	service := NewService(mockClient)

	// Test data
	req := &BalanceRequest{
		NetworkIdentifier: NetworkIdentifier{
			Blockchain: "ethereum",
			Network:    "mainnet",
		},
		AccountIdentifier: AccountIdentifier{
			Address: "0x1234567890123456789012345678901234567890",
		},
	}

	expectedMeshReq := &mesh.AccountBalanceRequest{
		NetworkIdentifier: &mesh.NetworkIdentifier{
			Blockchain: "ethereum",
			Network:    "mainnet",
		},
		AccountIdentifier: &mesh.AccountIdentifier{
			Address: "0x1234567890123456789012345678901234567890",
		},
	}

	expectedMeshResp := &mesh.AccountBalanceResponse{
		BlockIdentifier: &mesh.BlockIdentifier{
			Index: 12345,
			Hash:  "0xabcdef",
		},
		Balances: []*mesh.Amount{
			{
				Value: "1000000000000000000",
				Currency: &mesh.Currency{
					Symbol:   "ETH",
					Decimals: 18,
				},
			},
		},
	}

	mockClient.On("AccountBalance", expectedMeshReq).Return(expectedMeshResp, nil)

	// Execute
	resp, err := service.GetBalance(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(12345), resp.BlockIdentifier.Index)
	assert.Equal(t, "0xabcdef", resp.BlockIdentifier.Hash)
	assert.Len(t, resp.Balances, 1)
	assert.Equal(t, "1000000000000000000", resp.Balances[0].Value)
	assert.Equal(t, "ETH", resp.Balances[0].Currency.Symbol)
	assert.Equal(t, int32(18), resp.Balances[0].Currency.Decimals)

	mockClient.AssertExpectations(t)
}

func TestService_Preprocess(t *testing.T) {
	mockClient := &MockMeshClient{}
	service := NewService(mockClient)

	// Test data
	req := &PreprocessRequest{
		NetworkIdentifier: NetworkIdentifier{
			Blockchain: "ethereum",
			Network:    "mainnet",
		},
		Operations: []Operation{
			{
				OperationIdentifier: OperationIdentifier{
					Index: 0,
				},
				Type: "TRANSFER",
				Account: &AccountIdentifier{
					Address: "0x1234567890123456789012345678901234567890",
				},
				Amount: &Amount{
					Value: "1000000000000000000",
					Currency: Currency{
						Symbol:   "ETH",
						Decimals: 18,
					},
				},
			},
		},
	}

	expectedMeshResp := &mesh.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			"gas_limit": "21000",
			"gas_price": "20000000000",
		},
	}

	mockClient.On("ConstructionPreprocess", mock.AnythingOfType("*mesh.ConstructionPreprocessRequest")).Return(expectedMeshResp, nil)

	// Execute
	resp, err := service.Preprocess(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Options)
	assert.Equal(t, "21000", resp.Options["gas_limit"])
	assert.Equal(t, "20000000000", resp.Options["gas_price"])

	mockClient.AssertExpectations(t)
}