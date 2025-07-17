# Overledger Integration Guide

## Overview

This document describes the integration of Quant's Overledger API with the Quant-Mesh Connector. The integration provides OAuth2-authenticated access to Overledger's multi-DLT platform alongside the existing Coinbase Mesh API functionality.

## Architecture

### Components

1. **Overledger Client** (`internal/overledger/client.go`)
   - Handles OAuth2 authentication with Overledger API
   - Provides methods for network operations, balance queries, and transaction creation
   - Manages token refresh and error handling

2. **Overledger Models** (`internal/overledger/models.go`)
   - Defines data structures for Overledger API requests and responses
   - Includes models for networks, balances, transactions, and webhooks

3. **Enhanced Connector Service** (`internal/connector/service.go`)
   - Integrates both Mesh and Overledger clients
   - Provides unified interface for both APIs
   - Handles service orchestration and data mapping

4. **API Handlers** (`internal/api/handlers.go`)
   - Exposes Overledger functionality via REST endpoints
   - Handles request validation and error responses

## Configuration

### Environment Variables

Add the following variables to your `.env` file:

```env
# Overledger OAuth2 Configuration
OVERLEDGER_CLIENT_ID=your_client_id_here
OVERLEDGER_CLIENT_SECRET=your_client_secret_here
OVERLEDGER_AUTH_URL=https://auth.overledger.dev/oauth2/token
OVERLEDGER_BASE_URL=https://api.overledger.dev
```

### OAuth2 Credentials

To obtain OAuth2 credentials:

1. Register for a Quant Developer account at [developer.quant.network](https://developer.quant.network)
2. Create a new application in the developer portal
3. Copy the `Client ID` and `Client Secret` from your application settings
4. Use the provided credentials in your environment configuration

## API Endpoints

### Overledger-Specific Endpoints

All Overledger endpoints are prefixed with `/v1/overledger` and require API key authentication.

#### 1. Get Networks
```http
GET /v1/overledger/networks
Headers:
  X-API-Key: your_api_key
```

Returns available blockchain networks supported by Overledger.

**Response:**
```json
{
  "networks": [
    {
      "id": "ethereum-mainnet",
      "name": "Ethereum Mainnet",
      "type": "ethereum",
      "status": "active"
    }
  ]
}
```

#### 2. Get Balance
```http
GET /v1/overledger/balance/{networkId}/{address}
Headers:
  X-API-Key: your_api_key
```

Retrieves account balance for a specific address on a given network.

**Parameters:**
- `networkId`: Network identifier (e.g., "ethereum-mainnet")
- `address`: Account address to query

**Response:**
```json
{
  "address": "0x742d35Cc6634C0532925a3b8D4C9db96",
  "balances": [
    {
      "amount": "1000000000000000000",
      "symbol": "ETH",
      "decimals": 18
    }
  ]
}
```

#### 3. Create Transaction
```http
POST /v1/overledger/transaction
Headers:
  X-API-Key: your_api_key
  Content-Type: application/json
```

Creates a new transaction on the specified network.

**Request Body:**
```json
{
  "networkId": "ethereum-mainnet",
  "fromAddress": "0x742d35Cc6634C0532925a3b8D4C9db96",
  "toAddress": "0x8ba1f109551bD432803012645Hac136c",
  "amount": "1000000000000000000",
  "tokenSymbol": "ETH",
  "metadata": {
    "gasPrice": "20000000000",
    "gasLimit": "21000"
  }
}
```

**Response:**
```json
{
  "transactionId": "0x1234567890abcdef",
  "status": "pending",
  "networkId": "ethereum-mainnet",
  "hash": "0xabcdef1234567890"
}
```

#### 4. Test Connection
```http
GET /v1/overledger/test
Headers:
  X-API-Key: your_api_key
```

Tests connectivity to the Overledger API.

**Response:**
```json
{
  "status": "connected",
  "message": "Overledger API connection successful"
}
```

## Authentication Flow

### OAuth2 Client Credentials Flow

The integration uses OAuth2 Client Credentials flow for authentication:

1. **Token Request**: Client sends credentials to auth endpoint
   ```http
   POST https://auth.overledger.dev/oauth2/token
   Authorization: Basic base64(client_id:client_secret)
   Content-Type: application/x-www-form-urlencoded
   
   grant_type=client_credentials
   ```

2. **Token Response**: Server returns access token
   ```json
   {
     "access_token": "eyJhbGciOiJSUzI1NiIs...",
     "token_type": "Bearer",
     "expires_in": 3600
   }
   ```

3. **API Requests**: Client uses token for subsequent requests
   ```http
   GET https://api.overledger.dev/v2/networks
   Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
   ```

### Token Management

- Tokens are automatically refreshed when they expire
- Failed authentication triggers re-authentication
- Token storage is handled in-memory (consider persistent storage for production)

## Error Handling

### Common Error Responses

1. **Authentication Errors (401)**
   ```json
   {
     "error": "unauthorized",
     "message": "Invalid or expired access token",
     "code": 401
   }
   ```

2. **Network Errors (503)**
   ```json
   {
     "error": "overledger_connection_failed",
     "message": "Unable to connect to Overledger API",
     "code": 503
   }
   ```

3. **Validation Errors (400)**
   ```json
   {
     "error": "invalid_request",
     "message": "networkId and address are required",
     "code": 400
   }
   ```

## Development and Testing

### Local Development

1. Set up environment variables in `.env` file
2. Ensure valid Overledger credentials
3. Start the server: `go run cmd/main.go`
4. Test connection: `curl -H "X-API-Key: test" http://localhost:8080/v1/overledger/test`

### Testing Endpoints

Use the provided test script or curl commands:

```bash
# Test connection
curl -H "X-API-Key: test_api_key_12345" \
     http://localhost:8080/v1/overledger/test

# Get networks
curl -H "X-API-Key: test_api_key_12345" \
     http://localhost:8080/v1/overledger/networks

# Get balance
curl -H "X-API-Key: test_api_key_12345" \
     http://localhost:8080/v1/overledger/balance/ethereum-mainnet/0x742d35Cc6634C0532925a3b8D4C9db96
```

## Production Considerations

### Security

1. **Credential Management**
   - Store credentials securely (e.g., AWS Secrets Manager, HashiCorp Vault)
   - Rotate credentials regularly
   - Use environment-specific credentials

2. **Token Storage**
   - Consider persistent token storage for high-availability deployments
   - Implement token encryption at rest
   - Use secure token refresh mechanisms

3. **API Security**
   - Implement proper API key validation
   - Add rate limiting and request throttling
   - Use HTTPS for all communications

### Monitoring

1. **Health Checks**
   - Monitor Overledger API connectivity
   - Track authentication success/failure rates
   - Alert on service degradation

2. **Metrics**
   - Request/response times
   - Error rates by endpoint
   - Token refresh frequency

3. **Logging**
   - Log authentication events
   - Track API usage patterns
   - Monitor for security anomalies

### Scalability

1. **Connection Pooling**
   - Implement HTTP client connection pooling
   - Configure appropriate timeouts
   - Handle connection failures gracefully

2. **Caching**
   - Cache network information
   - Implement token caching strategies
   - Consider response caching for read-heavy operations

## Troubleshooting

### Common Issues

1. **Authentication Failures**
   - Verify client credentials are correct
   - Check if credentials have expired
   - Ensure auth URL is accessible

2. **Network Connectivity**
   - Verify firewall rules allow outbound HTTPS
   - Check DNS resolution for Overledger domains
   - Test connectivity from deployment environment

3. **Configuration Issues**
   - Validate environment variable names and values
   - Check for trailing spaces or special characters
   - Ensure proper URL formatting

### Debug Mode

Enable debug logging by setting:
```env
LOG_LEVEL=debug
```

This will provide detailed information about:
- OAuth2 token requests and responses
- API request/response details
- Error stack traces

## Future Enhancements

1. **Webhook Support**
   - Implement webhook endpoints for transaction notifications
   - Add webhook signature verification
   - Support multiple webhook destinations

2. **Advanced Features**
   - Multi-signature transaction support
   - Smart contract interaction capabilities
   - Cross-chain transaction orchestration

3. **Performance Optimizations**
   - Implement request batching
   - Add response compression
   - Optimize data serialization

## Support

For issues related to:
- **Overledger API**: Contact Quant support at [support@quant.network](mailto:support@quant.network)
- **Integration Code**: Create an issue in the project repository
- **Documentation**: Submit a pull request with improvements