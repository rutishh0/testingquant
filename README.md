# Quant Connector

_Unified Coinbase CDP × Quant Overledger bridge – Go backend · Next.js dashboard · Docker-ready for Koyeb_

---

Quant Connector exposes a single REST API plus a React dashboard that lets you:

* Create & manage Coinbase wallets/addresses
* Retrieve balances, assets and exchange-rates
* Relay signed transactions to multiple chains via Overledger
* View live system-health – all secured by an `X-API-Key` header

---

## Tech Stack

| Layer      | Technology                                                   |
|------------|--------------------------------------------------------------|
| Backend    | **Go 1.21**, Gin, CORS, OAuth2 (Overledger), Ed25519 JWT (CDP) |
| Frontend   | **Next.js 14** (app router), Tailwind CSS, Radix UI           |
| Container  | Multi-stage Docker (Node → Go → Alpine)                      |
| Hosting    | Koyeb (Docker image, zero-dyno scaling)                      |

---

## Quick Start (local)

```bash
# clone & install deps
git clone https://github.com/<your-org>/quant-connector.git
cd quant-connector
go mod download
npm --prefix web ci

# copy env template and fill in credentials
cp .env.example .env

# run backend (http://localhost:8080)
go run cmd/main.go &

# run frontend (http://localhost:3000)
npm --prefix web run dev
```

Visit `http://localhost:3000` – the dashboard proxies API calls to `localhost:8080`.

---

## Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `API_KEY` | Shared key required in `X-API-Key` | `dev-key-123` |
| `SERVER_ADDRESS` | Gin bind address | `:8080` |
| **Coinbase CDP** |||
| `COINBASE_API_KEY_ID` | Ed25519 key ID | `7ac8…` |
| `COINBASE_API_SECRET` | Base-64 seed | `mYB1…` |
| `COINBASE_API_URL` | CDP base URL | `https://api.cdp.coinbase.com` |
| **Overledger** |||
| `OVERLEDGER_CLIENT_ID` | OAuth2 client | `abc…` |
| `OVERLEDGER_CLIENT_SECRET` | OAuth2 secret | `xyz…` |
| `OVERLEDGER_AUTH_URL` | Token endpoint | `https://auth.overledger.dev/oauth2/token` |
| `OVERLEDGER_BASE_URL` | API base | `https://api.overledger.dev` |

---

## API Surface (v1)

| Method | Path | Description |
|--------|------|-------------|
| GET    | `/health` | Service & dependency health |
| GET    | `/v1/coinbase/wallets` | List wallets |
| POST   | `/v1/coinbase/wallets` | Create wallet |
| GET    | `/v1/coinbase/assets`  | List tradeable assets |
| …      | *(many more – see `internal/api/router.go`)* | |

All non-public routes require the `X-API-Key` header.

---

## Deployment on Koyeb

1. Push code → GitHub (Koyeb is set to auto-build the Dockerfile).
2. In the Koyeb service → **Settings → Environment** set the same variables shown above.
3. Wait for the build to finish and status to be **Running**.
4. Open `/health` in the browser – it should return HTTP 200 with `{"status":"healthy"}`.

---

## Project Layout & File Guide

| Path | What it does |
|------|--------------|
| **`cmd/main.go`** | Program entry; loads env, DI, starts Gin server |
| **`internal/api/router.go`** | Registers all REST routes & middleware |
| **`internal/api/handlers.go`** | Implementation of HTTP handlers (Coinbase & Overledger) |
| **`internal/config/config.go`** | Loads env vars into a typed `Config` struct |
| **`internal/clients/coinbase.go`** | Minimal Coinbase CDP client – signs Ed25519 JWT, prefixes `/platform` |
| **`internal/overledger/`** | Thin Overledger client + models |
| **`internal/connector/service.go`** | Business layer bridging clients with handlers |
| **`internal/utils/jwt.go`** | Helpers for creating CDP JWT + auth headers |
| **`web/`** | Next.js 14 dashboard (App Router) |
| &nbsp;&nbsp;`web/app/` | Top-level `layout.tsx`, global CSS, root page |
| &nbsp;&nbsp;`web/components/` | Reusable React components + `api-client.ts` (fetch wrapper) |
| **`Dockerfile`** | Multi-stage build → static frontend then Go binary on Alpine |
| **`docker-compose.yml`** | Local one-shot dev stack (backend + Postgres demo) |
| **`.env.example`** | Copy to `.env` – documents every variable |
| **`test_api.ps1`** | Sample PowerShell script calling endpoints |

For a deeper dive, see inline docstrings in each file.

---

## Development scripts

```bash
# run all Go tests
go test ./...
# lint (go vet + staticcheck suggested)
make lint
```

---

## Contributing

1. Fork → feature branch
2. Commit with conventional commits
3. Make sure `go test ./...` & `npm test` pass
4. Open PR – we squash-merge

---

## License

MIT © 2025 Quant Network

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
git clone https://github.com/rutishh0/quant-mesh-connector.git
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
MESH_API_URL=http://localhost:8081  # Optional: URL of a Coinbase Mesh implementation
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
│   ├── mesh/
│   │   ├── client.go        # Mesh API client
│   │   └── models.go        # Mesh data models
│   └── overledger/
│       ├── client.go        # Overledger API client
│       └── models.go        # Overledger data models
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
| `MESH_API_URL` | Coinbase Mesh API URL (optional) | Empty string (disabled) |
| `API_KEY` | API authentication key | - |
| `ENVIRONMENT` | Runtime environment | `development` |
| `LOG_LEVEL` | Logging level | `info` |
| `OVERLEDGER_CLIENT_ID` | Quant Overledger OAuth2 client ID | - |
| `OVERLEDGER_CLIENT_SECRET` | Quant Overledger OAuth2 client secret | - |
| `OVERLEDGER_AUTH_URL` | Quant Overledger OAuth2 token URL | `https://auth.overledger.dev/oauth2/token` |
| `OVERLEDGER_BASE_URL` | Quant Overledger API base URL | `https://api.overledger.dev` |

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

This project is licensed under the MIT License.
