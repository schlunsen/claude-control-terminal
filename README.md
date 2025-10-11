# Go Claude Templates

Go port of the [claude-code-templates](https://github.com/davila7/claude-code-templates) Node.js CLI tool.

## 🚀 Project Status

**Migration in Progress** - Porting from Node.js to Go for better performance and easier distribution.

### Why Go?

- **Fast compilation** - Builds in seconds
- **Single binary** - No node_modules, just one executable
- **Better performance** - 10-50x faster startup, 3-5x lower memory usage
- **Easy cross-compilation** - Build for Linux, macOS, Windows from anywhere
- **Great libraries** - Excellent ecosystem for CLI, web servers, and file operations

## 📁 Project Structure

```
go-claude-templates/
├── cmd/
│   └── cct/           # Main CLI application
├── internal/
│   ├── analytics/     # Analytics dashboard backend
│   ├── components/    # Component management (agents, commands, MCPs)
│   ├── fileops/       # File operations and template handling
│   ├── server/        # Web server (Fiber)
│   └── websocket/     # WebSocket server for real-time updates
├── pkg/
│   └── utils/         # Shared utilities
└── go.mod
```

## 🛠️ Tech Stack

- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Terminal UI**: [Pterm](https://github.com/pterm/pterm)
- **Web Framework**: [Fiber](https://github.com/gofiber/fiber)
- **WebSocket**: [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **File Watching**: [fsnotify](https://github.com/fsnotify/fsnotify)
- **System Info**: [gopsutil](https://github.com/shirou/gopsutil)

## 🏗️ Build & Run

```bash
# Build the binary
go build -o cct ./cmd/cct

# Run directly
go run ./cmd/cct

# Install globally
go install ./cmd/cct
```

## 📋 Migration Progress

- [x] Project structure created
- [ ] Core dependencies added
- [ ] CLI framework (Cobra)
- [ ] Terminal UI (Pterm)
- [ ] File operations
- [ ] Component management
- [ ] Analytics backend
- [ ] Web server
- [ ] WebSocket server
- [ ] Frontend integration
- [ ] Testing
- [ ] Documentation

## 📖 Original Project

Based on [claude-code-templates](https://github.com/davila7/claude-code-templates) by davila7.

## 📄 License

MIT
