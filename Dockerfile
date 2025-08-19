# Stage 1: Build the frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files and install dependencies
COPY web/package.json web/package-lock.json ./
RUN npm install

# Copy the rest of the frontend source code
COPY web/ . 

# Remove legacy construction PoC components that were deleted from the codebase but may linger in Docker cache
RUN rm -f ./components/construction-*.tsx || true

# Build the static frontend
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:1.21-alpine AS go-builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
# Also copy local replace module files so `go mod download` can resolve `replace github.com/rutishh0/mesh-server => ./mesh-server`
COPY mesh-server/go.mod ./mesh-server/go.mod
COPY mesh-server/go.sum ./mesh-server/go.sum
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Build test binaries from ./test to execute at runtime for diagnostics
RUN mkdir -p /app/tests && \
    go test -c -o /app/tests/mesh_tests ./test/conformance || true && \
    go test -c -o /app/tests/integration_tests ./test/integration || true && \
    go build -o /app/tests/mesh_config_validation ./test/validation/mesh_config_validation.go || true && \
    go build -o /app/tests/mesh_validation ./test/validation/mesh_validation.go || true

# Stage 3: Create the final image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Go binary from the go-builder stage
COPY --from=go-builder /app/main .

# Copy .env file for fallback local configuration (optional; environment variables in Koyeb override)
COPY --from=go-builder /app/.env .

# Copy compiled test binaries
COPY --from=go-builder /app/tests ./tests

# Copy the built frontend from the frontend-builder stage
# Place it under ./web/out so the Gin static paths match
COPY --from=frontend-builder /app/web/out ./web/out

# Copy documentation
#COPY --from=go-builder /app/docs ./docs

# Set environment variables for production
ENV GIN_MODE=release
ENV ENVIRONMENT=production

# Expose the port the app runs on
EXPOSE 8080

# Run the application
CMD ["./main"]