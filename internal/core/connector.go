package core

// Message is a generic representation of data exchanged between networks/models.
// It can contain structured transfer data, smart-contract payloads, etc.
// Concrete connectors may embed protocol-specific metadata inside the "Payload" field.
type Message struct {
    // Unique identifier within the originating network.
    ID string `json:"id,omitempty"`
    // Free-form payload – can be JSON, RLP, protobuf bytes, etc.
    Payload any `json:"payload"`
    // Optional headers / metadata shared across networks.
    Metadata map[string]any `json:"metadata,omitempty"`
}

// Tx is the canonical transaction abstraction visible to the framework.
// Each connector must map its underlying network transaction into this shape.
type Tx struct {
    // Network-unique transaction reference (hash / ID).
    Hash string `json:"hash"`
    // Human-readable status ("pending", "confirmed", "failed", etc.).
    Status string `json:"status"`
    // Raw network-specific receipt / payload if useful for higher layers.
    Raw any `json:"raw,omitempty"`
    // Additional arbitrary metadata.
    Metadata map[string]any `json:"metadata,omitempty"`
}

// Event is an observation produced by a network (new block, payment reception, etc.).
// Merged into a simple, composable struct so higher layers can reason generically.
type Event struct {
    Type string `json:"type"`
    Data any    `json:"data"`
    // Unix epoch when the event occurred (according to the producing connector).
    Timestamp int64 `json:"timestamp"`
}

// Connector is the minimal contract every adapter must implement to plug into the framework.
// It purposefully avoids business-logic terms like "Preprocess" so it can generalise across
// payments, messaging, or ML model invocations in the future.
type Connector interface {
    // ID returns the globally unique identifier of this connector (e.g., "ethereum-mainnet", "overledger", "mesh").
    ID() string

    // Init allows the connector to bootstrap – load credentials, open sockets, etc.
    // The config map is unmarshalled from YAML/JSON supplied by the user.
    Init(config map[string]any) error

    // HealthCheck verifies the connector can currently reach its backend / chain.
    HealthCheck() error

    // Send submits a Message to the underlying network and returns the resulting Tx abstraction.
    Send(msg *Message) (*Tx, error)

    // Receive fetches or subscribes to an application-level event / transaction confirmation.
    Receive(txID string) (*Event, error)
}

// registry holds all connectors registered at init() time.
var registry = make(map[string]Connector)

// Register makes a connector discoverable by the framework. Call from an init() func in the adapter.
// Duplicate registration panics – this happens during application startup so it is preferable to fail fast.
func Register(c Connector) {
    if c == nil {
        panic("core: Register connector is nil")
    }
    id := c.ID()
    if id == "" {
        panic("core: connector ID cannot be empty")
    }
    if _, dup := registry[id]; dup {
        panic("core: Register called twice for connector " + id)
    }
    registry[id] = c
}

// Get returns a connector by ID and a boolean indicating existence.
func Get(id string) (Connector, bool) {
    c, ok := registry[id]
    return c, ok
}

// All returns a slice of currently registered connectors. Useful for discovery endpoints.
func All() []Connector {
    conns := make([]Connector, 0, len(registry))
    for _, c := range registry {
        conns = append(conns, c)
    }
    return conns
}
