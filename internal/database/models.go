// Package database defines data models for command history and conversation tracking.
// This file contains struct definitions for shell commands, Claude tool invocations,
// conversations, command statistics, and user messages.
package database

import "time"

// ShellCommand represents a shell command execution record
type ShellCommand struct {
	ID               int64     `json:"id"`
	ConversationID   string    `json:"conversation_id"`
	Command          string    `json:"command"`
	Description      string    `json:"description,omitempty"`
	WorkingDirectory string    `json:"working_directory,omitempty"`
	GitBranch        string    `json:"git_branch,omitempty"`
	ExitCode         *int      `json:"exit_code,omitempty"`
	Stdout           string    `json:"stdout,omitempty"`
	Stderr           string    `json:"stderr,omitempty"`
	DurationMs       *int      `json:"duration_ms,omitempty"`
	ExecutedAt       time.Time `json:"executed_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// ClaudeCommand represents a Claude Code tool invocation
type ClaudeCommand struct {
	ID               int64     `json:"id"`
	ConversationID   string    `json:"conversation_id"`
	ToolName         string    `json:"tool_name"`
	Parameters       string    `json:"parameters,omitempty"` // JSON string
	Result           string    `json:"result,omitempty"`     // JSON string
	WorkingDirectory string    `json:"working_directory,omitempty"`
	GitBranch        string    `json:"git_branch,omitempty"`
	Success          bool      `json:"success"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	DurationMs       *int      `json:"duration_ms,omitempty"`
	ExecutedAt       time.Time `json:"executed_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// Conversation represents conversation metadata
type Conversation struct {
	ID                  string    `json:"id"`
	ProjectPath         string    `json:"project_path,omitempty"`
	StartedAt           time.Time `json:"started_at"`
	LastActivityAt      time.Time `json:"last_activity_at"`
	TotalCommands       int       `json:"total_commands"`
	TotalShellCommands  int       `json:"total_shell_commands"`
	TotalTokens         int       `json:"total_tokens"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CommandStat represents aggregated command statistics
type CommandStat struct {
	ID              int64     `json:"id"`
	CommandType     string    `json:"command_type"` // 'shell' or 'claude'
	CommandName     string    `json:"command_name"`
	ExecutionCount  int       `json:"execution_count"`
	SuccessCount    int       `json:"success_count"`
	FailureCount    int       `json:"failure_count"`
	AvgDurationMs   int       `json:"avg_duration_ms"`
	LastExecutedAt  time.Time `json:"last_executed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CommandHistoryQuery represents query parameters for filtering command history
type CommandHistoryQuery struct {
	ConversationID string
	Limit          int
	Offset         int
	StartDate      *time.Time
	EndDate        *time.Time
	ToolName       string
	CommandType    string // 'shell' or 'claude'
}

// UserMessage represents a user's input message
type UserMessage struct {
	ID               int64     `json:"id"`
	ConversationID   string    `json:"conversation_id,omitempty"`
	SessionName      string    `json:"session_name,omitempty"`
	Message          string    `json:"message"`
	WorkingDirectory string    `json:"working_directory,omitempty"`
	GitBranch        string    `json:"git_branch,omitempty"`
	MessageLength    int       `json:"message_length"`
	SubmittedAt      time.Time `json:"submitted_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// ProviderConfig represents an AI provider configuration
type ProviderConfig struct {
	ProviderID string    `json:"provider_id"`
	APIKey     string    `json:"api_key"`
	CustomURL  string    `json:"custom_url,omitempty"`
	ModelName  string    `json:"model_name,omitempty"`
	IsCurrent  bool      `json:"is_current"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
