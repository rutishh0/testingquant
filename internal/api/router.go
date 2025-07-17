package api

import (
	"strings"
	"time"

	"github.com/rutishh0/testingquant/internal/connector"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(connectorService connector.Service) *gin.Engine {
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
	router.Use(apiKeyMiddleware())

	// Initialize handlers
	handlers := NewHandlers(connectorService)

	// Health and status endpoints
	router.GET("/health", handlers.Health)
	router.GET("/status", handlers.Status)

	// Serve developer portal
	router.Static("/web", "./web")
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

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
func apiKeyMiddleware() gin.HandlerFunc {
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
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "API key is required",
				"code":    401,
			})
			c.Abort()
			return
		}

		// In a real implementation, validate the API key against a database or service
		// For now, we'll accept any non-empty API key
		if len(apiKey) < 10 {
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "Invalid API key format",
				"code":    401,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
