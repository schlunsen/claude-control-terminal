# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

**claude-control-terminal** is a high-performance Go port of the Node.js claude-code-templates CLI tool. It provides component templates, analytics dashboards, and real-time monitoring for Claude Code projects with superior performance and easy deployment.

### Key Features
- 🎮 **Control Center**: Comprehensive wrapper for Claude Code environments
- 🚀 **CLI Tool**: Component installation (agents, commands, MCPs, settings, hooks)
- 🤖 **Agent Server**: Go-based WebSocket server for real-time Claude agent conversations using claude-agent-sdk-go
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
- **Agent SDK**: [claude-agent-sdk-go](https://github.com/schlunsen/claude-agent-sdk-go) - Claude agent conversation SDK
- **File Watching**: [fsnotify](https://github.com/fsnotify/fsnotify) - Cross-platform file notifications
- **System Info**: [gopsutil](https://github.com/shirou/gopsutil) - Process detection

### Project Structure

```text
claude-control-terminal/
├── cmd/cct/                    # CLI entry point
│   └── main.go                 # Application bootstrap
├── internal/                   # Private application code
│   ├── server/                 # Web server & agent functionality
│   │   ├── server.go          # Fiber HTTP/HTTPS server
│   │   ├── config.go          # Configuration management
│   │   ├── tls.go             # TLS certificate generation
│   │   ├── auth.go            # API key authentication
│   │   ├── agents/            # Agent server implementation
│   │   │   ├── agent_handler.go    # WebSocket agent handler
│   │   │   ├── session_manager.go  # Agent session management
│   │   │   ├── messages.go         # Message types
│   │   │   └── config.go           # Agent configuration
│   │   ├── static.go          # Embedded static files
│   │   └── frontend/          # Nuxt 4 SPA frontend
│   │       ├── app/           # Nuxt app directory (IMPORTANT!)
│   │       │   ├── app.vue    # Root app component
│   │       │   ├── pages/     # Vue pages (index.vue, agents.vue)
│   │       │   └── composables/ # Vue composables (useAgentWebSocket.ts)
│   │       ├── components/    # Vue components (SessionMetrics.vue)
│   │       ├── types/         # TypeScript types
│   │       ├── nuxt.config.ts # Nuxt configuration
│   │       └── package.json   # Frontend dependencies
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
│   ├── database/               # Database layer (SQLite)
│   │   ├── database.go        # Database initialization & connection management
│   │   ├── schema.sql         # Complete unified database schema (embedded)
│   │   ├── models.go          # Data model structs
│   │   ├── repository.go      # Data access layer (CRUD operations)
│   │   └── git_utils.go       # Git metadata extraction helpers
│   ├── docker/                 # Docker support (NEW in v0.2.0)
│   │   ├── docker.go          # Docker operations
│   │   ├── dockerfile_generator.go  # Dockerfile generation
│   │   └── compose_generator.go     # docker-compose generation
│   ├── fileops/                # File operations
│   │   ├── github.go          # GitHub API downloads
│   │   ├── template.go        # Template processing
│   │   └── utils.go           # File utilities
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

## Database Architecture

### Overview

CCT uses **SQLite** as its embedded database for persistent storage of command history, user messages, provider configurations, agent sessions, and more. The database provides a unified data layer shared between the TUI, analytics server, and agent handler.

**Database Location**: `~/.claude/cct/cct.db`

**Database Engine**: SQLite 3 with WAL mode enabled for concurrent access

### Key Features

- ✅ **Single Database File**: All data consolidated in one location
- ✅ **WAL Mode**: Write-Ahead Logging for concurrent read/write access
- ✅ **Automatic Migrations**: Schema evolves automatically on startup
- ✅ **Singleton Pattern**: Single database connection shared across components
- ✅ **Repository Pattern**: Clean data access layer with type-safe operations
- ✅ **Embedded Schema**: Schema definition compiled into binary via `//go:embed`
- ✅ **Secure Permissions**: Database file has 0600 permissions (user read/write only)

### Database Tables

The database schema includes 11 core tables organized by functionality:

#### Command History Tables
- **`shell_commands`**: Records of all Bash tool executions
  - Tracks: command, exit code, stdout/stderr, duration, working directory, git branch
  - Used for: Command history, analytics, debugging

- **`claude_commands`**: Records of all Claude Code tool invocations
  - Tracks: tool name, parameters (JSON), results (JSON), success/error status
  - Used for: Tool usage analytics, audit trail

- **`command_stats`**: Aggregated command statistics
  - Tracks: execution count, success rate, average duration
  - Used for: Performance monitoring, most-used commands

#### Conversation & Session Tables
- **`conversations`**: Metadata for Claude Code conversation sessions
  - Tracks: project path, start/end times, total commands, token usage, status
  - Used for: Session management, analytics dashboard

- **`user_messages`**: User input messages intercepted by hooks
  - Tracks: message content, length, timestamps, context metadata
  - Used for: Prompt analytics, conversation history

- **`notifications`**: Permission requests and idle alerts
  - Tracks: notification type, tool name, command details, timestamps
  - Used for: Permission analytics, user interaction patterns

#### Agent Session Tables (Persistent Agent Conversations)
- **`agent_sessions`**: Agent conversation sessions from unified server
  - Tracks: session ID, status, cost, duration, message count, model info
  - Used for: Agent session persistence, cost tracking, session restoration

- **`agent_messages`**: Individual messages within agent sessions
  - Tracks: role (user/assistant/system), content, thinking, tool uses, tokens
  - Used for: Conversation history, message replay, debugging

#### Configuration Tables
- **`providers`**: AI provider configurations (Anthropic, OpenRouter, etc.)
  - Tracks: provider ID, API key, custom URL, model name, is_current flag
  - Used for: Multi-provider support, API key persistence

- **`user_settings`**: User preferences and application settings
  - Tracks: key/value pairs with type metadata
  - Used for: Settings persistence (e.g., diff display location)

### Database Schema

The complete schema is defined in `internal/database/schema.sql` and embedded into the binary:

```go
//go:embed schema.sql
var schemaSQL string
```

Key schema features:
- **Foreign Keys**: Enabled with `PRAGMA foreign_keys = ON`
- **Indexes**: 23+ indexes for optimal query performance
- **Constraints**: CHECK constraints for data validation
- **Timestamps**: Automatic `created_at` and `updated_at` tracking
- **JSON Support**: Stores complex data as JSON strings

### Database Initialization

Database initialization happens automatically when any component starts:

```go
// internal/tui/model.go (line 184)
dataDir := filepath.Join(homeDir, ".claude", "cct")
db, err := database.Initialize(dataDir)

// internal/server/server.go (line 238)
dataDir := filepath.Join(s.claudeDir, "cct")
db, err := database.Initialize(dataDir)
```

The `Initialize()` function:
1. Creates `~/.claude/cct/` directory if it doesn't exist
2. Opens/creates `cct.db` with SQLite driver
3. Sets secure file permissions (0600)
4. Enables WAL mode and performance pragmas
5. Executes embedded schema (idempotent with `IF NOT EXISTS`)
6. Runs migration system to upgrade existing databases
7. Returns singleton Database instance

### Migration System

CCT uses an **inline migration system** for schema evolution:

**Current Approach** (`internal/database/database.go:230-413`):
- Migrations run automatically on startup via `runMigrations()`
- Each migration checks if changes are needed before applying
- Uses `pragma_table_info()` to detect missing columns
- Migrations are idempotent (safe to run multiple times)

**Migration Examples**:
```go
// Migration 1: Add model_name column to providers table
var columnExists bool
query := `SELECT COUNT(*) > 0 FROM pragma_table_info('providers') WHERE name='model_name'`
db.QueryRow(query).Scan(&columnExists)
if !columnExists {
    db.Exec("ALTER TABLE providers ADD COLUMN model_name TEXT")
}

// Migration 7: Add model tracking to all tables
for _, table := range []string{"shell_commands", "claude_commands", "conversations", ...} {
    db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN model_provider TEXT", table))
    db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN model_name TEXT", table))
}
```

**Current Migrations** (7 total):
1. Add `model_name` to `providers` table
2. Add `session_name` to `user_messages` table
3. Add `session_name` to `shell_commands` table
4. Add `session_name` to `claude_commands` table
5. Create `notifications` table and indexes
6. Add `command_details` column to `notifications` table
7. Add `model_provider` and `model_name` to all tracking tables

### Repository Pattern

Data access uses the Repository pattern for clean separation of concerns:

```go
// Get repository instance
repo := database.NewRepository(db)

// Record a shell command
cmd := &database.ShellCommand{
    ConversationID:   "conv-123",
    Command:          "git status",
    WorkingDirectory: "/path/to/project",
    GitBranch:        "main",
    ExitCode:         &exitCode,
    ExecutedAt:       time.Now(),
}
repo.RecordShellCommand(cmd)

// Query commands with filters
query := &database.CommandHistoryQuery{
    ConversationID: "conv-123",
    Limit:          50,
    StartDate:      &startTime,
}
commands, err := repo.GetShellCommands(query)

// Provider management
provider := &database.ProviderConfig{
    ProviderID: "anthropic",
    APIKey:     "sk-ant-...",
    ModelName:  "claude-sonnet-4-20250514",
}
repo.SaveProvider(provider) // Auto-sets as current, unsets others
```

### Database Performance

**Optimizations**:
- **WAL Mode**: Allows concurrent readers with single writer
- **Cache Size**: 64MB cache (`PRAGMA cache_size = -64000`)
- **Synchronous Mode**: NORMAL for balance of safety and speed
- **Memory Temp Store**: Faster temporary table operations
- **Strategic Indexes**: 23+ indexes covering common query patterns

**Connection Pooling**:
- Single connection via singleton pattern
- Thread-safe with `sync.RWMutex` locking
- Shared across TUI, server, and agent handler

### Database Operations

#### Health Check
```go
db := database.GetInstance()
err := db.HealthCheck() // Verifies connectivity
```

#### Statistics
```go
stats, err := db.Stats()
// Returns: {
//   "shell_commands_count": 1234,
//   "claude_commands_count": 5678,
//   "conversations_count": 42,
//   "db_size_bytes": 2097152
// }
```

#### Vacuum (Space Reclamation)
```go
db.Vacuum() // Rebuilds database, reclaims unused space
```

### Database File Structure

```text
~/.claude/cct/
├── cct.db           # Main database file
├── cct.db-wal       # Write-Ahead Log (WAL mode)
└── cct.db-shm       # Shared memory file (WAL mode)
```

**File Permissions**:
- All files: 0600 (user read/write only)
- Set automatically on creation
- Protects sensitive data (API keys, command history)

### Hook Integration

The hook system automatically logs data to the database:

**user-prompt-logger.sh**: Records user messages
```bash
# Logs to user_messages table
curl -X POST https://localhost:3333/api/user-messages \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"message":"...", "conversation_id":"..."}'
```

**tool-logger.sh**: Records tool invocations
```bash
# Logs to claude_commands table
curl -X POST https://localhost:3333/api/claude-commands \
  -d '{"tool_name":"Read", "parameters":"{...}"}'
```

**notification-logger.sh**: Records permission requests
```bash
# Logs to notifications table
curl -X POST https://localhost:3333/api/notifications \
  -d '{"type":"permission_request", "tool_name":"Bash"}'
```

### Database Testing

```go
// internal/database/database_test.go

func TestDatabaseInitialization(t *testing.T) {
    tempDir := t.TempDir()
    db, err := database.Initialize(tempDir)
    assert.NoError(t, err)
    assert.NotNil(t, db)

    // Verify schema
    stats, _ := db.Stats()
    assert.Greater(t, stats["shell_commands_count"], 0)
}
```

### Troubleshooting Database Issues

**Database locked errors:**
```bash
# Check for stale lock
lsof ~/.claude/cct/cct.db

# Force checkpoint WAL
sqlite3 ~/.claude/cct/cct.db "PRAGMA wal_checkpoint(TRUNCATE);"
```

**Inspect database manually:**
```bash
sqlite3 ~/.claude/cct/cct.db

# List tables
.tables

# Check schema
.schema shell_commands

# Query data
SELECT * FROM conversations LIMIT 10;
```

**Reset database (nuclear option):**
```bash
# Backup first!
cp ~/.claude/cct/cct.db ~/.claude/cct/cct.db.backup

# Remove database
rm ~/.claude/cct/cct.db*

# Restart CCT - will recreate fresh database
cct
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

The analytics dashboard is a Nuxt 4 SPA frontend with a Go Fiber backend.

```bash
# Launch analytics server (backend)
./cct --analytics
# or
make run-analytics
# or
just analytics

# Access dashboard (HTTPS by default)
open https://localhost:3333

# API endpoints (use -k to accept self-signed cert)
curl -k https://localhost:3333/api/data
curl -k https://localhost:3333/api/conversations
curl -k https://localhost:3333/api/processes
curl -k https://localhost:3333/api/stats
```

### Unified Server (Analytics + Agents)

The unified server combines analytics dashboard and Claude agent functionality in a single Go-based Fiber server on port 3333.

#### Features
- **Analytics Dashboard**: Real-time conversation monitoring with WebSocket support
- **Agent Conversations**: WebSocket-based real-time Claude agent conversations using claude-agent-sdk-go
- **API Key Authentication**: Unified authentication for all endpoints
- **TLS/HTTPS**: Automatic self-signed certificate generation
- **Session Management**: Multiple concurrent agent conversations
- **Tool Support**: Full agent tool support (Read, Write, Edit, Bash, etc.)

#### Quick Start

```bash
# Start unified server (includes analytics + agents)
./cct --analytics

# Or in TUI, toggle "Server Status" (press 'A')
./cct
```

#### Unified Server Endpoints

**Port**: 3333 (HTTPS by default)

**Endpoints**:
- Analytics Dashboard: `https://localhost:3333/`
- Analytics WebSocket: `wss://localhost:3333/ws`
- Agent WebSocket: `wss://localhost:3333/agent/ws`
- API: `https://localhost:3333/api/*`

#### Agent Functionality

Agent conversations are now integrated into the unified server.

**WebSocket Connection**:
```javascript
const ws = new WebSocket('wss://localhost:3333/agent/ws?token=<api-key>')
```

The API key is stored in `~/.claude/analytics/.secret`.

**Message Types**:
- `create_session`: Create a new agent session
- `send_prompt`: Send a prompt to the agent
- `end_session`: End an agent session
- `list_sessions`: List all active sessions
- `kill_all_agents`: Kill all running agents
- `permission_response`: Respond to permission requests

**Frontend**: Access via Analytics Dashboard → "Live Agents" tab

#### Configuration

The unified server is configured via `~/.claude/analytics/config.json`:

```json
{
  "server": {
    "port": 3333,
    "host": "127.0.0.1"
  },
  "tls": {
    "enabled": true
  },
  "auth": {
    "enabled": true
  },
  "agent": {
    "model": "claude-sonnet-4-5-20250929",
    "max_concurrent_sessions": 10
  }
}
```

**Environment Variables**:
- `ANTHROPIC_API_KEY`: Required for agent functionality
- `CLAUDE_API_KEY`: Alternative to ANTHROPIC_API_KEY

#### Troubleshooting

**Port 3333 already in use:**
```bash
# Find process using port 3333
lsof -i :3333

# Kill the process
kill -9 <PID>
```

**Agent functionality not working:**
```bash
# Check if ANTHROPIC_API_KEY is set
echo $ANTHROPIC_API_KEY

# Set the API key
export ANTHROPIC_API_KEY=your-api-key-here
```

**WebSocket connection fails:**
```bash
# Check server is running
# In TUI, verify "Server: ON (Analytics + Agents)"

# Check firewall settings
# Verify CORS configuration in config.json

# For self-signed certificates, use -k flag with curl
curl -k https://localhost:3333/api/health
echo $ANTHROPIC_API_KEY
```

### Frontend Development

```bash
# Navigate to frontend directory
cd internal/server/frontend

# Install dependencies
npm install

# Run Nuxt dev server (development)
npm run dev
# Runs on http://localhost:3001 by default

# Build for production
npm run build

# Generate static files
npm run generate
```

**IMPORTANT**: Nuxt 4 requires all application code (pages, composables, etc.) to be inside the `app/` directory:

```text
internal/server/frontend/
├── app/                    # Main Nuxt app directory
│   ├── app.vue            # Root component
│   ├── pages/             # Vue pages (index.vue)
│   └── composables/       # Vue composables (useWebSocket.ts)
├── components/            # Vue components (outside app/)
├── types/                 # TypeScript types
├── nuxt.config.ts         # Nuxt configuration
└── package.json           # Frontend dependencies
```

The dev server proxies API calls to the Go backend on port 3333.

### Security Features

The analytics server includes comprehensive security features enabled by default:

#### TLS/HTTPS Encryption

**Automatic Configuration:**
- Self-signed TLS certificates are automatically generated on first run
- Certificates stored in `~/.claude/analytics/certs/`
- Valid for 1 year with automatic expiration warnings
- Server runs on HTTPS by default

**Certificate Management:**
```bash
# Certificates are auto-generated at:
~/.claude/analytics/certs/server.crt
~/.claude/analytics/certs/server.key

# The server will warn when certificates expire in < 30 days
# To regenerate: delete the cert files and restart the server
rm ~/.claude/analytics/certs/server.*
./cct --analytics
```

**Configuration File:**
```json
{
  "tls": {
    "enabled": true,
    "cert_path": "",  // Auto-detected if empty
    "key_path": ""    // Auto-detected if empty
  }
}
```

#### API Key Authentication

**Automatic Setup:**
- API key automatically generated on first run
- Stored in `~/.claude/analytics/.secret`
- Required for all POST/PUT/DELETE/PATCH requests
- GET requests allowed without authentication (for browser access)

**Hook Integration:**
- All hooks automatically read API key from `.secret` file
- Hooks send `Authorization: Bearer <api-key>` header
- Supports self-signed certificates with `-k` flag (curl) or `--no-check-certificate` (wget)

**Viewing Your API Key:**
```bash
# API key location
cat ~/.claude/analytics/.secret

# Hooks read from this file automatically
# No manual configuration needed
```

**Manual API Requests:**
```bash
# With authentication
API_KEY=$(cat ~/.claude/analytics/.secret)
curl -X POST https://localhost:3333/api/prompts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -k \
  -d '{"session_id":"123","prompt":"test"}'

# GET requests work without auth
curl https://localhost:3333/api/data -k
```

#### Security Configuration

**Config File Location:**
```bash
~/.claude/analytics/config.json
```

**Default Configuration:**
```json
{
  "tls": {
    "enabled": true
  },
  "auth": {
    "enabled": true,
    "api_key_path": "~/.claude/analytics/.secret"
  },
  "server": {
    "port": 3333,
    "host": "127.0.0.1",  // Localhost-only by default
    "quiet": false
  },
  "cors": {
    "allowed_origins": [
      "http://localhost:3333",
      "https://localhost:3333",
      "http://127.0.0.1:3333",
      "https://127.0.0.1:3333"
    ]
  },
  "agent": {
    "model": "claude-sonnet-4-5-20250929",
    "max_concurrent_sessions": 10
  }
}
```

**Disabling Security (Development Only):**
```json
{
  "tls": {
    "enabled": false  // Use HTTP instead of HTTPS
  },
  "auth": {
    "enabled": false  // Disable API key requirement
  },
  "server": {
    "host": "0.0.0.0"  // Bind to all interfaces (NOT RECOMMENDED)
  }
}
```

**Security Files:**
```text
~/.claude/analytics/
├── config.json          # Server configuration
├── .secret              # API key (keep private!)
└── certs/
    ├── server.crt       # TLS certificate
    └── server.key       # TLS private key
```

#### Hook Security

**Hook Environment Variables:**
- `CCT_ANALYTICS_URL`: Override analytics endpoint (default: `https://localhost:3333`)
- `CCT_API_KEY_FILE`: Override API key file path (default: `~/.claude/analytics/.secret`)

**Example Custom Configuration:**
```bash
# In your shell profile or .env
export CCT_ANALYTICS_URL="https://analytics.mycompany.com:8443"
export CCT_API_KEY_FILE="/path/to/custom/.secret"
```

**Hook Security Features:**
- Automatic API key authentication
- Support for self-signed certificates
- Silent failures (never block Claude Code)
- Background execution (non-blocking)

#### Security Best Practices

1. **Keep API Key Secret:**
   - Never commit `.secret` file to version control
   - Never share in logs or public places
   - Regenerate if compromised:
     ```bash
     rm ~/.claude/analytics/.secret
     ./cct --analytics  # Will generate new key
     ```

2. **Certificate Management:**
   - Self-signed certs are secure for localhost
   - For remote access, use proper CA-signed certificates
   - Monitor expiration warnings

3. **Network Security:**
   - Server binds to `127.0.0.1` by default (localhost-only)
   - For remote access, use SSH tunneling:
     ```bash
     ssh -L 3333:localhost:3333 user@remote-host
     ```
   - Or configure proper TLS certificates and update CORS origins

4. **Access Control:**
   - Browser access allowed via GET (read-only)
   - Hooks require API key (write access)
   - Consider IP-based restrictions for production

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
   - Test with `wscat -c wss://localhost:3333/ws --no-check`
   - For HTTP mode (if TLS disabled): `wscat -c ws://localhost:3333/ws`

3. **Component download fails**:
   - Check internet connection
   - Verify GitHub API rate limits
   - Check component name spelling

## Git Workflow

### Branch Protection Policy

**CRITICAL**: Never commit directly to the `main` branch unless it's during a release process!

**Rules**:
- All development work MUST be done on feature branches
- Feature branches should follow naming conventions:
  - `feature/` - New features (e.g., `feature/agent-session-manager`)
  - `fix/` - Bug fixes (e.g., `fix/websocket-connection-error`)
  - `docs/` - Documentation updates (e.g., `docs/update-readme`)
  - `refactor/` - Code refactoring (e.g., `refactor/analytics-module`)
  - `test/` - Test additions/updates (e.g., `test/add-agent-tests`)
  - `chore/` - Maintenance tasks (e.g., `chore/update-dependencies`)

**Workflow**:
1. Always create a feature branch before making changes
2. Make commits on the feature branch
3. Create a pull request to merge into `main`
4. Only during release processes (version bumps, release tags) should commits be made to `main`

**Exception**: Release commits are the ONLY commits allowed directly on `main`:
```bash
# ONLY during release process
git checkout main
git commit -m "chore: release v1.0.0"
git tag v1.0.0
git push origin main --tags
```

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
# ALWAYS create a feature branch first
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
