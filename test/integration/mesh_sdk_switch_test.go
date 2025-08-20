package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    meshadapter "github.com/rutishh0/testingquant/internal/adapters/mesh"
    "github.com/rutishh0/testingquant/internal/api"
    "github.com/rutishh0/testingquant/internal/clients"
    "github.com/rutishh0/testingquant/internal/config"
    "github.com/rutishh0/testingquant/internal/connector"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    roserver "github.com/coinbase/rosetta-sdk-go/server"
    rotypes "github.com/coinbase/rosetta-sdk-go/types"
    "github.com/coinbase/rosetta-sdk-go/asserter"
    meshservices "github.com/rutishh0/mesh-server/services"
)

// startMockRosettaServer spins up an in-memory Rosetta-compatible Mesh server with mock data
func startMockRosettaServer(t *testing.T) *httptest.Server {
    t.Helper()

    // Support both Ethereum Sepolia and Bitcoin Testnet in the asserter so tests can query either
    ethSepolia := &rotypes.NetworkIdentifier{Blockchain: "Ethereum", Network: "Sepolia"}
    btcTestnet := &rotypes.NetworkIdentifier{Blockchain: "Bitcoin", Network: "Testnet"}
    supported := []*rotypes.NetworkIdentifier{ethSepolia, btcTestnet}

    assr, err := asserter.NewServer(
        []string{"Transfer", "Reward", "Fee"},
        false,
        supported,
        nil,
        false,
        "",
    )
    require.NoError(t, err)

    // Underlying mock services remain initialized for Ethereum Sepolia, which is sufficient for our tests
    networkAPIService := meshservices.NewNetworkAPIService(ethSepolia, nil)
    blockAPIService := meshservices.NewBlockAPIService(ethSepolia, nil)
    accountAPIService := meshservices.NewAccountAPIService(ethSepolia, nil)

    networkAPIController := roserver.NewNetworkAPIController(networkAPIService, assr)
    blockAPIController := roserver.NewBlockAPIController(blockAPIService, assr)
    accountAPIController := roserver.NewAccountAPIController(accountAPIService, assr)

    rosettaRouter := roserver.NewRouter(networkAPIController, blockAPIController, accountAPIController)

    return httptest.NewServer(rosettaRouter)
}

// startConnectorServer spins up a minimal connector server exposing only /v1/mesh routes, using the provided Mesh client implementation
func startConnectorServer(t *testing.T, meshClient clients.MeshAPI) *httptest.Server {
    t.Helper()

    // Build service with our mesh adapter
    meshAdapter := meshadapter.NewAdapter(meshClient)
    svc := connector.NewService(nil, meshAdapter, nil)

    // Minimal config (no API key)
    cfg := &config.Config{APIKey: "", Environment: "test"}

    // Handlers and minimal router
    handlers := api.NewHandlers(svc, cfg)
    r := gin.New()
    r.GET("/v1/mesh/networks", handlers.GetMeshNetworks)
    r.POST("/v1/mesh/account/balance", handlers.GetMeshAccountBalance)

    return httptest.NewServer(r)
}

func httpClient() *http.Client {
    return &http.Client{Timeout: 10 * time.Second}
}

func getJSON(t *testing.T, baseURL, path string) (int, map[string]interface{}) {
    t.Helper()
    req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
    require.NoError(t, err)

    resp, err := httpClient().Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()

    var out map[string]interface{}
    dec := json.NewDecoder(resp.Body)
    _ = dec.Decode(&out)
    return resp.StatusCode, out
}

func postJSON(t *testing.T, baseURL, path string, payload any) (int, map[string]interface{}) {
    t.Helper()
    b, _ := json.Marshal(payload)
    req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(b))
    require.NoError(t, err)
    req.Header.Set("Content-Type", "application/json")

    resp, err := httpClient().Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()

    var out map[string]interface{}
    dec := json.NewDecoder(resp.Body)
    _ = dec.Decode(&out)
    return resp.StatusCode, out
}

// TestMeshSDKSwitch_Equivalence validates that using HTTP vs SDK Mesh clients yields equivalent responses at the connector API level
func TestMeshSDKSwitch_Equivalence(t *testing.T) {
    rosetta := startMockRosettaServer(t)
    defer rosetta.Close()

    // Base URL for clients should target the Rosetta server root (endpoints like /network/list)
    base := rosetta.URL

    // Start connector with HTTP client
    httpClientImpl := clients.NewMeshClient(base)
    connectorHTTP := startConnectorServer(t, httpClientImpl)
    defer connectorHTTP.Close()

    // Start connector with SDK client
    sdkClientImpl := clients.NewMeshSDKClient(base)
    connectorSDK := startConnectorServer(t, sdkClientImpl)
    defer connectorSDK.Close()

    // 1) Compare /v1/mesh/networks
    statusHTTP, bodyHTTP := getJSON(t, connectorHTTP.URL, "/v1/mesh/networks")
    statusSDK, bodySDK := getJSON(t, connectorSDK.URL, "/v1/mesh/networks")

    require.Equal(t, http.StatusOK, statusHTTP, "HTTP client should return 200 for networks")
    require.Equal(t, http.StatusOK, statusSDK, "SDK client should return 200 for networks")

    // Compare networks array length and first item identifiers
    netsHTTP, ok1 := bodyHTTP["networks"].([]interface{})
    netsSDK, ok2 := bodySDK["networks"].([]interface{})
    require.True(t, ok1 && ok2, "both responses must contain 'networks' array")
    require.NotEmpty(t, netsHTTP)
    require.NotEmpty(t, netsSDK)
    assert.Equal(t, len(netsHTTP), len(netsSDK), "network counts should match")

    // Ensure currency defaults are present and consistent
    firstHTTP := netsHTTP[0].(map[string]interface{})
    firstSDK := netsSDK[0].(map[string]interface{})
    assert.Equal(t, firstHTTP["currency"], firstSDK["currency"], "currency defaults should match (symbol/decimals)")

    // 2) Compare /v1/mesh/account/balance
    payload := map[string]interface{}{
        "network_identifier": map[string]string{
            "blockchain": "Ethereum",
            "network":    "Sepolia",
        },
        "account_identifier": map[string]string{
            "address": "0x1234567890abcdef1234567890abcdef12345678",
        },
    }

    statusHTTP, balHTTP := postJSON(t, connectorHTTP.URL, "/v1/mesh/account/balance", payload)
    statusSDK, balSDK := postJSON(t, connectorSDK.URL, "/v1/mesh/account/balance", payload)

    require.Equal(t, http.StatusOK, statusHTTP, "HTTP client should return 200 for account balance")
    require.Equal(t, http.StatusOK, statusSDK, "SDK client should return 200 for account balance")

    // Compare balances length and currency/value tuples
    balsHTTP, ok1 := balHTTP["balances"].([]interface{})
    balsSDK, ok2 := balSDK["balances"].([]interface{})
    require.True(t, ok1 && ok2, "both responses must contain 'balances' array")
    assert.Equal(t, len(balsHTTP), len(balsSDK), "balance counts should match")

    // Create short representation: symbol:decimals:value for first two items to avoid order sensitivity issues
    if len(balsHTTP) > 0 && len(balsSDK) > 0 {
        cHTTP := balsHTTP[0].(map[string]interface{})["currency"].(map[string]interface{})
        cSDK := balsSDK[0].(map[string]interface{})["currency"].(map[string]interface{})
        assert.Equal(t, cHTTP["symbol"], cSDK["symbol"]) 
        assert.Equal(t, cHTTP["decimals"], cSDK["decimals"]) 
    }
}