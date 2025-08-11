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
	Health() bool
}

type meshAdapter struct {
    client *clients.MeshClient
}

// NewAdapter creates a new mesh adapter
func NewAdapter(client *clients.MeshClient) Adapter {
    return &meshAdapter{client: client}
}

func (a *meshAdapter) ListNetworks() (*models.MeshNetworksResponse, error) {
	resp, err := a.client.ListNetworks()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var networks models.MeshNetworksResponse
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, err
	}
	return &networks, nil
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

func (a *meshAdapter) Health() bool {
	return a.client.Health()
}