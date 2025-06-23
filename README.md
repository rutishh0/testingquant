# Quant-to-Coinbase Mesh Connector

A middleware service that translates Quant Overledger API calls to Coinbase Mesh API requests, providing a unified interface for blockchain interactions.

## Overview

This connector serves as a translation layer between Quant Overledger and Coinbase Mesh APIs, enabling developers to:

- Execute cross-chain transactions through a unified interface
- Query blockchain data across multiple networks
- Interact with smart contracts using standardized endpoints
- Access comprehensive blockchain functionality without managing multiple API integrations

## Architecture

The connector consists of:

- **API Gateway**: HTTP server with authentication and routing
- **Translation Engine**: Converts between Overledger and Mesh data models
- **Mesh Client**: Handles communication with Coinbase Mesh API
- **Configuration Management**: Environment-based configuration

## Prerequisites

- Go 1.21 or higher
- Access to a Coinbase Mesh-compatible node
- Valid API credentials

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd quant-mesh-connector
```

2. Install dependencies:
```bash
go mod download
```

3. Copy environment configuration:
```bash
cp .env.example .env
```

4. Update `.env` with your configuration:
```env
SERVER_ADDRESS=:8080
MESH_API_URL=http://localhost:8080
API_KEY=your-api-key-here
ENVIRONMENT=development
LOG_LEVEL=info
```

## Usage

### Starting the Server

```bash
go run cmd/main.go
```

The server will start on the configured port (default: 8080).

### API Endpoints

#### Health Check
```bash
GET /health
```

#### Construction API

**Preprocess Transaction**
```bash
POST /v1/construction/preprocess
Content-Type: application/json
X-API-Key: your-api-key

{
  "network_identifier": {
    "blockchain": "ethereum",
    "network": "mainnet"
  },
  "operations": [
    {
      "operation_identifier": {
        "index": 0
      },
      "type": "TRANSFER",
      "account": {
        "address": "0x..."
      },
      "amount": {
        "value": "1000000000000000000",
        "currency": {
          "symbol": "ETH",
          "decimals": 18
        }
      }
    }
  ]
}
```

**Create Payloads**
```bash
POST /v1/construction/payloads
Content-Type: application/json
X-API-Key: your-api-key

{
  "network_identifier": {
    "blockchain": "ethereum",
    "network": "mainnet"
  },
  "operations": [...],
  "metadata": {...}
}
```

**Combine Signatures**
```bash
POST /v1/construction/combine
Content-Type: application/json
X-API-Key: your-api-key

{
  "network_identifier": {
    "blockchain": "ethereum",
    "network": "mainnet"
  },
  "unsigned_transaction": "0x...",
  "signatures": [...]
}
```

**Submit Transaction**
```bash
POST /v1/construction/submit
Content-Type: application/json
X-API-Key: your-api-key

{
  "network_identifier": {
    "blockchain": "ethereum",
    "network": "mainnet"
  },
  "signed_transaction": "0x..."
}
```

#### Account API

**Get Balance**
```bash
POST /v1/account/balance
Content-Type: application/json
X-API-Key: your-api-key

{
  "network_identifier": {
    "blockchain": "ethereum",
    "network": "mainnet"
  },
  "account_identifier": {
    "address": "0x..."
  }
}
```

#### Block API

**Get Block**
```bash
POST /v1/block/
Content-Type: application/json
X-API-Key: your-api-key

{
  "network_identifier": {
    "blockchain": "ethereum",
    "network": "mainnet"
  },
  "block_identifier": {
    "index": 12345678
  }
}
```

## Development

### Project Structure

```
.
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go      # HTTP handlers
│   │   └── router.go        # Route configuration
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── connector/
│   │   ├── models.go        # Overledger data models
│   │   └── service.go       # Translation service
│   └── mesh/
│       ├── client.go        # Mesh API client
│       └── models.go        # Mesh data models
├── .env.example             # Environment configuration template
├── go.mod                   # Go module definition
└── README.md               # This file
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/connector cmd/main.go
```

## Configuration

| Environment Variable | Description | Default |
|----------------------|-------------|----------|
| `SERVER_ADDRESS` | Server bind address | `:8080` |
| `MESH_API_URL` | Coinbase Mesh API URL | `http://localhost:8080` |
| `API_KEY` | API authentication key | - |
| `ENVIRONMENT` | Runtime environment | `development` |
| `LOG_LEVEL` | Logging level | `info` |

## Error Handling

The API returns standardized error responses:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "code": 400
}
```

Common error codes:
- `400`: Bad Request - Invalid input parameters
- `401`: Unauthorized - Missing or invalid API key
- `500`: Internal Server Error - Processing failure

## Security

- All API endpoints (except health checks) require authentication via `X-API-Key` header
- CORS is configured for cross-origin requests
- Input validation is performed on all requests
- Sensitive configuration is managed through environment variables

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License."# mesh" 
