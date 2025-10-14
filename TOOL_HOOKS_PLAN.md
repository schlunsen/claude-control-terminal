# Tool Logging Hooks Implementation Plan

## Overview

Add PostToolUse hook to capture all Claude Code tool usage (Bash commands and Claude tool invocations) for analytics tracking.

## Architecture

### Current State
- ✅ UserPromptSubmit hook captures user input
- ✅ Database tables exist for `shell_commands` and `claude_commands`
- ✅ API endpoints ready: `POST /api/commands/shell` and `POST /api/commands/claude`

### New Addition
- Single `tool-logger.sh` hook that captures ALL tool usage
- Routes internally based on `tool_name` field
- Posts to appropriate API endpoint

## Implementation Steps

### 1. Create Tool Logger Hook Script

**File**: `hooks/tool-logger.sh`

**Features**:
- Hook Type: PostToolUse with `"matcher": "*"`
- Parse JSON input from stdin
- Extract fields:
  - `tool_name` - determines routing
  - `session_id` - conversation identifier
  - `cwd` - working directory
  - `parameters` - tool input (JSON)
  - `result` - tool output (JSON)
  - `success` - boolean
  - `duration_ms` - execution time

**Routing Logic**:
```bash
if [[ "$TOOL_NAME" == "Bash" ]]; then
    # Extract: command, description, exit_code, stdout, stderr
    POST to /api/commands/shell
else
    # Extract: tool_name, parameters, result, success, error_message
    POST to /api/commands/claude
fi
```

**Session Naming**:
- Use same hash-based approach as user-prompt-logger
- 10 South Park character names for consistency

**Execution**:
- Run curl/wget in background (non-blocking)
- Silent failures (don't interrupt Claude Code)

### 2. Extend Hook Installer

**File**: `internal/components/hook.go`

**New Methods**:
- `InstallToolLogger()` - Install PostToolUse hook
- `UninstallToolLogger()` - Remove PostToolUse hook
- `CheckToolLoggerInstalled()` - Check installation status

**Updates to Existing Methods**:
- `addHookToSettingsAtPath()` - Support matcher-based hooks for PostToolUse
- Handle different event types (UserPromptSubmit vs PostToolUse)

**Settings Format**:
```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {"type": "command", "command": ".claude/hooks/user-prompt-logger.sh"}
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {"type": "command", "command": ".claude/hooks/tool-logger.sh"}
        ]
      }
    ]
  }
}
```

### 3. Add CLI Commands

**File**: `internal/cmd/root.go`

**New Flags**:
- `--install-tool-hook` - Install tool logger hook only
- `--uninstall-tool-hook` - Remove tool logger hook
- `--install-all-hooks` - Install both user-prompt and tool loggers
- `--uninstall-all-hooks` - Remove all hooks

**Integration**:
- Add to existing hook handling logic
- Show installation status and paths
- Provide analytics server startup instructions

### 4. Testing Plan

**Manual Testing**:
1. Install hook: `cct --install-tool-hook`
2. Start analytics server: `cct --analytics`
3. Start Claude Code session
4. Execute various tools: Bash, Read, Edit, Write
5. Verify data in dashboard at `http://localhost:3333`

**Verify**:
- Shell commands appear in shell_commands table
- Claude tool usage appears in claude_commands table
- Session names match across all three tables
- No blocking or slowdown of Claude Code
- Background curl processes complete successfully

### 5. Documentation Updates

**Files to Update**:
- `README.md` - Add hook installation instructions
- `CLAUDE.md` - Document hook architecture
- Add example settings.local.json

## Expected Outcome

After implementation:
- Complete activity tracking: user prompts + tool usage
- Unified analytics dashboard showing all actions
- Session-based grouping across all activity types
- No performance impact on Claude Code
- Easy installation with single command

## Technical Details

### PostToolUse JSON Schema (from Claude Code)

```json
{
  "tool_name": "Bash|Read|Edit|Write|...",
  "session_id": "uuid",
  "cwd": "/path/to/project",
  "parameters": {...},
  "result": {...},
  "success": true|false,
  "duration_ms": 123,
  "error_message": "..."
}
```

### API Endpoints

**POST /api/commands/shell**
```json
{
  "session_id": "uuid",
  "session_name": "Cartman",
  "command": "ls -la",
  "description": "List directory contents",
  "cwd": "/path",
  "branch": "main",
  "exit_code": 0,
  "stdout": "...",
  "stderr": "",
  "duration_ms": 45
}
```

**POST /api/commands/claude**
```json
{
  "session_id": "uuid",
  "session_name": "Cartman",
  "tool_name": "Read",
  "parameters": "{\"file_path\": \"...\"}",
  "result": "{\"content\": \"...\"}",
  "cwd": "/path",
  "branch": "main",
  "success": true,
  "error_message": "",
  "duration_ms": 12
}
```

## Benefits

1. **Complete Activity Tracking** - Every action Claude takes is logged
2. **Session Continuity** - Same session_name across prompts, shell, and tools
3. **Performance Analysis** - Track command durations and patterns
4. **Debugging Aid** - See exact sequence of events in a session
5. **Non-Invasive** - Hooks run in background, no impact on Claude Code
