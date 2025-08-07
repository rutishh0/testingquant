package clients

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

// MeshClient is a lightweight HTTP client for interacting with Coinbase Mesh compliant API servers.
// NOTE: This initial scaffold intentionally avoids introducing the external mesh-sdk-go dependency so the project continues to build offline.
//       Once the remote module can be fetched, this file can either be removed or refactored to delegate to the official SDK.
//
// The Mesh API mostly uses POST requests with a JSON body consisting of `network_identifier`, `block_identifier`, etc. This
// client provides thin wrappers for common endpoints so the rest of the application can be migrated incrementally.

type MeshClient struct {
    BaseURL string
    Client  *http.Client
}

func NewMeshClient(baseURL string) *MeshClient {
    if baseURL == "" {
        baseURL = "https://mesh.coinbase.com" // Public reference implementation. Replace with appropriate host.
    }
    return &MeshClient{
        BaseURL: strings.TrimSuffix(baseURL, "/"),
        Client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// do executes an HTTP request against the Mesh API. The body value will be JSON encoded when non-nil.
func (m *MeshClient) do(method, path string, body interface{}) (*http.Response, error) {
    if !strings.HasPrefix(path, "/") {
        path = "/" + path
    }
    var reader io.Reader
    if body != nil {
        data, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request body: %w", err)
        }
        reader = bytes.NewReader(data)
    }
    req, err := http.NewRequest(method, m.BaseURL+path, reader)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := m.Client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    if resp.StatusCode >= 400 {
        defer resp.Body.Close()
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("mesh API error (%d): %s", resp.StatusCode, string(b))
    }
    return resp, nil
}

// ListNetworks maps to POST /network/list
func (m *MeshClient) ListNetworks() (*http.Response, error) {
    return m.do(http.MethodPost, "/network/list", map[string]interface{}{})
}

// NetworkStatus maps to POST /network/status
func (m *MeshClient) NetworkStatus(networkIdentifier interface{}, blockIdentifier interface{}) (*http.Response, error) {
    body := map[string]interface{}{
        "network_identifier": networkIdentifier,
    }
    if blockIdentifier != nil {
        body["block_identifier"] = blockIdentifier
    }
    return m.do(http.MethodPost, "/network/status", body)
}

// NetworkOptions maps to POST /network/options
func (m *MeshClient) NetworkOptions(networkIdentifier interface{}) (*http.Response, error) {
    body := map[string]interface{}{
        "network_identifier": networkIdentifier,
    }
    return m.do(http.MethodPost, "/network/options", body)
}

// AccountBalance maps to POST /account/balance
func (m *MeshClient) AccountBalance(networkIdentifier, accountIdentifier interface{}) (*http.Response, error) {
    body := map[string]interface{}{
        "network_identifier":  networkIdentifier,
        "account_identifier":  accountIdentifier,
    }
    return m.do(http.MethodPost, "/account/balance", body)
}

// Health checks the health of the mesh client
func (m *MeshClient) Health() bool {
	_, err := m.ListNetworks()
	return err == nil
}