# Claude Agent WebSocket Server

A WebSocket server for managing Claude agent conversations with authentication and session management.

## Features

- = Authentication using the analytics API key (`~/.claude/analytics/.secret`)
- <¯ Session management for multiple concurrent agent conversations
- = Real-time streaming responses from Claude agents
- =à Tool configuration (Read, Write, Edit, Bash, etc.)
- =Ê Session tracking and cleanup

## Installation

```bash
# Install dependencies
uv sync
```

## Configuration

The server uses the same API key as the analytics server:
- Location: `~/.claude/analytics/.secret`
- Format: Plain text API key

Environment variables (optional):
```bash
AGENT_SERVER_HOST=127.0.0.1
AGENT_SERVER_PORT=8001
AGENT_SERVER_LOG_LEVEL=INFO
AGENT_SERVER_AUTH_ENABLED=true
AGENT_SERVER_MAX_CONCURRENT_SESSIONS=10
```

## Running the Server

```bash
# Run with uv
uv run python main.py

# Or activate venv and run directly
source .venv/bin/activate
python main.py
```

The server will start on `http://localhost:8001` by default.

## WebSocket API

### Authentication

Connect to `ws://localhost:8001/ws` with one of:
1. Query parameter: `ws://localhost:8001/ws?token=<api_key>`
2. First message: `{"type": "auth", "token": "<api_key>"}`

### Message Types

#### Create Session
```json
{
  "type": "create_session",
  "session_id": "optional-uuid",
  "options": {
    "system_prompt": "optional",
    "tools": ["Read", "Write", "Bash"],
    "working_directory": "/path/to/dir"
  }
}
```

#### Send Prompt
```json
{
  "type": "send_prompt",
  "session_id": "uuid",
  "prompt": "Your message to Claude"
}
```

#### End Session
```json
{
  "type": "end_session",
  "session_id": "uuid"
}
```

#### List Sessions
```json
{
  "type": "list_sessions"
}
```

### Response Types

- `auth_success`: Authentication successful
- `session_created`: Session created with details
- `agent_message`: Streaming content from agent
- `agent_thinking`: Agent is processing
- `agent_tool_use`: Agent is using a tool
- `error`: Error occurred

## Development

```bash
# Install dev dependencies
uv add --dev pytest pytest-asyncio black ruff

# Run tests
uv run pytest

# Format code
uv run black src/
uv run ruff src/
```

## Architecture

- `src/main.py`: FastAPI application and WebSocket endpoint
- `src/auth.py`: Authentication middleware
- `src/agent_manager.py`: Claude Agent SDK integration
- `src/session.py`: Session management
- `src/models.py`: Pydantic models for messages
- `src/config.py`: Configuration management

## Security

- API key authentication required for all connections
- Session timeout after 1 hour of inactivity
- Rate limiting and concurrent session limits
- Automatic cleanup of abandoned sessions