package main

import (
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/coinbase/rosetta-sdk-go/asserter"
    "github.com/coinbase/rosetta-sdk-go/server"
    "github.com/coinbase/rosetta-sdk-go/types"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "github.com/rutishh0/mesh-server/services"
)

func getPort() string {
    p := os.Getenv("PORT")
    if p == "" {
        p = "8080"
    }
    return p
}

// NewBlockchainRouter creates a Mux http.Handler from a collection
// of server controllers.
func NewBlockchainRouter(
	network *types.NetworkIdentifier,
	asserter *asserter.Asserter,
) http.Handler {
    // Initialize Ethereum RPC client from environment (INFURA_RPC_URL or ETH_RPC_URL)
    rpc, rpcErr := services.NewEthRPCFromEnv()
    if rpcErr != nil {
        log.Printf("Mesh RPC not configured or unreachable, using mock fallback: %v", rpcErr)
        rpc = nil
    }

	networkAPIService := services.NewNetworkAPIService(network, rpc)
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	blockAPIService := services.NewBlockAPIService(network, rpc)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	accountAPIService := services.NewAccountAPIService(network, rpc)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	return server.NewRouter(networkAPIController, blockAPIController, accountAPIController)
}

func main() {
	// Load environment variables from .env if present (development use)
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("⚠️  Error loading .env: %v", err)
		}
	}

	network := &types.NetworkIdentifier{
        Blockchain: "Ethereum",
        Network:    "Sepolia",
	}

	// The asserter automatically rejects incorrectly formatted
	// requests.
	asserter, err := asserter.NewServer(
		[]string{"Transfer", "Reward", "Fee"},
		false,
		[]*types.NetworkIdentifier{network},
		nil,
		false,
		"",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create the main router handler then apply the logger and Cors
	// middlewares in sequence.
	router := NewBlockchainRouter(network, asserter)
	
	// Create Gin router for additional endpoints
	ginRouter := gin.Default()
	
	// Add CORS middleware
	ginRouter.Use(cors.Default())
	
	// Add health check endpoint
	ginRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "coinbase-mesh-server",
		})
	})
	
    // Mount the Mesh API router at /mesh with path rewriting so the wrapped
    // Rosetta router receives paths like /network/list (without the /mesh
    // prefix). Without this rewrite, requests would reach the wrapped router as
    // /mesh/network/list and return 404.
    ginRouter.Any("/mesh/*path", func(c *gin.Context) {
        r := c.Request.Clone(c.Request.Context())
        r.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/mesh")
        router.ServeHTTP(c.Writer, r)
    })
	
    port := getPort()
    log.Printf("🚀 Starting Coinbase Mesh Server on port %s", port)
    log.Printf("📊 Network: %s %s", network.Blockchain, network.Network)
    log.Printf("🔗 Mesh API available at: http://localhost:%s/mesh", port)
    log.Printf("🏥 Health check available at: http://localhost:%s/health", port)
    
    if err := ginRouter.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}