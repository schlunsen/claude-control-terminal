.PHONY: build run clean install test help

# Binary name
BINARY_NAME=cct
BUILD_DIR=./cmd/cct
OUTPUT_DIR=.

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(OUTPUT_DIR)/$(BINARY_NAME) $(BUILD_DIR)
	@echo "✅ Build complete: ./$(BINARY_NAME)"

# Run the application
run:
	@go run $(BUILD_DIR)/main.go

# Run with specific flags
run-analytics:
	@go run $(BUILD_DIR)/main.go --analytics

run-agents:
	@go run $(BUILD_DIR)/main.go --agents

run-chats:
	@go run $(BUILD_DIR)/main.go --chats

run-help:
	@go run $(BUILD_DIR)/main.go --help

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(OUTPUT_DIR)/$(BINARY_NAME)
	@rm -rf dist/
	@echo "✅ Clean complete"

# Install the binary to GOPATH
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(BUILD_DIR)
	@echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 $(BUILD_DIR)
	@GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)-linux-arm64 $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe $(BUILD_DIR)
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

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

# Display help
help:
	@echo "Available commands:"
	@echo "  make build          - Build the binary"
	@echo "  make run            - Run the application"
	@echo "  make run-analytics  - Run with --analytics flag"
	@echo "  make run-agents     - Run with --agents flag"
	@echo "  make run-chats      - Run with --chats flag"
	@echo "  make run-help       - Show help"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install        - Install to GOPATH/bin"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Lint code"
	@echo "  make deps           - Download and tidy dependencies"
