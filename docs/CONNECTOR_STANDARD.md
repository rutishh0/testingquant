# Quant-Mesh Connector Standard

This document defines the canonical contract that every connector **(adapter)** must satisfy in order to plug into the Quant-Mesh interoperability framework.

It is intentionally minimal, technology-agnostic, and suitable for future ISO standardisation.

---

## 1. Core Abstractions (Go types)

```
// Message – generic request/intent sent to a network (Tx construction, smart-contract call, etc.)
type Message struct {
  ID       string         `json:"id,omitempty"`  // optional client-supplied reference
  Payload  any            `json:"payload"`       // free-form payload understood by the adapter
  Metadata map[string]any `json:"metadata,omitempty"`
}

// Tx – canonical network transaction representation
type Tx struct {
  Hash     string         `json:"hash"`          // unique identifier / hash
  Status   string         `json:"status"`        // pending | submitted | confirmed | failed
  Raw      any            `json:"raw,omitempty"` // concrete receipt / response
  Metadata map[string]any `json:"metadata,omitempty"`
}

// Event – any observation worth surfacing (blocks, tx status, log entry, …)
type Event struct {
  Type      string `json:"type"`
  Data      any    `json:"data"`
  Timestamp int64  `json:"timestamp"` // unix epoch (seconds)
}

// Connector – the mandatory interface each adapter implements
// NOTE: all **public** behaviour is captured here; no hidden expectations.

type Connector interface {
  ID() string                                // globally-unique, URL-safe identifier (e.g. "mesh", "overledger")
  Init(cfg map[string]any) error             // bootstrap using arbitrary key/val map (unmarshalled from YAML/JSON)
  HealthCheck() error                        // liveness check (fail fast on startup / probes)
  Send(msg *Message) (*Tx, error)            // translate + submit Message -> Tx
  Receive(txID string) (*Event, error)       // fetch or subscribe to event(s) given a tx hash / opaque ref
}
```

Adapters register themselves automatically:

```go
func init() {
  core.Register(&MyConnector{})
}
```

---

## 2. YAML Configuration (`connectors.yaml`)

The application boots with a single YAML file at project root. Each **top-level key** equals the adapter `ID()`. Values are arbitrary and passed verbatim to `Init`.

```yaml
mesh:
  base_url: "http://localhost:8081"    # Rosetta/ Mesh endpoint
  network:  "mainnet"                 # optional default network

overledger:
  base_url:      "https://api.overledger.dev"
  client_id:     "..."
  client_secret: "..."
  auth_url:      "https://auth.overledger.dev/oauth2/token"
```

Adapters **must NOT** access environment variables directly – all runtime configuration flows through this map.

---

## 3. Send / Receive Contract

1. **Mesh Adapter**
   • `Send` expects `Message.Payload` = `map[string]any{ "network": "ethereum", "signed_tx": "0xf9…" }`  
   • Translates to Rosetta `/construction/submit` and returns `Tx{Hash, Status="submitted"}`.  
   • `Receive` – for PoC returns `submission_ack` event; production should poll `/transaction/status` or use WS.

2. **Overledger Adapter**
   • `Send` payload map:
     ```yaml
     network_id:   "ethereum-mainnet"
     from_address: "0xabc…"
     to_address:   "0xdef…"
     amount:       "1000000000000000000" # wei
     token_id:     "optional ERC20/721 ID"
     ```
   • Calls `POST /v2/networks/{id}/transactions` and maps response -> `Tx`.  
   • `Receive(txID)` expects format `networkID:hash` and proxy-fetches `/transactions/{hash}/status`.

Adapters MAY extend the payload contract but **must document** any extra fields.

---

## 4. Conformance Test Suite

Located under `internal/adapters/`. Every new adapter must satisfy:

* Unique, non-empty `ID()` (see `TestConnectorIDsUnique`).
* `connectors.yaml` must parse with the adapter’s section present (`TestConnectorsYAML`).
* Future: standardised behaviour mocks (happy path Send/Receive via httpmock / httptest).

Run:

```bash
go test ./internal/adapters/...
```

---

## 5. Registry & Runtime

`cmd/main.go`:
1. Loads `connectors.yaml` (fallback to env vars).  
2. Loops through map, finds connectors via `core.Get(id)`.  
3. Calls `Init` + `HealthCheck` on each – application fails fast if any adapter misbehaves.

At runtime you can discover available connectors via `core.All()` or expose an API endpoint.

---

## 6. Versioning & ISO Path

* Semantic version this document (`v0.1.0`).  
* Each breaking change MUST bump major and update adapters/tests accordingly.  
* For ISO submission include:
  1. This specification
  2. Reference Go implementation (current repo)
  3. Conformance test results across at least two networks

---

## 7. Glossary

| Term        | Description                                     |
|-------------|-------------------------------------------------|
| **Adapter** | Concrete implementation of `Connector`          |
| **Message** | Abstract request/intent sent to network         |
| **Tx**      | Canonical transaction representation            |
| **Event**   | Observation reported back by network            |

---

*Last updated: 2025-07-17*
