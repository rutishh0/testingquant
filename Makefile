# Makefile for Quant-to-Coinbase Mesh Connector

# Variables
BINARY_NAME=quant-mesh-connector
BUILD_DIR=bin
MAIN_FILE=cmd/main.go
GO_FILES=$(shell find . -name "*.go" -type f)
ENV_FILE=.env

# Go commands
GO=go
GOFMT=gofmt
GOLINT=golint
GOVET=go vet
GOTEST=go test

# Docker commands
DOCKER=docker
DOCKER_COMPOSE=docker-compose

.PHONY: all build clean run test lint fmt vet docker docker-run docker-stop help

all: clean build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w $(GO_FILES)

# Lint code
lint:
	@echo "Linting code..."
	$(GOLINT) ./...

# Vet code
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	$(DOCKER) build -t $(BINARY_NAME) .

# Run with Docker Compose
docker-run:
	@echo "Starting with Docker Compose..."
	$(DOCKER_COMPOSE) up -d

# Stop Docker Compose
docker-stop:
	@echo "Stopping Docker Compose..."
	$(DOCKER_COMPOSE) down

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@if [ ! -f $(ENV_FILE) ]; then cp .env.example $(ENV_FILE); fi
	$(GO) mod download
	$(GO) mod tidy

# Help
help:
	@echo "Makefile commands:"
	@echo "  make build         - Build the application"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Lint code"
	@echo "  make vet           - Vet code"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run with Docker Compose"
	@echo "  make docker-stop   - Stop Docker Compose"
	@echo "  make setup         - Setup development environment"