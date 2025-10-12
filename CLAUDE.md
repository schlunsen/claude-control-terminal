# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

**claude-control-terminal** is a high-performance Go port of the Node.js claude-code-templates CLI tool. It provides component templates, analytics dashboards, and real-time monitoring for Claude Code projects with superior performance and easy deployment.

### Key Features
- 🎮 **Control Center**: Comprehensive wrapper for Claude Code environments
- 🚀 **CLI Tool**: Component installation (agents, commands, MCPs, settings, hooks)
- 🐳 **Docker Support**: Containerize Claude environments with one command
- 📊 **Analytics Dashboard**: Real-time conversation monitoring with WebSocket support
- 🔧 **Component Management**: 600+ agents, 200+ commands, MCPs from GitHub
- ⚡ **Performance**: 10-50x faster startup, 3-5x lower memory vs Node.js
- 📦 **Single Binary**: No dependencies, just one executable
- 🌐 **Web Server**: Fiber-based REST API with real-time updates

## Technology Stack

### Core Technologies
- **Language**: Go 1.23+ (using go1.24.8 toolchain)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) - Industry-standard CLI
- **Terminal UI**: [Pterm](https://github.com/pterm/pterm) - Beautiful terminal output
- **Web Framework**: [Fiber v2](https://github.com/gofiber/fiber) - Express-like HTTP framework
- **WebSocket**: [Gorilla WebSocket](https://github.com/gorilla/websocket) + Fiber WebSocket
- **File Watching**: [fsnotify](https://github.com/fsnotify/fsnotify) - Cross-platform file notifications
- **System Info**: [gopsutil](https://github.com/shirou/gopsutil) - Process detection

### Project Structure

```text
claude-control-terminal/
├── cmd/cct/                    # CLI entry point
│   └── main.go                 # Application bootstrap
├── internal/                   # Private application code
│   ├── analytics/              # Analytics backend modules
│   │   ├── state_calculator.go       # Conversation state logic
│   │   ├── process_detector.go       # Process monitoring
│   │   ├── conversation_analyzer.go  # JSONL parsing
│   │   └── file_watcher.go          # Real-time file watching
│   ├── cmd/                    # CLI commands & UI
│   │   ├── root.go            # Cobra root command
│   │   └── banner.go          # Pterm UI helpers
│   ├── components/             # Component installers
│   │   ├── agent.go           # Agent installation
│   │   ├── command.go         # Command installation
│   │   └── mcp.go             # MCP installation
│   ├── docker/                 # Docker support (NEW in v0.2.0)
│   │   ├── docker.go          # Docker operations
│   │   ├── dockerfile_generator.go  # Dockerfile generation
│   │   └── compose_generator.go     # docker-compose generation
│   ├── fileops/                # File operations
│   │   ├── github.go          # GitHub API downloads
│   │   ├── template.go        # Template processing
│   │   └── utils.go           # File utilities
│   ├── server/                 # Web server
│   │   ├── server.go          # Fiber HTTP server
│   │   ├── static.go          # Embedded static files
│   │   └── static/            # Frontend HTML/JS
│   └── websocket/              # Real-time updates
│       └── websocket.go       # WebSocket hub
├── pkg/                        # Public libraries (future)
│   └── utils/
├── Makefile                    # Make build automation
├── justfile                    # Just task runner
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
└── README.md                   # User documentation
```

## Development Commands

### Building & Running

```bash
# Build binary (fast - ~2 seconds)
make build
# or
just build

# Run directly
go run ./cmd/cct

# Install globally
go install ./cmd/cct
# or
make install
```

### Component Installation

```bash
# Install agents
./cct --agent security-auditor
./cct --agent "api-tester,code-reviewer,debug-assistant"

# Install commands
./cct --command check-file
./cct --command "deploy,test,build"

# Install MCPs
./cct --mcp postgresql
./cct --mcp "github,supabase,filesystem"

# Mix components
./cct --agent security-auditor --command vulnerability-scan --mcp postgres
```

### Analytics Dashboard

```bash
# Launch analytics server
./cct --analytics
# or
make run-analytics
# or
just analytics

# Access dashboard
open http://localhost:3333

# API endpoints
curl http://localhost:3333/api/data
curl http://localhost:3333/api/conversations
curl http://localhost:3333/api/processes
curl http://localhost:3333/api/stats
```

### Development Workflow

```bash
# Format code
make fmt
just fmt

# Run tests
make test
just test

# Test with coverage
make test-coverage

# Cross-platform builds
make build-all
just build-all
# Outputs: dist/cct-{linux,darwin,windows}-{amd64,arm64}

# Clean build artifacts
make clean
just clean
```

## Code Style & Best Practices

### Go Idioms

1. **Error Handling**: Always check and handle errors explicitly
   ```go
   if err != nil {
       return fmt.Errorf("failed to do X: %w", err)
   }
   ```

2. **Struct Initialization**: Use composite literals
   ```go
   conversation := Conversation{
       ID:       id,
       Status:   "active",
       Tokens:   tokens,
   }
   ```

3. **Goroutines**: Use for concurrent operations
   ```go
   go fileWatcher.Start()
   go wsHub.Run()
   ```

4. **Channels**: For communication between goroutines
   ```go
   stopChan := make(chan bool)
   broadcast := make(chan []byte, 256)
   ```

### Project Conventions

1. **Package Organization**:
   - `internal/` for private code (main application)
   - `pkg/` for public libraries (reusable code)
   - `cmd/` for executable entry points

2. **Naming**:
   - Packages: lowercase, single word (`analytics`, `server`)
   - Structs: PascalCase (`ConversationAnalyzer`, `ProcessDetector`)
   - Functions: camelCase for private, PascalCase for exported
   - Constants: PascalCase or UPPER_SNAKE_CASE for package-level

3. **File Naming**:
   - Use snake_case for Go files (`state_calculator.go`)
   - Group related functions in same file
   - Keep files focused on single responsibility

4. **Comments**:
   - Document all exported types, functions, methods
   - Use godoc format
   ```go
   // ConversationAnalyzer handles conversation data loading and analysis.
   // It provides methods for parsing JSONL files and extracting metrics.
   type ConversationAnalyzer struct { ... }
   ```

### Testing Guidelines

```go
// Test file naming: *_test.go
// Test function naming: TestFunctionName

func TestStateCalculator_DetermineState(t *testing.T) {
    sc := NewStateCalculator()

    // Arrange
    messages := []Message{...}
    lastModified := time.Now()

    // Act
    state := sc.DetermineConversationState(messages, lastModified, nil)

    // Assert
    if state != "Claude Code working..." {
        t.Errorf("expected 'Claude Code working...', got '%s'", state)
    }
}
```

## Architecture & Design Patterns

### Analytics Backend

The analytics system is modular and follows the Single Responsibility Principle:

1. **StateCalculator**: Determines conversation state based on timestamps and messages
2. **ProcessDetector**: Monitors running Claude CLI processes
3. **ConversationAnalyzer**: Parses JSONL conversation files
4. **FileWatcher**: Monitors file changes for real-time updates

### Concurrent Patterns

```go
// Hub pattern for WebSocket connections
type Hub struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan []byte
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
    mutex      sync.RWMutex
}

// Run hub in goroutine
go hub.Run()

// File watcher with channels
go fileWatcher.watchLoop()
go fileWatcher.periodicRefresh()
```

### Server Architecture

The Fiber server follows middleware patterns:

```go
app := fiber.New()
app.Use(cors.New())
app.Use(logger.New())

// API routes
api := app.Group("/api")
api.Get("/data", handleGetData)

// WebSocket endpoint
app.Get("/ws", websocket.New(handler))
```

## Common Tasks

### Adding a New API Endpoint

1. Add handler method to `internal/server/server.go`:
   ```go
   func (s *Server) handleNewEndpoint(c *fiber.Ctx) error {
       data := s.getData()
       return c.JSON(fiber.Map{
           "result": data,
       })
   }
   ```

2. Register route in `setupRoutes()`:
   ```go
   api.Get("/new-endpoint", s.handleNewEndpoint)
   ```

### Adding a New CLI Command

1. Add flag in `internal/cmd/root.go`:
   ```go
   var newCommand bool
   rootCmd.Flags().BoolVar(&newCommand, "new-command", false, "description")
   ```

2. Add handler in `handleCommand()`:
   ```go
   if newCommand {
       ShowSpinner("Executing new command...")
       // Implementation
       return
   }
   ```

### Adding a New Component Type

1. Create installer in `internal/components/`:
   ```go
   type NewComponentInstaller struct {
       config *fileops.GitHubConfig
   }

   func (nci *NewComponentInstaller) Install(name, targetDir string) error {
       // Download from GitHub
       // Install to appropriate directory
   }
   ```

2. Integrate in `internal/cmd/root.go`

### Embedding Static Files

```go
//go:embed static/file.html
var fileHTML []byte

func ServeFile(c *fiber.Ctx) error {
    c.Set("Content-Type", "text/html")
    return c.Send(fileHTML)
}
```

## Performance Considerations

### Benchmarks (vs Node.js version)

| Metric | Node.js | Go | Improvement |
|--------|---------|-----|-------------|
| Build Time | npm install (minutes) | 2-5 seconds | 50-100x faster |
| Binary Size | 50MB+ (node_modules) | ~15MB | 3x smaller |
| Startup Time | ~500ms | <10ms | 50x faster |
| Memory Usage | ~80MB baseline | ~15MB | 5x lower |
| Concurrent Connections | Event loop | Goroutines | Unlimited scaling |

### Optimization Tips

1. **Avoid Allocations**: Reuse structs and slices
2. **Use sync.Pool**: For frequently allocated objects
3. **Buffer Channels**: Use buffered channels for high-throughput
4. **Context Timeouts**: Set timeouts for long-running operations
5. **Profile**: Use `pprof` for performance analysis

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Debugging & Troubleshooting

### Enable Verbose Logging

```bash
./cct --analytics --verbose
```

### Check Build Issues

```bash
# Verify Go version
go version  # Should be 1.23+

# Check dependencies
go mod verify
go mod tidy

# Clear cache
go clean -cache -modcache -i -r
```

### Common Issues

1. **Port 3333 in use**:
   ```bash
   lsof -i :3333
   kill -9 <PID>
   ```

2. **WebSocket connection fails**:
   - Check firewall settings
   - Verify CORS configuration
   - Test with `wscat -c ws://localhost:3333/ws`

3. **Component download fails**:
   - Check internet connection
   - Verify GitHub API rate limits
   - Check component name spelling

## Git Workflow

### Commit Message Format

```text
<type>: <subject>

<body>

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

### Creating Pull Requests

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes and commit
git add .
git commit -m "feat: add new feature"

# Push and create PR
git push origin feature/new-feature
gh pr create --title "Add new feature" --body "Description"
```

## Deployment

### Building for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o cct ./cmd/cct

# Cross-compile for all platforms
make build-all

# Outputs:
# - dist/cct-linux-amd64
# - dist/cct-linux-arm64
# - dist/cct-darwin-amd64
# - dist/cct-darwin-arm64
# - dist/cct-windows-amd64.exe
```

### Installation Methods

```bash
# Direct binary
curl -L https://github.com/schlunsen/claude-control-terminal/releases/latest/download/cct-<platform> -o cct
chmod +x cct
sudo mv cct /usr/local/bin/

# Go install
go install github.com/davila7/claude-control-terminal/cmd/cct@latest

# From source
git clone https://github.com/schlunsen/claude-control-terminal
cd claude-control-terminal
make install
```

## Resources

### Documentation
- [Cobra CLI](https://github.com/spf13/cobra)
- [Fiber Framework](https://docs.gofiber.io/)
- [Pterm](https://github.com/pterm/pterm)
- [fsnotify](https://github.com/fsnotify/fsnotify)

### Original Project
- [claude-code-templates (Node.js)](https://github.com/davila7/claude-code-templates)

### Go Resources
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

## License

MIT License - See LICENSE file for details

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Write tests
5. Run `make fmt && make test`
6. Submit a pull request

---

**Version**: 2.0.0-go
**Author**: Port by Claude Code
**Original**: davila7/claude-code-templates
