package mesh

import (
	"encoding/json"
	"github.com/rutishh0/testingquant/internal/clients"
	"github.com/rutishh0/testingquant/internal/models"
)

// Adapter defines the interface for mesh client operations
type Adapter interface {
	ListNetworks() (*models.MeshNetworksResponse, error)
	AccountBalance(networkIdentifier, accountIdentifier interface{}) (*models.MeshBalanceResponse, error)
	// New: block and transaction retrieval
	Block(networkIdentifier, blockIdentifier interface{}) (map[string]interface{}, error)
	BlockTransaction(networkIdentifier, blockIdentifier, transactionIdentifier interface{}) (map[string]interface{}, error)
	Health() bool
}

type meshAdapter struct {
    client clients.MeshAPI
}

// NewAdapter creates a new mesh adapter
func NewAdapter(client clients.MeshAPI) Adapter {
    return &meshAdapter{client: client}
}

func (a *meshAdapter) ListNetworks() (*models.MeshNetworksResponse, error) {
	resp, err := a.client.ListNetworks()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode Rosetta NetworkListResponse and map to internal model with currency metadata
	var rosettaResp struct {
		NetworkIdentifiers []struct {
			Blockchain string `json:"blockchain"`
			Network    string `json:"network"`
		} `json:"network_identifiers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rosettaResp); err != nil {
		return nil, err
	}

	networks := make([]models.MeshNetwork, 0, len(rosettaResp.NetworkIdentifiers))
	for _, ni := range rosettaResp.NetworkIdentifiers {
		var mn models.MeshNetwork
		mn.NetworkIdentifier.Blockchain = ni.Blockchain
		mn.NetworkIdentifier.Network = ni.Network
		// Provide sensible defaults for currency metadata for display
		mn.Currency.Symbol = "ETH"
		mn.Currency.Decimals = 18
		networks = append(networks, mn)
	}

	return &models.MeshNetworksResponse{Networks: networks}, nil
}

func (a *meshAdapter) AccountBalance(networkIdentifier, accountIdentifier interface{}) (*models.MeshBalanceResponse, error) {
	resp, err := a.client.AccountBalance(networkIdentifier, accountIdentifier)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var balance models.MeshBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, err
	}
	return &balance, nil
}

// Block returns the Rosetta BlockResponse as a generic map to keep adapter layer decoupled from SDK types
func (a *meshAdapter) Block(networkIdentifier, blockIdentifier interface{}) (map[string]interface{}, error) {
	resp, err := a.client.Block(networkIdentifier, blockIdentifier)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// BlockTransaction returns the Rosetta BlockTransactionResponse as a generic map
func (a *meshAdapter) BlockTransaction(networkIdentifier, blockIdentifier, transactionIdentifier interface{}) (map[string]interface{}, error) {
	resp, err := a.client.BlockTransaction(networkIdentifier, blockIdentifier, transactionIdentifier)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *meshAdapter) Health() bool {
	return a.client.Health()
}