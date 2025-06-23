# Railway Deployment Guide

## Prerequisites

1. Railway account
2. GitHub repository connected to Railway

## Environment Variables Required

Set these environment variables in your Railway project:

```
# Required for production
ENVIRONMENT=production
GIN_MODE=release

# Mesh API Configuration
MESH_API_URL=https://rosetta-ethereum.coinbase.com:8080
API_KEY=your-coinbase-api-key

# Overledger Configuration
OVERLEDGER_CLIENT_ID=your-client-id
OVERLEDGER_CLIENT_SECRET=your-client-secret
OVERLEDGER_AUTH_URL=https://auth.overledger.dev/oauth2/token
OVERLEDGER_BASE_URL=https://api.overledger.dev

# Logging
LOG_LEVEL=info
```

## Deployment Steps

1. Connect your GitHub repository to Railway
2. Set the environment variables listed above
3. Railway will automatically detect the Dockerfile and build the application
4. The application will be available on the provided Railway URL

## Troubleshooting

- Ensure all environment variables are set correctly
- Check Railway logs for any startup errors
- Verify that the PORT environment variable is handled correctly (Railway sets this automatically)