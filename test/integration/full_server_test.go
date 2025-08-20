package integration

import (
    "context"
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/rutishh0/testingquant/internal/api"
    "github.com/rutishh0/testingquant/internal/clients"
    "github.com/rutishh0/testingquant/internal/config"
    "github.com/rutishh0/testingquant/internal/connector"
    meshadapter "github.com/rutishh0/testingquant/internal/adapters/mesh"
)

// TestFullServerMeshSDKIntegration validates a full server integration with MESH_USE_SDK flag
func TestFullServerMeshSDKIntegration(t *testing.T) {
    rosetta := startMockRosettaServer(t)
    defer rosetta.Close()

    tests := []struct {
        name      string
        meshUseSDK bool
    }{
        {"HTTP_Client", false},
        {"SDK_Client", true},
    }

    var httpResult, sdkResult map[string]interface{}
    var mu sync.Mutex

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := startFullServerWithSDKFlag(t, rosetta.URL, tt.meshUseSDK)
            defer server.Close()

            // Test /health endpoint
            statusCode, _ := getJSON(t, server.URL, "/health")
            require.Equal(t, http.StatusOK, statusCode, "health endpoint should work")

            // Test /mesh/network/list endpoint (Rosetta expects POST)
            statusCode, _ = postJSON(t, server.URL, "/mesh/network/list", map[string]any{})
            require.Equal(t, http.StatusOK, statusCode, "Rosetta /mesh/network/list should work")

            // Test connector's /v1/mesh/networks
            statusCode, networks := getJSON(t, server.URL, "/v1/mesh/networks")
            require.Equal(t, http.StatusOK, statusCode, "connector /v1/mesh/networks should work")

            mu.Lock()
            if tt.meshUseSDK {
                sdkResult = networks
            } else {
                httpResult = networks
            }
            mu.Unlock()
        })
    }

    // Compare results between SDK and HTTP clients
    require.NotNil(t, httpResult, "HTTP client results should be captured")
    require.NotNil(t, sdkResult, "SDK client results should be captured")

    // Verify networks array length and structure consistency
    httpNets, ok1 := httpResult["networks"].([]interface{})
    sdkNets, ok2 := sdkResult["networks"].([]interface{})
    require.True(t, ok1 && ok2, "both results should have networks array")
    assert.Equal(t, len(httpNets), len(sdkNets), "network counts should match between SDK and HTTP")

    if len(httpNets) > 0 && len(sdkNets) > 0 {
        httpFirst := httpNets[0].(map[string]interface{})
        sdkFirst := sdkNets[0].(map[string]interface{})
        assert.Equal(t, httpFirst["currency"], sdkFirst["currency"], "currency defaults should match")
    }
}

// startFullServerWithSDKFlag creates a minimal full server with the mesh client determined by SDK flag
func startFullServerWithSDKFlag(t *testing.T, meshBaseURL string, useMeshSDK bool) *httptest.Server {
    t.Helper()

    // Temporarily set environment variable to control SDK usage
    oldEnv := os.Getenv("MESH_USE_SDK")
    if useMeshSDK {
        os.Setenv("MESH_USE_SDK", "true")
    } else {
        os.Setenv("MESH_USE_SDK", "false") 
    }
    defer func() { os.Setenv("MESH_USE_SDK", oldEnv) }()

    // Load config to reflect SDK flag
    cfg := config.LoadConfig()
    cfg.MeshAPIURL = meshBaseURL // Override with mock server URL
    cfg.APIKey = "" // No API key for tests

    // Create mesh client based on flag
    var meshClient clients.MeshAPI
    if cfg.MeshUseSDK {
        meshClient = clients.NewMeshSDKClient(meshBaseURL)
    } else {
        meshClient = clients.NewMeshClient(meshBaseURL)
    }

    // Create mesh adapter
    meshAdapter := meshadapter.NewAdapter(meshClient)

    // Create service and router (coinbase and overledger adapters are nil for this test)
    service := connector.NewService(nil, meshAdapter, nil)
    router := api.SetupRouter(service, cfg)

    return httptest.NewServer(router)
}

// TestConcurrentMeshRequests validates the server handles concurrent mesh requests properly
func TestConcurrentMeshRequests(t *testing.T) {
    rosetta := startMockRosettaServer(t)
    defer rosetta.Close()

    server := startFullServerWithSDKFlag(t, rosetta.URL, true) // Use SDK client
    defer server.Close()

    const numRequests = 10
    var wg sync.WaitGroup
    results := make([]int, numRequests)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Send concurrent requests to /v1/mesh/networks
    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            
            req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/v1/mesh/networks", nil)
            if err != nil {
                results[idx] = -1
                return
            }

            resp, err := http.DefaultClient.Do(req)
            if err != nil || resp == nil {
                results[idx] = -1
                return
            }
            defer resp.Body.Close()
            results[idx] = resp.StatusCode
        }(i)
    }

    wg.Wait()

    // All requests should succeed
    for i, status := range results {
        assert.Equal(t, http.StatusOK, status, fmt.Sprintf("request %d should succeed", i))
    }
}