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

	dbPath := filepath.Join(dataDir, "cct_history.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
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
