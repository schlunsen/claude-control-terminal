.PHONY: build run clean install test help test-verbose test-coverage-html coverage-badge test-race lint lint-fix fmt deps build-all build-frontend build-go run-tui run-analytics run-agents run-chats run-help

# Binary name
BINARY_NAME=cct
BUILD_DIR=./cmd/cct
OUTPUT_DIR=.

# Build the application (with frontend)
build: build-frontend
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(OUTPUT_DIR)/$(BINARY_NAME) $(BUILD_DIR)
	@echo "✅ Build complete: ./$(BINARY_NAME)"

# Build frontend only
build-frontend:
	@echo "Building Nuxt frontend..."
	@cd internal/server/frontend && npm run generate
	@echo "✅ Frontend build complete"

# Build Go binary only (assumes frontend already built)
build-go:
	@echo "Building $(BINARY_NAME) (Go only)..."
	@go build -o $(OUTPUT_DIR)/$(BINARY_NAME) $(BUILD_DIR)
	@echo "✅ Build complete: ./$(BINARY_NAME)"

# Run the application
run:
	@go run $(BUILD_DIR)/main.go

# Run TUI (interactive mode - default)
run-tui:
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
	@rm -f coverage.out coverage-filtered.out coverage.html
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

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v -race ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@./scripts/filter-coverage.sh coverage.out > coverage-filtered.out
	@go tool cover -html=coverage-filtered.out -o coverage.html
	@echo ""
	@echo "📊 Coverage Summary:"
	@go tool cover -func=coverage.out | grep total | awk '{print "   Total Coverage:    " $$3}'
	@go tool cover -func=coverage-filtered.out | grep total | awk '{print "   Filtered Coverage: " $$3 " (excludes main.go, static files, interactive TUI)"}'
	@echo ""
	@echo "✅ Coverage report generated:"
	@echo "   HTML: coverage.html (filtered)"
	@echo "   Data: coverage.out"
	@echo "   Filtered: coverage-filtered.out"

# Open coverage report in browser
test-coverage-html: test-coverage
	@echo "Opening coverage report in browser..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		open coverage.html; \
	elif [ "$$(uname)" = "Linux" ]; then \
		xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html manually"; \
	else \
		echo "Please open coverage.html manually"; \
	fi

# Generate coverage badge locally (uses filtered coverage)
coverage-badge:
	@echo "Generating coverage badge..."
	@go test -coverprofile=coverage.out ./... >/dev/null 2>&1
	@./scripts/filter-coverage.sh coverage.out > coverage-filtered.out
	@COVERAGE=$$(go tool cover -func=coverage-filtered.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	COLOR="red"; \
	if [ $$(echo "$$COVERAGE > 80" | bc) -eq 1 ]; then COLOR="brightgreen"; \
	elif [ $$(echo "$$COVERAGE > 60" | bc) -eq 1 ]; then COLOR="yellow"; \
	elif [ $$(echo "$$COVERAGE > 40" | bc) -eq 1 ]; then COLOR="orange"; \
	fi; \
	echo "Filtered Coverage: $$COVERAGE% ($$COLOR)"; \
	echo "Badge URL: https://img.shields.io/badge/coverage-$$COVERAGE%25%20filtered-$$COLOR"

# Build for multiple platforms
build-all: build-frontend
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
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not found, falling back to go vet"; \
		echo "   Install golangci-lint: https://golangci-lint.run/usage/install/"; \
		go vet ./...; \
	fi
	@echo "✅ Lint complete"

# Lint with autofix
lint-fix:
	@echo "Linting code with autofix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not found"; \
		echo "   Install golangci-lint: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
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
	@echo ""
	@echo "Build & Run:"
	@echo "  make build               - Build the binary"
	@echo "  make run                 - Run the application (launches TUI)"
	@echo "  make run-tui             - Run the interactive TUI"
	@echo "  make run-analytics       - Run with --analytics flag"
	@echo "  make run-agents          - Run with --agents flag"
	@echo "  make run-chats           - Run with --chats flag"
	@echo "  make run-help            - Show help"
	@echo "  make build-all           - Build for all platforms"
	@echo ""
	@echo "Testing & Coverage:"
	@echo "  make test                - Run tests"
	@echo "  make test-verbose        - Run tests with verbose output and race detector"
	@echo "  make test-race           - Run tests with race detector"
	@echo "  make test-coverage       - Run tests with coverage report"
	@echo "  make test-coverage-html  - Run tests and open coverage report in browser"
	@echo "  make coverage-badge      - Generate coverage badge URL"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt                 - Format code"
	@echo "  make lint                - Lint code"
	@echo "  make lint-fix            - Lint code with autofix"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean               - Remove build artifacts"
	@echo "  make install             - Install to GOPATH/bin"
	@echo "  make deps                - Download and tidy dependencies"
