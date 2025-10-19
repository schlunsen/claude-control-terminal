package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	claude "github.com/schlunsen/claude-agent-sdk-go"
	"github.com/schlunsen/claude-agent-sdk-go/types"
	"github.com/schlunsen/claude-control-terminal/internal/logging"
)

// SessionManager manages agent sessions
type SessionManager struct {
	sessions map[uuid.UUID]*AgentSession
	mu       sync.RWMutex
	config   *Config
}

// AgentSession represents an active agent session
type AgentSession struct {
	Session
	ctx           context.Context
	cancel        context.CancelFunc
	responseChan  chan types.Message
	permissionReq chan *PermissionRequestMessage
	active        bool
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *Config) *SessionManager {
	return &SessionManager{
		sessions: make(map[uuid.UUID]*AgentSession),
		config:   config,
	}
}

// CreateSession creates a new agent session
func (sm *SessionManager) CreateSession(sessionID uuid.UUID, options SessionOptions) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	logging.Debug("CreateSession called for session: %s", sessionID)

	// Check if session already exists
	if _, exists := sm.sessions[sessionID]; exists {
		logging.Warning("Session already exists: %s", sessionID)
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	// Create session
	session := &AgentSession{
		Session: Session{
			ID:           sessionID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			Status:       SessionStatusIdle,
			Options:      options,
			MessageCount: 0,
		},
		active: true,
	}

	// Create context for session
	session.ctx, session.cancel = context.WithCancel(context.Background())

	// Create response and permission channels
	session.responseChan = make(chan types.Message, 10)
	session.permissionReq = make(chan *PermissionRequestMessage, 10)

	sm.sessions[sessionID] = session

	logging.Info("Session created: %s (total sessions: %d)", sessionID, len(sm.sessions))

	return &session.Session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID uuid.UUID) (*AgentSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// ListSessions returns all active sessions
func (sm *SessionManager) ListSessions() []Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		sessions = append(sessions, s.Session)
	}

	return sessions
}

// EndSession ends a session
func (sm *SessionManager) EndSession(sessionID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Cancel context (will stop any ongoing queries)
	if session.cancel != nil {
		session.cancel()
	}

	// Update status
	session.Status = SessionStatusEnded
	session.active = false

	// Remove from sessions
	delete(sm.sessions, sessionID)

	return nil
}

// EndAllSessions ends all active sessions
func (sm *SessionManager) EndAllSessions() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	count := 0
	for sessionID, session := range sm.sessions {
		if session.cancel != nil {
			session.cancel()
		}
		delete(sm.sessions, sessionID)
		count++
	}

	return count
}

// SendPrompt sends a prompt to an agent session using claude.Query
func (sm *SessionManager) SendPrompt(sessionID uuid.UUID, prompt string) error {
	logging.Debug("SendPrompt: Getting session %s", sessionID)
	session, err := sm.GetSession(sessionID)
	if err != nil {
		logging.Error("SendPrompt: Failed to get session: %v", err)
		return err
	}

	// Update session status
	sm.mu.Lock()
	session.Status = SessionStatusProcessing
	session.UpdatedAt = time.Now()
	session.MessageCount++
	sm.mu.Unlock()

	logging.Debug("SendPrompt: Executing query for session %s", sessionID)

	// Determine permission mode
	permMode := types.PermissionModeDefault
	if session.Options.PermissionMode != nil {
		switch *session.Options.PermissionMode {
		case "allow-all":
			permMode = types.PermissionModeBypassPermissions
		case "read-only":
			permMode = types.PermissionModeDefault
		default:
			permMode = types.PermissionModeDefault
		}
	}

	// Build SDK options
	logging.Debug("SendPrompt: Building SDK options (model: %s, permMode: %v, verbose: %v)", sm.config.Model, permMode, sm.config.Verbose)
	opts := types.NewClaudeAgentOptions().
		WithModel(sm.config.Model).
		WithPermissionMode(permMode).
		WithEnvVar("ANTHROPIC_API_KEY", sm.config.APIKey).
		WithVerbose(sm.config.Verbose)

	// Set working directory if provided
	if session.Options.WorkingDirectory != nil && *session.Options.WorkingDirectory != "" {
		logging.Debug("SendPrompt: Setting working directory: %s", *session.Options.WorkingDirectory)
		opts = opts.WithCWD(*session.Options.WorkingDirectory)
	}

	// Execute query (uses non-streaming mode, no control protocol initialization)
	logging.Debug("SendPrompt: Executing claude.Query...")
	logging.Debug("SendPrompt: API Key length: %d", len(sm.config.APIKey))
	logging.Debug("Executing claude.Query for session %s with options: model=%s, permMode=%v",
		sessionID, sm.config.Model, permMode)

	messages, err := claude.Query(session.ctx, prompt, opts)
	if err != nil {
		logging.Error("SendPrompt: Failed to execute query: %v", err)
		sm.mu.Lock()
		errMsg := err.Error()
		session.ErrorMessage = &errMsg
		session.Status = SessionStatusError
		sm.mu.Unlock()
		return fmt.Errorf("failed to execute query: %w", err)
	}
	if messages == nil {
		logging.Error("SendPrompt: messages channel is nil for session %s", sessionID)
		return fmt.Errorf("messages channel is nil")
	}
	logging.Info("SendPrompt: Query executed successfully, messages channel ready")
	logging.Info("Query executed successfully for session %s, starting response stream", sessionID)

	// Start receiving responses in background
	go sm.receiveQueryResponses(session, messages)

	logging.Debug("SendPrompt: Completed successfully for session %s", sessionID)
	return nil
}

// receiveQueryResponses receives responses from a Query and sends them to the response channel
func (sm *SessionManager) receiveQueryResponses(session *AgentSession, messages <-chan types.Message) {
	defer func() {
		if r := recover(); r != nil {
			logging.Error("Session %s: PANIC in receiveQueryResponses: %v", session.ID, r)
		}
		sm.mu.Lock()
		session.Status = SessionStatusIdle
		session.UpdatedAt = time.Now()
		sm.mu.Unlock()
		logging.Debug("Session %s: Query response receiving completed", session.ID)
	}()

	logging.Debug("Session %s: Starting to receive query responses", session.ID)

	messageCount := 0
	timeout := time.After(60 * time.Second) // 60 second timeout for first message

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				logging.Info("Session %s: Messages channel closed after %d messages", session.ID, messageCount)
				return
			}

			messageCount++
			logging.Debug("Session %s: Received message #%d, type: %s", session.ID, messageCount, msg.GetMessageType())

			select {
			case session.responseChan <- msg:
				logging.Debug("Session %s: Message #%d forwarded to response channel", session.ID, messageCount)
			case <-session.ctx.Done():
				logging.Info("Session %s: Context cancelled after %d messages", session.ID, messageCount)
				return
			}

			// Reset timeout after first message
			timeout = time.After(60 * time.Second)

		case <-timeout:
			logging.Warning("Session %s: TIMEOUT waiting for messages (received %d so far)", session.ID, messageCount)
			return

		case <-session.ctx.Done():
			logging.Info("Session %s: Context cancelled while waiting for messages", session.ID)
			return
		}
	}
}

// GetResponseChannel returns the response channel for a session
func (sm *SessionManager) GetResponseChannel(sessionID uuid.UUID) (chan types.Message, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session.responseChan, nil
}
