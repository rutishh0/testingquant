# Coinbase Mesh Server

A Coinbase Mesh API server implementation that provides blockchain data through the Mesh API specification.

## Features

- **Network API**: Provides network information and status
- **Block API**: Retrieves block and transaction data
- **Account API**: Retrieves account balances and coin information
- **Health Check**: Built-in health monitoring endpoint
- **CORS Support**: Cross-origin resource sharing enabled

## API Endpoints

### Mesh API Endpoints

- `GET /mesh/network/list` - List supported networks
- `GET /mesh/network/options` - Get network options and capabilities
- `GET /mesh/network/status` - Get current network status
- `GET /mesh/block` - Get block information
- `GET /mesh/block/transaction` - Get transaction information
- `GET /mesh/account/balance` - Get account balance
- `GET /mesh/account/coins` - Get account coins

### Additional Endpoints

- `GET /health` - Health check endpoint

## Quick Start

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd mesh-server
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run the server**
   ```bash
   go run main.go
   ```

4. **Test the server**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/mesh/network/list
   ```

### Docker

1. **Build the image**
   ```bash
   docker build -t coinbase-mesh-server .
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 coinbase-mesh-server
   ```

## Configuration

The server uses environment variables for configuration:

- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment (development/production)

## API Examples

### Get Network Status
```bash
curl -X POST http://localhost:8080/mesh/network/status \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "Coinbase",
      "network": "Mainnet"
    }
  }'
```

### Get Account Balance
```bash
curl -X POST http://localhost:8080/mesh/account/balance \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "Coinbase",
      "network": "Mainnet"
    },
    "account_identifier": {
      "address": "0x1234567890abcdef1234567890abcdef1234567890"
    }
  }'
```

### Get Block Information
```bash
curl -X POST http://localhost:8080/mesh/block \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "Coinbase",
      "network": "Mainnet"
    },
    "block_identifier": {
      "index": 1000000
    }
  }'
```

## Deployment

### Koyeb Deployment

1. **Create a new app on Koyeb**
2. **Connect your GitHub repository**
3. **Set the build command**: `go build -o mesh-server .`
4. **Set the run command**: `./mesh-server`
5. **Set the port**: `8080`

### Environment Variables for Koyeb

- `PORT`: `8080`
- `ENVIRONMENT`: `production`

## Development

### Project Structure

```
mesh-server/
├── main.go              # Main application entry point
├── go.mod               # Go module file
├── go.sum               # Go module checksums
├── Dockerfile           # Docker configuration
├── README.md            # This file
└── services/            # API service implementations
    ├── network_service.go
    ├── block_service.go
    └── account_service.go
```

### Adding New Endpoints

1. Create a new service file in the `services/` directory
2. Implement the required interface methods
3. Add the controller to the router in `main.go`

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## License

This project is licensed under the Apache 2.0 License. 