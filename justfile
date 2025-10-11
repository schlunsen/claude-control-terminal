# Go Claude Templates - Just Commands
# Install just: https://github.com/casey/just

# Default recipe to display help
default:
    @just --list

# Build the application
build:
    @echo "Building cct..."
    @go build -o cct ./cmd/cct
    @echo "✅ Build complete: ./cct"

# Run the application
run:
    @go run ./cmd/cct

# Run with analytics flag
analytics:
    @go run ./cmd/cct --analytics

# Run with agents flag
agents:
    @go run ./cmd/cct --agents

# Run with chats flag
chats:
    @go run ./cmd/cct --chats

# Run help
help:
    @go run ./cmd/cct --help

# Install specific agent
install-agent agent:
    @go run ./cmd/cct --agent {{agent}}

# Install specific command
install-command command:
    @go run ./cmd/cct --command {{command}}

# Install specific MCP
install-mcp mcp:
    @go run ./cmd/cct --mcp {{mcp}}

# Clean build artifacts
clean:
    @echo "Cleaning..."
    @rm -f cct
    @rm -rf dist/
    @echo "✅ Clean complete"

# Install to GOPATH/bin
install:
    @echo "Installing cct..."
    @go install ./cmd/cct
    @echo "✅ Installed to $(go env GOPATH)/bin/cct"

# Run all tests
test:
    @echo "Running tests..."
    @go test -v ./...

# Run tests with coverage
test-coverage:
    @echo "Running tests with coverage..."
    @go test -v -coverprofile=coverage.out ./...
    @go tool cover -html=coverage.out -o coverage.html
    @echo "✅ Coverage report: coverage.html"

# Build for all platforms
build-all:
    @echo "Building for multiple platforms..."
    @mkdir -p dist
    @GOOS=linux GOARCH=amd64 go build -o dist/cct-linux-amd64 ./cmd/cct
    @GOOS=linux GOARCH=arm64 go build -o dist/cct-linux-arm64 ./cmd/cct
    @GOOS=darwin GOARCH=amd64 go build -o dist/cct-darwin-amd64 ./cmd/cct
    @GOOS=darwin GOARCH=arm64 go build -o dist/cct-darwin-arm64 ./cmd/cct
    @GOOS=windows GOARCH=amd64 go build -o dist/cct-windows-amd64.exe ./cmd/cct
    @echo "✅ Build complete for all platforms in ./dist/"

# Format code
fmt:
    @echo "Formatting code..."
    @go fmt ./...
    @echo "✅ Format complete"

# Lint code
lint:
    @echo "Linting code..."
    @golangci-lint run || go vet ./...
    @echo "✅ Lint complete"

# Download and tidy dependencies
deps:
    @echo "Downloading dependencies..."
    @go mod download
    @go mod tidy
    @echo "✅ Dependencies updated"

# Run the app with verbose logging
verbose:
    @go run ./cmd/cct --verbose

# Quick test - build and run
quick: build
    @./cct --help

# Development mode - build and test
dev: fmt build test
    @echo "✅ Development checks passed"
