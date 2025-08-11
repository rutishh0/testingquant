package api

import (
    "net/http"
	"strings"
	"time"

	"github.com/rutishh0/testingquant/internal/config"
	"github.com/rutishh0/testingquant/internal/connector"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(connectorService connector.Service, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// CORS middleware for production deployment
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API Key middleware
	router.Use(apiKeyMiddleware(cfg))

	// Initialize handlers
	handlers := NewHandlers(connectorService, cfg)

	// Health and status endpoints
	router.GET("/health", handlers.Health)
	router.GET("/status", handlers.Status)
    // Automated tests endpoint
    router.GET("/tests", handlers.RunTests)
	
	// Serve Next.js web application
	router.Static("/web", "./web")
	
	// Serve the Next.js app at root for production
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})
	
	// Handle Next.js static assets
	router.Static("/_next", "./web/_next")
	router.Static("/favicon.ico", "./web/favicon.ico")

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
			coinbase.GET("/wallets/:walletId/transactions", handlers.GetCoinbaseTransactions)
			coinbase.POST("/wallets/:walletId/transactions/estimate-fee", handlers.EstimateCoinbaseTransactionFee)
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
		if cfg.APIKey == "" {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, connector.ErrorResponse{
				Error:   "missing_api_key",
				Message: "API key is required. Please provide X-API-Key header.",
				Code:    401,
			})
			c.Abort()
			return
		}

		if apiKey != cfg.APIKey {
			c.JSON(http.StatusUnauthorized, connector.ErrorResponse{
				Error:   "invalid_api_key",
				Message: "Invalid API key provided.",
				Code:    401,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPublicPath determines if a path should be accessible without API key
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/health",
		"/status", 
		"/",
		"/web",
		"/_next",
		"/favicon.ico",
		"/docs",
        "/tests",
	}

	for _, publicPath := range publicPaths {
		if path == publicPath || strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}
