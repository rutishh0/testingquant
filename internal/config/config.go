package config

import (
	"os"
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

	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", port),
		APIKey:        getEnv("API_KEY", "development-api-key-12345678901234567890"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),

		// Coinbase API Configuration
		CoinbaseAPIKeyID:  getEnv("COINBASE_API_KEY_ID", "64b6c1ef-d093-42e7-b811-4cc19a342696"),
		CoinbaseAPISecret: getEnv("COINBASE_API_SECRET", "32IrsqS1B8PcoN+iy+P/i3mNlZzXFPY66zLtJV/xsDQ+WphMqeOTTUSlb05FbaM9itHkZAjIzR3dPjgyl3ZiCg=="),
		CoinbaseAPIURL:    getEnv("COINBASE_API_URL", "https://api.cdp.coinbase.com"),

		// Overledger OAuth2 Configuration
		OverledgerClientID:     getEnv("OVERLEDGER_CLIENT_ID", "3nhqpst935v0kqumc3s76jcq46"),
		OverledgerClientSecret: getEnv("OVERLEDGER_CLIENT_SECRET", "102l0eabrqum4pald0mv7l0o6e25i73utvn9htdv0rjjkusrblje"),
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
