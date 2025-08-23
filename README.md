# TestingQuant Connector (Go backend + Next.js web)

A full-stack project that provides a Go (Gin) backend exposing Coinbase, Overledger, and Rosetta-compatible Mesh endpoints, plus a Next.js web UI in `web/`. For local development on Windows, a PowerShell helper script `start.ps1` launches both the backend and the frontend for you.

---

## Quick start (Windows, PowerShell)

1) Prerequisites
- Git
- Go 1.21+
- Node.js 18.18+ or 20+ and npm
- PowerShell 5+ (or PowerShell 7)

2) Clone and install frontend deps
```powershell
git clone https://github.com/rutishh0/testingquant.git
cd testingquant
# Install frontend dependencies once
cd web
npm install
cd ..
```

3) Run using the helper script
```powershell
# From the repo root
# If your script execution policy blocks running scripts, use the second command instead
.\start.ps1
# or
powershell -ExecutionPolicy Bypass -File .\start.ps1
```
What happens:
- A new PowerShell window runs the backend: `go run ./cmd/main.go` on http://localhost:8080
- A second PowerShell window runs the frontend: `npm run dev -- -p 3001` with `NEXT_PUBLIC_BACKEND_URL=http://localhost:8080`

4) Open the app
- Frontend: http://localhost:3001
- Backend health: http://localhost:8080/health

> Tip: On first run, the frontend will fail if `node_modules` are missing. Make sure you executed `npm install` in the `web/` directory before running the script.

---

## Manual run (macOS/Linux/WSL or without the script)

Terminal A – backend:
```bash
# from repo root
# optionally set a different port; default is :8080
# PORT=8080 or SERVER_ADDRESS=:8080 are supported
# API_KEY is optional for dev; leave empty to disable auth
PORT=8080 go run ./cmd/main.go
```

Terminal B – frontend:
```bash
cd web
export NEXT_PUBLIC_BACKEND_URL="http://localhost:8080"
npm install
npm run dev -- -p 3001
```

Open http://localhost:3001

---

## Environment variables

You can run the backend with no credentials for a simple local demo. Features that need external services (Coinbase, Overledger, external Ethereum RPC) will automatically disable when their vars are missing.

Backend (read in `internal/config/config.go`):
- SERVER_ADDRESS: Listening address, default `:8080`. Alternative: set PORT.
- API_KEY: Optional. If set, all non-public API endpoints require the header `X-API-Key: <API_KEY>`.
- ENVIRONMENT: `development` (default) or `production`.
- LOG_LEVEL: `info` (default).
- COINBASE_API_KEY_ID, COINBASE_API_SECRET, COINBASE_API_URL: Optional. Enable Coinbase features.
- OVERLEDGER_CLIENT_ID, OVERLEDGER_CLIENT_SECRET, OVERLEDGER_AUTH_URL, OVERLEDGER_BASE_URL, OVERLEDGER_TX_SIGNING_KEY_ID: Optional. Enable Overledger features.
- MESH_API_URL: Optional. Default `http://localhost:8080/mesh`. Leave empty to use the embedded Mesh Rosetta API served by this backend under `/mesh`. If you point to an external service, ensure the URL includes the `/mesh` path.
- MESH_USE_SDK: Optional, `true` to use the Mesh SDK client instead of HTTP.
- ETH_RPC_URL or INFURA_RPC_URL: Optional. Used by the embedded Mesh services. If omitted, a Sepolia RPC fallback is used.

Frontend (`web/components/api-client.ts`):
- NEXT_PUBLIC_BACKEND_URL: Base URL for the backend; set by `start.ps1` to `http://localhost:8080`.
- NEXT_PUBLIC_API_KEY: Optional. If you set `API_KEY` on the backend, set this to the same value so the UI sends `X-API-Key` with requests.

Sample `.env` for backend (optional):
```env
# Server
PORT=8080
API_KEY=
ENVIRONMENT=development
LOG_LEVEL=info

# Coinbase (optional)
COINBASE_API_KEY_ID=
COINBASE_API_SECRET=
COINBASE_API_URL=https://api.coinbase.com

# Overledger (optional)
OVERLEDGER_CLIENT_ID=
OVERLEDGER_CLIENT_SECRET=
OVERLEDGER_AUTH_URL=https://auth.overledger.dev/oauth2/token
OVERLEDGER_BASE_URL=https://api.overledger.dev
OVERLEDGER_TX_SIGNING_KEY_ID=

# Mesh (optional)
MESH_API_URL=
MESH_USE_SDK=false

# Ethereum RPC for Mesh (optional)
ETH_RPC_URL=
# or
INFURA_RPC_URL=
```

---

## Useful endpoints

Public (no API key):
- `GET /health`
- `GET /status`
- `GET /tests`
- Rosetta Mesh (embedded): `GET /mesh/*` (e.g. `/mesh/network/list`)

Versioned API (often requires API key if `API_KEY` is set):
- Coinbase: `/v1/coinbase/*`
- Overledger: `/v1/overledger/*`
- Mesh helper: `/v1/mesh/*`

---
