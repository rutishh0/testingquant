# Progress Update Meeting Notes

**Project:** Quant-Coinbase Mesh Connector

**Date:** 01/07/2025

**Attendees:** Rutishkrishna Srinivasaraghavan, Luke Riley, Firas Dahi

---

## 1. Executive Summary & Project Goals

**Objective:** The primary goal of this project is to develop a robust and scalable **Quant Mesh Connector**. This service acts as a middleware bridge, translating Quant Overledger's standardized API requests into the format required by a Coinbase Mesh (Rosetta-compatible) node. This allows any application using the Overledger SDK to seamlessly interact with a new DLT network that exposes a Rosetta API, without requiring native integration.

**Core Functionality:**

*   **API Compatibility:** Expose a set of API endpoints that are fully compatible with the Quant Overledger v2 API specification.
*   **Request Translation:** Map incoming Overledger requests to the corresponding Mesh/Rosetta API calls.
*   **Response Mapping:** Transform the responses from the Mesh API back into the Overledger format.
*   **Authentication:** Securely handle authentication with the Overledger platform using OAuth2.
*   **Scalability & Reliability:** Built with Go and the Gin web framework for high performance and concurrency, and containerized with Docker for easy deployment and scalability.

---

## 2. Milestones Planning

Over the recent development period, I have compiled the whole list of milestones to be addressed:

1.  **Full Project Scaffolding:** Establish a clean and scalable project structure, separating concerns into distinct packages for API handling, configuration, core business logic, and external service clients.
2.  **Configuration Management:** Implement a robust configuration system that loads settings from environment variables, with sensible defaults, making the application adaptable to different environments (development, production).
3.  **API Layer Implementation:** Develop a complete API layer using the Gin web framework, including:
    *   **Routing:** Define all necessary endpoints for construction, account, block, and transaction operations.
    *   **Middleware:** Implement crucial middleware for CORS (Cross-Origin Resource Sharing) and API key authentication.
    *   **Request Handling:** Create handlers that manage the full lifecycle of an API request: validation, service logic execution, and response generation.
4.  **Core Connector Service:** Build the central business logic that orchestrates the entire translation process between the Overledger and Mesh APIs.
5.  **External API Clients:**
    *   **Mesh Client:** A dedicated client for interacting with the Rosetta-compatible Mesh API, featuring a generic request handler and type-safe methods for all required endpoints.
    *   **Overledger Client:** A client for the Overledger API, with built-in, thread-safe OAuth2 authentication to automatically manage access tokens.
6.  **Comprehensive Data Modeling:** Define exhaustive Go structs for all API request/response bodies for the connector's own API, the Mesh API, and the Overledger API, ensuring type safety and data integrity throughout the application.
7.  **Containerization:** Create a `Dockerfile` and `docker-compose.yml` to containerize the application, ensuring consistent, reproducible deployments and simplifying the development setup.
8.  **Documentation:**
    *   **`README.md`**: High-level overview of the project.
    *   **`Project notes.md`**: Detailed explanation of the project's function and architecture.
    *   **`Code notes.md`**: A line-by-line, in-depth walkthrough of the entire codebase.

---

## 3. Detailed Architectural & Code Breakdown

This section provides a detailed, file-by-file breakdown of the source code, including explanations of key functions, structs, and logic. This will serve as the core of the technical discussion.

### 3.1. `cmd/main.go` - The Application Entry Point

**Purpose:** This file is the main entry point for the Quant Mesh Connector application. It is responsible for initializing the application, loading the configuration, setting up the clients for Mesh and Overledger, initializing the connector service, and starting the web server.

**Key Components:**

- **`main()` function:** The main function orchestrates the entire application startup process.
- **Environment Variable Loading:** It uses `godotenv` to load environment variables from a `.env` file, which is crucial for managing configuration in different environments (development, production, etc.).
- **Configuration Loading:** It calls `config.LoadConfig()` to load the application configuration from environment variables.
- **Client Initialization:** It initializes the `mesh.Client` and `overledger.Client` with the appropriate configuration.
- **Service Initialization:** It creates an instance of the `connector.Service`, which contains the core business logic of the application.
- **Gin Router Setup:** It sets up the Gin web server and its routes by calling `api.SetupRouter()`.
- **Server Start:** It starts the HTTP server and listens for incoming requests on the configured address.

**Code Snippet:**

```go
package main

import (
	"log"
	"os"

	"github.com/rutishh0/quant-mesh-connector/tree/main/internal/api"
	"github.com/rutishh0/quant-mesh-connector/tree/main/internal/config"
	"github.com/rutishh0/quant-mesh-connector/tree/main/internal/connector"
	"github.com/rutishh0/quant-mesh-connector/tree/main/internal/mesh"
	"github.com/rutishh0/quant-mesh-connector/tree/main/internal/overledger"

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
```

### 3.2. `internal/config/config.go` - Application Configuration

**Purpose:** This file defines the configuration structure for the application and provides a function to load the configuration from environment variables. This centralizes configuration management and makes it easy to change settings without modifying the code.

**Key Components:**

- **`Config` struct:** This struct holds all the configuration parameters for the application, such as the server address, Mesh API URL, API key, and Overledger OAuth2 credentials.
- **`LoadConfig()` function:** This function is responsible for reading environment variables and populating the `Config` struct. It provides default values for some parameters, which is useful for development.
- **`getEnv()` helper function:** This is a small utility function to get an environment variable with a fallback value.
- **`IsProduction()` and `IsDevelopment()` methods:** These methods provide a convenient way to check the current environment.

**Code Snippet:**

```go
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
```


### 3.3. `internal/api/router.go` - API Routing

**Purpose:** This file is responsible for setting up the API routes and middleware for the application. It uses the Gin web framework to define the endpoints and their handlers.

**Key Components:**

- **`SetupRouter()` function:** This function configures and returns the Gin router. It sets up CORS middleware, an API key validation middleware, and defines all the API endpoints.
- **CORS Middleware:** The `cors.New()` middleware is configured to allow cross-origin requests, which is essential for web applications that interact with the API from a different domain.
- **API Key Middleware:** The `apiKeyMiddleware()` function provides a simple API key validation mechanism. It checks for the `X-API-Key` header in the request and aborts the request if the key is missing or invalid.
- **Route Groups:** The routes are organized into groups (`/v1`, `/construction`, `/account`, etc.) for better organization and versioning.
- **Handler Registration:** Each route is associated with a handler function from the `Handlers` struct.

**Code Snippet:**

```go
package api

import (
	"strings"
	"time"

	"github.com/quant-mesh-connector/internal/connector"

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
```
### 3.4. `internal/api/handlers.go` - API Handlers

**Purpose:** This file contains the implementation of the HTTP handlers for the API endpoints. Each handler is responsible for parsing the request, calling the appropriate method on the `connectorService`, and sending the response back to the client.

**Key Components:**

- **`Handlers` struct:** This struct holds a reference to the `connectorService`, which it uses to process requests.
- **`NewHandlers()` function:** This function creates a new instance of the `Handlers` struct.
- **Handler Functions:** Each public method on the `Handlers` struct corresponds to an API endpoint (e.g., `Preprocess`, `Payloads`, `GetBalance`, etc.).
- **Request Binding:** The handlers use `c.ShouldBindJSON()` to parse the JSON request body into the appropriate Go struct.
- **Error Handling:** The handlers perform error handling and return appropriate JSON error responses with status codes.
- **Response Writing:** The handlers write the successful JSON response to the client using `c.JSON()`.

**Code Snippet:**

```go
package api

import (
	"net/http"
	"time"

	"github.com/rutishh0/quant-mesh-connector/tree/main/internal/connector"
	"github.com/rutishh0/quant-mesh-connector/tree/main/internal/overledger"

	"github.com/gin-gonic/gin"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	connectorService connector.Service
}

// NewHandlers creates a new handlers instance
func NewHandlers(connectorService connector.Service) *Handlers {
	return &Handlers{
		connectorService: connectorService,
	}
}

// Health handles health check requests
func (h *Handlers) Health(c *gin.Context) {
	response := connector.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

// Status handles status requests
func (h *Handlers) Status(c *gin.Context) {
	response := connector.StatusResponse{
		Service:   "quant-mesh-connector",
		Status:    "running",
		Uptime:    "N/A", // In a real implementation, calculate actual uptime
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// Preprocess handles construction preprocess requests
func (h *Handlers) Preprocess(c *gin.Context) {
	var req connector.PreprocessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, connector.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	resp, err := h.connectorService.Preprocess(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, connector.ErrorResponse{
			Error:   "preprocess_failed",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ... (other handlers for Payloads, Combine, Submit, GetBalance, etc.)
```

This concludes the progress update. I am confident in the progress made so far and look forward to continuing to build out the Quant Mesh Connector.