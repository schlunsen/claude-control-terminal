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

# Create a new release (tags, builds, updates Homebrew formula)
release version:
    #!/usr/bin/env bash
    set -euo pipefail
    \
    echo "🚀 Creating release v{{version}}..."; \
    \
    if [[ "{{version}}" != v* ]]; then \
        VERSION="v{{version}}"; \
    else \
        VERSION="{{version}}"; \
    fi; \
    \
    if [[ -n $(git status -s) ]]; then \
        echo "❌ Working directory is not clean. Please commit or stash changes first."; \
        exit 1; \
    fi; \
    \
    echo "📝 Creating tag $VERSION..."; \
    git tag -a "$VERSION" -m "Release $VERSION"; \
    git push origin "$VERSION"; \
    \
    echo "⏳ Waiting 60 seconds for GitHub Actions to build binaries..."; \
    sleep 60; \
    \
    echo "📦 Downloading binaries and calculating checksums..."; \
    mkdir -p /tmp/cct-release; \
    cd /tmp/cct-release; \
    \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-arm64" -o cct-darwin-arm64; \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-amd64" -o cct-darwin-amd64; \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-arm64" -o cct-linux-arm64; \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-amd64" -o cct-linux-amd64; \
    \
    SHA_DARWIN_ARM64=$(shasum -a 256 cct-darwin-arm64 | awk '{print $1}'); \
    SHA_DARWIN_AMD64=$(shasum -a 256 cct-darwin-amd64 | awk '{print $1}'); \
    SHA_LINUX_ARM64=$(shasum -a 256 cct-linux-arm64 | awk '{print $1}'); \
    SHA_LINUX_AMD64=$(shasum -a 256 cct-linux-amd64 | awk '{print $1}'); \
    \
    echo "✅ Checksums calculated"; \
    \
    echo "🍺 Updating Homebrew formula..."; \
    VERSION_NUM="${VERSION#v}"; \
    \
    cd /Users/schlunsen/projects/homebrew-cct; \
    \
    printf '%s\n' \
        'class Cct '"<"' Formula' \
        '  desc "High-performance CLI tool for Claude Code component templates and analytics"' \
        '  homepage "https://github.com/schlunsen/claude-templates-go"' \
        "  version \"$VERSION_NUM\"" \
        '' \
        '  # This is a precompiled binary, no build tools required' \
        '  uses_from_macos "unzip" '"=>"' :build' \
        '' \
        '  on_macos do' \
        '    if Hardware::CPU.arm?' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-arm64\"" \
        "      sha256 \"$SHA_DARWIN_ARM64\"" \
        '    else' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-amd64\"" \
        "      sha256 \"$SHA_DARWIN_AMD64\"" \
        '    end' \
        '  end' \
        '' \
        '  on_linux do' \
        '    if Hardware::CPU.arm?' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-arm64\"" \
        "      sha256 \"$SHA_LINUX_ARM64\"" \
        '    else' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-amd64\"" \
        "      sha256 \"$SHA_LINUX_AMD64\"" \
        '    end' \
        '  end' \
        '' \
        '  def install' \
        '    # The downloaded file is a precompiled binary' \
        '    downloaded_file = Dir["cct-*"].first' \
        '    bin.install downloaded_file '"=>"' "cct"' \
        '    chmod 0755, bin/"cct"' \
        '  end' \
        '' \
        '  test do' \
        '    system "#{bin}/cct", "--help"' \
        '  end' \
        'end' \
        > Formula/cct.rb; \
    \
    git add Formula/cct.rb; \
    git commit -m "chore: update cct formula to $VERSION"; \
    git push origin main; \
    \
    rm -rf /tmp/cct-release; \
    \
    echo ""; \
    echo "✅ Release $VERSION complete!"; \
    echo ""; \
    echo "📦 GitHub Release: https://github.com/schlunsen/claude-templates-go/releases/tag/$VERSION"; \
    echo ""; \
    echo "🍺 Homebrew users can upgrade with:"; \
    echo "   brew update && brew upgrade cct"; \
    echo ""; \
    echo "📝 Or force cache refresh:"; \
    echo "   brew untap schlunsen/cct && brew tap schlunsen/cct && brew install cct"; \
    echo "";

# Update Homebrew formula only (use after manual release)
update-homebrew version:
    #!/usr/bin/env bash
    set -euo pipefail
    \
    echo "🍺 Updating Homebrew formula for v{{version}}..."; \
    \
    if [[ "{{version}}" != v* ]]; then \
        VERSION="v{{version}}"; \
    else \
        VERSION="{{version}}"; \
    fi; \
    \
    echo "📦 Downloading binaries and calculating checksums..."; \
    mkdir -p /tmp/cct-release; \
    cd /tmp/cct-release; \
    \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-arm64" -o cct-darwin-arm64; \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-amd64" -o cct-darwin-amd64; \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-arm64" -o cct-linux-arm64; \
    curl -sL "https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-amd64" -o cct-linux-amd64; \
    \
    SHA_DARWIN_ARM64=$(shasum -a 256 cct-darwin-arm64 | awk '{print $1}'); \
    SHA_DARWIN_AMD64=$(shasum -a 256 cct-darwin-amd64 | awk '{print $1}'); \
    SHA_LINUX_ARM64=$(shasum -a 256 cct-linux-arm64 | awk '{print $1}'); \
    SHA_LINUX_AMD64=$(shasum -a 256 cct-linux-amd64 | awk '{print $1}'); \
    \
    echo "✅ Checksums calculated"; \
    \
    VERSION_NUM="${VERSION#v}"; \
    \
    cd /Users/schlunsen/projects/homebrew-cct; \
    \
    printf '%s\n' \
        'class Cct '"<"' Formula' \
        '  desc "High-performance CLI tool for Claude Code component templates and analytics"' \
        '  homepage "https://github.com/schlunsen/claude-templates-go"' \
        "  version \"$VERSION_NUM\"" \
        '' \
        '  # This is a precompiled binary, no build tools required' \
        '  uses_from_macos "unzip" '"=>"' :build' \
        '' \
        '  on_macos do' \
        '    if Hardware::CPU.arm?' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-arm64\"" \
        "      sha256 \"$SHA_DARWIN_ARM64\"" \
        '    else' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-darwin-amd64\"" \
        "      sha256 \"$SHA_DARWIN_AMD64\"" \
        '    end' \
        '  end' \
        '' \
        '  on_linux do' \
        '    if Hardware::CPU.arm?' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-arm64\"" \
        "      sha256 \"$SHA_LINUX_ARM64\"" \
        '    else' \
        "      url \"https://github.com/schlunsen/claude-templates-go/releases/download/$VERSION/cct-linux-amd64\"" \
        "      sha256 \"$SHA_LINUX_AMD64\"" \
        '    end' \
        '  end' \
        '' \
        '  def install' \
        '    # The downloaded file is a precompiled binary' \
        '    downloaded_file = Dir["cct-*"].first' \
        '    bin.install downloaded_file '"=>"' "cct"' \
        '    chmod 0755, bin/"cct"' \
        '  end' \
        '' \
        '  test do' \
        '    system "#{bin}/cct", "--help"' \
        '  end' \
        'end' \
        > Formula/cct.rb; \
    \
    git add Formula/cct.rb; \
    git commit -m "chore: update cct formula to $VERSION"; \
    git push origin main; \
    \
    rm -rf /tmp/cct-release; \
    \
    echo ""; \
    echo "✅ Homebrew formula updated to $VERSION!"; \
    echo ""; \
    echo "🍺 Users can upgrade with:"; \
    echo "   brew update && brew upgrade cct"; \
    echo ""; \
    echo "📝 Or force cache refresh:"; \
    echo "   brew untap schlunsen/cct && brew tap schlunsen/cct && brew install cct"; \
    echo "";
