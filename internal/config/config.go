package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	ServerAddress string
	MeshAPIURL    string
	APIKey        string
	Environment   string
	LogLevel      string
	
	// Overledger OAuth2 Configuration
	OverledgerClientID     string
	OverledgerClientSecret string
	OverledgerAuthURL      string
	OverledgerBaseURL      string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Railway provides PORT environment variable
	port := getEnv("PORT", "8080")
	if port[0] != ':' {
		port = ":" + port
	}
	
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", port),
		MeshAPIURL:    getEnv("MESH_API_URL", "http://localhost:8081"),
		APIKey:        getEnv("API_KEY", ""),
		Environment:   getEnv("ENVIRONMENT", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		
		// Overledger OAuth2 Configuration
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