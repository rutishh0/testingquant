package config

import (
    "os"
    "strconv"
)

// Config holds the application configuration
type Config struct {
	ServerAddress string
	APIKey        string
	Environment   string
	LogLevel      string

	// Coinbase API Configuration
	CoinbaseAPIKeyID  string
	CoinbaseAPISecret string
	CoinbaseAPIURL    string

    // Mesh API Configuration
    MeshAPIURL string

	// Overledger OAuth2 Configuration
	OverledgerClientID     string
	OverledgerClientSecret string
	OverledgerAuthURL      string
	OverledgerBaseURL      string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Railway/Koyeb provides PORT environment variable
	port := getEnv("PORT", "8080")
	if port[0] != ':' {
		port = ":" + port
	}

    // Derive a safe server address. Some platforms don't expand ${PORT} in
    // custom envs, so accept values like "${PORT}", "$PORT", plain numbers,
    // or an empty SERVER_ADDRESS and fallback to PORT.
    serverAddress := os.Getenv("SERVER_ADDRESS")
    if serverAddress == "" || serverAddress == "${PORT}" || serverAddress == "$PORT" || serverAddress == os.Getenv("PORT") {
        serverAddress = port
    } else {
        // If the value is a plain number like "8000", prepend a colon
        if _, err := strconv.Atoi(serverAddress); err == nil {
            serverAddress = ":" + serverAddress
        }
    }

    return &Config{
        ServerAddress: serverAddress,
		APIKey:        getEnv("API_KEY", ""),
		Environment:   getEnv("ENVIRONMENT", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),

		// Coinbase API Configuration - NO HARDCODED CREDENTIALS
		CoinbaseAPIKeyID:  getEnv("COINBASE_API_KEY_ID", ""),
		CoinbaseAPISecret: getEnv("COINBASE_API_SECRET", ""),
		CoinbaseAPIURL:    getEnv("COINBASE_API_URL", "https://api.coinbase.com"),

        // Mesh API Configuration
        MeshAPIURL:        getEnv("MESH_API_URL", "http://localhost:8080/mesh"),

		// Overledger OAuth2 Configuration - NO HARDCODED CREDENTIALS
		OverledgerClientID:     getEnv("OVERLEDGER_CLIENT_ID", ""),
		OverledgerClientSecret: getEnv("OVERLEDGER_CLIENT_SECRET", ""),
		OverledgerAuthURL:      getEnv("OVERLEDGER_AUTH_URL", "https://auth.overledger.dev/oauth2/token"),
		OverledgerBaseURL:      getEnv("OVERLEDGER_BASE_URL", "https://api.overledger.dev"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// HasCoinbaseCredentials returns true if Coinbase API credentials are configured
func (c *Config) HasCoinbaseCredentials() bool {
	return c.CoinbaseAPIKeyID != "" && c.CoinbaseAPISecret != ""
}

// HasOverledgerCredentials returns true if Overledger API credentials are configured
func (c *Config) HasOverledgerCredentials() bool {
	return c.OverledgerClientID != "" && c.OverledgerClientSecret != ""
}
