package agents

import (
	"context"
	"database/sql"
	"encoding/json"
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
	storage  SessionStorage
	db       *sql.DB // Database connection for loading provider configs
}

// PermissionRequest represents a pending permission request
type PermissionRequest struct {
	RequestID   string
	ToolName    string
	Input       map[string]interface{}
	Context     types.ToolPermissionContext
	ResponseChan chan PermissionResponse
}

// PermissionResponse represents the user's response to a permission request
type PermissionResponse struct {
	Approved      bool
	UpdatedInput  *map[string]interface{}
	DenyMessage   string
}

// AgentSession represents an active agent session
type AgentSession struct {
	Session
	ctx                    context.Context
	cancel                 context.CancelFunc
	responseChan           chan types.Message
	permissionReqChan      chan *PermissionRequest  // Outgoing permission requests to frontend
	permissionRespChan     chan *PermissionResponse // Incoming permission responses from frontend
	pendingPermissions     map[string]chan PermissionResponse // Map of request_id -> response channel
	permMu                 sync.Mutex
	permForwarderRunning   bool // Track if permission forwarder goroutine is running
	permForwarderMu        sync.Mutex
	active                 bool
	client                 *claude.Client // Streaming client for this session
	mu                     sync.Mutex     // Protects client field
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *Config, db *sql.DB) (*SessionManager, error) {
	// Initialize storage
	storage, err := NewSQLiteSessionStorage(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize session storage: %w", err)
	}

	sm := &SessionManager{
		sessions: make(map[uuid.UUID]*AgentSession),
		config:   config,
		storage:  storage,
		db:       db,
	}

	// Load active sessions from database
	if err := sm.loadSessionsFromDB(); err != nil {
		logging.Warning("Failed to load sessions from database: %v", err)
		// Don't fail initialization, just log the warning
	}

	return sm, nil
}

// loadSessionsFromDB loads active sessions from the database into memory
func (sm *SessionManager) loadSessionsFromDB() error {
	sessions, err := sm.storage.ListSessions("active")
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, sessionMeta := range sessions {
		// Create an in-memory session object
		session := &AgentSession{
			Session: Session{
				ID:              sessionMeta.ID,
				CreatedAt:       sessionMeta.CreatedAt,
				UpdatedAt:       sessionMeta.UpdatedAt,
				Status:          SessionStatus(sessionMeta.Status),
				MessageCount:    sessionMeta.MessageCount,
				CostUSD:         sessionMeta.CostUSD,
				NumTurns:        sessionMeta.NumTurns,
				DurationMS:      sessionMeta.DurationMS,
				ModelName:       sessionMeta.ModelName,
				ClaudeSessionID: sessionMeta.ClaudeSessionID,
			},
			active: true,
		}

		if sessionMeta.ErrorMessage != "" {
			session.ErrorMessage = &sessionMeta.ErrorMessage
		}

		// Create context for session
		session.ctx, session.cancel = context.WithCancel(context.Background())

		// Create channels
		session.responseChan = make(chan types.Message, 10)
		session.permissionReqChan = make(chan *PermissionRequest, 10)
		session.permissionRespChan = make(chan *PermissionResponse, 10)
		session.pendingPermissions = make(map[string]chan PermissionResponse)

		sm.sessions[sessionMeta.ID] = session

		logging.Info("Loaded session from database: %s (status: %s, messages: %d)",
			sessionMeta.ID, sessionMeta.Status, sessionMeta.MessageCount)
	}

	if len(sessions) > 0 {
		logging.Info("Loaded %d active sessions from database", len(sessions))
	}

	return nil
}

// StartCleanupJob starts a background goroutine that periodically cleans up old sessions
func (sm *SessionManager) StartCleanupJob() {
	if !sm.config.CleanupEnabled {
		logging.Info("Session cleanup job disabled")
		return
	}

	logging.Info("Starting session cleanup job (retention: %d days, interval: %d hours)",
		sm.config.SessionRetentionDays, sm.config.CleanupIntervalHours)

	go func() {
		ticker := time.NewTicker(time.Duration(sm.config.CleanupIntervalHours) * time.Hour)
		defer ticker.Stop()

		// Run cleanup once immediately
		sm.runCleanup()

		// Then run on ticker
		for range ticker.C {
			sm.runCleanup()
		}
	}()
}

// runCleanup performs the actual cleanup of old sessions
func (sm *SessionManager) runCleanup() {
	deleted, err := sm.storage.DeleteOldSessions(sm.config.SessionRetentionDays)
	if err != nil {
		logging.Error("Failed to cleanup old sessions: %v", err)
		return
	}

	if deleted > 0 {
		logging.Info("Cleaned up %d old sessions (retention: %d days)", deleted, sm.config.SessionRetentionDays)
	}
}

// CreateSession creates a new agent session
func (sm *SessionManager) CreateSession(sessionID uuid.UUID, options SessionOptions) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	logging.Debug("CreateSession called for session: %s", sessionID)

	// Check if session already exists in memory
	if _, exists := sm.sessions[sessionID]; exists {
		logging.Warning("Session already exists in memory: %s", sessionID)
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	// Check if session exists in database (from previous server run or browser refresh)
	existingMeta, err := sm.storage.GetSession(sessionID)
	if err == nil && existingMeta != nil {
		logging.Info("Restoring session from database: %s (Claude session: %s)", sessionID, existingMeta.ClaudeSessionID)

		// Restore options from database (if available), otherwise use request options
		restoredOptions := options // Start with request options as fallback
		if existingMeta.OptionsJSON != "" {
			var dbOptions SessionOptions
			if err := json.Unmarshal([]byte(existingMeta.OptionsJSON), &dbOptions); err == nil {
				logging.Debug("Restored options from database for session %s", sessionID)
				restoredOptions = dbOptions
			} else {
				logging.Warning("Failed to deserialize session options from DB for session %s: %v", sessionID, err)
			}
		}

		// Detect git branch if working directory is provided
		gitBranch := existingMeta.GitBranch // Use stored value first
		if gitBranch == "" && restoredOptions.WorkingDirectory != nil && *restoredOptions.WorkingDirectory != "" {
			// If not stored, try to detect it now
			gitBranch = GetGitBranch(*restoredOptions.WorkingDirectory)
		}

		// Restore session to memory with data from database
		session := &AgentSession{
			Session: Session{
				ID:              existingMeta.ID,
				CreatedAt:       existingMeta.CreatedAt,
				UpdatedAt:       time.Now(), // Update to current time
				Status:          SessionStatus(existingMeta.Status),
				Options:         restoredOptions, // Use restored options from database
				MessageCount:    existingMeta.MessageCount,
				CostUSD:         existingMeta.CostUSD,
				NumTurns:        existingMeta.NumTurns,
				DurationMS:      existingMeta.DurationMS,
				ModelName:       existingMeta.ModelName,
				ClaudeSessionID: existingMeta.ClaudeSessionID, // CRITICAL: Restore Claude session ID
				GitBranch:       gitBranch,
			},
			active: true,
		}

		if existingMeta.ErrorMessage != "" {
			session.ErrorMessage = &existingMeta.ErrorMessage
		}

		// Create context for restored session
		session.ctx, session.cancel = context.WithCancel(context.Background())

		// Create response and permission channels
		session.responseChan = make(chan types.Message, 10)
		session.permissionReqChan = make(chan *PermissionRequest, 10)
		session.permissionRespChan = make(chan *PermissionResponse, 10)
		session.pendingPermissions = make(map[string]chan PermissionResponse)

		sm.sessions[sessionID] = session

		logging.Info("Session restored from database: %s (total sessions: %d)", sessionID, len(sm.sessions))
		return &session.Session, nil
	}

	// Session doesn't exist anywhere, create new one
	logging.Debug("Creating new session: %s", sessionID)
	now := time.Now()

	// Detect git branch if working directory is provided
	gitBranch := ""
	if options.WorkingDirectory != nil && *options.WorkingDirectory != "" {
		gitBranch = GetGitBranch(*options.WorkingDirectory)
	}

	session := &AgentSession{
		Session: Session{
			ID:           sessionID,
			CreatedAt:    now,
			UpdatedAt:    now,
			Status:       SessionStatusIdle,
			Options:      options,
			MessageCount: 0,
			CostUSD:      0.0,
			NumTurns:     0,
			DurationMS:   0,
			ModelName:    sm.config.Model,
			GitBranch:    gitBranch,
		},
		active: true,
	}

	// Create context for session
	session.ctx, session.cancel = context.WithCancel(context.Background())

	// Create response and permission channels
	session.responseChan = make(chan types.Message, 10)
	session.permissionReqChan = make(chan *PermissionRequest, 10)
	session.permissionRespChan = make(chan *PermissionResponse, 10)
	session.pendingPermissions = make(map[string]chan PermissionResponse)

	sm.sessions[sessionID] = session

	// Save to database
	if err := sm.saveSessionToDB(&session.Session); err != nil {
		logging.Error("Failed to save session to database: %v", err)
		// Don't fail the creation, just log the error
	}

	logging.Info("Session created: %s (total sessions: %d)", sessionID, len(sm.sessions))

	return &session.Session, nil
}

// sessionToMetadata converts a Session to SessionMetadata
func (sm *SessionManager) sessionToMetadata(session *Session) *SessionMetadata {
	metadata := &SessionMetadata{
		ID:              session.ID,
		Status:          string(session.Status),
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
		EndedAt:         nil,
		MessageCount:    session.MessageCount,
		CostUSD:         session.CostUSD,
		NumTurns:        session.NumTurns,
		DurationMS:      session.DurationMS,
		ModelName:       session.ModelName,
		ClaudeSessionID: session.ClaudeSessionID,
		GitBranch:       session.GitBranch,
	}

	if session.ErrorMessage != nil {
		metadata.ErrorMessage = *session.ErrorMessage
	}

	// Serialize Options to JSON
	if optionsBytes, err := json.Marshal(session.Options); err == nil {
		metadata.OptionsJSON = string(optionsBytes)
	}

	// Check if session is ended
	if session.Status == SessionStatusEnded {
		now := time.Now()
		metadata.EndedAt = &now
	}

	return metadata
}

// saveSessionToDB persists a session to the database
func (sm *SessionManager) saveSessionToDB(session *Session) error {
	return sm.storage.SaveSession(sm.sessionToMetadata(session))
}

// updateSessionInDB updates an existing session in the database
func (sm *SessionManager) updateSessionInDB(session *Session) error {
	return sm.storage.UpdateSession(sm.sessionToMetadata(session))
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

// ListAllSessions returns all sessions (active and ended) from database
func (sm *SessionManager) ListAllSessions(statusFilter string) ([]Session, error) {
	sessionMetas, err := sm.storage.ListSessions(statusFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions from storage: %w", err)
	}

	sessions := make([]Session, 0, len(sessionMetas))
	for _, meta := range sessionMetas {
		session := Session{
			ID:              meta.ID,
			CreatedAt:       meta.CreatedAt,
			UpdatedAt:       meta.UpdatedAt,
			Status:          SessionStatus(meta.Status),
			MessageCount:    meta.MessageCount,
			CostUSD:         meta.CostUSD,
			NumTurns:        meta.NumTurns,
			DurationMS:      meta.DurationMS,
			ModelName:       meta.ModelName,
			ClaudeSessionID: meta.ClaudeSessionID,
			GitBranch:       meta.GitBranch,
		}

		if meta.ErrorMessage != "" {
			session.ErrorMessage = &meta.ErrorMessage
		}

		// Deserialize Options from JSON
		if meta.OptionsJSON != "" {
			var options SessionOptions
			if err := json.Unmarshal([]byte(meta.OptionsJSON), &options); err == nil {
				session.Options = options
			} else {
				logging.Warning("Failed to deserialize session options for session %s: %v", meta.ID, err)
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// EndSession ends a session
func (sm *SessionManager) EndSession(sessionID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Close streaming client if exists
	session.mu.Lock()
	if session.client != nil {
		session.client.Close(session.ctx)
		session.client = nil
	}
	session.mu.Unlock()

	// Cancel context (will stop any ongoing queries)
	if session.cancel != nil {
		session.cancel()
	}

	// Update status
	session.Status = SessionStatusEnded
	session.UpdatedAt = time.Now()
	session.active = false

	// Calculate duration
	session.DurationMS = time.Since(session.CreatedAt).Milliseconds()

	// Update in database
	if err := sm.updateSessionInDB(&session.Session); err != nil {
		logging.Error("Failed to update ended session in database: %v", err)
	}

	// Remove from active sessions map
	delete(sm.sessions, sessionID)

	logging.Info("Session ended: %s (duration: %dms, messages: %d)",
		sessionID, session.DurationMS, session.MessageCount)

	return nil
}

// DeleteSession deletes a session from the database
func (sm *SessionManager) DeleteSession(sessionID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// If session is still active, end it first
	if session, exists := sm.sessions[sessionID]; exists {
		// Close streaming client if exists
		session.mu.Lock()
		if session.client != nil {
			session.client.Close(session.ctx)
			session.client = nil
		}
		session.mu.Unlock()

		// Cancel context
		if session.cancel != nil {
			session.cancel()
		}

		// Remove from active sessions
		delete(sm.sessions, sessionID)
	}

	// Delete from database
	if err := sm.storage.DeleteSession(sessionID); err != nil {
		return fmt.Errorf("failed to delete session from storage: %w", err)
	}

	logging.Info("Session deleted: %s", sessionID)
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

// DeleteAllSessions deletes all sessions from the database
func (sm *SessionManager) DeleteAllSessions() (int, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// First, end all active sessions
	for sessionID, session := range sm.sessions {
		// Close streaming client if exists
		session.mu.Lock()
		if session.client != nil {
			session.client.Close(session.ctx)
			session.client = nil
		}
		session.mu.Unlock()

		// Cancel context
		if session.cancel != nil {
			session.cancel()
		}

		// Remove from active sessions
		delete(sm.sessions, sessionID)
	}

	// Get all sessions from database to count them
	allSessions, err := sm.storage.ListSessions("all")
	if err != nil {
		return 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	count := 0
	// Delete each session from database
	for _, session := range allSessions {
		if err := sm.storage.DeleteSession(session.ID); err != nil {
			logging.Error("Failed to delete session %s: %v", session.ID, err)
			continue
		}
		count++
	}

	logging.Info("Deleted %d sessions from database", count)
	return count, nil
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
	userMsgSequence := session.MessageCount
	sm.mu.Unlock()

	// Save user prompt message to database
	if err := sm.saveMessageToDB(session.ID, userMsgSequence, "user", prompt, "", nil); err != nil {
		logging.Error("Failed to save user message to database: %v", err)
	}

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

	// Create permission callback
	canUseTool := func(ctx context.Context, toolName string, input map[string]interface{}, permCtx types.ToolPermissionContext) (interface{}, error) {
		requestID := uuid.New().String()
		logging.Info("🔐🔐🔐 CALLBACK INVOKED: tool=%s, requestID=%s, input=%+v", toolName, requestID, input)

		// Create response channel for this specific request
		responseChan := make(chan PermissionResponse, 1)

		// Store in pending permissions map
		session.permMu.Lock()
		session.pendingPermissions[requestID] = responseChan
		session.permMu.Unlock()

		// Clean up when done
		defer func() {
			session.permMu.Lock()
			delete(session.pendingPermissions, requestID)
			session.permMu.Unlock()
		}()

		// Send permission request to frontend via channel
		permReq := &PermissionRequest{
			RequestID:    requestID,
			ToolName:     toolName,
			Input:        input,
			Context:      permCtx,
			ResponseChan: responseChan,
		}

		logging.Info("⏳ Sending permission request to channel...")

		select {
		case session.permissionReqChan <- permReq:
			logging.Info("✅ Permission request sent to channel successfully: %s", requestID)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			logging.Warning("Timeout sending permission request to frontend")
			return types.PermissionResultDeny{Message: "Permission request timeout"}, nil
		}

		// Wait for response from frontend
		select {
		case response := <-responseChan:
			logging.Info("Permission response received: approved=%v, requestID=%s", response.Approved, requestID)
			if response.Approved {
				result := types.PermissionResultAllow{}
				if response.UpdatedInput != nil {
					result.UpdatedInput = response.UpdatedInput
				}
				return result, nil
			} else {
				return types.PermissionResultDeny{
					Message: response.DenyMessage,
				}, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(300 * time.Second): // 5 minute timeout for user response
			logging.Warning("Timeout waiting for permission response from user")
			return types.PermissionResultDeny{Message: "Permission response timeout"}, nil
		}
	}

	// Define available tools (Claude Code standard tools)
	allowedTools := []string{
		"Bash", "Read", "Write", "Edit", "Glob", "Grep",
		"WebSearch", "WebFetch",
	}

	// If session options specify tools, use those instead
	if len(session.Options.Tools) > 0 {
		allowedTools = session.Options.Tools
	}

	logging.Info("Allowed tools for session %s: %v", sessionID, allowedTools)
	logging.Info("Permission mode: %v", permMode)

	// Determine model to use: session-specific > config default
	modelToUse := sm.config.Model
	if session.Options.Model != nil && *session.Options.Model != "" {
		modelToUse = *session.Options.Model
		logging.Info("Using session-specific model: %s", modelToUse)
	}

	opts := types.NewClaudeAgentOptions().
		WithModel(modelToUse).
		WithPermissionMode(permMode).
		WithVerbose(sm.config.Verbose).
		WithCanUseTool(canUseTool).
		// Don't set AllowedTools - let permission mode control tool access
		// WithAllowedTools(allowedTools...).
		WithSystemPrompt("code")

	// Set base URL if specified in session options (for custom providers)
	if session.Options.BaseURL != nil && *session.Options.BaseURL != "" {
		logging.Info("Using session-specific base URL: %s", *session.Options.BaseURL)
		opts = opts.WithBaseURL(*session.Options.BaseURL)
	}

	// Set API key: session-specific > database provider config > config default
	apiKeyToUse := sm.config.APIKey

	// If session specifies a provider, try to get API key from database
	if session.Options.Provider != nil && *session.Options.Provider != "" {
		// Query providers table directly
		var apiKey string
		err := sm.db.QueryRow("SELECT api_key FROM providers WHERE provider_id = ? LIMIT 1", *session.Options.Provider).Scan(&apiKey)
		if err == nil && apiKey != "" {
			apiKeyToUse = apiKey
			logging.Info("Using API key from provider database: %s (masked)", *session.Options.Provider)
		} else if err == sql.ErrNoRows {
			logging.Warning("No API key found for provider: %s - configure it via TUI first", *session.Options.Provider)
		} else if err != nil {
			logging.Warning("Failed to load provider config from database: %v", err)
		}
	}

	// Session-specific API key overrides everything
	if session.Options.APIKey != nil && *session.Options.APIKey != "" {
		apiKeyToUse = *session.Options.APIKey
		logging.Info("Using session-specific API key (masked)")
	}

	// Only set API key if provided (don't override SDK's default detection)
	if apiKeyToUse != "" {
		opts = opts.WithEnvVar("ANTHROPIC_API_KEY", apiKeyToUse)
		logging.Debug("API key configured (length: %d)", len(apiKeyToUse))
	} else {
		logging.Warning("No API key configured for this session - make sure ANTHROPIC_API_KEY is set in environment or configure provider in TUI")
	}

	// Set working directory if provided
	if session.Options.WorkingDirectory != nil && *session.Options.WorkingDirectory != "" {
		logging.Debug("SendPrompt: Setting working directory: %s", *session.Options.WorkingDirectory)
		opts = opts.WithCWD(*session.Options.WorkingDirectory)
	}

	// Resume existing conversation if Claude session ID exists
	if session.ClaudeSessionID != "" {
		logging.Debug("SendPrompt: Resuming conversation from Claude session: %s", session.ClaudeSessionID)
		opts = opts.WithResume(session.ClaudeSessionID)
	}

	// Reuse existing client if available (preserves conversation context)
	// Otherwise create a new client
	session.mu.Lock()
	client := session.client
	hasClaudeSession := session.ClaudeSessionID != ""
	session.mu.Unlock()

	if client == nil {
		// Execute query using streaming mode (required for permission callbacks via control protocol)
		if hasClaudeSession {
			logging.Info("SendPrompt: Creating new client for restored session %s (Claude session: %s)", sessionID, session.ClaudeSessionID)
		} else {
			logging.Debug("SendPrompt: Creating streaming client...")
		}
		logging.Debug("SendPrompt: API Key length: %d", len(sm.config.APIKey))
		logging.Debug("Creating streaming client for session %s with options: model=%s, permMode=%v",
			sessionID, sm.config.Model, permMode)

		newClient, err := claude.NewClient(session.ctx, opts)
		if err != nil {
			logging.Error("SendPrompt: Failed to create client: %v", err)
			sm.mu.Lock()
			errMsg := err.Error()
			session.ErrorMessage = &errMsg
			session.Status = SessionStatusError
			sm.mu.Unlock()
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Connect to Claude
		if err := newClient.Connect(session.ctx); err != nil {
			logging.Error("SendPrompt: Failed to connect client: %v", err)
			sm.mu.Lock()
			errMsg := err.Error()
			session.ErrorMessage = &errMsg
			session.Status = SessionStatusError
			sm.mu.Unlock()
			return fmt.Errorf("failed to connect client: %w", err)
		}

		// Store client reference
		session.mu.Lock()
		session.client = newClient
		session.mu.Unlock()
		client = newClient
	} else {
		logging.Info("SendPrompt: Reusing existing client for session %s (preserves conversation context)", sessionID)
	}

	// Send the query
	if err := client.Query(session.ctx, prompt); err != nil {
		logging.Error("SendPrompt: Failed to send query: %v", err)
		sm.mu.Lock()
		errMsg := err.Error()
		session.ErrorMessage = &errMsg
		session.Status = SessionStatusError
		sm.mu.Unlock()
		return fmt.Errorf("failed to send query: %w", err)
	}

	// Get response channel
	messages := client.ReceiveResponse(session.ctx)
	logging.Info("SendPrompt: Client connected and query sent, starting response stream")
	logging.Info("Streaming client created for session %s, starting response stream", sessionID)

	go sm.receiveQueryResponses(session, messages)

	logging.Debug("SendPrompt: Completed successfully for session %s", sessionID)
	return nil
}

// SendPromptWithContent sends structured content (text + images) to an agent session
// This method bypasses the SDK's Query method to support image content blocks
func (sm *SessionManager) SendPromptWithContent(sessionID uuid.UUID, content []ContentBlock) error {
	logging.Debug("SendPromptWithContent: Getting session %s", sessionID)
	session, err := sm.GetSession(sessionID)
	if err != nil {
		logging.Error("SendPromptWithContent: Failed to get session: %v", err)
		return err
	}

	// Update session status
	sm.mu.Lock()
	session.Status = SessionStatusProcessing
	session.UpdatedAt = time.Now()
	session.MessageCount++
	userMsgSequence := session.MessageCount
	sm.mu.Unlock()

	// Convert content blocks to JSON string for database storage
	contentJSON, err := json.Marshal(content)
	if err != nil {
		logging.Error("Failed to marshal content for database: %v", err)
		contentJSON = []byte("[]")
	}

	// Save user prompt message to database (with structured content)
	if err := sm.saveMessageToDB(session.ID, userMsgSequence, "user", string(contentJSON), "", nil); err != nil {
		logging.Error("Failed to save user message to database: %v", err)
	}

	logging.Debug("SendPromptWithContent: Building stream-json message for session %s", sessionID)

	// Get or create client (same as SendPrompt method)
	session.mu.Lock()
	client := session.client
	session.mu.Unlock()

	// If no client exists, we need to create one first
	if client == nil {
		// Create client using the same logic as SendPrompt
		// For brevity, we'll call SendPrompt with an empty string to initialize the client
		// then send our structured content
		logging.Info("SendPromptWithContent: No client exists, initializing session first")

		// We need to create the client but not send a query yet
		// Let's replicate the client creation logic here
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

		// Create permission callback (same as SendPrompt)
		canUseTool := sm.createPermissionCallback(session)

		// Determine model to use
		modelToUse := sm.config.Model
		if session.Options.Model != nil && *session.Options.Model != "" {
			modelToUse = *session.Options.Model
		}

		opts := types.NewClaudeAgentOptions().
			WithModel(modelToUse).
			WithPermissionMode(permMode).
			WithVerbose(sm.config.Verbose).
			WithCanUseTool(canUseTool).
			WithSystemPrompt("code")

		// Set other options (base URL, API key, working directory, resume)
		if session.Options.BaseURL != nil && *session.Options.BaseURL != "" {
			opts = opts.WithBaseURL(*session.Options.BaseURL)
		}

		apiKeyToUse := sm.config.APIKey
		if session.Options.Provider != nil && *session.Options.Provider != "" {
			var apiKey string
			err := sm.db.QueryRow("SELECT api_key FROM providers WHERE provider_id = ? LIMIT 1", *session.Options.Provider).Scan(&apiKey)
			if err == nil && apiKey != "" {
				apiKeyToUse = apiKey
			}
		}
		if session.Options.APIKey != nil && *session.Options.APIKey != "" {
			apiKeyToUse = *session.Options.APIKey
		}
		if apiKeyToUse != "" {
			opts = opts.WithEnvVar("ANTHROPIC_API_KEY", apiKeyToUse)
		}

		if session.Options.WorkingDirectory != nil && *session.Options.WorkingDirectory != "" {
			opts = opts.WithCWD(*session.Options.WorkingDirectory)
		}

		if session.ClaudeSessionID != "" {
			opts = opts.WithResume(session.ClaudeSessionID)
		}

		// Create new client
		newClient, err := claude.NewClient(session.ctx, opts)
		if err != nil {
			logging.Error("SendPromptWithContent: Failed to create client: %v", err)
			sm.mu.Lock()
			errMsg := err.Error()
			session.ErrorMessage = &errMsg
			session.Status = SessionStatusError
			sm.mu.Unlock()
			return fmt.Errorf("failed to create client: %w", err)
		}

		if err := newClient.Connect(session.ctx); err != nil {
			logging.Error("SendPromptWithContent: Failed to connect client: %v", err)
			sm.mu.Lock()
			errMsg := err.Error()
			session.ErrorMessage = &errMsg
			session.Status = SessionStatusError
			sm.mu.Unlock()
			return fmt.Errorf("failed to connect client: %w", err)
		}

		session.mu.Lock()
		session.client = newClient
		session.mu.Unlock()
		client = newClient
	}

	// Convert ContentBlock array to interface{} for SDK
	// The SDK's QueryWithContent accepts interface{} which can be a content array
	contentInterface := make([]interface{}, len(content))
	for i, block := range content {
		blockMap := make(map[string]interface{})
		blockMap["type"] = block.Type

		if block.Type == "text" {
			blockMap["text"] = block.Text
		} else if block.Type == "image" && block.Source != nil {
			blockMap["source"] = map[string]interface{}{
				"type":       block.Source.Type,
				"media_type": block.Source.MediaType,
				"data":       block.Source.Data,
			}
		}

		contentInterface[i] = blockMap
	}

	logging.Info("SendPromptWithContent: Sending %d content blocks to Claude CLI", len(content))

	// Use the new QueryWithContent method to send structured content
	if err := client.QueryWithContent(session.ctx, contentInterface); err != nil {
		logging.Error("SendPromptWithContent: Failed to send query: %v", err)
		sm.mu.Lock()
		errMsg := err.Error()
		session.ErrorMessage = &errMsg
		session.Status = SessionStatusError
		sm.mu.Unlock()
		return fmt.Errorf("failed to send query: %w", err)
	}

	// Get response channel
	messages := client.ReceiveResponse(session.ctx)
	logging.Info("SendPromptWithContent: Starting response stream for session %s", sessionID)

	go sm.receiveQueryResponses(session, messages)

	logging.Debug("SendPromptWithContent: Completed successfully for session %s", sessionID)
	return nil
}

// createPermissionCallback creates the permission callback function for a session
func (sm *SessionManager) createPermissionCallback(session *AgentSession) types.CanUseToolFunc {
	return func(ctx context.Context, toolName string, input map[string]interface{}, permCtx types.ToolPermissionContext) (interface{}, error) {
		requestID := uuid.New().String()
		logging.Info("🔐 PERMISSION CALLBACK: tool=%s, requestID=%s", toolName, requestID)

		responseChan := make(chan PermissionResponse, 1)

		session.permMu.Lock()
		session.pendingPermissions[requestID] = responseChan
		session.permMu.Unlock()

		defer func() {
			session.permMu.Lock()
			delete(session.pendingPermissions, requestID)
			session.permMu.Unlock()
		}()

		// Send permission request to frontend via channel
		session.permissionReqChan <- &PermissionRequest{
			RequestID:    requestID,
			ToolName:     toolName,
			Input:        input,
			Context:      permCtx,
			ResponseChan: responseChan,
		}

		// Wait for response from frontend
		select {
		case response := <-responseChan:
			if response.Approved {
				logging.Info("✅ Permission APPROVED for %s (request %s)", toolName, requestID)
				if response.UpdatedInput != nil {
					return *response.UpdatedInput, nil
				}
				return nil, nil
			} else {
				logging.Info("❌ Permission DENIED for %s (request %s): %s", toolName, requestID, response.DenyMessage)
				return nil, fmt.Errorf("permission denied: %s", response.DenyMessage)
			}
		case <-ctx.Done():
			logging.Warning("⏱️ Permission request TIMEOUT for %s (request %s)", toolName, requestID)
			return nil, fmt.Errorf("permission request timeout")
		}
	}
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
	timeout := time.After(300 * time.Second) // 5 minute timeout for first message

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				logging.Info("Session %s: Messages channel closed after %d messages", session.ID, messageCount)

				// Refresh git branch after conversation turn completes
				if _, changed, err := sm.RefreshGitBranch(session.ID); err == nil && changed {
					logging.Debug("Session %s: Git branch updated after conversation turn", session.ID)
				}

				// Update session in database before finishing
				sm.mu.Lock()
				session.UpdatedAt = time.Now()
				sm.updateSessionInDB(&session.Session)
				sm.mu.Unlock()
				return
			}

			messageCount++
			logging.Debug("Session %s: Received message #%d, type: %s", session.ID, messageCount, msg.GetMessageType())

			// Refresh git branch before forwarding message (especially after tool execution)
			// This ensures the current message will have the updated git branch
			if _, _, err := sm.RefreshGitBranch(session.ID); err != nil {
				logging.Debug("Session %s: Failed to refresh git branch: %v", session.ID, err)
			}

			// Increment session message count atomically and get sequence number
			sm.mu.Lock()
			session.MessageCount++
			sequenceNum := session.MessageCount
			sm.mu.Unlock()

			// Save message to database based on type with proper sequence number
			sm.persistSDKMessage(session.ID, sequenceNum, msg)

			select {
			case session.responseChan <- msg:
				logging.Debug("Session %s: Message #%d forwarded to response channel", session.ID, messageCount)
			case <-session.ctx.Done():
				logging.Info("Session %s: Context cancelled after %d messages", session.ID, messageCount)
				return
			}

			// Reset timeout after each message
			timeout = time.After(300 * time.Second)

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

// RefreshGitBranch checks and updates the git branch for a session
// Returns the new branch name and whether it changed
func (sm *SessionManager) RefreshGitBranch(sessionID uuid.UUID) (newBranch string, changed bool, err error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return "", false, err
	}

	// Only refresh if we have a working directory
	if session.Options.WorkingDirectory == nil || *session.Options.WorkingDirectory == "" {
		return session.GitBranch, false, nil
	}

	// Detect current git branch
	currentBranch := GetGitBranch(*session.Options.WorkingDirectory)

	// Check if it changed
	changed = currentBranch != session.GitBranch

	if changed {
		logging.Info("Git branch changed for session %s: %s -> %s", sessionID, session.GitBranch, currentBranch)
		sm.mu.Lock()
		session.GitBranch = currentBranch
		session.UpdatedAt = time.Now()
		sm.updateSessionInDB(&session.Session)
		sm.mu.Unlock()
	}

	return currentBranch, changed, nil
}

// StartPermissionForwarder marks that the permission forwarder is running
// Returns true if this call started it, false if it was already running
func (s *AgentSession) StartPermissionForwarder() bool {
	s.permForwarderMu.Lock()
	defer s.permForwarderMu.Unlock()

	if s.permForwarderRunning {
		return false // Already running
	}

	s.permForwarderRunning = true
	return true // Started by this call
}

// saveMessageToDB persists a message to the database
func (sm *SessionManager) saveMessageToDB(sessionID uuid.UUID, sequence int, role, content, thinkingContent string, toolUses interface{}) error {
	var toolUsesJSON []byte
	if toolUses != nil {
		var err error
		toolUsesJSON, err = json.Marshal(toolUses)
		if err != nil {
			return fmt.Errorf("failed to marshal tool uses: %w", err)
		}
	}

	msg := &MessageRecord{
		ID:              uuid.New(),
		SessionID:       sessionID,
		Sequence:        sequence,
		Role:            role,
		Content:         content,
		ThinkingContent: thinkingContent,
		ToolUses:        toolUsesJSON,
		Timestamp:       time.Now(),
		TokensUsed:      0, // TODO: Extract from SDK response if available
	}

	return sm.storage.SaveMessage(msg)
}

// GetMessages retrieves messages for a session with pagination
func (sm *SessionManager) GetMessages(sessionID uuid.UUID, limit, offset int) ([]*MessageRecord, bool, error) {
	return sm.storage.GetMessages(sessionID, limit, offset)
}

// persistSDKMessage saves an SDK message to the database
func (sm *SessionManager) persistSDKMessage(sessionID uuid.UUID, sequence int, msg types.Message) {
	messageType := msg.GetMessageType()

	switch messageType {
	case "assistant":
		// Assistant message contains multiple content blocks
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			var textContent, thinkingContent string
			var toolUses []map[string]interface{}

			// Extract content from blocks
			for _, block := range assistantMsg.Content {
				switch b := block.(type) {
				case *types.TextBlock:
					if textContent != "" {
						textContent += "\n"
					}
					textContent += b.Text

				case *types.ThinkingBlock:
					if thinkingContent != "" {
						thinkingContent += "\n"
					}
					thinkingContent += b.Thinking

				case *types.ToolUseBlock:
					toolUses = append(toolUses, map[string]interface{}{
						"id":    b.ID,
						"name":  b.Name,
						"input": b.Input,
					})
				}
			}

			// Save the combined assistant message
			var toolUsesData interface{}
			if len(toolUses) > 0 {
				toolUsesData = toolUses
			}

			if err := sm.saveMessageToDB(sessionID, sequence, "assistant", textContent, thinkingContent, toolUsesData); err != nil {
				logging.Error("Failed to save assistant message: %v", err)
			}
		}

	case "result":
		// Result message with cost and usage info
		if resultMsg, ok := msg.(*types.ResultMessage); ok {
			content := ""
			if resultMsg.Result != nil {
				content = *resultMsg.Result
			}

			resultData := map[string]interface{}{
				"duration_ms":     resultMsg.DurationMs,
				"duration_api_ms": resultMsg.DurationAPIMs,
				"is_error":        resultMsg.IsError,
				"num_turns":       resultMsg.NumTurns,
			}
			if resultMsg.TotalCostUSD != nil {
				resultData["total_cost_usd"] = *resultMsg.TotalCostUSD
			}
			if resultMsg.Usage != nil {
				resultData["usage"] = resultMsg.Usage
			}

			if err := sm.saveMessageToDB(sessionID, sequence, "system", content, "", resultData); err != nil {
				logging.Error("Failed to save result message: %v", err)
			}

			// Update session with cost and turn info
			sm.mu.Lock()
			if session, exists := sm.sessions[sessionID]; exists {
				if resultMsg.TotalCostUSD != nil {
					session.CostUSD = *resultMsg.TotalCostUSD
				}
				session.NumTurns = resultMsg.NumTurns
				session.DurationMS = int64(resultMsg.DurationMs)

				// Extract and store Claude CLI session ID for resuming conversations
				if resultMsg.SessionID != "" && session.ClaudeSessionID == "" {
					session.ClaudeSessionID = resultMsg.SessionID
					logging.Debug("Extracted Claude session ID for session %s: %s", sessionID, resultMsg.SessionID)

					// Persist to database
					metadata := sm.sessionToMetadata(&session.Session)
					if err := sm.storage.UpdateSession(metadata); err != nil {
						logging.Error("Failed to persist Claude session ID: %v", err)
					}
				}
			}
			sm.mu.Unlock()
		}

	case "user":
		// User message (shouldn't normally come through here, but handle it)
		if userMsg, ok := msg.(*types.UserMessage); ok {
			content := ""
			if str, ok := userMsg.Content.(string); ok {
				content = str
			}
			if err := sm.saveMessageToDB(sessionID, sequence, "user", content, "", nil); err != nil {
				logging.Error("Failed to save user message: %v", err)
			}
		}

	default:
		// Log unhandled message types (system, stream_event, etc.)
		logging.Debug("Unhandled message type for persistence: %s", messageType)
	}
}
