# Railway.app Deployment Guide

This guide provides step-by-step instructions for deploying the Quant-Mesh Connector on Railway.app, a modern application hosting platform with built-in database and secrets management.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Railway Project Setup](#railway-project-setup)
3. [Database Setup](#database-setup)
4. [Environment Variables](#environment-variables)
5. [Deploying the Application](#deploying-the-application)
6. [Custom Domain Setup](#custom-domain-setup)
7. [Monitoring and Logs](#monitoring-and-logs)
8. [Scaling](#scaling)
9. [Troubleshooting](#troubleshooting)

## Prerequisites

- A [Railway account](https://railway.app/)
- [Git](https://git-scm.com/) installed locally
- [Railway CLI](https://docs.railway.app/develop/cli) (optional)
- [Docker](https://www.docker.com/) (for local testing)

## Railway Project Setup

1. **Create a new Railway project**:
   - Go to [Railway Dashboard](https://railway.app/dashboard)
   - Click "New Project"
   - Select "Deploy from GitHub repo" (recommended) or "Deploy from template"

2. **Connect your GitHub repository**:
   - Select your repository from the list
   - Choose the branch you want to deploy (e.g., `main` or `master`)
   - Click "Deploy Now"

## Database Setup

1. **Add a PostgreSQL database**:
   - In your Railway project, click "New"
   - Select "Database" > "PostgreSQL"
   - Wait for the database to be provisioned

2. **Get the database connection string**:
   - Click on the PostgreSQL service
   - Go to the "Connect" tab
   - Copy the "Postgres Connection URL"
   - It should look like: `postgresql://postgres:password@containers-us-west-xxx.railway.app:5432/railway`

## Environment Variables

1. **Set up environment variables** in Railway:
   - In your Railway project, go to the "Variables" tab
   - Add the following variables from your `.env` file:
     ```
     # Server Configuration
     SERVER_ADDRESS=:8080
     ENVIRONMENT=production
     LOG_LEVEL=info
     
     # Database Configuration (use the connection string from above)
     DATABASE_URL=postgresql://postgres:password@containers-us-west-xxx.railway.app:5432/railway
     
     # Coinbase API
     COINBASE_API_KEY_ID=your-api-key-id
     COINBASE_API_SECRET=your-api-secret
     COINBASE_API_URL=https://api.cdp.coinbase.com
     
     # Overledger Configuration
     OVERLEDGER_CLIENT_ID=your-client-id
     OVERLEDGER_CLIENT_SECRET=your-client-secret
     OVERLEDGER_AUTH_URL=https://auth.overledger.dev/oauth2/token
     OVERLEDGER_BASE_URL=https://api.overledger.dev
     OVERLEDGER_TX_SIGNING_KEY_ID=your-signing-key-id
     ```

2. **Mark sensitive variables as secret** (recommended for API keys and secrets)
   - Toggle the "Encrypt" switch for sensitive values
   - This ensures they are stored securely and not exposed in logs

## Deploying the Application

### Option 1: Deploy from GitHub (Recommended)
1. Push your code to your connected GitHub repository
2. Railway will automatically detect the changes and trigger a new deployment
3. Monitor the deployment in the Railway dashboard

### Option 2: Deploy using Railway CLI
1. Install the Railway CLI:
   ```bash
   npm i -g @railway/cli
   ```
2. Login to your Railway account:
   ```bash
   railway login
   ```
3. Link your project:
   ```bash
   railway link
   ```
4. Deploy your application:
   ```bash
   railway up
   ```

## Custom Domain Setup

1. **Add a custom domain**:
   - Go to your project in Railway
   - Click on "Settings" > "Domains"
   - Click "Add Domain"
   - Follow the instructions to verify domain ownership

2. **Configure DNS records**:
   - Add the provided CNAME or A records to your domain's DNS settings
   - Wait for DNS propagation (may take up to 48 hours)

## Monitoring and Logs

1. **View application logs**:
   - In the Railway dashboard, go to your service
   - Click on the "Logs" tab
   - You can filter logs by level (info, error, etc.)

2. **Set up monitoring**:
   - Railway provides basic metrics in the "Metrics" tab
   - For advanced monitoring, connect to an external service like Datadog or New Relic

## Scaling

1. **Adjust resources**:
   - Go to your service in the Railway dashboard
   - Click on "Settings" > "Resources"
   - Adjust CPU and memory allocation as needed

2. **Enable auto-scaling**:
   - In the "Resources" section
   - Toggle "Auto Scale" to enable automatic scaling based on load

## Troubleshooting

### Common Issues

1. **Application not starting**:
   - Check the logs for error messages
   - Verify all required environment variables are set
   - Ensure the `SERVER_ADDRESS` matches the port Railway expects (usually `:8080`)

2. **Database connection issues**:
   - Verify the `DATABASE_URL` is correct
   - Check if the database is running and accessible
   - Ensure your IP is whitelisted if using IP restrictions

3. **Environment variables not loading**:
   - Make sure all variables are set in the Railway dashboard
   - Check for typos in variable names
   - Restart the service after changing environment variables

### Getting Help

- [Railway Documentation](https://docs.railway.app/)
- [Join Railway Discord](https://discord.gg/railway)
- Check the "Troubleshooting" section in Railway dashboard

## Next Steps

- Set up automated backups for your database
- Configure alerts for downtime or errors
- Set up a staging environment for testing changes
- Implement a CI/CD pipeline for automated testing and deployment
