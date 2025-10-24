// Package database provides SQLite database management for command history and conversation tracking.
// It implements a singleton pattern for database access and handles schema migrations,
// connection pooling, and WAL mode for concurrent access.
package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// Database represents the SQLite database connection
type Database struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

var (
	instance *Database
	once     sync.Once
	mu       sync.Mutex
)

// Initialize creates and initializes the database
func Initialize(dataDir string) (*Database, error) {
	mu.Lock()
	defer mu.Unlock()

	// For testing: allow re-initialization if instance is nil
	if instance != nil {
		return instance, nil
	}

	var initErr error

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "cct.db")

	// Check if database file exists before opening
	dbExists := false
	if _, err := os.Stat(dbPath); err == nil {
		dbExists = true
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping to force database file creation
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set strict permissions on database file (0600 - user read/write only)
	// This ensures sensitive command history is only readable by the user
	if !dbExists {
		if err := os.Chmod(dbPath, 0600); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set database file permissions: %w", err)
		}
	}

	// Enable foreign keys and set pragmas for performance
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -64000", // 64MB cache
		"PRAGMA temp_store = MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			initErr = fmt.Errorf("failed to set pragma: %w", err)
			db.Close()
			return nil, initErr
		}
	}

	// Run schema migrations
	if _, err := db.Exec(schemaSQL); err != nil {
		initErr = fmt.Errorf("failed to execute schema: %w", err)
		db.Close()
		return nil, initErr
	}

	// Run additional migrations for existing databases
	if err := runMigrations(db); err != nil {
		initErr = fmt.Errorf("failed to run migrations: %w", err)
		db.Close()
		return nil, initErr
	}

	// Set permissions on WAL and SHM files created by SQLite (WAL mode)
	// These files may be created after opening the database
	walPath := dbPath + "-wal"
	shmPath := dbPath + "-shm"

	if _, err := os.Stat(walPath); err == nil {
		os.Chmod(walPath, 0600)
	}
	if _, err := os.Stat(shmPath); err == nil {
		os.Chmod(shmPath, 0600)
	}

	instance = &Database{
		db:   db,
		path: dbPath,
	}

	return instance, nil
}

// GetInstance returns the singleton database instance
func GetInstance() *Database {
	return instance
}

// Close closes the database connection
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// ResetInstance resets the singleton instance (for testing only)
func ResetInstance() {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		instance.Close()
		instance = nil
	}
	once = sync.Once{}
}

// GetDB returns the underlying sql.DB for direct access
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// Path returns the database file path
func (d *Database) Path() string {
	return d.path
}

// HealthCheck verifies database connectivity
func (d *Database) HealthCheck() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.db == nil {
		return fmt.Errorf("database not initialized")
	}

	return d.db.Ping()
}

// Stats returns database statistics
func (d *Database) Stats() (map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := make(map[string]interface{})

	// Get table counts
	tables := []string{"shell_commands", "claude_commands", "conversations", "command_stats"}
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := d.db.QueryRow(query).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to get count for %s: %w", table, err)
		}
		stats[table+"_count"] = count
	}

	// Get database file size
	if fileInfo, err := os.Stat(d.path); err == nil {
		stats["db_size_bytes"] = fileInfo.Size()
	}

	return stats, nil
}

// Vacuum reclaims unused disk space by rebuilding the database file
func (d *Database) Vacuum() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// First, run VACUUM to rebuild the database
	_, err := d.db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	// Then checkpoint the WAL to ensure all data is written to main DB file
	_, err = d.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		return fmt.Errorf("failed to checkpoint WAL after vacuum: %w", err)
	}

	return nil
}

// runMigrations runs database migrations for existing databases
func runMigrations(db *sql.DB) error {
	// Migration 1: Add model_name column to providers table if it doesn't exist
	var columnExists bool
	query := `
		SELECT COUNT(*) > 0
		FROM pragma_table_info('providers')
		WHERE name='model_name'
	`
	if err := db.QueryRow(query).Scan(&columnExists); err != nil {
		// If the providers table doesn't exist yet, that's fine - it will be created by schema.sql
		return nil
	}

	if !columnExists {
		// Add the model_name column
		_, err := db.Exec("ALTER TABLE providers ADD COLUMN model_name TEXT")
		if err != nil {
			return fmt.Errorf("failed to add model_name column: %w", err)
		}
	}

	// Migration 2: Add session_name column to user_messages table if it doesn't exist
	var sessionNameExists bool
	sessionQuery := `
		SELECT COUNT(*) > 0
		FROM pragma_table_info('user_messages')
		WHERE name='session_name'
	`
	if err := db.QueryRow(sessionQuery).Scan(&sessionNameExists); err != nil {
		// If the user_messages table doesn't exist yet, that's fine - it will be created by schema.sql
		return nil
	}

	if !sessionNameExists {
		// Add the session_name column
		_, err := db.Exec("ALTER TABLE user_messages ADD COLUMN session_name TEXT")
		if err != nil {
			return fmt.Errorf("failed to add session_name column to user_messages: %w", err)
		}
	}

	// Migration 3: Add session_name column to shell_commands table if it doesn't exist
	var shellSessionNameExists bool
	shellSessionQuery := `
		SELECT COUNT(*) > 0
		FROM pragma_table_info('shell_commands')
		WHERE name='session_name'
	`
	if err := db.QueryRow(shellSessionQuery).Scan(&shellSessionNameExists); err == nil {
		if !shellSessionNameExists {
			_, err := db.Exec("ALTER TABLE shell_commands ADD COLUMN session_name TEXT")
			if err != nil {
				return fmt.Errorf("failed to add session_name column to shell_commands: %w", err)
			}
		}
	}

	// Migration 4: Add session_name column to claude_commands table if it doesn't exist
	var claudeSessionNameExists bool
	claudeSessionQuery := `
		SELECT COUNT(*) > 0
		FROM pragma_table_info('claude_commands')
		WHERE name='session_name'
	`
	if err := db.QueryRow(claudeSessionQuery).Scan(&claudeSessionNameExists); err == nil {
		if !claudeSessionNameExists {
			_, err := db.Exec("ALTER TABLE claude_commands ADD COLUMN session_name TEXT")
			if err != nil {
				return fmt.Errorf("failed to add session_name column to claude_commands: %w", err)
			}
		}
	}

	// Migration 5: Create notifications table if it doesn't exist
	var notificationsTableExists bool
	notificationsQuery := `
		SELECT COUNT(*) > 0
		FROM sqlite_master
		WHERE type='table' AND name='notifications'
	`
	if err := db.QueryRow(notificationsQuery).Scan(&notificationsTableExists); err == nil {
		if !notificationsTableExists {
			createNotificationsTable := `
				CREATE TABLE IF NOT EXISTS notifications (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					conversation_id TEXT NOT NULL,
					session_name TEXT,
					notification_type TEXT NOT NULL,
					message TEXT NOT NULL,
					tool_name TEXT,
					command_details TEXT,
					working_directory TEXT,
					git_branch TEXT,
					notified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_notifications_conversation
					ON notifications(conversation_id, notified_at DESC);

				CREATE INDEX IF NOT EXISTS idx_notifications_notified_at
					ON notifications(notified_at DESC);

				CREATE INDEX IF NOT EXISTS idx_notifications_type
					ON notifications(notification_type, notified_at DESC);

				CREATE INDEX IF NOT EXISTS idx_notifications_tool
					ON notifications(tool_name, notified_at DESC) WHERE tool_name IS NOT NULL;
			`
			_, err := db.Exec(createNotificationsTable)
			if err != nil {
				return fmt.Errorf("failed to create notifications table: %w", err)
			}
		}
	}

	// Migration 6: Add command_details column to notifications table if it doesn't exist
	var commandDetailsExists bool
	commandDetailsQuery := `
		SELECT COUNT(*) > 0
		FROM pragma_table_info('notifications')
		WHERE name='command_details'
	`
	if err := db.QueryRow(commandDetailsQuery).Scan(&commandDetailsExists); err == nil {
		if !commandDetailsExists {
			_, err := db.Exec("ALTER TABLE notifications ADD COLUMN command_details TEXT")
			if err != nil {
				return fmt.Errorf("failed to add command_details column to notifications: %w", err)
			}
		}
	}

	// Migration 7: Add model_provider and model_name columns to all tables
	tables := []string{"shell_commands", "claude_commands", "conversations", "user_messages", "notifications"}
	for _, table := range tables {
		// Check if model_provider exists
		var modelProviderExists bool
		providerQuery := fmt.Sprintf(`
			SELECT COUNT(*) > 0
			FROM pragma_table_info('%s')
			WHERE name='model_provider'
		`, table)
		if err := db.QueryRow(providerQuery).Scan(&modelProviderExists); err == nil {
			if !modelProviderExists {
				alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN model_provider TEXT", table)
				if _, err := db.Exec(alterQuery); err != nil {
					return fmt.Errorf("failed to add model_provider column to %s: %w", table, err)
				}
			}
		}

		// Check if model_name exists
		var modelNameExists bool
		nameQuery := fmt.Sprintf(`
			SELECT COUNT(*) > 0
			FROM pragma_table_info('%s')
			WHERE name='model_name'
		`, table)
		if err := db.QueryRow(nameQuery).Scan(&modelNameExists); err == nil {
			if !modelNameExists {
				alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN model_name TEXT", table)
				if _, err := db.Exec(alterQuery); err != nil {
					return fmt.Errorf("failed to add model_name column to %s: %w", table, err)
				}
			}
		}
	}

	// Create indexes for model columns if they don't exist
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_shell_commands_model ON shell_commands(model_provider, model_name)",
		"CREATE INDEX IF NOT EXISTS idx_claude_commands_model ON claude_commands(model_provider, model_name)",
		"CREATE INDEX IF NOT EXISTS idx_user_messages_model ON user_messages(model_provider, model_name)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_model ON notifications(model_provider, model_name)",
		"CREATE INDEX IF NOT EXISTS idx_conversations_model ON conversations(model_provider, model_name)",
	}
	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
