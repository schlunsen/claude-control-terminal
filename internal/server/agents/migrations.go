package agents

import (
	"database/sql"
	"fmt"
	"log"
)

// Migration system for agent session persistence tables

const (
	// Schema version for tracking migrations
	currentSchemaVersion = 1
)

// InitializeAgentTables creates the necessary tables for agent session persistence
// if they don't already exist. This is idempotent and safe to call multiple times.
func InitializeAgentTables(db *sql.DB) error {
	log.Printf("InitializeAgentTables: Starting agent table initialization")

	// Check if tables already exist
	exists, err := tablesExist(db)
	if err != nil {
		log.Printf("ERROR: Failed to check if tables exist: %v", err)
		return fmt.Errorf("failed to check if tables exist: %w", err)
	}

	if exists {
		log.Printf("InitializeAgentTables: Tables already exist, skipping creation")
		// Tables already exist, no need to create
		return nil
	}

	log.Printf("InitializeAgentTables: Tables don't exist, creating them now")

	// Create tables using transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		log.Printf("ERROR: Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Create agent_sessions table
	log.Printf("InitializeAgentTables: Creating agent_sessions table")
	if err := createSessionsTable(tx); err != nil {
		log.Printf("ERROR: Failed to create agent_sessions table: %v", err)
		return fmt.Errorf("failed to create agent_sessions table: %w", err)
	}

	// Create agent_messages table
	log.Printf("InitializeAgentTables: Creating agent_messages table")
	if err := createMessagesTable(tx); err != nil {
		log.Printf("ERROR: Failed to create agent_messages table: %v", err)
		return fmt.Errorf("failed to create agent_messages table: %w", err)
	}

	// Create indexes
	log.Printf("InitializeAgentTables: Creating indexes")
	if err := createIndexes(tx); err != nil {
		log.Printf("ERROR: Failed to create indexes: %v", err)
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Commit transaction
	log.Printf("InitializeAgentTables: Committing transaction")
	if err := tx.Commit(); err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("InitializeAgentTables: Successfully created agent tables")

	return nil
}

// tablesExist checks if the agent persistence tables already exist
func tablesExist(db *sql.DB) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type='table' AND name IN ('agent_sessions', 'agent_messages')
	`
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return false, err
	}

	// Both tables should exist (count = 2) or neither (count = 0)
	return count == 2, nil
}

// createSessionsTable creates the agent_sessions table
func createSessionsTable(tx *sql.Tx) error {
	schema := `
		CREATE TABLE IF NOT EXISTS agent_sessions (
			id TEXT PRIMARY KEY,
			status TEXT NOT NULL DEFAULT 'idle',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			ended_at TIMESTAMP,
			message_count INTEGER NOT NULL DEFAULT 0,
			cost_usd REAL NOT NULL DEFAULT 0.0,
			num_turns INTEGER NOT NULL DEFAULT 0,
			duration_ms INTEGER NOT NULL DEFAULT 0,
			error_message TEXT,
			model_name TEXT,
			transcript_path TEXT,
			CONSTRAINT status_check CHECK (status IN ('idle', 'active', 'processing', 'error', 'ended'))
		);
	`

	_, err := tx.Exec(schema)
	return err
}

// createMessagesTable creates the agent_messages table
func createMessagesTable(tx *sql.Tx) error {
	schema := `
		CREATE TABLE IF NOT EXISTS agent_messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			sequence INTEGER NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			thinking_content TEXT,
			tool_uses TEXT,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			tokens_used INTEGER DEFAULT 0,
			FOREIGN KEY (session_id) REFERENCES agent_sessions(id) ON DELETE CASCADE,
			CONSTRAINT role_check CHECK (role IN ('user', 'assistant', 'system'))
		);
	`

	_, err := tx.Exec(schema)
	return err
}

// createIndexes creates indexes for efficient querying
func createIndexes(tx *sql.Tx) error {
	indexes := []string{
		// Index for listing sessions by status and update time
		`CREATE INDEX IF NOT EXISTS idx_agent_sessions_status
		 ON agent_sessions(status, updated_at DESC)`,

		// Index for finding sessions by creation time
		`CREATE INDEX IF NOT EXISTS idx_agent_sessions_created
		 ON agent_sessions(created_at DESC)`,

		// Index for ended sessions (for cleanup)
		`CREATE INDEX IF NOT EXISTS idx_agent_sessions_ended
		 ON agent_sessions(ended_at DESC) WHERE ended_at IS NOT NULL`,

		// Index for messages by session (most important for pagination)
		`CREATE INDEX IF NOT EXISTS idx_agent_messages_session
		 ON agent_messages(session_id, sequence ASC)`,

		// Index for messages by timestamp
		`CREATE INDEX IF NOT EXISTS idx_agent_messages_timestamp
		 ON agent_messages(timestamp DESC)`,

		// Index for message sequence ordering
		`CREATE INDEX IF NOT EXISTS idx_agent_messages_sequence
		 ON agent_messages(session_id, sequence DESC)`,
	}

	for _, indexSQL := range indexes {
		if _, err := tx.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// CleanupOldSessions deletes sessions older than the specified number of days
func CleanupOldSessions(db *sql.DB, retentionDays int) (int64, error) {
	query := `
		DELETE FROM agent_sessions
		WHERE ended_at IS NOT NULL
		AND ended_at < datetime('now', '-' || ? || ' days')
	`

	result, err := db.Exec(query, retentionDays)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetSchemaVersion returns the current schema version (for future migrations)
func GetSchemaVersion() int {
	return currentSchemaVersion
}
