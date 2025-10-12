-- SQLite schema for command history tracking

-- Table for shell commands executed via Bash tool
CREATE TABLE IF NOT EXISTS shell_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL,
    command TEXT NOT NULL,
    description TEXT,
    working_directory TEXT,
    git_branch TEXT,
    exit_code INTEGER,
    stdout TEXT,
    stderr TEXT,
    duration_ms INTEGER,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for Claude Code commands (tool invocations)
CREATE TABLE IF NOT EXISTS claude_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    parameters TEXT, -- JSON string
    result TEXT, -- JSON string
    working_directory TEXT,
    git_branch TEXT,
    success BOOLEAN DEFAULT 1,
    error_message TEXT,
    duration_ms INTEGER,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for conversation metadata
CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    project_path TEXT,
    started_at TIMESTAMP,
    last_activity_at TIMESTAMP,
    total_commands INTEGER DEFAULT 0,
    total_shell_commands INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for command statistics (aggregated data)
CREATE TABLE IF NOT EXISTS command_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command_type TEXT NOT NULL, -- 'shell' or 'claude'
    command_name TEXT NOT NULL,
    execution_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    avg_duration_ms INTEGER DEFAULT 0,
    last_executed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(command_type, command_name)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_shell_commands_conversation
    ON shell_commands(conversation_id, executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_shell_commands_executed_at
    ON shell_commands(executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_claude_commands_conversation
    ON claude_commands(conversation_id, executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_claude_commands_executed_at
    ON claude_commands(executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_claude_commands_tool
    ON claude_commands(tool_name, executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_conversations_status
    ON conversations(status, last_activity_at DESC);

CREATE INDEX IF NOT EXISTS idx_command_stats_type_name
    ON command_stats(command_type, command_name);
