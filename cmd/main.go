package main

import (
	"log"
	"os"

	"github.com/rutishh0/testingquant/internal/adapters/coinbase"
	"github.com/rutishh0/testingquant/internal/adapters/mesh"
	"github.com/rutishh0/testingquant/internal/api"
	"github.com/rutishh0/testingquant/internal/clients"
	"github.com/rutishh0/testingquant/internal/config"
	"github.com/rutishh0/testingquant/internal/connector"
	"github.com/rutishh0/testingquant/internal/overledger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env if present (development use)
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("⚠️  Error loading .env: %v", err)
		}
	}

	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting Quant Connector Service...")

	// Initialize Coinbase client (only if credentials are provided)
	var coinbaseClient *clients.CoinbaseClient
	if cfg.CoinbaseAPIKeyID != "" && cfg.CoinbaseAPISecret != "" {
		coinbaseClient = clients.NewCoinbaseClient()
		log.Printf("Coinbase client initialized")
	} else {
		log.Println("Coinbase credentials not configured, Coinbase functionality disabled")
	}

	// Initialize Overledger client (only if credentials are provided)
	var overledgerClient *overledger.Client
	if cfg.OverledgerClientID != "" && cfg.OverledgerClientSecret != "" {
		overledgerClient = overledger.NewClient(cfg)
		log.Printf("Overledger client initialized")
		
		// Test Overledger connection
		if err := overledgerClient.TestConnection(); err != nil {
			log.Fatalf("❌ Critical: Overledger connection test failed, cannot start service: %v", err)
		} else {
			log.Printf("✅ Overledger connection successful")
		}
	} else {
		log.Println("Overledger credentials not configured, Overledger functionality disabled")
	}

	// Initialize connector service
    // Initialize Mesh client (configurable)
    meshClient := clients.NewMeshClient(cfg.MeshAPIURL)

    // Initialize adapters
    var coinbaseAdapter coinbase.Adapter
    if coinbaseClient != nil {
        coinbaseAdapter = coinbase.NewAdapter(coinbaseClient)
    } else {
        coinbaseAdapter = nil
    }
    meshAdapter := mesh.NewAdapter(meshClient)

	// Initialize connector service with all clients
	connectorService := connector.NewService(coinbaseAdapter, meshAdapter, overledgerClient)

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	router := api.SetupRouter(connectorService, cfg)

	// Start server
	log.Printf("🚀 Starting Quant Connector Service on %s", cfg.ServerAddress)
	log.Printf("📊 Environment: %s", cfg.Environment)
	
	if coinbaseClient != nil {
		log.Printf("✅ Coinbase integration: ENABLED")
	} else {
		log.Printf("❌ Coinbase integration: DISABLED (missing credentials)")
	}
	
	if overledgerClient != nil {
		log.Printf("✅ Overledger integration: ENABLED")
	} else {
		log.Printf("❌ Overledger integration: DISABLED (missing credentials)")
	}
	
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}