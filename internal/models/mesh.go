package models

// MeshNetwork represents a single network in the Mesh
type MeshNetwork struct {
	NetworkIdentifier struct {
		Blockchain string `json:"blockchain"`
		Network    string `json:"network"`
	} `json:"network_identifier"`
	Currency struct {
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"currency"`
}

// MeshNetworksResponse represents the response for a list of networks
type MeshNetworksResponse struct {
	Networks []MeshNetwork `json:"networks"`
}

// MeshBalance represents the balance of a single asset in an account
type MeshBalance struct {
	Value    string `json:"value"`
	Currency struct {
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"currency"`
}

// MeshBalanceResponse represents the response for an account's balance
type MeshBalanceResponse struct {
	Balances []MeshBalance `json:"balances"`
}