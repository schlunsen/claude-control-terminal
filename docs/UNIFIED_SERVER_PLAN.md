# Unified Server Implementation Plan

**Status**: Ready for Implementation
**Date**: 2025-10-18
**Goal**: Merge standalone agent server (port 8001) into analytics server (port 3333)

---

## Current Architecture

```
┌─────────────────────────────────────────┐
│  Analytics Server (port 3333)           │
│  - Dashboard                            │
│  - WebSocket (/ws)                      │
│  - API endpoints                        │
│  - Proxy to agent server (/agent/ws)   │ ──┐
└─────────────────────────────────────────┘   │
                                              │ HTTP Proxy
┌─────────────────────────────────────────┐   │
│  Agent Server (port 8001)               │ ◄─┘
│  - WebSocket (/ws)                      │
│  - Claude Agent SDK                     │
│  - Session management                   │
└─────────────────────────────────────────┘
```

## Target Architecture

```
┌─────────────────────────────────────────┐
│  Unified Server (port 3333)             │
│  - Dashboard                            │
│  - Analytics WebSocket (/ws)            │
│  - Agent WebSocket (/agent/ws) ◄─────── │ Direct handler
│  - API endpoints                        │
│  - Claude Agent SDK (integrated)        │
│  - Session management (integrated)      │
└─────────────────────────────────────────┘
```

---

## Implementation Steps

### Phase 1: Backend Integration

#### 1.1 Update Server Struct (`internal/server/server.go`)

**Add fields** (after line 45):
```go
agentHandler *agents.AgentHandler
agentConfig  *agents.AgentConfig
```

**In Setup() method** (after line 101):
```go
// Initialize agent configuration
agentAPIKey := os.Getenv("ANTHROPIC_API_KEY")
if agentAPIKey == "" {
    agentAPIKey = os.Getenv("CLAUDE_API_KEY")
}

agentConfig := &agents.Config{
    Model:                 config.Agent.Model,
    APIKey:                agentAPIKey,
    MaxConcurrentSessions: config.Agent.MaxConcurrentSessions,
}
s.agentConfig = agentConfig

// Initialize agent handler
s.agentHandler = agents.NewAgentHandler(agentConfig)

if !s.quiet {
    fmt.Printf("🤖 Agent handler initialized (model: %s, max sessions: %d)\n",
        agentConfig.Model, agentConfig.MaxConcurrentSessions)
}
```

#### 1.2 Update Routes (`internal/server/server.go`)

**In setupRoutes() method** (replace lines 219-223):

**DELETE:**
```go
// Agent endpoints (serve agents from project directory)
api.Get("/agents", s.handleListAgents)
api.Get("/agents/:name", s.handleGetAgentDetail)

// Agent server proxy endpoints (if agent server is running)
api.All("/agent/*", s.handleAgentProxy)

// Agent WebSocket proxy
s.app.Get("/agent/ws", websocket.New(s.handleAgentWebSocketProxy()))
```

**REPLACE WITH:**
```go
// Agent endpoints (serve agents from project directory)
api.Get("/agents", s.handleListAgents)
api.Get("/agents/:name", s.handleGetAgentDetail)

// Agent WebSocket endpoint (direct, not proxied)
s.app.Get("/agent/ws", s.agentHandler.HandleWebSocket)
```

**DELETE methods** (lines 1682-1798):
- `handleAgentProxy()`
- `handleAgentWebSocketProxy()`

#### 1.3 Update Shutdown (`internal/server/server.go`)

**In Shutdown() method** (before line 567):
```go
// Cleanup agent sessions
if s.agentHandler != nil {
    if err := s.agentHandler.Cleanup(); err != nil && !s.quiet {
        fmt.Printf("⚠️  Error cleaning up agent sessions: %v\n", err)
    }
}
```

#### 1.4 Add Agent Cleanup Method (`internal/server/agents/agent_handler.go`)

**Add method** (after HandleWebSocket):
```go
// Cleanup ends all active sessions gracefully
func (h *AgentHandler) Cleanup() error {
    count := h.SessionManager.EndAllSessions()
    log.Printf("Cleaned up %d active agent sessions", count)
    return nil
}
```

---

### Phase 2: Configuration Updates

#### 2.1 Update Config Struct (`internal/server/config.go`)

**Add to Config struct** (after line 44):
```go
Agent AgentSettings `json:"agent"`
```

**Add new struct** (after CORSSettings):
```go
// AgentSettings holds agent configuration
type AgentSettings struct {
    Model                 string `json:"model"`
    MaxConcurrentSessions int    `json:"max_concurrent_sessions"`
}
```

#### 2.2 Update Default Config (`internal/server/config.go`)

**In getDefaultConfig()** (after line 177):
```go
Agent: AgentSettings{
    Model:                 "claude-3-5-sonnet-latest",
    MaxConcurrentSessions: 10,
},
```

**Result config.json**:
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
    "host": "127.0.0.1",
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
    "model": "claude-3-5-sonnet-latest",
    "max_concurrent_sessions": 10
  }
}
```

---

### Phase 3: TUI Cleanup

#### 3.1 Update Model Struct (`internal/tui/model.go`)

**DELETE fields** (lines 95-97):
```go
// Agent server state
agentServerEnabled bool   // Whether agent server is running
agentServerPID     int    // PID of agent server process
```

**DELETE constructor** (rename `NewModelWithServers` to existing):
```go
// DELETE this function entirely - use NewModelWithServer instead
func NewModelWithServers(...) Model { ... }
```

**UPDATE NewModelWithServer** (lines 145-241):
Remove all agent server parameters and initialization.

#### 3.2 Update Model Handlers (`internal/tui/model.go`)

**DELETE message handler** (lines 362-366):
```go
case toggleAgentServerMsg:
    // Handle agent server toggle result
    m.agentServerEnabled = msg.enabled
    m.agentServerPID = msg.pid
    return m, nil
```

**DELETE toggle handler** (lines 497-498):
```go
case "s", "S":
    // Toggle agent server on/off
    return m, toggleAgentServerCmd(m.agentServerEnabled, m.agentServerPID)
```

**DELETE command** (lines 1729-1773):
```go
type toggleAgentServerMsg struct { ... }
func toggleAgentServerCmd(...) tea.Cmd { ... }
```

#### 3.3 Update Main Menu Display (`internal/tui/model.go`)

**UPDATE viewMainScreen()** (lines 935-943):

**DELETE:**
```go
// Agent server status
agentServerStatus := "OFF"
agentServerStyle := StatusErrorStyle
if m.agentServerEnabled {
    agentServerStatus = fmt.Sprintf("ON (PID: %d)", m.agentServerPID)
    agentServerStyle = StatusSuccessStyle
}
b.WriteString(SubtitleStyle.Render("Agent Server: ") + agentServerStyle.Render(agentServerStatus))
b.WriteString(SubtitleStyle.Render(" (http://localhost:8001)") + "\n")
```

**REPLACE WITH:**
```go
// Server status (unified)
serverStatus := "OFF"
serverStyle := StatusErrorStyle
serverDesc := ""
if m.analyticsEnabled {
    serverStatus = "ON"
    serverStyle = StatusSuccessStyle
    serverDesc = " (Analytics + Agents)"
}
b.WriteString(SubtitleStyle.Render("Server: ") + serverStyle.Render(serverStatus) + SubtitleStyle.Render(serverDesc))
b.WriteString(SubtitleStyle.Render(" (https://localhost:3333)") + "\n")
```

**UPDATE help text** (lines 946-950):
```go
// Remove "S: Agent Server" from help text
b.WriteString(HelpStyle.Render("↑/↓: Navigate • Enter: Select • T: Theme • A: Analytics • H: Logging Hooks • Q/Esc: Quit"))
```

#### 3.4 Update TUI Startup (`internal/tui/tui.go`)

**DELETE agent server startup** (lines 47-71):
```go
// Start agent server in background (enabled by default)
var agentServerPID int
var agentServerEnabled bool

agentConfig := agentspkg.DefaultConfig()
agentLauncher := agentspkg.NewLauncher(agentConfig, true, true) // quiet mode, background mode

// Try to start agent server
if err := agentLauncher.Start(); err != nil {
    pterm.Warning.Printf("Failed to start agent server: %v\n", err)
    agentServerEnabled = false
} else {
    // Wait briefly for server to start
    time.Sleep(500 * time.Millisecond)

    // Get the PID
    running, pid, _ := agentLauncher.IsRunning()
    if running {
        agentServerEnabled = true
        agentServerPID = pid
    }
}

defer func() {
    // Cleanup agent server on exit
    agentLauncher.Cleanup()
}()
```

**UPDATE model creation** (line 84):
```go
// Before:
m := NewModelWithServers(targetDir, claudeDir, analyticsServer, agentServerEnabled, agentServerPID)

// After:
m := NewModelWithServer(targetDir, claudeDir, analyticsServer)
```

**DELETE state syncing** (lines 105-107):
```go
// Sync agent server state from model
agentServerEnabled = model.agentServerEnabled
agentServerPID = model.agentServerPID
```

---

### Phase 4: Frontend Updates

#### 4.1 Update WebSocket Composable (`internal/server/frontend/app/composables/useAgentWebSocket.ts`)

**UPDATE comments** (lines 49-50):

**Before:**
```typescript
// Connect through analytics server proxy at /agent/ws
// The analytics server will proxy to the agents server on port 8001
```

**After:**
```typescript
// Connect to unified server's agent WebSocket endpoint at /agent/ws
// The analytics server directly handles agent functionality (no proxy)
```

**No other changes needed** - the endpoint URL stays the same (`/agent/ws`)

---

### Phase 5: Documentation Updates

#### 5.1 Update CLAUDE.md

**Section: "Unified Server"** - Update to reflect direct integration:

```markdown
### Unified Server (Analytics + Agents)

The unified server combines analytics dashboard and Claude agent functionality in a single Go-based Fiber server on port 3333.

#### Features
- **Analytics Dashboard**: Real-time conversation monitoring with WebSocket support
- **Agent Conversations**: WebSocket-based real-time Claude agent conversations using claude-agent-sdk-go (integrated)
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
    "model": "claude-3-5-sonnet-latest",
    "max_concurrent_sessions": 10
  }
}
```

**Environment Variables**:
- `ANTHROPIC_API_KEY`: Required for agent functionality
- `CLAUDE_API_KEY`: Alternative to ANTHROPIC_API_KEY
```

**DELETE all references to:**
- Port 8001
- Separate agent server process
- Agent server commands
- Proxy functionality

#### 5.2 Update README.md

- Update architecture diagram to show single server
- Remove agent server CLI commands
- Update all port references (8001 → 3333)
- Update examples

---

## Testing Checklist

### Backend Tests
- [ ] Server starts successfully on port 3333
- [ ] Agent handler initializes without errors
- [ ] Analytics endpoints work (`/api/data`, `/api/conversations`, etc.)
- [ ] Analytics WebSocket connects (`/ws`)
- [ ] Agent WebSocket connects (`/agent/ws`)
- [ ] Server shutdown cleanly ends agent sessions

### Frontend Tests
- [ ] Analytics Dashboard loads
- [ ] "Live Agents" tab loads
- [ ] WebSocket connection status shows "Connected"
- [ ] Can create new agent session
- [ ] Can send messages to agent
- [ ] Agent responses stream correctly
- [ ] Tool use displays properly
- [ ] Permission requests work

### TUI Tests
- [ ] TUI starts without errors
- [ ] Server status shows correctly (OFF initially)
- [ ] Toggle server (A) starts unified server
- [ ] Server status updates to "ON (Analytics + Agents)"
- [ ] Toggle server (A) stops unified server
- [ ] No separate agent server toggle

### Integration Tests
- [ ] API key authentication works for both analytics and agents
- [ ] TLS works for both analytics and agents
- [ ] Multiple concurrent agent sessions work
- [ ] Agent sessions clean up on server shutdown
- [ ] No orphaned processes on port 8001

---

## Benefits

✅ **Simplified Architecture**
- Single server process
- Single port (3333)
- Single configuration file

✅ **Better Performance**
- No proxy overhead
- Direct WebSocket handling
- Reduced latency

✅ **Easier Deployment**
- One process to manage
- No port conflicts
- Unified lifecycle

✅ **Better UX**
- TUI shows correct server state
- Single toggle for all features
- Consistent authentication

---

## Breaking Changes

⚠️ **WebSocket Endpoint**: Agent connections must use port 3333 (was 8001)
⚠️ **CLI Commands**: Agent server commands no longer exist
⚠️ **Configuration**: Agent config merged into main config file
⚠️ **PID Files**: Agent PID file no longer used

---

## Migration for Existing Users

### If you were using agent server separately:

**CLI Usage:**
```bash
# Before:
cct agents start
cct agents status
cct agents stop

# After:
cct --analytics  # Starts unified server with agents
# Check TUI for status
```

**Direct WebSocket:**
```javascript
// Before:
new WebSocket('ws://localhost:8001/ws')

// After:
new WebSocket('ws://localhost:3333/agent/ws')
```

**Frontend:**
- No changes needed
- Already connects through port 3333

---

## Files Modified

### Created
- `docs/UNIFIED_SERVER_PLAN.md` (this file)

### Modified
- `internal/server/server.go` - Integrate agent handler
- `internal/server/config.go` - Add agent configuration
- `internal/server/agents/agent_handler.go` - Add Cleanup method
- `internal/tui/model.go` - Remove agent server management
- `internal/tui/tui.go` - Remove agent server startup
- `internal/server/frontend/app/composables/useAgentWebSocket.ts` - Update comments
- `CLAUDE.md` - Update documentation
- `README.md` - Update documentation

### Deleted (functions/code blocks)
- `internal/server/server.go::handleAgentProxy()`
- `internal/server/server.go::handleAgentWebSocketProxy()`
- `internal/tui/model.go::toggleAgentServerCmd()`
- `internal/tui/model.go::agentServerEnabled` field
- `internal/tui/model.go::agentServerPID` field
- Agent server startup code in `internal/tui/tui.go`

---

## Implementation Order

1. ✅ Write this plan document
2. Backend integration (server.go, config.go)
3. Agent handler cleanup method
4. TUI cleanup (model.go, tui.go)
5. Frontend comment updates
6. Documentation updates
7. Testing
8. Git commit and tag release

---

## Notes

- The agent handler code already exists in `internal/server/agents/`
- The frontend already connects through port 3333
- This is mostly removing proxy code and wiring up the direct handler
- No protocol changes needed - just endpoint consolidation
- Existing `claude-agent-sdk-go` integration is preserved
