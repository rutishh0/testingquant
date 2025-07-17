# GCP Deployment Guide

This guide provides step-by-step instructions for deploying the Quant-Mesh Connector on Google Cloud Platform (GCP).

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [GCP Project Setup](#gcp-project-setup)
3. [Database Setup](#database-setup)
4. [Service Account Setup](#service-account-setup)
5. [Environment Configuration](#environment-configuration)
6. [Building the Application](#building-the-application)
7. [Deploying to Cloud Run](#deploying-to-cloud-run)
8. [Setting Up Cloud SQL Proxy](#setting-up-cloud-sql-proxy)
9. [Configuring Secrets](#configuring-secrets)
10. [Setting Up CI/CD](#setting-up-cicd)
11. [Monitoring and Logging](#monitoring-and-logging)
12. [Scaling](#scaling)

## Prerequisites

- Google Cloud SDK installed and configured
- Docker installed locally
- Go 1.21 or later
- A GCP project with billing enabled
- Required APIs enabled (see next section)

## GCP Project Setup

1. **Create a new GCP project** or select an existing one
2. **Enable required APIs**:
   ```bash
   gcloud services enable \
     run.googleapis.com \
     sqladmin.googleapis.com \
     secretmanager.googleapis.com \
     cloudbuild.googleapis.com \
     containerregistry.googleapis.com \
     logging.googleapis.com \
     monitoring.googleapis.com
   ```

## Database Setup

1. **Create a Cloud SQL instance**:
   ```bash
   gcloud sql instances create quant-mesh-db \
     --database-version=POSTGRES_14 \
     --cpu=2 \
     --memory=7680MB \
     --region=europe-west2 \
     --root-password=your-secure-password
   ```

2. **Create a database**:
   ```bash
   gcloud sql databases create quantmesh \
     --instance=quant-mesh-db \
     --charset=utf8 \
     --collation=utf8_general_ci
   ```

3. **Create a database user**:
   ```bash
   gcloud sql users create quantmesh \
     --instance=quant-mesh-db \
     --password=your-secure-password
   ```

## Service Account Setup

1. **Create a service account**:
   ```bash
   gcloud iam service-accounts create quantmesh-sa \
     --display-name="Quant Mesh Service Account"
   ```

2. **Assign required roles**:
   ```bash
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:quantmesh-sa@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/cloudsql.client"

   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:quantmesh-sa@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/secretmanager.secretAccessor"
   ```

## Environment Configuration

1. **Create a `.env.production` file** with your production settings:
   ```env
   # Server Configuration
   SERVER_ADDRESS=:8080
   ENVIRONMENT=production
   LOG_LEVEL=info
   
   # Database Configuration
   DB_HOST=/cloudsql/your-project-id:europe-west2:quant-mesh-db
   DB_USER=quantmesh
   DB_PASSWORD=your-secure-password
   DB_NAME=quantmesh
   DB_SSLMODE=disable
   
   # Coinbase API
   COINBASE_API_KEY_ID=your-api-key-id
   COINBASE_API_SECRET=your-api-secret
   
   # Overledger Configuration
   OVERLEDGER_CLIENT_ID=your-client-id
   OVERLEDGER_CLIENT_SECRET=your-client-secret
   OVERLEDGER_AUTH_URL=https://auth.overledger.dev/oauth2/token
   OVERLEDGER_BASE_URL=https://api.overledger.dev
   ```

2. **Store secrets in Secret Manager**:
   ```bash
   # Store database password
   echo -n "your-secure-password" | gcloud secrets create DB_PASSWORD --data-file=-
   
   # Store API keys
   echo -n "your-api-key-id" | gcloud secrets create COINBASE_API_KEY_ID --data-file=-
   echo -n "your-api-secret" | gcloud secrets create COINBASE_API_SECRET --data-file=-
   
   # Grant access to the service account
   gcloud secrets add-iam-policy-binding DB_PASSWORD \
     --member="serviceAccount:quantmesh-sa@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/secretmanager.secretAccessor"
   ```

## Building the Application

1. **Build the Docker image**:
   ```bash
   docker build -t gcr.io/$PROJECT_ID/quant-mesh-connector:latest .
   ```

2. **Push the image to Container Registry**:
   ```bash
   docker push gcr.io/$PROJECT_ID/quant-mesh-connector:latest
   ```

## Deploying to Cloud Run

1. **Deploy the service**:
   ```bash
   gcloud run deploy quant-mesh-connector \
     --image gcr.io/$PROJECT_ID/quant-mesh-connector:latest \
     --platform managed \
     --region europe-west2 \
     --allow-unauthenticated \
     --add-cloudsql-instances $PROJECT_ID:europe-west2:quant-mesh-db \
     --service-account quantmesh-sa@$PROJECT_ID.iam.gserviceaccount.com \
     --set-env-vars "ENVIRONMENT=production" \
     --set-secrets=DB_PASSWORD=DB_PASSWORD:latest \
     --set-secrets=COINBASE_API_KEY_ID=COINBASE_API_KEY_ID:latest \
     --set-secrets=COINBASE_API_SECRET=COINBASE_API_SECRET:latest
   ```

2. **Update the service with environment variables**:
   ```bash
   gcloud run services update quant-mesh-connector \
     --region europe-west2 \
     --update-env-vars "DB_HOST=/cloudsql/$PROJECT_ID:europe-west2:quant-mesh-db" \
     --update-env-vars "DB_USER=quantmesh" \
     --update-env-vars "DB_NAME=quantmesh" \
     --update-env-vars "DB_SSLMODE=disable"
   ```

## Setting Up Cloud SQL Proxy

For local development with the Cloud SQL instance:

1. **Install the Cloud SQL Auth Proxy**:
   ```bash
   curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.6.1/cloud-sql-proxy.linux.amd64
   chmod +x cloud-sql-proxy
   ```

2. **Run the proxy**:
   ```bash
   ./cloud-sql-proxy --address 0.0.0.0 --port 5432 "$PROJECT_ID:europe-west2:quant-mesh-db"
   ```

## Configuring Secrets

1. **Update secrets**:
   ```bash
   # Update a secret
   echo -n "new-secret-value" | gcloud secrets versions add SECRET_NAME --data-file=-
   ```

2. **Redeploy the service** to pick up secret changes:
   ```bash
   gcloud run services update-traffic quant-mesh-connector \
     --region europe-west2 \
     --to-latest
   ```

## Setting Up CI/CD

1. **Create a `cloudbuild.yaml` file**:
   ```yaml
   steps:
     # Build the container image
     - name: 'gcr.io/cloud-builders/docker'
       args: ['build', '-t', 'gcr.io/$PROJECT_ID/quant-mesh-connector:$COMMIT_SHA', '.']
     
     # Push the container image to Container Registry
     - name: 'gcr.io/cloud-builders/docker'
       args: ['push', 'gcr.io/$PROJECT_ID/quant-mesh-connector:$COMMIT_SHA']
     
     # Deploy container image to Cloud Run
     - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
       entrypoint: gcloud
       args:
         - 'run'
         - 'deploy'
         - 'quant-mesh-connector'
         - '--image'
         - 'gcr.io/$PROJECT_ID/quant-mesh-connector:$COMMIT_SHA'
         - '--region'
         - 'europe-west2'
         - '--platform'
         - 'managed'
         - '--add-cloudsql-instances'
         - '$PROJECT_ID:europe-west2:quant-mesh-db'
         - '--service-account'
         - 'quantmesh-sa@$PROJECT_ID.iam.gserviceaccount.com'
         - '--set-env-vars'
         - 'ENVIRONMENT=production'
         - '--set-secrets=DB_PASSWORD=DB_PASSWORD:latest'
         - '--set-secrets=COINBASE_API_KEY_ID=COINBASE_API_KEY_ID:latest'
         - '--set-secrets=COINBASE_API_SECRET=COINBASE_API_SECRET:latest'
   ```

2. **Create a trigger** in Cloud Build to build on git push

## Monitoring and Logging

1. **View logs**:
   ```bash
   gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=quant-mesh-connector" --limit=50
   ```

2. **Set up monitoring dashboards** in Cloud Monitoring

## Scaling

1. **Configure auto-scaling**:
   ```bash
   gcloud run services update quant-mesh-connector \
     --region europe-west2 \
     --min-instances=1 \
     --max-instances=10 \
     --concurrency=80
   ```

## Troubleshooting

- **Connection issues**: Check Cloud SQL Proxy logs and IAM permissions
- **Deployment failures**: Check Cloud Build logs
- **Application errors**: Check Cloud Run logs
- **Secret access issues**: Verify IAM permissions for the service account

## Next Steps

- Set up a custom domain with Cloud Load Balancer
- Configure SSL certificates
- Set up alerting policies
- Implement backup strategy for the database
