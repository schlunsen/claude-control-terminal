# Plan: Merge Agent Server into Analytics Server

## Overview
Consolidate the standalone agent server (port 8001) into the analytics server (port 3333) to create a unified server architecture. The frontend "Live Agents" feature will continue to work with minimal changes since it already connects through the analytics server.

---

## Phase 1: Code Reorganization (Move Agent Code)

### 1.1 Move agent package
- **Move**: `internal/agents/` → `internal/server/agents/`
- **Files to keep**:
  - `agent_handler.go` - WebSocket handler and message routing
  - `session_manager.go` - Session management
  - `messages.go` - Message type definitions
  - `config.go` - Agent configuration (simplified)
- **DELETE**: `launcher.go` - No longer needed, server manages lifecycle
- **UPDATE**: Package declarations from `package agents` to `package agents`

### 1.2 Update imports across codebase
Change all imports from:
```go
github.com/schlunsen/claude-control-terminal/internal/agents
```
To:
```go
github.com/schlunsen/claude-control-terminal/internal/server/agents
```

**Files to update**:
- `internal/cmd/root.go`
- `internal/tui/tui.go`
- `internal/tui/model.go`

---

## Phase 2: Server Integration (Backend)

### 2.1 Update `internal/server/server.go`
**Add to Server struct**:
```go
agentHandler *agents.AgentHandler
agentConfig  *agents.Config
```

**In `Setup()` method**:
```go
// Initialize agent configuration
agentConfig := &agents.Config{
    Model:                 config.Agent.Model,
    APIKey:                config.Agent.APIKey,
    MaxConcurrentSessions: config.Agent.MaxConcurrentSessions,
}

// Initialize agent handler
s.agentHandler = agents.NewAgentHandler(agentConfig)
```

**In `setupRoutes()` method** (add after line 208):
```go
// Agent WebSocket endpoint (direct, not proxied)
s.app.Get("/agent/ws", websocket.New(s.agentHandler.HandleWebSocket))
```

**DELETE proxy endpoints** (lines 219-223, 1682-1798):
- Remove `handleAgentProxy` method
- Remove `handleAgentWebSocketProxy` method
- Remove `/api/agent/*` route

**Update `Shutdown()` method**:
```go
// Cleanup agent sessions
if s.agentHandler != nil {
    // End all active agent sessions gracefully
    s.agentHandler.Cleanup()
}
```

### 2.2 Update `internal/server/config.go`
**Add to Config struct**:
```go
Agent struct {
    Model                 string `json:"model"`
    MaxConcurrentSessions int    `json:"max_concurrent_sessions"`
    APIKey                string `json:"-"` // Don't serialize, read from env
} `json:"agent"`
```

**Update default config** in `LoadOrCreateConfig()`:
```go
Agent: struct {
    Model                 string `json:"model"`
    MaxConcurrentSessions int    `json:"max_concurrent_sessions"`
    APIKey                string `json:"-"`
}{
    Model:                 "claude-3-5-sonnet-latest",
    MaxConcurrentSessions: 10,
    APIKey:                getAgentAPIKey(), // Helper function
},
```

**Add helper function**:
```go
func getAgentAPIKey() string {
    // Try ANTHROPIC_API_KEY first, then CLAUDE_API_KEY
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        apiKey = os.Getenv("CLAUDE_API_KEY")
    }
    return apiKey
}
```

### 2.3 Simplify `internal/server/agents/config.go`
**Remove** server-related fields (keep only agent-specific):
```go
type Config struct {
    Model                 string
    APIKey                string
    MaxConcurrentSessions int
}
```

**DELETE** all helper functions:
- `DefaultConfig()` - No longer needed
- `GetServerDir()` - No longer needed
- `getEnvOrDefault()` - No longer needed
- `getEnvIntOrDefault()` - No longer needed
- `getEnvBoolOrDefault()` - No longer needed

### 2.4 Update `internal/server/agents/agent_handler.go`
**Add Cleanup method**:
```go
// Cleanup ends all active sessions gracefully
func (h *AgentHandler) Cleanup() error {
    count := h.sessionManager.EndAllSessions()
    log.Printf("Cleaned up %d active agent sessions", count)
    return nil
}
```

---

## Phase 3: Remove Agent CLI Commands

### 3.1 DELETE file entirely
**DELETE**: `internal/cmd/agents.go` (all 178 lines)

### 3.2 Update `internal/cmd/root.go`

**Remove flag** (line 44):
```go
agents       bool  // DELETE THIS LINE
```

**Remove flag initialization** (line 176):
```go
rootCmd.Flags().BoolVar(&agents, "agents", false, "launch agents dashboard")  // DELETE
```

**Remove function** (lines 150-177):
```go
func LaunchAgentServer() { ... }  // DELETE ENTIRE FUNCTION
```

**Remove agents handling** (lines 267-271):
```go
// Agents dashboard
if agents {
    LaunchAgentServer()
    return
}
// DELETE THIS ENTIRE BLOCK
```

---

## Phase 4: TUI Updates

### 4.1 Update `internal/tui/model.go`

**Remove fields** (lines 95-96):
```go
agentServerEnabled bool   // DELETE
agentServerPID     int    // DELETE
```

**Simplify constructor**:
- Rename `NewModelWithServers` to `NewModelWithServer`
- Remove agent server parameters
- **Signature change**:
  ```go
  // Before:
  func NewModelWithServers(targetDir, claudeDir string, analyticsServer *server.Server, agentServerEnabled bool, agentServerPID int) Model

  // After:
  func NewModelWithServer(targetDir, claudeDir string, analyticsServer *server.Server) Model
  ```

**Remove agent toggle logic** (lines 1760-1772):
```go
// DELETE entire toggleAgentServer function
```

### 4.2 Update `internal/tui/tui.go`

**Remove agent server startup** (lines 47-71):
```go
// Start agent server in background (enabled by default)
var agentServerPID int
var agentServerEnabled bool
...
// DELETE ALL THIS CODE
```

**Remove agent cleanup** (line 79):
```go
agentLauncher.Cleanup()  // DELETE
```

**Update model creation** (line 84):
```go
// Before:
m := NewModelWithServers(targetDir, claudeDir, analyticsServer, agentServerEnabled, agentServerPID)

// After:
m := NewModelWithServer(targetDir, claudeDir, analyticsServer)
```

**Remove agent state syncing** (lines 105-107):
```go
// Sync agent server state from model
agentServerEnabled = model.agentServerEnabled
agentServerPID = model.agentServerPID
// DELETE THESE LINES
```

### 4.3 Update main menu display

In the TUI main menu rendering code, update server status to show:
```
Server Status: Running (Analytics + Agents on port 3333)
```
Instead of separate "Analytics Server" and "Agent Server" entries.

---

## Phase 5: Frontend Updates (Nuxt.js)

### 5.1 Update `app/composables/useAgentWebSocket.ts`

**Update comments** (lines 49-52):
```typescript
// Before:
// Connect through analytics server proxy at /agent/ws
// The analytics server will proxy to the agents server on port 8001

// After:
// Connect to unified server's agent WebSocket endpoint at /agent/ws
// The analytics server directly handles agent functionality (no proxy)
```

**No URL changes needed** - endpoint stays `/agent/ws`

### 5.2 No changes needed for `app/pages/agents.vue`
The "Live Agents" page will continue to work as-is since:
- It connects via `useAgentWebSocket` composable
- Endpoint URL stays the same (`/agent/ws`)
- Message protocol remains unchanged
- All functionality preserved

---

## Phase 6: Documentation Updates

### 6.1 Update `CLAUDE.md`

**Section: "Agent Server"** → Rename to **"Unified Server (Analytics + Agents)"**

Update port documentation:
```markdown
### Unified Server Endpoints

**Port**: 3333 (HTTPS by default)

**Endpoints**:
- Analytics Dashboard: `https://localhost:3333/`
- Analytics WebSocket: `wss://localhost:3333/ws`
- Agent WebSocket: `wss://localhost:3333/agent/ws`
- API: `https://localhost:3333/api/*`

### Starting the Server

```bash
# Start unified server (includes analytics + agents)
./cct --analytics

# Or in TUI, toggle "Server Status"
./cct
```
```

**DELETE** all references to:
- `./cct agents start/stop/status`
- Port 8001
- Separate agent server process
- PID file management for agents

**UPDATE** "Agent Server" section:
```markdown
## Agent Functionality

Agent conversations are now integrated into the unified server.

**WebSocket Connection**:
```javascript
const ws = new WebSocket('wss://localhost:3333/agent/ws?token=<api-key>')
```

**Features**:
- Real-time Claude agent conversations
- Full Claude Agent SDK integration
- Session management
- Tool support (Read, Write, Edit, Bash, etc.)
- Permission handling

**Frontend**: Access via Analytics Dashboard → "Live Agents" tab
```

### 6.2 Update `README.md`
- Update architecture diagram (single server)
- Update installation instructions
- Remove agent server commands
- Update port references (8001 → 3333)

---

## Phase 7: Configuration Migration

### 7.1 Update default `config.json` schema
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

### 7.2 Clean up old agent files
**DELETE** (if they exist):
- `~/.claude/agents_server/.pid`
- `~/.claude/agents_server/server.log`
- Note: Keep `~/.claude/analytics/.secret` for API key

---

## Phase 8: Testing & Validation

### 8.1 Backend Testing
**Test steps**:
1. **Build**: `make build` or `just build`
2. **Run TUI**: `./cct`
   - ✅ TUI should show single "Server Status: Stopped"
   - ✅ Toggle server on - should start on port 3333
   - ✅ TUI should show "Server Status: Running"
3. **Run server directly**: `./cct --analytics`
   - ✅ Should start on port 3333
   - ✅ Should log: "Agent handler initialized"

### 8.2 WebSocket Testing
**Test steps**:
1. **Analytics WebSocket**: Connect to `ws://localhost:3333/ws`
   - ✅ Should receive `{"type":"connected",...}`
2. **Agent WebSocket**: Connect to `ws://localhost:3333/agent/ws`
   - ✅ Should connect successfully
   - ✅ Send `{"type":"ping"}` → Should receive pong

### 8.3 Frontend Testing
**Test steps**:
1. **Open dashboard**: `https://localhost:3333/`
2. **Navigate to "Live Agents"** tab
3. **Check connection status**: Should show "Connected"
4. **Create session**: Click "New Session"
   - ✅ Session should be created
   - ✅ Should appear in sessions list
5. **Send message**: Type and send a prompt
   - ✅ Should receive Claude agent response
   - ✅ Messages should stream in real-time

---

## Benefits

✅ **Single server process** - Simplified architecture, easier deployment
✅ **Single port (3333)** - No port conflicts, easier firewall configuration
✅ **Unified lifecycle** - Start/stop everything with one command
✅ **TUI status works** - Correct detection of server state
✅ **Reduced complexity** - Less code to maintain, fewer moving parts
✅ **Better performance** - No proxy overhead, direct WebSocket handling
✅ **Consistent auth** - Same API key for both analytics and agents

---

## Breaking Changes

⚠️ **WebSocket endpoint**: Agent connections must use port 3333 (was 8001)
⚠️ **CLI commands removed**: `cct agents start/stop/status/logs` no longer exist
⚠️ **Configuration**: Agent config merged into main config file
⚠️ **PID files**: Agent PID file no longer used

---

## Migration Guide for Users

### If you were using:

**CLI**:
```bash
# Before:
cct agents start
cct agents status
cct agents stop

# After:
cct --analytics  # Starts unified server with agents
# Status shown in TUI or check port 3333
# Stop with Ctrl+C or toggle in TUI
```

**WebSocket URL**:
```javascript
// Before:
new WebSocket('ws://localhost:8001/ws')

// After:
new WebSocket('ws://localhost:3333/agent/ws')
```

**Frontend**:
- No changes needed - frontend already connects through port 3333
- "Live Agents" feature continues to work seamlessly

---

## Implementation Order

1. **Phase 1** - Move code (no functional changes)
2. **Phase 2** - Integrate into server
3. **Phase 5** - Update frontend comments
4. **Phase 3** - Remove CLI commands
5. **Phase 4** - Update TUI
6. **Phase 6** - Update docs
7. **Phase 7** - Configuration updates
8. **Phase 8** - Testing with user

This order minimizes breakage and allows incremental testing.

---

## Current Issues Being Fixed

1. **Agent server not starting**: The standalone agent server has lifecycle management issues
2. **TUI status incorrect**: TUI can't properly detect agent server state (separate PID file)
3. **Port complexity**: Two servers on different ports (3333 + 8001)
4. **Proxy overhead**: Frontend connects through analytics server which proxies to agent server

All of these issues will be resolved by the unified architecture.
