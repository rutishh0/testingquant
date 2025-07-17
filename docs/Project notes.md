# Project Notes: Quant-to-Coinbase Mesh Connector

This document provides a detailed explanation of the Quant-to-Coinbase Mesh Connector, including its overall functionality and a breakdown of each file's purpose and code.

## Table of Contents

- [Project Overview](#project-overview)
- [File-by-File Breakdown](#file-by-file-breakdown)

## Project Overview

The Quant-to-Coinbase Mesh Connector is a middleware service written in Go. It acts as a translation layer between Quant's Overledger API and the Coinbase Mesh API (formerly Rosetta). The primary goal of this project is to allow developers to interact with various blockchains through a single, unified interface, abstracting away the complexities of individual blockchain integrations.

### Core Functionality

- **API Translation**: It translates API calls from the Overledger standard to the Mesh standard.
- **Cross-Chain Operations**: Enables cross-chain transactions, balance inquiries, and block data retrieval.
- **Authentication**: Manages OAuth2 authentication with the Overledger API.
- **Configuration**: Uses environment variables for easy configuration of API endpoints and credentials.

### Architecture

The application is built with a modular architecture, separating concerns into different packages:

- **`cmd`**: The main entry point of the application.
- **`internal/api`**: Handles HTTP routing and request handling.
- **`internal/config`**: Manages application configuration.
- **`internal/connector`**: Contains the core business logic for connecting and translating between the two APIs.
- **`internal/mesh`**: A client for interacting with the Coinbase Mesh API.
- **`internal/overledger`**: A client for interacting with the Quant Overledger API.

## File-by-File Code Explanation

### `cmd/main.go`

**Purpose:** This is the main entry point of the application.

**Code Explanation:**

- **Package and Imports:** It imports necessary internal packages (`api`, `config`, `connector`, `mesh`, `overledger`) and external libraries like `gin` (for the web server) and `godotenv` (for managing environment variables).
- **`main()` function:**
    - It first attempts to load environment variables from a `.env` file. This is useful for local development.
    - It loads the application configuration using the `config.LoadConfig()` function.
    - It initializes the `meshClient` and `overledgerClient` with the loaded configuration. These clients are responsible for communicating with the Coinbase Mesh and Quant Overledger APIs, respectively.
    - The `connectorService` is initialized, which contains the core business logic for translating requests between Overledger and Mesh.
    - It sets the Gin web server's mode (e.g., `release` or `debug`) based on the environment.
    - The `api.SetupRouter()` function is called to define all the API endpoints.
    - Finally, it starts the web server on the configured address and port, logging the startup information.

### `internal/api/router.go`

**Purpose:** This file defines the API routes and sets up the Gin web server.

**Code Explanation:**

- **`SetupRouter()` function:**
    - It creates a new Gin router instance.
    - **CORS Middleware:** It configures Cross-Origin Resource Sharing (CORS) to allow requests from any origin. This is important for web-based clients.
    - **API Key Middleware:** It adds a middleware (`apiKeyMiddleware`) to protect the API endpoints. Most endpoints will require a valid `X-API-Key` header.
    - It initializes the API handlers, which contain the logic for each endpoint.
    - **Endpoints:**
        - `/health` and `/status`: Basic endpoints to check the health and status of the service.
        - `/` and `/web`: Serves the static files for a developer portal/UI.
        - `/v1/...`: All version 1 API endpoints are grouped under this path.
            - `/construction/...`: Endpoints for the transaction construction process (preprocess, payloads, combine, submit).
            - `/account/...`: Endpoint for checking account balances.
            - `/block/...`: Endpoint for retrieving block information.
            - `/transaction/...`: Endpoint for retrieving transaction information.
            - `/overledger/...`: Endpoints that interact directly with the Overledger API.
- **`apiKeyMiddleware()` function:**
    - This function checks for the `X-API-Key` in the request header.
    - It allows requests to the health, status, and web endpoints to bypass the API key check.
    - If the API key is missing or invalid, it returns a `401 Unauthorized` error.

### `internal/config/config.go`

**Purpose:** This file defines the configuration structure for the application and loads values from environment variables.

**Code Explanation:**

- **`Config` struct:** This struct holds all the configuration parameters for the application, such as server address, API URLs, API keys, and Overledger OAuth2 credentials.
- **`LoadConfig()` function:**
    - It reads environment variables using the `getEnv` helper function.
    - It provides default values for essential parameters if the environment variables are not set. This is useful for getting the application running quickly.
    - It handles the `PORT` environment variable specifically, which is often provided by deployment platforms like Railway.
- **`getEnv()` function:** A simple helper to read an environment variable or return a fallback value.
- **`IsProduction()` and `IsDevelopment()`:** Helper methods on the `Config` struct to easily check the current environment.

### `internal/connector/service.go`

**Purpose:** This is the core of the application, containing the business logic that connects and translates between the Overledger and Mesh APIs.

**Code Explanation:**

- **`Service` interface:** Defines the contract for the connector service, listing all the methods it must implement. This includes methods for each step of the transaction construction flow, as well as for fetching data like balances and blocks.
- **`service` struct:** The concrete implementation of the `Service` interface. It holds clients for both the Mesh and Overledger APIs.
- **`NewService()`:** The constructor function that creates a new instance of the service.
- **Method Implementations (`Preprocess`, `Payloads`, `Combine`, `Submit`, `GetBalance`, `GetBlock`):**
    - Each of these methods takes a request in a format compatible with Overledger.
    - It maps the Overledger-style request to a Mesh-style request.
    - It calls the appropriate method on the `meshClient`.
    - It then maps the response from the `meshClient` back into an Overledger-compatible format before returning it.
- **Overledger-specific Methods (`GetOverledgerNetworks`, etc.):** These methods are pass-through calls to the `overledgerClient`, allowing the connector to also expose some native Overledger functionality.
- **Helper Functions (`mapOperations`, `mapBalances`, etc.):** These are private functions used to perform the detailed mapping between the data models of Overledger and Mesh. They handle the translation of fields, structures, and data types.

### `internal/connector/models.go`

**Purpose:** This file defines all the data structures (models) used for requests and responses in the connector service. These models are designed to be compatible with the Overledger API specification.

**Code Explanation:**

- The file contains a series of Go structs that represent the JSON objects for each API call.
- **Request and Response Structs:** For each major function (e.g., `Preprocess`, `Payloads`, `Balance`), there are corresponding `...Request` and `...Response` structs.
- **JSON Tags:** Each field in the structs has a `json` tag (e.g., `json:"dlt"`). These tags control how the struct is serialized to and deserialized from JSON, ensuring the keys in the JSON match the expected API format.
- **Binding Tags:** Some fields have `binding:"required"` tags. These are used by the Gin framework to automatically validate that required fields are present in incoming requests.
- **Shared Structures:** The file also defines common, reusable structures like `Transfer`, `PublicKey`, `Signature`, `Balance`, and `BlockInfo` that are used within the larger request/response models.

### `internal/mesh/client.go`

**Purpose:** This file provides a client for interacting with a Coinbase Mesh-compatible API.

**Code Explanation:**

- **`Client` struct:** Holds the base URL for the Mesh API and an `http.Client` for making requests.
- **`NewClient()`:** A constructor to create a new Mesh client.
- **`makeRequest()`:** A generic helper method for making HTTP requests to the Mesh API. It handles marshalling the request payload into JSON, setting the correct headers, executing the request, and unmarshalling the JSON response. It also includes error handling for non-successful status codes.
- **API Methods (`NetworkStatus`, `AccountBalance`, `Block`, `ConstructionPreprocess`, etc.):** Each of these methods corresponds to a specific endpoint in the Mesh API. They prepare the request, call the `makeRequest` helper, and return the response. This abstracts the raw HTTP communication away from the rest of the application.
- **`Health()`:** A specific method to check the `/health` endpoint of the Mesh API, ensuring it's reachable and operational.

### `internal/mesh/models.go`

**Purpose:** This file defines the Go data structures that correspond to the JSON objects used in the Coinbase Mesh API (which is based on the Rosetta API standard).

**Code Explanation:**

- The file contains a comprehensive set of structs that model the entire Mesh/Rosetta API specification.
- **Core Identifiers:** Structs like `NetworkIdentifier`, `AccountIdentifier`, `BlockIdentifier`, and `TransactionIdentifier` are fundamental for uniquely identifying resources.
- **Data Structures:** `Amount`, `Currency`, `Operation`, `Transaction`, and `Block` represent the core data objects.
- **Construction API Models:** A series of request and response structs (`ConstructionPreprocessRequest`, `ConstructionPayloadsResponse`, etc.) are defined specifically for the multi-step transaction construction process.
- **JSON Tags:** Like other model files, `json` tags are used extensively to map the Go struct fields to the `snake_case` JSON fields used by the Mesh API.

### `internal/overledger/client.go`

**Purpose:** This file provides a client for interacting with the Quant Overledger API, including handling its OAuth2 authentication.

**Code Explanation:**

- **`Client` struct:** Holds the application configuration, an `http.Client`, the OAuth2 access token, the token's expiry time, and a mutex for thread-safe token handling.
- **`NewClient()`:** A constructor to create a new Overledger client.
- **`authenticate()`:** This is a critical method that handles the OAuth2 client credentials grant flow. It requests an access token from the Overledger authentication server using the client ID and secret. It stores the token and its expiry time and ensures that a new token is fetched only when the current one is expired or non-existent. The `sync.Mutex` prevents race conditions when multiple concurrent requests try to refresh the token simultaneously.
- **`makeRequest()`:** Similar to the Mesh client, this is a generic helper for making authenticated requests. Before making the actual API call, it ensures a valid access token is available by calling `authenticate()`. It then adds the `Authorization: Bearer <token>` header to the request.
- **API Methods (`GetNetworks`, `GetAccountBalance`, etc.):** These methods provide a clean interface for calling specific Overledger API endpoints. They construct the correct endpoint URL, call `makeRequest`, and return the parsed response.
- **`TestConnection()`:** A simple method that calls `authenticate()` to verify that the client can successfully obtain an access token, which serves as a connection test.

### `internal/overledger/models.go`

**Purpose:** This file defines the Go data structures that correspond to the JSON objects used in the Quant Overledger API.

**Code Explanation:**

- The structs in this file are tailored to the specific responses provided by the Overledger v2 API.
- **Response Structs:** `NetworksResponse`, `BalanceResponse`, `TransactionResponse`, and `TransactionStatusResponse` model the JSON returned from the primary Overledger endpoints.
- **Data Structures:** `Network`, `Balance`, and other nested structs define the shape of the data within the responses.
- **JSON Tags:** `json` tags are used to map the Go struct fields to the `camelCase` JSON fields used by the Overledger API.

### `internal/api/handlers.go`

**Purpose:** This file contains the handler functions for each of the API endpoints defined in `router.go`.

**Code Explanation:**

- **`Handlers` struct:** Holds a reference to the `connectorService`, which it uses to perform the core business logic.
- **`NewHandlers()`:** A constructor to create a new `Handlers` instance.
- **Handler Functions (`Health`, `Status`, `Preprocess`, `Payloads`, etc.):**
    - Each function corresponds to an API endpoint.
    - It has a `(c *gin.Context)` parameter, which is the context for the incoming HTTP request.
    - **Request Binding:** It uses `c.ShouldBindJSON(&req)` to parse the JSON body of the incoming request and populate a corresponding Go struct (e.g., `connector.PreprocessRequest`). It includes error handling for invalid request bodies.
    - **Service Call:** It calls the appropriate method on the `connectorService` to execute the request's logic.
    - **Response Writing:** It uses `c.JSON()` to serialize the response from the service (or an error response) into JSON and writes it back to the client with the appropriate HTTP status code (e.g., `http.StatusOK` for success, `http.StatusInternalServerError` for an internal error).