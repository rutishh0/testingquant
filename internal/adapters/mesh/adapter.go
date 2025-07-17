package meshadapter

import (
    "errors"
    "time"

    core "github.com/your-username/quant-mesh-connector/internal/core"
    "github.com/your-username/quant-mesh-connector/internal/mesh"
)

// Adapter implements the core.Connector interface for Coinbase Mesh API.
type Adapter struct {
    network string       // cached network identifier from Init optionally
    client  *mesh.Client // Mesh API client
}

// ID returns the unique identifier for this connector implementation.
func (a *Adapter) ID() string { return "mesh" }

// Init bootstraps the adapter using a generic config map. Expected keys:
//   base_url: string – HTTP(S) endpoint of the Mesh API
func (a *Adapter) Init(cfg map[string]any) error {
    urlAny, ok := cfg["base_url"]
    if !ok {
        return errors.New("mesh adapter: missing base_url in config")
    }
    baseURL, ok := urlAny.(string)
    if !ok {
        return errors.New("mesh adapter: base_url must be a string")
    }
    a.client = mesh.NewClient(baseURL)

    // Perform a quick health check so we fail fast during startup.
    // Cache network if provided for Receive polling
    if n, ok := cfg["network"].(string); ok {
        a.network = n
    }
    return a.client.Health()
}

// HealthCheck verifies connectivity with the remote Mesh node.
func (a *Adapter) HealthCheck() error {
    if a.client == nil {
        return errors.New("mesh adapter: not initialised")
    }
    return a.client.Health()
}

// Send submits a core.Message to Mesh.
// The Message.Payload must be a map with:
//   - "network":     string blockchain network (e.g., "ethereum")
//   - "signed_tx":   string hex-encoded signed transaction
func (a *Adapter) Send(msg *core.Message) (*core.Tx, error) {
    if a.client == nil {
        return nil, errors.New("mesh adapter: not initialised")
    }
    payload, ok := msg.Payload.(map[string]any)
    if !ok {
        return nil, errors.New("mesh adapter: payload must be a map[string]any")
    }
    signed, ok := payload["signed_tx"].(string)
    if !ok || signed == "" {
        return nil, errors.New("mesh adapter: payload missing signed_tx")
    }
    netName := a.network
    if n, ok := payload["network"].(string); ok && n != "" {
        netName = n
    }
    if netName == "" {
        return nil, errors.New("mesh adapter: network not specified")
    }
    req := &mesh.ConstructionSubmitRequest{
        NetworkIdentifier: mesh.NetworkIdentifier{Blockchain: "ethereum", Network: netName},
        SignedTransaction: signed,
    }
    resp, err := a.client.ConstructionSubmit(req)
    if err != nil {
        return nil, err
    }
    return &core.Tx{
        Hash:   resp.TransactionIdentifier.Hash,
        Status: "submitted",
        Raw:    resp,
    }, nil
}

// Receive returns a stub confirmation event. A production implementation would poll
// the Mesh `/transaction/status` endpoint or subscribe to websockets.
func (a *Adapter) Receive(txID string) (*core.Event, error) {
    return &core.Event{
        Type:      "submission_ack",
        Data:      map[string]string{"tx": txID},
        Timestamp: time.Now().Unix(),
    }, nil
}

// Ensure registration at package init.
func init() {
    core.Register(&Adapter{})
}
