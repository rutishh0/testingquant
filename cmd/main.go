package main

import (
	"log"
	"os"

	"github.com/quant-mesh-connector/internal/api"
	"github.com/quant-mesh-connector/internal/config"
	"github.com/quant-mesh-connector/internal/connector"
	"github.com/quant-mesh-connector/internal/mesh"
	"github.com/quant-mesh-connector/internal/overledger"

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