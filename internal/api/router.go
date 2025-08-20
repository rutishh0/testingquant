package api

import (
	"net/http"
	"strings"
	"time"
	"log"

	"github.com/rutishh0/testingquant/internal/config"
	"github.com/rutishh0/testingquant/internal/connector"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// Rosetta + Mesh services
	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/rutishh0/mesh-server/services"
)

// SetupRouter configures the Gin router and routes
func SetupRouter(connectorService connector.Service, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Configure CORS to allow frontend to access APIs
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true // For development; restrict in production
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "X-API-Key"}
	router.Use(cors.New(corsConfig))

	// Initialize handlers
	handlers := NewHandlers(connectorService, cfg)

	// Public endpoints that don't require API key
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services":  map[string]interface{}{},
		})
	})
	// Keep status public for simple uptime checks
	router.GET("/status", handlers.Status)
	// Tests endpoint can be public in development; protect via API key in production by setting X-API-Key
	router.GET("/tests", handlers.RunTests)

	// Serve static Next.js export from ./web/out (copied to /root/web/out in Dockerfile)
	router.Static("/_next", "./web/out/_next")
	router.Static("/static", "./web/out/static")
	router.StaticFile("/favicon.ico", "./web/out/favicon.ico")
	router.StaticFile("/", "./web/out/index.html")

	// Mount Rosetta-compliant Mesh API under /mesh using services from mesh-server module
	// This enables validation tests to call /mesh/network/*, /mesh/account/*, /mesh/block*, etc.
	{
		network := &types.NetworkIdentifier{Blockchain: "Ethereum", Network: "Sepolia"}
		assr, err := asserter.NewServer(
			[]string{"Transfer", "Reward", "Fee"},
			false,
			[]*types.NetworkIdentifier{network},
			nil,
			false,
			"",
		)
		if err != nil {
			log.Printf("Failed to initialize Mesh Rosetta asserter: %v", err)
		} else {
			networkAPIService := services.NewNetworkAPIService(network)
			networkAPIController := server.NewNetworkAPIController(networkAPIService, assr)

			blockAPIService := services.NewBlockAPIService(network)
			blockAPIController := server.NewBlockAPIController(blockAPIService, assr)

			accountAPIService := services.NewAccountAPIService(network)
			accountAPIController := server.NewAccountAPIController(accountAPIService, assr)

			rosettaRouter := server.NewRouter(networkAPIController, blockAPIController, accountAPIController)

			// Path rewrite so wrapped router sees /network/list (no /mesh prefix)
			router.Any("/mesh/*path", func(c *gin.Context) {
				r := c.Request.Clone(c.Request.Context())
				r.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/mesh")
				rosettaRouter.ServeHTTP(c.Writer, r)
			})
		}
	}

	// Apply API key middleware to all routes except public ones
	router.Use(apiKeyMiddleware(cfg))

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Coinbase API endpoints
		coinbase := v1.Group("/coinbase")
		{
			// Wallet operations
			coinbase.GET("/wallets", handlers.GetCoinbaseWallets)
			coinbase.POST("/wallets", handlers.CreateCoinbaseWallet)
			coinbase.GET("/wallets/:walletId/balance", handlers.GetCoinbaseWalletBalance)

			// Address operations
			coinbase.GET("/wallets/:walletId/addresses", handlers.GetCoinbaseWalletAddresses)
			coinbase.POST("/wallets/:walletId/addresses", handlers.CreateCoinbaseWalletAddress)

			// Transaction operations
			coinbase.POST("/wallets/:walletId/transactions", handlers.CreateCoinbaseTransaction)
			coinbase.POST("/wallets/:walletId/transactions/estimate-fee", handlers.EstimateCoinbaseTransactionFee)
			coinbase.GET("/wallets/:walletId/transactions", handlers.GetCoinbaseTransactions)
			coinbase.GET("/wallets/:walletId/transactions-paginated", handlers.GetCoinbaseTransactionsPaginated)
			coinbase.POST("/transactions", handlers.CreateCoinbaseTransaction)
			coinbase.GET("/transactions/:transactionId", handlers.GetCoinbaseTransaction)

			// General information endpoints
			coinbase.GET("/assets", handlers.GetCoinbaseAssets)
			coinbase.GET("/networks", handlers.GetCoinbaseNetworks)
			coinbase.GET("/exchange-rates", handlers.GetCoinbaseExchangeRates)
		}

		// Exchange API endpoints
		exchange := v1.Group("/exchange")
		{
			exchange.GET("/products", handlers.GetExchangeProducts)
			exchange.GET("/accounts", handlers.GetExchangeAccounts)
		}

		// Overledger API endpoints
		overledger := v1.Group("/overledger")
		{
			// Network information
			overledger.GET("/networks", handlers.GetOverledgerNetworks)

			// Account balance operations
			overledger.GET("/networks/:networkId/addresses/:address/balance", handlers.GetOverledgerBalance)

			// Transaction operations
			overledger.POST("/transactions", handlers.CreateOverledgerTransaction)
			overledger.GET("/networks/:networkId/transactions/:txHash/status", handlers.GetOverledgerTransactionStatus)

			// Connection test
			overledger.GET("/test", handlers.TestOverledgerConnection)
		}

		// Mesh API endpoints
		mesh := v1.Group("/mesh")
		{
			mesh.GET("/networks", handlers.GetMeshNetworks)
			mesh.POST("/account/balance", handlers.GetMeshAccountBalance)
			// New: block and transaction retrieval
			mesh.POST("/block", handlers.GetMeshBlock)
			mesh.POST("/block/transaction", handlers.GetMeshBlockTransaction)
		}
	}

	return router
}

// apiKeyMiddleware validates API key from X-API-Key header
func apiKeyMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip API key check for health endpoints and static files
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Skip API key check if not configured (development mode)
		apiKey := cfg.APIKey
		if strings.TrimSpace(apiKey) == "" {
			c.Next()
			return
		}

		// Verify API key
		if c.GetHeader("X-API-Key") != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid API key",
			})
			return
		}

		c.Next()
	}
}

// isPublicPath checks if the path is public and does not require API key
func isPublicPath(path string) bool {
	if path == "/health" || path == "/status" || path == "/tests" || strings.HasPrefix(path, "/_next/") || strings.HasPrefix(path, "/static/") || path == "/" || strings.HasPrefix(path, "/mesh/") {
		return true
	}
	return false
}
