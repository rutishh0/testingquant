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

	// CORS middleware
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
	handlers := NewHandlers(connectorService)

	// Health and status endpoints
	router.GET("/health", handlers.Health)
	router.GET("/status", handlers.Status)
	
	// Serve documentation
	router.GET("/docs", func(c *gin.Context) {
		c.File("./docs/index.html")
	})

	// Serve Next.js static files
	router.Static("/web", "./web")
	
	// Serve the Next.js app at root
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})
	
	// Handle Next.js static assets
	router.Static("/_next", "./web/_next")

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Construction API endpoints
		construction := v1.Group("/construction")
		{
			construction.POST("/preprocess", handlers.Preprocess)
			construction.POST("/payloads", handlers.Payloads)
			construction.POST("/combine", handlers.Combine)
			construction.POST("/submit", handlers.Submit)
		}

		// Account API endpoints
		account := v1.Group("/account")
		{
			account.POST("/balance", handlers.GetBalance)
		}

		// Block API endpoints
		block := v1.Group("/block")
		{
			block.POST("/", handlers.GetBlock)
		}

		// Transaction API endpoints
		transaction := v1.Group("/transaction")
		{
			transaction.POST("/", handlers.GetTransaction)
		}

		// Overledger API endpoints
		overledger := v1.Group("/overledger")
		{
			// Network endpoints
			overledger.GET("/networks", handlers.GetOverledgerNetworks)

			// Balance endpoints
			overledger.GET("/balance/:networkId/:address", handlers.GetOverledgerBalance)

			// Transaction endpoints
			overledger.POST("/transaction", handlers.CreateOverledgerTransaction)

			// Connection test endpoint
			overledger.GET("/test", handlers.TestOverledgerConnection)
		}
	}

	return router
}

// apiKeyMiddleware validates API key from X-API-Key header
func apiKeyMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip API key validation for health, status, root, and web endpoints
		path := c.Request.URL.Path
		if path == "/health" || path == "/status" || path == "/" ||
			strings.HasPrefix(path, "/web/") {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "API key is required",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate the API key
		if cfg.APIKey != "" && apiKey != cfg.APIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid API key",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
