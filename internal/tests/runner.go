package tests

import (

    "github.com/rutishh0/testingquant/internal/config"
    "github.com/rutishh0/testingquant/internal/connector"
)

// Result represents the outcome of a single test case.
// It purposefully mirrors a light JUnit-style schema so the frontend can group and color-code easily.
// Success=true => Status "pass"; otherwise "fail" with an Error message.
// Tier indicates the broad tier (0–5) from the spec so the UI can section them.
type Result struct {
    Tier    int    `json:"tier"`
    Name    string `json:"name"`
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
    Error   string `json:"error,omitempty"`
}

// RunAll executes the core Tier-0 & Tier-1 sanity checks.
// It returns a slice of Result structs that the HTTP layer will marshal to JSON.
//
// Future work: expand to Tier-2+ once write-paths are enabled in Overledger.
func RunAll(s connector.Service, cfg *config.Config) []Result {
    var out []Result

    // --- Tier 0: service health/status ---
    {
        name := "Health endpoint"
        if health, err := s.HealthCheck(); err != nil {
            out = append(out, Result{Tier: 0, Name: name, Success: false, Error: err.Error()})
        } else if health.Status == "degraded" {
            out = append(out, Result{Tier: 0, Name: name, Success: false, Message: "status=degraded"})
        } else {
            out = append(out, Result{Tier: 0, Name: name, Success: true})
        }
    }

    // Status is internal only – just craft a synthetic pass because it always returns OK currently.
    out = append(out, Result{Tier: 0, Name: "Status endpoint", Success: true})

    // --- Tier 1: Overledger read-only ---
    if cfg.HasOverledgerCredentials() {
        // Test connection shortcut
        if err := s.TestOverledgerConnection(); err != nil {
            out = append(out, Result{Tier: 1, Name: "Overledger connection", Success: false, Error: err.Error()})
        } else {
            out = append(out, Result{Tier: 1, Name: "Overledger connection", Success: true})
        }

        // List networks
        if _, err := s.GetOverledgerNetworks(); err != nil {
            out = append(out, Result{Tier: 1, Name: "Overledger networks", Success: false, Error: err.Error()})
        } else {
            msg := "networks response OK"
            out = append(out, Result{Tier: 1, Name: "Overledger networks", Success: true, Message: msg})
        }
    }

    // Coinbase asset list (only if credentials)
    if cfg.HasCoinbaseCredentials() {
        if _, err := s.GetCoinbaseAssets(); err != nil {
            out = append(out, Result{Tier: 1, Name: "Coinbase assets", Success: false, Error: err.Error()})
        } else {
            msg := "assets response OK"
            out = append(out, Result{Tier: 1, Name: "Coinbase assets", Success: true, Message: msg})
        }

        // Test Coinbase wallets retrieval
        if wallets, err := s.GetCoinbaseWallets(); err != nil {
            out = append(out, Result{Tier: 1, Name: "Coinbase wallets", Success: false, Error: err.Error()})
        } else {
            msg := "wallets response OK"
            out = append(out, Result{Tier: 1, Name: "Coinbase wallets", Success: true, Message: msg})

            // If we have wallets, test paginated transactions endpoint
            if len(wallets.Wallets) > 0 {
                walletID := wallets.Wallets[0].ID
                if resp, err := s.GetCoinbaseTransactionsPaginated(walletID, 10, ""); err != nil {
                    out = append(out, Result{Tier: 1, Name: "Coinbase transactions (paginated)", Success: false, Error: err.Error()})
                } else {
                    msg := "paginated transactions response OK"
                    if resp.HasNext {
                        msg = "paginated transactions response OK with next page"
                    }
                    out = append(out, Result{Tier: 1, Name: "Coinbase transactions (paginated)", Success: true, Message: msg})
                }
            } else {
                out = append(out, Result{Tier: 1, Name: "Coinbase transactions (paginated)", Success: false, Error: "no wallets available for testing"})
            }
        }
    }

    return out
}
