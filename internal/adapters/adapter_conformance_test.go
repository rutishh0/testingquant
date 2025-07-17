package adapters_test

import (
    "testing"
    "path/filepath"

    core "github.com/your-username/quant-mesh-connector/internal/core"
)

// TestConnectorIDsUnique ensures each adapter registers with a non-empty, unique ID.
func TestConnectorIDsUnique(t *testing.T) {
    conns := core.All()
    if len(conns) == 0 {
        t.Fatal("no connectors registered – ensure blank imports in cmd/main.go")
    }
    seen := map[string]struct{}{}
    for _, c := range conns {
        id := c.ID()
        if id == "" {
            t.Fatalf("connector %T returned empty ID", c)
        }
        if _, dup := seen[id]; dup {
            t.Fatalf("duplicate connector ID detected: %s", id)
        }
        seen[id] = struct{}{}
    }
}

// TestConnectorsYAML parses the default connectors.yaml present at project root.
func TestConnectorsYAML(t *testing.T) {
    // Resolve path relative to this test file (internal/adapters)
    root := filepath.Join("..", "..")
    path := filepath.Join(root, "connectors.yaml")
    if _, err := core.LoadConnectorConfigs(path); err != nil {
        t.Fatalf("failed to parse connectors.yaml: %v", err)
    }
}
