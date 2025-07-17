package overledgeradapter

import (
    "errors"
    "strings"
    "time"

    core "github.com/your-username/quant-mesh-connector/internal/core"
    "github.com/your-username/quant-mesh-connector/internal/overledger"
    "github.com/your-username/quant-mesh-connector/internal/config"
)

// Adapter implements the core.Connector interface for Quant Overledger API.
type Adapter struct {
    client *overledger.Client
}

// ID returns the unique identifier for this connector implementation.
func (a *Adapter) ID() string { return "overledger" }

// Init bootstraps the adapter using a generic config map. Expected keys:
//   base_url         : string – Overledger gateway URL
//   tls_skip_verify? : bool   – optional, skip TLS verification for dev
//   client_id        : string – Overledger client ID
//   client_secret    : string – Overledger client secret
//   auth_url         : string – Overledger auth URL
func (a *Adapter) Init(cfg map[string]any) error {
    baseURLAny, ok := cfg["base_url"]
    if !ok {
        return errors.New("overledger adapter: missing base_url in config")
    }
    baseURL, ok := baseURLAny.(string)
    if !ok {
        return errors.New("overledger adapter: base_url must be a string")
    }


    // Build minimal config.Config for Overledger client
    olCfg := &config.Config{
        OverledgerBaseURL:      baseURL,
        OverledgerClientID:     cfgValueString(cfg, "client_id"),
        OverledgerClientSecret: cfgValueString(cfg, "client_secret"),
        OverledgerAuthURL:      cfgValueString(cfg, "auth_url"),
    }

    a.client = overledger.NewClient(olCfg)

    return a.client.TestConnection()
}

// HealthCheck verifies connectivity with Overledger.
func (a *Adapter) HealthCheck() error {
    if a.client == nil {
        return errors.New("overledger adapter: not initialised")
    }
    return a.client.TestConnection()
}

// Send submits a core.Message to Overledger.
// Expected Message.Payload structure (map[string]any):
//   network_id   : string – target network identifier understood by Overledger
//   from_address : string – sender address
//   to_address   : string – recipient address
//   amount       : string – amount in minimal unit (wei, satoshi, etc.)
//   token_id?    : string – optional token identifier
func (a *Adapter) Send(msg *core.Message) (*core.Tx, error) {
    if a.client == nil {
        return nil, errors.New("overledger adapter: not initialised")
    }
    payload, ok := msg.Payload.(map[string]any)
    if !ok {
        return nil, errors.New("overledger adapter: payload must be a map[string]any")
    }
    networkID, ok := payload["network_id"].(string)
    if !ok || networkID == "" {
        return nil, errors.New("overledger adapter: missing network_id in payload")
    }
    fromAddr, ok := payload["from_address"].(string)
    if !ok || fromAddr == "" {
        return nil, errors.New("overledger adapter: missing from_address in payload")
    }
    toAddr, ok := payload["to_address"].(string)
    if !ok || toAddr == "" {
        return nil, errors.New("overledger adapter: missing to_address in payload")
    }
    amount, ok := payload["amount"].(string)
    if !ok || amount == "" {
        return nil, errors.New("overledger adapter: missing amount in payload")
    }
    req := &overledger.TransactionRequest{
        NetworkID:   networkID,
        FromAddress: fromAddr,
        ToAddress:   toAddr,
        Amount:      amount,
    }
    if token, ok := payload["token_id"].(string); ok && token != "" {
        req.TokenID = token
    }
    resp, err := a.client.CreateTransaction(req)
    if err != nil {
        return nil, err
    }
    return &core.Tx{
        Hash:   resp.Hash,
        Status: resp.Status,
        Raw:    resp,
    }, nil
}

// Receive waits for a confirmation or related event.
func (a *Adapter) Receive(txID string) (*core.Event, error) {
    if a.client == nil {
        return nil, errors.New("overledger adapter: not initialised")
    }
    // For Receive we need network ID; assume caller passes it via txID as "networkID:hash"
    parts := strings.SplitN(txID, ":", 2)
    if len(parts) != 2 {
        return nil, errors.New("overledger adapter: txID must be in format <networkID>:<hash>")
    }
    networkID, hash := parts[0], parts[1]
    statusResp, err := a.client.GetTransactionStatus(networkID, hash)
    if err != nil {
        return nil, err
    }
    return &core.Event{
        Type:      "tx_status",
        Data:      statusResp,
        Timestamp: time.Now().Unix(),
    }, nil
}

// cfgValueString safely extracts a string value from a generic map.
func cfgValueString(m map[string]any, key string) string {
    if v, ok := m[key]; ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

// Automatic registration.
func init() {
    core.Register(&Adapter{})
}
