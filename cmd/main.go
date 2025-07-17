package main

import (
    "log"
    "os"

    "github.com/your-username/quant-mesh-connector/internal/api"
    "github.com/your-username/quant-mesh-connector/internal/config"
    "github.com/your-username/quant-mesh-connector/internal/connector"
    "github.com/your-username/quant-mesh-connector/internal/mesh"
    "github.com/your-username/quant-mesh-connector/internal/overledger"

    core "github.com/your-username/quant-mesh-connector/internal/core"
    _ "github.com/your-username/quant-mesh-connector/internal/adapters/mesh"
    _ "github.com/your-username/quant-mesh-connector/internal/adapters/overledger"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Mesh client
	meshClient := mesh.NewClient(cfg.MeshAPIURL)

	// Initialize Overledger client
	overledgerClient := overledger.NewClient(cfg)
	log.Printf("Overledger API URL: %s", cfg.OverledgerBaseURL)

	// Initialize connector service
    connectorService := connector.NewService(meshClient, overledgerClient)

    // Initialize modular adapters via registry
    // Load connector configs from YAML if available
    connConfigs, err := core.LoadConnectorConfigs("connectors.yaml")
    if err != nil {
        log.Printf("could not read connectors.yaml: %v – falling back to env vars", err)
        connConfigs = map[string]map[string]any{
            "mesh": {
                "base_url": cfg.MeshAPIURL,
            },
            "overledger": {
                "base_url":      cfg.OverledgerBaseURL,
                "client_id":     cfg.OverledgerClientID,
                "client_secret": cfg.OverledgerClientSecret,
                "auth_url":      cfg.OverledgerAuthURL,
            },
        }
    }

    for id, conf := range connConfigs {
        if c, ok := core.Get(id); ok {
            if err := c.Init(conf); err != nil {
                log.Fatalf("failed to init connector %s: %v", id, err)
            }
            if err := c.HealthCheck(); err != nil {
                log.Fatalf("connector %s health check failed: %v", id, err)
            }
            log.Printf("Connector %s initialised and healthy", id)
        } else {
            log.Printf("Connector %s not registered", id)
        }
    }

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	router := api.SetupRouter(connectorService)

	// Start server
	log.Printf("Starting Quant-Mesh Connector on %s", cfg.ServerAddress)
	log.Printf("Mesh API URL: %s", cfg.MeshAPIURL)
	
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}