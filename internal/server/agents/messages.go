package agents

import (
	"time"

	"github.com/google/uuid"
)

// MessageType represents WebSocket message types
type MessageType string

const (
	// Authentication
	MessageTypeAuth        MessageType = "auth"
	MessageTypeAuthSuccess MessageType = "auth_success"

	// Session management
	MessageTypeCreateSession MessageType = "create_session"
	MessageTypeSessionCreated MessageType = "session_created"
	MessageTypeEndSession    MessageType = "end_session"
	MessageTypeSessionEnded  MessageType = "session_ended"
	MessageTypeInterruptSession MessageType = "interrupt_session"
	MessageTypeSessionInterrupted MessageType = "session_interrupted"
	MessageTypeDeleteSession MessageType = "delete_session"
	MessageTypeSessionDeleted MessageType = "session_deleted"
	MessageTypeListSessions  MessageType = "list_sessions"
	MessageTypeSessionsList  MessageType = "sessions_list"
	MessageTypeLoadMessages  MessageType = "load_messages"
	MessageTypeMessagesLoaded MessageType = "messages_loaded"

	// Agent interaction
	MessageTypeSendPrompt     MessageType = "send_prompt"
	MessageTypeAgentMessage   MessageType = "agent_message"
	MessageTypeAgentThinking  MessageType = "agent_thinking"
	MessageTypeAgentToolUse   MessageType = "agent_tool_use"
	MessageTypeAgentError     MessageType = "agent_error"

	// Permission requests
	MessageTypePermissionRequest      MessageType = "permission_request"
	MessageTypePermissionResponse     MessageType = "permission_response"
	MessageTypePermissionAcknowledged MessageType = "permission_acknowledged"

	// Kill switch
	MessageTypeKillAllAgents     MessageType = "kill_all_agents"
	MessageTypeAgentsKilled      MessageType = "agents_killed"
	MessageTypeDeleteAllSessions MessageType = "delete_all_sessions"
	MessageTypeAllSessionsDeleted MessageType = "all_sessions_deleted"

	// Session updates
	MessageTypeSessionUpdated MessageType = "session_updated"

	// System
	MessageTypeError MessageType = "error"
	MessageTypePing  MessageType = "ping"
	MessageTypePong  MessageType = "pong"
)

// SessionStatus represents agent session status
type SessionStatus string

const (
	SessionStatusActive     SessionStatus = "active"
	SessionStatusIdle       SessionStatus = "idle"
	SessionStatusProcessing SessionStatus = "processing"
	SessionStatusError      SessionStatus = "error"
	SessionStatusEnded      SessionStatus = "ended"
)

// SessionOptions holds options for creating an agent session
type SessionOptions struct {
	SystemPrompt     *string  `json:"system_prompt,omitempty"`
	AgentName        *string  `json:"agent_name,omitempty"`
	Tools            []string `json:"tools,omitempty"`
	WorkingDirectory *string  `json:"working_directory,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	PermissionMode   *string  `json:"permission_mode,omitempty"`
	Provider         *string  `json:"provider,omitempty"`  // Provider ID (e.g., "glm", "deepseek")
	Model            *string  `json:"model,omitempty"`     // Model name
	BaseURL          *string  `json:"base_url,omitempty"`  // API base URL for custom providers
	APIKey           *string  `json:"api_key,omitempty"`   // API key for the provider
}

// Session represents an agent conversation session
type Session struct {
	ID             uuid.UUID      `json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Status         SessionStatus  `json:"status"`
	Options        SessionOptions `json:"options"`
	MessageCount   int            `json:"message_count"`
	ErrorMessage   *string        `json:"error_message,omitempty"`
	CostUSD          float64        `json:"cost_usd"`
	NumTurns         int            `json:"num_turns"`
	DurationMS       int64          `json:"duration_ms"`
	ModelName        string         `json:"model_name,omitempty"`
	ClaudeSessionID  string         `json:"claude_session_id,omitempty"`  // Claude CLI session ID for resuming conversations
	GitBranch        string         `json:"git_branch,omitempty"`         // Git branch of working directory (if applicable)
}

// BaseMessage represents a base WebSocket message
type BaseMessage struct {
	Type MessageType `json:"type"`
}

// AuthMessage represents authentication request
type AuthMessage struct {
	BaseMessage
	Token string `json:"token"`
}

// CreateSessionMessage represents a session creation request
type CreateSessionMessage struct {
	BaseMessage
	SessionID uuid.UUID      `json:"session_id"`
	Options   SessionOptions `json:"options"`
}

// SessionCreatedMessage represents a session creation response
type SessionCreatedMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
	Session   Session   `json:"session"` // Full session object for frontend
	Status    string    `json:"status"`
}

// ContentBlock represents a piece of content (text or image)
type ContentBlock struct {
	Type   string       `json:"type"`             // "text" or "image"
	Text   string       `json:"text,omitempty"`   // For text blocks
	Source *ImageSource `json:"source,omitempty"` // For image blocks
}

// ImageSource represents base64-encoded image data
type ImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/png", "image/jpeg", "image/gif", "image/webp"
	Data      string `json:"data"`       // Base64 encoded image data
}

// SendPromptMessage represents sending a prompt to an agent
type SendPromptMessage struct {
	BaseMessage
	SessionID uuid.UUID      `json:"session_id"`
	Prompt    string         `json:"prompt,omitempty"`  // Legacy text-only support
	Content   []ContentBlock `json:"content,omitempty"` // New structured content (text + images)
}

// AgentMessageResponse represents a message from the agent
type AgentMessageResponse struct {
	BaseMessage
	SessionID uuid.UUID   `json:"session_id"`
	Content   interface{} `json:"content"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// EndSessionMessage represents ending a session
type EndSessionMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
}

// SessionEndedMessage represents a session end response
type SessionEndedMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
	Status    string    `json:"status"`
}

// InterruptSessionMessage represents interrupting a session
type InterruptSessionMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
}

// SessionInterruptedMessage represents a session interrupt response
type SessionInterruptedMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
	Status    string    `json:"status"`
}

// DeleteSessionMessage represents deleting a session
type DeleteSessionMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
}

// SessionDeletedMessage represents a session deletion response
type SessionDeletedMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
	Status    string    `json:"status"`
}

// ListSessionsMessage represents a request to list sessions
type ListSessionsMessage struct {
	BaseMessage
}

// SessionsListMessage represents a list of sessions response
type SessionsListMessage struct {
	BaseMessage
	Sessions []Session `json:"sessions"`
}

// LoadMessagesMessage represents a request to load messages for a session
type LoadMessagesMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// MessagesLoadedMessage represents a response with loaded messages
type MessagesLoadedMessage struct {
	BaseMessage
	SessionID uuid.UUID        `json:"session_id"`
	Messages  []MessageRecord  `json:"messages"`
	HasMore   bool             `json:"has_more"`
	Count     int              `json:"count"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

// KillAllAgentsMessage represents killing all agents
type KillAllAgentsMessage struct {
	BaseMessage
}

// AgentsKilledMessage represents agents killed response
type AgentsKilledMessage struct {
	BaseMessage
	Count int `json:"count"`
}

// DeleteAllSessionsMessage represents deleting all sessions
type DeleteAllSessionsMessage struct {
	BaseMessage
}

// AllSessionsDeletedMessage represents all sessions deleted response
type AllSessionsDeletedMessage struct {
	BaseMessage
	Count int `json:"count"`
}

// ErrorMessage represents an error response
type ErrorMessage struct {
	BaseMessage
	Content interface{} `json:"content,omitempty"`
	Message string      `json:"message"` // Changed from "error" to match frontend expectation
}

// PermissionRequestMessage represents a permission request
type PermissionRequestMessage struct {
	BaseMessage
	SessionID      uuid.UUID   `json:"session_id"`
	PermissionID   string      `json:"permission_id"`
	Tool           string      `json:"tool"`
	Action         string      `json:"action"`
	Details        interface{} `json:"details,omitempty"`
	Description    string      `json:"description"` // Human-readable description of the permission request
}

// PermissionResponseMessage represents a permission response
type PermissionResponseMessage struct {
	BaseMessage
	SessionID    uuid.UUID `json:"session_id"`
	PermissionID string    `json:"permission_id"`
	Approved     bool      `json:"approved"`
}

// SessionUpdatedMessage represents a session update notification
type SessionUpdatedMessage struct {
	BaseMessage
	SessionID uuid.UUID `json:"session_id"`
	GitBranch *string   `json:"git_branch,omitempty"`
}
