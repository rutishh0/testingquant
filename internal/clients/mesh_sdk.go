package clients

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"

    roclient "github.com/coinbase/rosetta-sdk-go/client"
    rotypes "github.com/coinbase/rosetta-sdk-go/types"
)

// MeshSDKClient implements MeshAPI using the official Rosetta (Mesh) SDK client.
// It adapts typed SDK responses back into http.Response objects to keep the
// rest of the code unchanged.
type MeshSDKClient struct {
    baseURL   string
    apiClient *roclient.APIClient
}

func NewMeshSDKClient(baseURL string) *MeshSDKClient {
    if baseURL == "" {
        baseURL = "http://localhost:8080/mesh"
    }
    // rosetta-sdk-go v0.9.0 NewConfiguration signature: (serverURL, userAgent string, httpClient *http.Client)
    cfg := roclient.NewConfiguration(strings.TrimSuffix(baseURL, "/"), "", nil)
    return &MeshSDKClient{
        baseURL:   strings.TrimSuffix(baseURL, "/"),
        apiClient: roclient.NewAPIClient(cfg),
    }
}

// ListNetworks calls /network/list via SDK and wraps the result into an http.Response
func (m *MeshSDKClient) ListNetworks() (*http.Response, error) {
    ctx := context.Background()
    // MetadataRequest is empty for /network/list
    resp, _, err := m.apiClient.NetworkAPI.NetworkList(ctx, &rotypes.MetadataRequest{})
    if err != nil {
        return nil, err
    }
    return wrapJSONResponse(resp)
}

// NetworkStatus calls /network/status via SDK and wraps the result
func (m *MeshSDKClient) NetworkStatus(networkIdentifier interface{}, blockIdentifier interface{}) (*http.Response, error) { // blockIdentifier ignored (not in request schema)
    ctx := context.Background()
    ni, err := toNetworkIdentifier(networkIdentifier)
    if err != nil {
        return nil, err
    }
    req := &rotypes.NetworkRequest{NetworkIdentifier: ni}
    resp, _, err := m.apiClient.NetworkAPI.NetworkStatus(ctx, req)
    if err != nil {
        return nil, err
    }
    return wrapJSONResponse(resp)
}

// NetworkOptions calls /network/options via SDK and wraps the result
func (m *MeshSDKClient) NetworkOptions(networkIdentifier interface{}) (*http.Response, error) {
    ctx := context.Background()
    ni, err := toNetworkIdentifier(networkIdentifier)
    if err != nil {
        return nil, err
    }
    req := &rotypes.NetworkRequest{NetworkIdentifier: ni}
    resp, _, err := m.apiClient.NetworkAPI.NetworkOptions(ctx, req)
    if err != nil {
        return nil, err
    }
    return wrapJSONResponse(resp)
}

// AccountBalance calls /account/balance via SDK and wraps the result
func (m *MeshSDKClient) AccountBalance(networkIdentifier, accountIdentifier interface{}) (*http.Response, error) {
    ctx := context.Background()
    ni, err := toNetworkIdentifier(networkIdentifier)
    if err != nil {
        return nil, err
    }
    ai, err := toAccountIdentifier(accountIdentifier)
    if err != nil {
        return nil, err
    }
    req := &rotypes.AccountBalanceRequest{
        NetworkIdentifier: ni,
        AccountIdentifier: ai,
    }
    resp, _, err := m.apiClient.AccountAPI.AccountBalance(ctx, req)
    if err != nil {
        return nil, err
    }
    return wrapJSONResponse(resp)
}

// Block calls /block via SDK and wraps the result
func (m *MeshSDKClient) Block(networkIdentifier interface{}, blockIdentifier interface{}) (*http.Response, error) {
    ctx := context.Background()
    ni, err := toNetworkIdentifier(networkIdentifier)
    if err != nil {
        return nil, err
    }
    pbi, err := toPartialBlockIdentifier(blockIdentifier)
    if err != nil {
        return nil, err
    }
    req := &rotypes.BlockRequest{
        NetworkIdentifier: ni,
        BlockIdentifier:   pbi,
    }
    resp, _, err := m.apiClient.BlockAPI.Block(ctx, req)
    if err != nil {
        return nil, err
    }
    return wrapJSONResponse(resp)
}

// BlockTransaction calls /block/transaction via SDK and wraps the result
func (m *MeshSDKClient) BlockTransaction(networkIdentifier interface{}, blockIdentifier interface{}, transactionIdentifier interface{}) (*http.Response, error) {
    ctx := context.Background()
    ni, err := toNetworkIdentifier(networkIdentifier)
    if err != nil {
        return nil, err
    }
    bi, err := toFullBlockIdentifier(blockIdentifier)
    if err != nil {
        return nil, err
    }
    ti, err := toTransactionIdentifier(transactionIdentifier)
    if err != nil {
        return nil, err
    }
    req := &rotypes.BlockTransactionRequest{
        NetworkIdentifier:     ni,
        BlockIdentifier:       bi,
        TransactionIdentifier: ti,
    }
    resp, _, err := m.apiClient.BlockAPI.BlockTransaction(ctx, req)
    if err != nil {
        return nil, err
    }
    return wrapJSONResponse(resp)
}

// Health checks the Mesh network list via SDK
func (m *MeshSDKClient) Health() bool {
    if _, err := m.ListNetworks(); err != nil {
        return false
    }
    return true
}

// Helpers

func wrapJSONResponse(v interface{}) (*http.Response, error) {
    b, err := json.Marshal(v)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal response: %w", err)
    }
    r := &http.Response{
        Status:     "200 OK",
        StatusCode: 200,
        Header:     make(http.Header),
        Body:       io.NopCloser(bytes.NewReader(b)),
    }
    r.Header.Set("Content-Type", "application/json")
    return r, nil
}

func toNetworkIdentifier(v interface{}) (*rotypes.NetworkIdentifier, error) {
    switch t := v.(type) {
    case *rotypes.NetworkIdentifier:
        return t, nil
    case rotypes.NetworkIdentifier:
        return &t, nil
    case map[string]interface{}:
        blockchain, _ := t["blockchain"].(string)
        network, _ := t["network"].(string)
        if blockchain == "" || network == "" {
            return nil, fmt.Errorf("invalid network_identifier map: missing fields")
        }
        return &rotypes.NetworkIdentifier{Blockchain: blockchain, Network: network}, nil
    default:
        return nil, fmt.Errorf("unsupported network_identifier type: %T", v)
    }
}

func toAccountIdentifier(v interface{}) (*rotypes.AccountIdentifier, error) {
    switch t := v.(type) {
    case *rotypes.AccountIdentifier:
        return t, nil
    case rotypes.AccountIdentifier:
        return &t, nil
    case map[string]interface{}:
        address, _ := t["address"].(string)
        if address == "" {
            return nil, fmt.Errorf("invalid account_identifier map: missing address")
        }
        return &rotypes.AccountIdentifier{Address: address}, nil
    default:
        return nil, fmt.Errorf("unsupported account_identifier type: %T", v)
    }
}

// Converts input to PartialBlockIdentifier for /block
func toPartialBlockIdentifier(v interface{}) (*rotypes.PartialBlockIdentifier, error) {
    switch t := v.(type) {
    case nil:
        return &rotypes.PartialBlockIdentifier{}, nil
    case *rotypes.PartialBlockIdentifier:
        return t, nil
    case rotypes.PartialBlockIdentifier:
        return &t, nil
    case map[string]interface{}:
        var (
            idxPtr *int64
            hashStr string
        )
        if raw, ok := t["index"]; ok {
            switch n := raw.(type) {
            case float64:
                v := int64(n)
                idxPtr = &v
            case int64:
                v := n
                idxPtr = &v
            case int:
                v := int64(n)
                idxPtr = &v
            }
        }
        if h, ok := t["hash"].(string); ok {
            hashStr = h
        }
        pbi := &rotypes.PartialBlockIdentifier{}
        if idxPtr != nil {
            pbi.Index = idxPtr
        }
        if hashStr != "" {
            pbi.Hash = &hashStr
        }
        return pbi, nil
    default:
        return nil, fmt.Errorf("unsupported block_identifier type for /block: %T", v)
    }
}

// Converts input to BlockIdentifier for /block/transaction
func toFullBlockIdentifier(v interface{}) (*rotypes.BlockIdentifier, error) {
    switch t := v.(type) {
    case *rotypes.BlockIdentifier:
        return t, nil
    case rotypes.BlockIdentifier:
        return &t, nil
    case map[string]interface{}:
        var (
            idx int64
            hasIdx bool
            hash string
        )
        if raw, ok := t["index"]; ok {
            switch n := raw.(type) {
            case float64:
                idx = int64(n)
                hasIdx = true
            case int64:
                idx = n
                hasIdx = true
            case int:
                idx = int64(n)
                hasIdx = true
            }
        }
        if h, ok := t["hash"].(string); ok {
            hash = h
        }
        // If neither provided, return empty identifier error
        if !hasIdx && hash == "" {
            return nil, fmt.Errorf("invalid block_identifier map: require index and/or hash")
        }
        return &rotypes.BlockIdentifier{Index: idx, Hash: hash}, nil
    default:
        return nil, fmt.Errorf("unsupported block_identifier type for /block/transaction: %T", v)
    }
}

func toTransactionIdentifier(v interface{}) (*rotypes.TransactionIdentifier, error) {
    switch t := v.(type) {
    case *rotypes.TransactionIdentifier:
        return t, nil
    case rotypes.TransactionIdentifier:
        return &t, nil
    case map[string]interface{}:
        hash, _ := t["hash"].(string)
        if hash == "" {
            return nil, fmt.Errorf("invalid transaction_identifier map: missing hash")
        }
        return &rotypes.TransactionIdentifier{Hash: hash}, nil
    default:
        return nil, fmt.Errorf("unsupported transaction_identifier type: %T", v)
    }
}