-- SQLite schema for command history tracking

-- Table for shell commands executed via Bash tool
CREATE TABLE IF NOT EXISTS shell_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL,
    session_name TEXT,
    command TEXT NOT NULL,
    description TEXT,
    working_directory TEXT,
    git_branch TEXT,
    model_provider TEXT,
    model_name TEXT,
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
    session_name TEXT,
    tool_name TEXT NOT NULL,
    parameters TEXT, -- JSON string
    result TEXT, -- JSON string
    working_directory TEXT,
    git_branch TEXT,
    model_provider TEXT,
    model_name TEXT,
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
    model_provider TEXT,
    model_name TEXT,
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

-- Table for user messages (intercepted input)
CREATE TABLE IF NOT EXISTS user_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT,
    session_name TEXT,
    message TEXT NOT NULL,
    working_directory TEXT,
    git_branch TEXT,
    model_provider TEXT,
    model_name TEXT,
    message_length INTEGER,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for AI provider configurations
CREATE TABLE IF NOT EXISTS providers (
    provider_id TEXT PRIMARY KEY,
    api_key TEXT NOT NULL,
    custom_url TEXT,
    model_name TEXT,
    is_current BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for notifications (permission requests and idle alerts)
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id TEXT NOT NULL,
    session_name TEXT,
    notification_type TEXT NOT NULL, -- 'permission_request', 'idle_alert', 'other'
    message TEXT NOT NULL,
    tool_name TEXT, -- extracted from permission requests
    command_details TEXT, -- actual command/parameters that required permission
    working_directory TEXT,
    git_branch TEXT,
    model_provider TEXT,
    model_name TEXT,
    notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for user settings
CREATE TABLE IF NOT EXISTS user_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    value_type TEXT DEFAULT 'string', -- 'string', 'boolean', 'number', 'json'
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT OR IGNORE INTO user_settings (key, value, value_type, description) VALUES
('diff_display_location', 'chat', 'string', 'Where to display file diffs: "chat" or "options"');

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

CREATE INDEX IF NOT EXISTS idx_user_messages_conversation
    ON user_messages(conversation_id, submitted_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_messages_submitted_at
    ON user_messages(submitted_at DESC);

CREATE INDEX IF NOT EXISTS idx_providers_is_current
    ON providers(is_current) WHERE is_current = 1;

CREATE INDEX IF NOT EXISTS idx_notifications_conversation
    ON notifications(conversation_id, notified_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_notified_at
    ON notifications(notified_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_type
    ON notifications(notification_type, notified_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_tool
    ON notifications(tool_name, notified_at DESC) WHERE tool_name IS NOT NULL;

-- Indexes for model filtering
CREATE INDEX IF NOT EXISTS idx_shell_commands_model
    ON shell_commands(model_provider, model_name);

CREATE INDEX IF NOT EXISTS idx_claude_commands_model
    ON claude_commands(model_provider, model_name);

CREATE INDEX IF NOT EXISTS idx_user_messages_model
    ON user_messages(model_provider, model_name);

CREATE INDEX IF NOT EXISTS idx_notifications_model
    ON notifications(model_provider, model_name);

CREATE INDEX IF NOT EXISTS idx_conversations_model
    ON conversations(model_provider, model_name);
