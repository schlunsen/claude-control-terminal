package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDatabaseInitialization(t *testing.T) {
	// Reset singleton for test
	ResetInstance()

	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "cct_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize database
	db, err := Initialize(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify health
	if err := db.HealthCheck(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Verify database file exists
	dbPath := filepath.Join(tempDir, "cct_history.db")
	fileInfo, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		t.Errorf("Database file was not created")
	}

	// Verify strict permissions (0600 - user read/write only)
	if fileInfo != nil {
		mode := fileInfo.Mode()
		expectedMode := os.FileMode(0600)
		if mode.Perm() != expectedMode {
			t.Errorf("Expected file permissions %v, got %v", expectedMode, mode.Perm())
		}
	}

	// Get stats
	stats, err := db.Stats()
	if err != nil {
		t.Errorf("Failed to get stats: %v", err)
	}

	if stats["shell_commands_count"] != 0 {
		t.Errorf("Expected 0 shell commands, got %v", stats["shell_commands_count"])
	}
}

func TestShellCommandRecording(t *testing.T) {
	// Reset singleton for test
	ResetInstance()

	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "cct_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize database
	db, err := Initialize(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	// Create test command
	exitCode := 0
	duration := 150
	cmd := &ShellCommand{
		ConversationID:   "test-conv-123",
		Command:          "git status",
		Description:      "Check git status",
		WorkingDirectory: "/test/dir",
		GitBranch:        "feat/test",
		ExitCode:         &exitCode,
		Stdout:           "On branch main",
		Stderr:           "",
		DurationMs:       &duration,
		ExecutedAt:       time.Now(),
	}

	// Record command
	if err := repo.RecordShellCommand(cmd); err != nil {
		t.Fatalf("Failed to record shell command: %v", err)
	}

	// Verify command was saved
	if cmd.ID == 0 {
		t.Error("Command ID was not set")
	}

	// Query commands
	query := &CommandHistoryQuery{
		ConversationID: "test-conv-123",
		Limit:          10,
	}

	commands, err := repo.GetShellCommands(query)
	if err != nil {
		t.Fatalf("Failed to get shell commands: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(commands))
	}

	retrieved := commands[0]
	if retrieved.Command != "git status" {
		t.Errorf("Expected 'git status', got '%s'", retrieved.Command)
	}

	if retrieved.GitBranch != "feat/test" {
		t.Errorf("Expected 'feat/test', got '%s'", retrieved.GitBranch)
	}
}

func TestClaudeCommandRecording(t *testing.T) {
	// Reset singleton for test
	ResetInstance()

	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "cct_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize database
	db, err := Initialize(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	// Create test command
	duration := 500
	cmd := &ClaudeCommand{
		ConversationID:   "test-conv-456",
		ToolName:         "Read",
		Parameters:       `{"file_path": "/test/file.go"}`,
		Result:           `{"success": true}`,
		WorkingDirectory: "/test/project",
		GitBranch:        "main",
		Success:          true,
		DurationMs:       &duration,
		ExecutedAt:       time.Now(),
	}

	// Record command
	if err := repo.RecordClaudeCommand(cmd); err != nil {
		t.Fatalf("Failed to record claude command: %v", err)
	}

	// Verify command was saved
	if cmd.ID == 0 {
		t.Error("Command ID was not set")
	}

	// Query commands
	query := &CommandHistoryQuery{
		ConversationID: "test-conv-456",
		Limit:          10,
	}

	commands, err := repo.GetClaudeCommands(query)
	if err != nil {
		t.Fatalf("Failed to get claude commands: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(commands))
	}

	retrieved := commands[0]
	if retrieved.ToolName != "Read" {
		t.Errorf("Expected 'Read', got '%s'", retrieved.ToolName)
	}

	if retrieved.GitBranch != "main" {
		t.Errorf("Expected 'main', got '%s'", retrieved.GitBranch)
	}
}
