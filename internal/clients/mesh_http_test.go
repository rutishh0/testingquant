package clients

import (
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

// helper to decode JSON body into a generic map
func decodeBody(t *testing.T, r *http.Request) map[string]any {
    t.Helper()
    b, err := io.ReadAll(r.Body)
    if err != nil {
        t.Fatalf("failed reading body: %v", err)
    }
    defer r.Body.Close()
    var m map[string]any
    if len(b) > 0 {
        if err := json.Unmarshal(b, &m); err != nil {
            t.Fatalf("invalid JSON body: %v; raw=%s", err, string(b))
        }
    } else {
        m = map[string]any{}
    }
    return m
}

func TestMeshClient_Block_NilBlockIdentifierSendsEmptyObject(t *testing.T) {
    var captured map[string]any
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/block" && r.Method == http.MethodPost {
            captured = decodeBody(t, r)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{}`))
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    client := NewMeshClient(srv.URL)
    networkID := map[string]any{"blockchain": "Ethereum", "network": "Sepolia"}

    resp, err := client.Block(networkID, nil)
    if err != nil {
        t.Fatalf("Block returned error: %v", err)
    }
    resp.Body.Close()

    if captured == nil {
        t.Fatalf("server did not capture request body")
    }

    // Ensure block_identifier exists and is an empty object
    bi, ok := captured["block_identifier"].(map[string]any)
    if !ok {
        t.Fatalf("expected block_identifier to be object, got: %T (%v)", captured["block_identifier"], captured["block_identifier"])
    }
    if len(bi) != 0 {
        t.Fatalf("expected empty block_identifier object, got: %v", bi)
    }

    // Ensure network_identifier is passed through
    ni, ok := captured["network_identifier"].(map[string]any)
    if !ok || ni["blockchain"] != "Ethereum" || ni["network"] != "Sepolia" {
        t.Fatalf("unexpected network_identifier: %v", ni)
    }
}

func TestMeshClient_BlockTransaction_WithValues(t *testing.T) {
    var captured map[string]any
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/block/transaction" && r.Method == http.MethodPost {
            captured = decodeBody(t, r)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            _, _ = w.Write([]byte(`{}`))
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    client := NewMeshClient(srv.URL)
    networkID := map[string]any{"blockchain": "Ethereum", "network": "Sepolia"}
    blockID := map[string]any{"index": 1}
    txID := map[string]any{"hash": "0xabc"}

    resp, err := client.BlockTransaction(networkID, blockID, txID)
    if err != nil {
        t.Fatalf("BlockTransaction returned error: %v", err)
    }
    resp.Body.Close()

    if captured == nil {
        t.Fatalf("server did not capture request body")
    }

    // Validate fields presence and values
    if _, ok := captured["network_identifier"].(map[string]any); !ok {
        t.Fatalf("network_identifier missing or wrong type: %T", captured["network_identifier"])
    }
    bi, ok := captured["block_identifier"].(map[string]any)
    if !ok {
        t.Fatalf("block_identifier missing or wrong type: %T", captured["block_identifier"])
    }
    if v, vok := bi["index"].(float64); !vok || v != 1 {
        t.Fatalf("expected block_identifier.index == 1, got: %v (%T)", bi["index"], bi["index"])
    }
    ti, ok := captured["transaction_identifier"].(map[string]any)
    if !ok {
        t.Fatalf("transaction_identifier missing or wrong type: %T", captured["transaction_identifier"])
    }
    if ti["hash"] != "0xabc" {
        t.Fatalf("unexpected transaction_identifier.hash: %v", ti["hash"])
    }
}

func TestMeshClient_ErrorStatusIncludesBody(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/network/options" && r.Method == http.MethodPost {
            w.WriteHeader(http.StatusInternalServerError)
            _, _ = w.Write([]byte("boom"))
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer srv.Close()

    client := NewMeshClient(srv.URL)
    networkID := map[string]any{"blockchain": "Ethereum", "network": "Sepolia"}

    resp, err := client.NetworkOptions(networkID)
    if err == nil || !strings.Contains(err.Error(), "mesh API error (500): boom") {
        if err == nil {
            if resp != nil { resp.Body.Close() }
        }
        t.Fatalf("expected 500 error including body, got: %v", err)
    }
}

func TestMeshClient_RequestMarshalError(t *testing.T) {
    client := NewMeshClient("http://127.0.0.1:0") // baseURL won't be used because marshal fails first

    // Introduce an unmarshalable value (func) in the body via parameters
    bad := func() {}
    resp, err := client.AccountBalance(bad, map[string]any{"address": "0xdead"})
    if resp != nil { resp.Body.Close() }
    if err == nil || !strings.Contains(err.Error(), "failed to marshal request body") {
        t.Fatalf("expected marshal error, got: %v", err)
    }
}

func TestMeshClient_Health_PositiveAndNegative(t *testing.T) {
    healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/network/list" && r.Method == http.MethodPost {
            w.WriteHeader(http.StatusOK)
            _, _ = w.Write([]byte(`{"network_identifiers":[]}`))
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer healthy.Close()

    unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/network/list" && r.Method == http.MethodPost {
            w.WriteHeader(http.StatusInternalServerError)
            _, _ = w.Write([]byte("fail"))
            return
        }
        w.WriteHeader(http.StatusNotFound)
    }))
    defer unhealthy.Close()

    hc := NewMeshClient(healthy.URL)
    if !hc.Health() {
        t.Fatalf("expected Health() to be true against healthy server")
    }

    uc := NewMeshClient(unhealthy.URL)
    if uc.Health() {
        t.Fatalf("expected Health() to be false against unhealthy server")
    }
}