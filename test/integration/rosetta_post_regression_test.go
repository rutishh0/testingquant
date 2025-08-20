package integration

import (
    "testing"
    "net/http"

    "github.com/stretchr/testify/require"
)

// TestRosettaNetworkListRequiresPOST ensures the Rosetta /network/list endpoint
// requires POST and does not accept GET, preventing regressions where the wrong
// method might be used.
func TestRosettaNetworkListRequiresPOST(t *testing.T) {
    srv := startMockRosettaServer(t)
    defer srv.Close()

    // GET should NOT succeed (Rosetta requires POST)
    status, _ := getJSON(t, srv.URL, "/network/list")
    require.NotEqual(t, http.StatusOK, status, "GET /network/list should not return 200")

    // POST should succeed
    status, _ = postJSON(t, srv.URL, "/network/list", map[string]any{})
    require.Equal(t, http.StatusOK, status, "POST /network/list should return 200")
}