package agents

import (
	"database/sql"
	"fmt"
	"log"
)

// Migration system for agent session persistence tables

const (
	// Current schema version - increment when adding new migrations
	currentSchemaVersion = 2
)

// InitializeAgentTables creates the necessary tables for agent session persistence
// if they don't already exist, and runs any pending migrations.
// This is idempotent and safe to call multiple times.
func InitializeAgentTables(db *sql.DB) error {
	log.Printf("InitializeAgentTables: Starting agent table initialization")

	// First, ensure schema version table exists
	if err := ensureSchemaVersionTable(db); err != nil {
		log.Printf("ERROR: Failed to create schema version table: %v", err)
		return fmt.Errorf("failed to create schema version table: %w", err)
	}

	// Get current schema version
	currentVersion, err := getSchemaVersion(db)
	if err != nil {
		log.Printf("ERROR: Failed to get schema version: %v", err)
		return fmt.Errorf("failed to get schema version: %w", err)
	}
	log.Printf("InitializeAgentTables: Current schema version: %d, Target version: %d", currentVersion, currentSchemaVersion)

	// Check if tables already exist
	exists, err := tablesExist(db)
	if err != nil {
		log.Printf("ERROR: Failed to check if tables exist: %v", err)
		return fmt.Errorf("failed to check if tables exist: %w", err)
	}

	if !exists {
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

		// Set schema version to latest after initial creation
		if err := setSchemaVersion(db, currentSchemaVersion); err != nil {
			log.Printf("ERROR: Failed to set schema version: %v", err)
			return fmt.Errorf("failed to set schema version: %w", err)
		}
	} else {
		// Tables exist, run any pending migrations
		log.Printf("InitializeAgentTables: Tables exist, checking for pending migrations")
		if err := runMigrations(db, currentVersion, currentSchemaVersion); err != nil {
			log.Printf("ERROR: Failed to run migrations: %v", err)
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	log.Printf("InitializeAgentTables: Agent tables initialized successfully (schema version: %d)", currentSchemaVersion)
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
			claude_session_id TEXT,
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

// ensureSchemaVersionTable creates the schema_version table if it doesn't exist
func ensureSchemaVersionTable(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS schema_version (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			version INTEGER NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	return nil
}

// getSchemaVersion returns the current schema version from the database
// Returns 0 if no version is set (brand new database)
func getSchemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT version FROM schema_version WHERE id = 1").Scan(&version)

	if err == sql.ErrNoRows {
		// No version record exists, check if tables exist
		exists, checkErr := tablesExist(db)
		if checkErr != nil {
			return 0, checkErr
		}

		if exists {
			// Tables exist but no version record - this is a v1 database
			return 1, nil
		}

		// No tables and no version - brand new database
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}

// setSchemaVersion updates the schema version in the database
func setSchemaVersion(db *sql.DB, version int) error {
	// Use INSERT OR REPLACE to handle both initial insert and updates
	query := `
		INSERT OR REPLACE INTO schema_version (id, version, updated_at)
		VALUES (1, ?, CURRENT_TIMESTAMP)
	`

	if _, err := db.Exec(query, version); err != nil {
		return fmt.Errorf("failed to set schema version to %d: %w", version, err)
	}

	log.Printf("Schema version updated to %d", version)
	return nil
}

// runMigrations executes all pending migrations
func runMigrations(db *sql.DB, fromVersion, toVersion int) error {
	if fromVersion >= toVersion {
		log.Printf("No migrations needed (current: %d, target: %d)", fromVersion, toVersion)
		return nil
	}

	log.Printf("Running migrations from version %d to %d", fromVersion, toVersion)

	// Run each migration in sequence
	for version := fromVersion; version < toVersion; version++ {
		nextVersion := version + 1
		log.Printf("Migrating from version %d to %d", version, nextVersion)

		switch nextVersion {
		case 2:
			if err := migrate_v1_to_v2(db); err != nil {
				return fmt.Errorf("migration v1->v2 failed: %w", err)
			}
		default:
			return fmt.Errorf("unknown migration version: %d", nextVersion)
		}

		// Update schema version after successful migration
		if err := setSchemaVersion(db, nextVersion); err != nil {
			return fmt.Errorf("failed to update schema version to %d: %w", nextVersion, err)
		}
	}

	log.Printf("Migrations completed successfully")
	return nil
}

// migrate_v1_to_v2 renames transcript_path column to claude_session_id
func migrate_v1_to_v2(db *sql.DB) error {
	log.Printf("Migration v1->v2: Renaming transcript_path to claude_session_id")

	// Check if transcript_path column exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM pragma_table_info('agent_sessions')
		WHERE name = 'transcript_path'
	`).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check for transcript_path column: %w", err)
	}

	if count == 0 {
		log.Printf("Migration v1->v2: transcript_path column not found, assuming already migrated")
		return nil
	}

	// Rename column using ALTER TABLE (SQLite 3.25.0+)
	// If this fails, we'll fall back to the copy approach
	_, err = db.Exec("ALTER TABLE agent_sessions RENAME COLUMN transcript_path TO claude_session_id")

	if err == nil {
		log.Printf("Migration v1->v2: Successfully renamed column using ALTER TABLE")
		return nil
	}

	// Fallback: SQLite version doesn't support RENAME COLUMN
	// Use the copy approach: create new table, copy data, replace
	log.Printf("Migration v1->v2: ALTER TABLE RENAME COLUMN not supported, using copy approach")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create new table with correct column name
	_, err = tx.Exec(`
		CREATE TABLE agent_sessions_new (
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
			claude_session_id TEXT,
			CONSTRAINT status_check CHECK (status IN ('idle', 'active', 'processing', 'error', 'ended'))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// Copy data from old table to new table
	_, err = tx.Exec(`
		INSERT INTO agent_sessions_new
		SELECT
			id, status, created_at, updated_at, ended_at,
			message_count, cost_usd, num_turns, duration_ms,
			error_message, model_name, transcript_path
		FROM agent_sessions
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE agent_sessions")
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Rename new table to original name
	_, err = tx.Exec("ALTER TABLE agent_sessions_new RENAME TO agent_sessions")
	if err != nil {
		return fmt.Errorf("failed to rename new table: %w", err)
	}

	// Recreate indexes
	if err := createIndexes(tx); err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Migration v1->v2: Successfully renamed column using copy approach")
	return nil
}
