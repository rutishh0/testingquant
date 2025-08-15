package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type NetworkIdentifier struct {
	Blockchain string `json:"blockchain"`
	Network    string `json:"network"`
}

type MeshConfig struct {
	Network   NetworkIdentifier `json:"network"`
	OnlineURL string            `json:"online_url"`
	Data      DataConfig        `json:"data"`
}

type DataConfig struct {
	StartIndex int `json:"start_index"`
	EndIndex   int `json:"end_index"`
	// Add other data configuration fields as needed
}

type NetworkListRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
}

type NetworkStatusRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
}

type AccountBalanceRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	AccountIdentifier map[string]string `json:"account_identifier"`
}

type BlockRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	BlockIdentifier   *BlockIdentifier  `json:"block_identifier,omitempty"`
}

type BlockIdentifier struct {
	Index *int64  `json:"index,omitempty"`
	Hash  *string `json:"hash,omitempty"`
}

type ValidationConfig struct {
	Network NetworkIdentifier `json:"network"`
	BaseURL string            `json:"base_url"`
	Timeout time.Duration     `json:"timeout"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run mesh_config_validation.go <command> [config-file]")
		fmt.Println("Commands:")
		fmt.Println("  check:data      - Run data validation tests")
		fmt.Println("  check:config    - Validate mesh configuration file")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run mesh_config_validation.go check:data")
		fmt.Println("  go run mesh_config_validation.go check:data config/mesh-cli-config.json")
		os.Exit(1)
	}

	command := os.Args[1]
	configFile := "config/mesh-cli-config.json"

	if len(os.Args) > 2 {
		configFile = os.Args[2]
	}

	switch command {
	case "check:data":
		runDataValidation(configFile)
	case "check:config":
		runConfigValidation(configFile)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func loadConfig(configFile string) (*MeshConfig, error) {
	// Try to find config file in multiple locations
	possiblePaths := []string{
		configFile,
		filepath.Join("config", "mesh-cli-config.json"),
		"mesh-cli-config.json",
	}

	var configData []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		configData, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not find config file in any of these locations: %v", possiblePaths)
	}

	fmt.Printf("📋 Using config file: %s\n", foundPath)

	var config MeshConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func runConfigValidation(configFile string) {
	fmt.Println("🔧 Starting Mesh Configuration Validation...")

	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Config validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Configuration loaded successfully\n")
	fmt.Printf("   Network: %s/%s\n", config.Network.Blockchain, config.Network.Network)
	fmt.Printf("   Online URL: %s\n", config.OnlineURL)
	fmt.Printf("   Data range: %d - %d\n", config.Data.StartIndex, config.Data.EndIndex)
	fmt.Println("🎉 Configuration validation passed!")
}

func runDataValidation(configFile string) {
	fmt.Println("🚀 Starting Mesh API Data Validation...")

	config, err := loadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	validationConfig := ValidationConfig{
		Network: config.Network,
		BaseURL: config.OnlineURL, // Use the configured online_url directly
		Timeout: 10 * time.Second,
	}

	fmt.Printf("   Network: %s/%s\n", validationConfig.Network.Blockchain, validationConfig.Network.Network)
	fmt.Printf("   Base URL: %s\n", validationConfig.BaseURL)
	fmt.Println()

	tests := []struct {
		name string
		fn   func(ValidationConfig) error
	}{
		{"Network List", testNetworkList},
		{"Network Status", testNetworkStatus},
		{"Network Options", testNetworkOptions},
		{"Account Balance", testAccountBalance},
		{"Block Retrieval", testBlock},
	}

	passed := 0
	total := len(tests)

	for _, test := range tests {
		fmt.Printf("📋 Running %s test...", test.name)
		if err := test.fn(validationConfig); err != nil {
			fmt.Printf(" ❌ FAILED\n")
			fmt.Printf("   Error: %v\n", err)
		} else {
			fmt.Printf(" ✅ PASSED\n")
			passed++
		}
	}

	fmt.Println()
	fmt.Printf("📊 Results: %d/%d tests passed\n", passed, total)
	if passed == total {
		fmt.Println("🎉 All validation tests passed!")
	} else {
		fmt.Printf("❌ %d tests failed\n", total-passed)
		os.Exit(1)
	}
}

func testNetworkList(config ValidationConfig) error {
	req := NetworkListRequest{
		NetworkIdentifier: config.Network,
	}

	resp, err := makeRequest(config.BaseURL+"/network/list", req, config.Timeout)
	if err != nil {
		return fmt.Errorf("network list request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse network list response: %w", err)
	}

	// Verify response contains network_identifiers
	if _, ok := result["network_identifiers"]; !ok {
		return fmt.Errorf("response missing network_identifiers field")
	}

	return nil
}

func testNetworkStatus(config ValidationConfig) error {
	req := NetworkStatusRequest{
		NetworkIdentifier: config.Network,
	}

	resp, err := makeRequest(config.BaseURL+"/network/status", req, config.Timeout)
	if err != nil {
		return fmt.Errorf("network status request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse network status response: %w", err)
	}

	// Verify response contains required fields
	requiredFields := []string{"current_block_identifier", "current_block_timestamp", "genesis_block_identifier"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			return fmt.Errorf("response missing required field: %s", field)
		}
	}

	return nil
}

func testNetworkOptions(config ValidationConfig) error {
	req := NetworkListRequest{
		NetworkIdentifier: config.Network,
	}

	resp, err := makeRequest(config.BaseURL+"/network/options", req, config.Timeout)
	if err != nil {
		return fmt.Errorf("network options request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse network options response: %w", err)
	}

	// Verify response contains version and operation_types
	requiredFields := []string{"version", "allow"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			return fmt.Errorf("response missing required field: %s", field)
		}
	}

	return nil
}

func testAccountBalance(config ValidationConfig) error {
	req := AccountBalanceRequest{
		NetworkIdentifier: config.Network,
		AccountIdentifier: map[string]string{
			"address": "test-address",
		},
	}

	resp, err := makeRequest(config.BaseURL+"/account/balance", req, config.Timeout)
	if err != nil {
		return fmt.Errorf("account balance request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse account balance response: %w", err)
	}

	// Verify response contains balances
	if _, ok := result["balances"]; !ok {
		return fmt.Errorf("response missing balances field")
	}

	return nil
}

func testBlock(config ValidationConfig) error {
	// Test with block index 0 (genesis block)
	index := int64(0)
	req := BlockRequest{
		NetworkIdentifier: config.Network,
		BlockIdentifier: &BlockIdentifier{
			Index: &index,
		},
	}

	resp, err := makeRequest(config.BaseURL+"/block", req, config.Timeout)
	if err != nil {
		return fmt.Errorf("block request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse block response: %w", err)
	}

	// Verify response contains block
	if _, ok := result["block"]; !ok {
		return fmt.Errorf("response missing block field")
	}

	return nil
}

func makeRequest(url string, payload interface{}, timeout time.Duration) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
