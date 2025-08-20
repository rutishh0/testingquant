package mesh

import (
    "bytes"
    "errors"
    "io"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// mockMeshAPI implements clients.MeshAPI for testing the adapter mapping
type mockMeshAPI struct {
    listResp       *http.Response
    listErr        error
    balanceResp    *http.Response
    balanceErr     error
    health         bool
}

func (m *mockMeshAPI) ListNetworks() (*http.Response, error) { return m.listResp, m.listErr }
func (m *mockMeshAPI) NetworkStatus(networkIdentifier interface{}, blockIdentifier interface{}) (*http.Response, error) {
    return nil, nil
}
func (m *mockMeshAPI) NetworkOptions(networkIdentifier interface{}) (*http.Response, error) { return nil, nil }
func (m *mockMeshAPI) AccountBalance(networkIdentifier, accountIdentifier interface{}) (*http.Response, error) {
    return m.balanceResp, m.balanceErr
}
// Added to satisfy clients.MeshAPI interface after Block methods were introduced
func (m *mockMeshAPI) Block(networkIdentifier interface{}, blockIdentifier interface{}) (*http.Response, error) {
    return nil, nil
}
func (m *mockMeshAPI) BlockTransaction(networkIdentifier interface{}, blockIdentifier interface{}, transactionIdentifier interface{}) (*http.Response, error) {
    return nil, nil
}
func (m *mockMeshAPI) Health() bool { return m.health }

func newHTTPResponse(status int, body string) *http.Response {
    return &http.Response{
        StatusCode: status,
        Body:       io.NopCloser(bytes.NewBufferString(body)),
        Header:     http.Header{"Content-Type": []string{"application/json"}},
    }
}

func TestAdapter_ListNetworks_MapsCurrencyDefaults(t *testing.T) {
    rosettaList := `{
        "network_identifiers": [
            {"blockchain": "Ethereum", "network": "Sepolia"},
            {"blockchain": "Bitcoin", "network": "Testnet3"}
        ]
    }`

    mockClient := &mockMeshAPI{listResp: newHTTPResponse(200, rosettaList), health: true}
    a := NewAdapter(mockClient)

    resp, err := a.ListNetworks()
    require.NoError(t, err)
    require.NotNil(t, resp)
    require.Len(t, resp.Networks, 2)

    // First network mapping
    n0 := resp.Networks[0]
    assert.Equal(t, "Ethereum", n0.NetworkIdentifier.Blockchain)
    assert.Equal(t, "Sepolia", n0.NetworkIdentifier.Network)
    // Currency defaults provided by adapter
    assert.Equal(t, "ETH", n0.Currency.Symbol)
    assert.Equal(t, 18, n0.Currency.Decimals)

    // Second network mapping
    n1 := resp.Networks[1]
    assert.Equal(t, "Bitcoin", n1.NetworkIdentifier.Blockchain)
    assert.Equal(t, "Testnet3", n1.NetworkIdentifier.Network)
    // Still uses defaults as per current adapter implementation
    assert.Equal(t, "ETH", n1.Currency.Symbol)
    assert.Equal(t, 18, n1.Currency.Decimals)

    // Health should reflect underlying client
    assert.True(t, a.Health())
}

func TestAdapter_ListNetworks_Error(t *testing.T) {
    mockClient := &mockMeshAPI{listErr: errors.New("boom")}
    a := NewAdapter(mockClient)

    resp, err := a.ListNetworks()
    require.Error(t, err)
    assert.Nil(t, resp)
}

func TestAdapter_ListNetworks_BadJSON(t *testing.T) {
    mockClient := &mockMeshAPI{listResp: newHTTPResponse(200, "not json")}
    a := NewAdapter(mockClient)

    resp, err := a.ListNetworks()
    require.Error(t, err)
    assert.Nil(t, resp)
}

func TestAdapter_AccountBalance_PassThroughDecode(t *testing.T) {
    balanceJSON := `{
        "balances": [
            {"value": "1234567890000000000", "currency": {"symbol": "ETH", "decimals": 18}}
        ]
    }`
    mockClient := &mockMeshAPI{balanceResp: newHTTPResponse(200, balanceJSON)}
    a := NewAdapter(mockClient)

    resp, err := a.AccountBalance(map[string]string{"blockchain": "Ethereum", "network": "Sepolia"}, map[string]string{"address": "0xabc"})
    require.NoError(t, err)
    require.NotNil(t, resp)

    require.Len(t, resp.Balances, 1)
    b := resp.Balances[0]
    assert.Equal(t, "ETH", b.Currency.Symbol)
    assert.Equal(t, 18, b.Currency.Decimals)
    assert.Equal(t, "1234567890000000000", b.Value)
}

func TestAdapter_AccountBalance_Error(t *testing.T) {
    mockClient := &mockMeshAPI{balanceErr: errors.New("downstream error")}
    a := NewAdapter(mockClient)

    resp, err := a.AccountBalance(map[string]string{"blockchain": "Ethereum", "network": "Sepolia"}, map[string]string{"address": "0xabc"})
    require.Error(t, err)
    assert.Nil(t, resp)
}

func TestAdapter_AccountBalance_BadJSON(t *testing.T) {
    mockClient := &mockMeshAPI{balanceResp: newHTTPResponse(200, "{" )}
    a := NewAdapter(mockClient)

    resp, err := a.AccountBalance(map[string]string{"blockchain": "Ethereum", "network": "Sepolia"}, map[string]string{"address": "0xabc"})
    require.Error(t, err)
    assert.Nil(t, resp)
}

func TestAdapter_Health_Delegates(t *testing.T) {
    aTrue := NewAdapter(&mockMeshAPI{health: true})
    aFalse := NewAdapter(&mockMeshAPI{health: false})
    assert.True(t, aTrue.Health())
    assert.False(t, aFalse.Health())
}