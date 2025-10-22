package agents

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionStorage defines the interface for persisting agent sessions and messages
type SessionStorage interface {
	// Session operations
	SaveSession(session *SessionMetadata) error
	UpdateSession(session *SessionMetadata) error
	GetSession(sessionID uuid.UUID) (*SessionMetadata, error)
	ListSessions(statusFilter string) ([]*SessionMetadata, error)
	DeleteSession(sessionID uuid.UUID) error

	// Message operations
	SaveMessage(msg *MessageRecord) error
	GetMessages(sessionID uuid.UUID, limit, offset int) ([]*MessageRecord, bool, error)
	GetMessageCount(sessionID uuid.UUID) (int, error)

	// Cleanup
	DeleteOldSessions(retentionDays int) (int64, error)
}

// SessionMetadata represents a persisted agent session
type SessionMetadata struct {
	ID             uuid.UUID       `json:"id"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	EndedAt        *time.Time      `json:"ended_at,omitempty"`
	MessageCount    int             `json:"message_count"`
	CostUSD         float64         `json:"cost_usd"`
	NumTurns        int             `json:"num_turns"`
	DurationMS      int64           `json:"duration_ms"`
	ErrorMessage    string          `json:"error_message,omitempty"`
	ModelName       string          `json:"model_name,omitempty"`
	ClaudeSessionID string          `json:"claude_session_id,omitempty"`  // Claude CLI session ID for resuming
	GitBranch       string          `json:"git_branch,omitempty"`         // Git branch of working directory
	OptionsJSON     string          `json:"options_json,omitempty"`       // JSON-serialized SessionOptions
}

// MessageRecord represents a persisted message
type MessageRecord struct {
	ID              uuid.UUID       `json:"id"`
	SessionID       uuid.UUID       `json:"session_id"`
	Sequence        int             `json:"sequence"`
	Role            string          `json:"role"` // user, assistant, system
	Content         string          `json:"content"`
	ThinkingContent string          `json:"thinking_content,omitempty"`
	ToolUses        json.RawMessage `json:"tool_uses,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
	TokensUsed      int             `json:"tokens_used"`
}

// SQLiteSessionStorage implements SessionStorage using SQLite
type SQLiteSessionStorage struct {
	db *sql.DB
}

// NewSQLiteSessionStorage creates a new SQLite session storage
func NewSQLiteSessionStorage(db *sql.DB) (*SQLiteSessionStorage, error) {
	storage := &SQLiteSessionStorage{db: db}

	// Initialize tables if they don't exist
	if err := InitializeAgentTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize agent tables: %w", err)
	}

	// Run migration to fix message sequences (idempotent)
	if err := storage.FixMessageSequences(); err != nil {
		// Log warning but don't fail initialization
		fmt.Printf("Warning: Failed to fix message sequences: %v\n", err)
	}

	return storage, nil
}

// SaveSession inserts a new session into the database
func (s *SQLiteSessionStorage) SaveSession(session *SessionMetadata) error {
	query := `
		INSERT INTO agent_sessions (
			id, status, created_at, updated_at, ended_at,
			message_count, cost_usd, num_turns, duration_ms,
			error_message, model_name, claude_session_id, git_branch, options
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		session.ID.String(),
		session.Status,
		session.CreatedAt,
		session.UpdatedAt,
		session.EndedAt,
		session.MessageCount,
		session.CostUSD,
		session.NumTurns,
		session.DurationMS,
		session.ErrorMessage,
		session.ModelName,
		session.ClaudeSessionID,
		session.GitBranch,
		session.OptionsJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// UpdateSession updates an existing session in the database
func (s *SQLiteSessionStorage) UpdateSession(session *SessionMetadata) error {
	query := `
		UPDATE agent_sessions
		SET status = ?, updated_at = ?, ended_at = ?,
		    message_count = ?, cost_usd = ?, num_turns = ?,
		    duration_ms = ?, error_message = ?, model_name = ?,
		    claude_session_id = ?, git_branch = ?, options = ?
		WHERE id = ?
	`

	result, err := s.db.Exec(
		query,
		session.Status,
		session.UpdatedAt,
		session.EndedAt,
		session.MessageCount,
		session.CostUSD,
		session.NumTurns,
		session.DurationMS,
		session.ErrorMessage,
		session.ModelName,
		session.ClaudeSessionID,
		session.GitBranch,
		session.OptionsJSON,
		session.ID.String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	return nil
}

// GetSession retrieves a session by ID
func (s *SQLiteSessionStorage) GetSession(sessionID uuid.UUID) (*SessionMetadata, error) {
	query := `
		SELECT id, status, created_at, updated_at, ended_at,
		       message_count, cost_usd, num_turns, duration_ms,
		       error_message, model_name, claude_session_id, git_branch, options
		FROM agent_sessions
		WHERE id = ?
	`

	session := &SessionMetadata{}
	var idStr string
	var endedAt sql.NullTime
	var errorMsg, modelName, claudeSessionID, gitBranch, optionsJSON sql.NullString

	err := s.db.QueryRow(query, sessionID.String()).Scan(
		&idStr,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
		&endedAt,
		&session.MessageCount,
		&session.CostUSD,
		&session.NumTurns,
		&session.DurationMS,
		&errorMsg,
		&modelName,
		&claudeSessionID,
		&gitBranch,
		&optionsJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Parse UUID
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID in database: %w", err)
	}
	session.ID = parsedID

	// Handle nullable fields
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	if errorMsg.Valid {
		session.ErrorMessage = errorMsg.String
	}
	if modelName.Valid {
		session.ModelName = modelName.String
	}
	if claudeSessionID.Valid {
		session.ClaudeSessionID = claudeSessionID.String
	}
	if gitBranch.Valid {
		session.GitBranch = gitBranch.String
	}
	if optionsJSON.Valid {
		session.OptionsJSON = optionsJSON.String
	}

	return session, nil
}

// ListSessions retrieves sessions filtered by status
// statusFilter can be: "all", "active", "idle", "processing", "error", "ended"
func (s *SQLiteSessionStorage) ListSessions(statusFilter string) ([]*SessionMetadata, error) {
	var query string
	var args []interface{}

	if statusFilter == "all" || statusFilter == "" {
		query = `
			SELECT id, status, created_at, updated_at, ended_at,
			       message_count, cost_usd, num_turns, duration_ms,
			       error_message, model_name, claude_session_id, git_branch, options
			FROM agent_sessions
			ORDER BY updated_at DESC
		`
	} else if statusFilter == "active" {
		// Active means any session that hasn't ended
		query = `
			SELECT id, status, created_at, updated_at, ended_at,
			       message_count, cost_usd, num_turns, duration_ms,
			       error_message, model_name, claude_session_id, git_branch, options
			FROM agent_sessions
			WHERE status != 'ended'
			ORDER BY updated_at DESC
		`
	} else {
		query = `
			SELECT id, status, created_at, updated_at, ended_at,
			       message_count, cost_usd, num_turns, duration_ms,
			       error_message, model_name, claude_session_id, git_branch, options
			FROM agent_sessions
			WHERE status = ?
			ORDER BY updated_at DESC
		`
		args = append(args, statusFilter)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*SessionMetadata
	for rows.Next() {
		session := &SessionMetadata{}
		var idStr string
		var endedAt sql.NullTime
		var errorMsg, modelName, claudeSessionID, gitBranch, optionsJSON sql.NullString

		err := rows.Scan(
			&idStr,
			&session.Status,
			&session.CreatedAt,
			&session.UpdatedAt,
			&endedAt,
			&session.MessageCount,
			&session.CostUSD,
			&session.NumTurns,
			&session.DurationMS,
			&errorMsg,
			&modelName,
			&claudeSessionID,
			&gitBranch,
			&optionsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Parse UUID
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid session ID in database: %w", err)
		}
		session.ID = parsedID

		// Handle nullable fields
		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}
		if errorMsg.Valid {
			session.ErrorMessage = errorMsg.String
		}
		if modelName.Valid {
			session.ModelName = modelName.String
		}
		if claudeSessionID.Valid {
			session.ClaudeSessionID = claudeSessionID.String
		}
		if gitBranch.Valid {
			session.GitBranch = gitBranch.String
		}
		if optionsJSON.Valid {
			session.OptionsJSON = optionsJSON.String
		}

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// DeleteSession removes a session and its messages
func (s *SQLiteSessionStorage) DeleteSession(sessionID uuid.UUID) error {
	query := `DELETE FROM agent_sessions WHERE id = ?`

	result, err := s.db.Exec(query, sessionID.String())
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Messages are automatically deleted via CASCADE

	return nil
}

// SaveMessage inserts a new message into the database
func (s *SQLiteSessionStorage) SaveMessage(msg *MessageRecord) error {
	query := `
		INSERT INTO agent_messages (
			id, session_id, sequence, role, content,
			thinking_content, tool_uses, timestamp, tokens_used
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var toolUsesStr sql.NullString
	if len(msg.ToolUses) > 0 {
		toolUsesStr = sql.NullString{String: string(msg.ToolUses), Valid: true}
	}

	_, err := s.db.Exec(
		query,
		msg.ID.String(),
		msg.SessionID.String(),
		msg.Sequence,
		msg.Role,
		msg.Content,
		msg.ThinkingContent,
		toolUsesStr,
		msg.Timestamp,
		msg.TokensUsed,
	)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// GetMessages retrieves messages for a session with pagination
// Returns: messages, hasMore, error
func (s *SQLiteSessionStorage) GetMessages(sessionID uuid.UUID, limit, offset int) ([]*MessageRecord, bool, error) {
	// Query limit+1 to check if there are more messages
	query := `
		SELECT id, session_id, sequence, role, content,
		       thinking_content, tool_uses, timestamp, tokens_used
		FROM agent_messages
		WHERE session_id = ?
		ORDER BY sequence ASC, timestamp ASC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, sessionID.String(), limit+1, offset)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*MessageRecord
	for rows.Next() {
		msg := &MessageRecord{}
		var idStr, sessionIDStr string
		var thinkingContent sql.NullString
		var toolUses sql.NullString

		err := rows.Scan(
			&idStr,
			&sessionIDStr,
			&msg.Sequence,
			&msg.Role,
			&msg.Content,
			&thinkingContent,
			&toolUses,
			&msg.Timestamp,
			&msg.TokensUsed,
		)
		if err != nil {
			return nil, false, fmt.Errorf("failed to scan message: %w", err)
		}

		// Parse UUIDs
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, false, fmt.Errorf("invalid message ID in database: %w", err)
		}
		msg.ID = parsedID

		parsedSessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			return nil, false, fmt.Errorf("invalid session ID in database: %w", err)
		}
		msg.SessionID = parsedSessionID

		// Handle nullable fields
		if thinkingContent.Valid {
			msg.ThinkingContent = thinkingContent.String
		}
		if toolUses.Valid {
			msg.ToolUses = json.RawMessage(toolUses.String)
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, false, fmt.Errorf("error iterating messages: %w", err)
	}

	// Check if there are more messages
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit] // Trim to requested limit
	}

	return messages, hasMore, nil
}

// GetMessageCount returns the total number of messages for a session
func (s *SQLiteSessionStorage) GetMessageCount(sessionID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM agent_messages WHERE session_id = ?`

	var count int
	err := s.db.QueryRow(query, sessionID.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}

	return count, nil
}

// DeleteOldSessions removes sessions older than retentionDays
func (s *SQLiteSessionStorage) DeleteOldSessions(retentionDays int) (int64, error) {
	return CleanupOldSessions(s.db, retentionDays)
}

// FixMessageSequences resequences all messages based on timestamp order
// This is an idempotent migration that can be run multiple times safely
func (s *SQLiteSessionStorage) FixMessageSequences() error {
	// Use a transaction for atomicity
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create temp table with correct sequences
	_, err = tx.Exec(`
		CREATE TEMP TABLE IF NOT EXISTS temp_sequences AS
		SELECT
			id,
			ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY timestamp ASC, sequence ASC) as new_sequence
		FROM agent_messages
	`)
	if err != nil {
		return fmt.Errorf("failed to create temp sequences table: %w", err)
	}

	// Update all messages with correct sequence numbers
	result, err := tx.Exec(`
		UPDATE agent_messages
		SET sequence = (
			SELECT new_sequence
			FROM temp_sequences
			WHERE temp_sequences.id = agent_messages.id
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to update message sequences: %w", err)
	}

	// Update session message counts to match actual message count
	_, err = tx.Exec(`
		UPDATE agent_sessions
		SET message_count = (
			SELECT MAX(sequence)
			FROM agent_messages
			WHERE agent_messages.session_id = agent_sessions.id
		)
		WHERE EXISTS (
			SELECT 1
			FROM agent_messages
			WHERE agent_messages.session_id = agent_sessions.id
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to update session message counts: %w", err)
	}

	// Clean up temp table
	_, err = tx.Exec(`DROP TABLE IF EXISTS temp_sequences`)
	if err != nil {
		return fmt.Errorf("failed to drop temp sequences table: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Fixed sequence numbers for %d messages\n", rowsAffected)
	}

	return nil
}
