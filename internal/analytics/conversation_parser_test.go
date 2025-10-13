package analytics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schlunsen/claude-control-terminal/internal/database"
)

func TestNewConversationParser(t *testing.T) {
	// Create a temporary database for testing
	tempDB := setupTestDB(t)
	defer tempDB.Close()

	repo := database.NewRepository(tempDB)
	cp := NewConversationParser(repo)

	if cp == nil {
		t.Fatal("NewConversationParser returned nil")
	}

	if cp.repo == nil {
		t.Error("repo should not be nil")
	}

	if cp.maxToolMapSize != 10000 {
		t.Errorf("expected maxToolMapSize 10000, got %d", cp.maxToolMapSize)
	}

	if cp.maxScannerBufferMB != 10 {
		t.Errorf("expected maxScannerBufferMB 10, got %d", cp.maxScannerBufferMB)
	}
}

func TestParseConversationFileNonExistent(t *testing.T) {
	tempDB := setupTestDB(t)
	defer tempDB.Close()

	repo := database.NewRepository(tempDB)
	cp := NewConversationParser(repo)

	err := cp.ParseConversationFile("/non/existent/file.jsonl")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseConversationFileEmpty(t *testing.T) {
	tempDB := setupTestDB(t)
	defer tempDB.Close()

	repo := database.NewRepository(tempDB)
	cp := NewConversationParser(repo)

	// Create empty temp file
	tempFile := createTempFile(t, "")
	defer os.Remove(tempFile)

	err := cp.ParseConversationFile(tempFile)
	if err != nil {
		t.Errorf("unexpected error parsing empty file: %v", err)
	}
}

func TestParseConversationFileWithUserMessage(t *testing.T) {
	tempDB := setupTestDB(t)
	defer tempDB.Close()

	repo := database.NewRepository(tempDB)
	cp := NewConversationParser(repo)

	// Create conversation with user message
	now := time.Now()
	msg := ConversationMessage{
		Type:      "user",
		UUID:      "test-uuid",
		Timestamp: now.Format(time.RFC3339),
		CWD:       "/test/path",
		GitBranch: "main",
		SessionID: "session-123",
	}
	msg.Message.Role = "user"
	msg.Message.Content = make([]struct {
		Type      string                 `json:"type"`
		Text      string                 `json:"text,omitempty"`
		ID        string                 `json:"id,omitempty"`
		Name      string                 `json:"name,omitempty"`
		Input     map[string]interface{} `json:"input,omitempty"`
		ToolUseID string                 `json:"tool_use_id,omitempty"`
		Content   interface{}            `json:"content,omitempty"`
	}, 1)
	msg.Message.Content[0].Type = "text"
	msg.Message.Content[0].Text = "Hello, Claude!"

	jsonData, _ := json.Marshal(msg)
	tempFile := createTempFile(t, string(jsonData))
	defer os.Remove(tempFile)

	err := cp.ParseConversationFile(tempFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseConversationFileWithToolUse(t *testing.T) {
	tempDB := setupTestDB(t)
	defer tempDB.Close()

	repo := database.NewRepository(tempDB)
	cp := NewConversationParser(repo)

	now := time.Now()

	// Create assistant message with tool_use
	assistantMsg := ConversationMessage{
		Type:      "assistant",
		UUID:      "test-uuid-1",
		Timestamp: now.Format(time.RFC3339),
		CWD:       "/test/path",
		GitBranch: "main",
		SessionID: "session-123",
	}
	assistantMsg.Message.Role = "assistant"
	assistantMsg.Message.Content = make([]struct {
		Type      string                 `json:"type"`
		Text      string                 `json:"text,omitempty"`
		ID        string                 `json:"id,omitempty"`
		Name      string                 `json:"name,omitempty"`
		Input     map[string]interface{} `json:"input,omitempty"`
		ToolUseID string                 `json:"tool_use_id,omitempty"`
		Content   interface{}            `json:"content,omitempty"`
	}, 1)
	assistantMsg.Message.Content[0].Type = "tool_use"
	assistantMsg.Message.Content[0].ID = "tool-123"
	assistantMsg.Message.Content[0].Name = "Read"
	assistantMsg.Message.Content[0].Input = map[string]interface{}{
		"file_path": "/test/file.go",
	}

	// Create user message with tool_result
	userMsg := ConversationMessage{
		Type:      "user",
		UUID:      "test-uuid-2",
		Timestamp: now.Add(1 * time.Second).Format(time.RFC3339),
		CWD:       "/test/path",
		GitBranch: "main",
		SessionID: "session-123",
	}
	userMsg.Message.Role = "user"
	userMsg.Message.Content = make([]struct {
		Type      string                 `json:"type"`
		Text      string                 `json:"text,omitempty"`
		ID        string                 `json:"id,omitempty"`
		Name      string                 `json:"name,omitempty"`
		Input     map[string]interface{} `json:"input,omitempty"`
		ToolUseID string                 `json:"tool_use_id,omitempty"`
		Content   interface{}            `json:"content,omitempty"`
	}, 1)
	userMsg.Message.Content[0].Type = "tool_result"
	userMsg.Message.Content[0].ToolUseID = "tool-123"
	userMsg.Message.Content[0].Content = "file contents"

	jsonData1, _ := json.Marshal(assistantMsg)
	jsonData2, _ := json.Marshal(userMsg)
	tempFile := createTempFile(t, string(jsonData1)+"\n"+string(jsonData2))
	defer os.Remove(tempFile)

	err := cp.ParseConversationFile(tempFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseConversationFileWithBashCommand(t *testing.T) {
	tempDB := setupTestDB(t)
	defer tempDB.Close()

	repo := database.NewRepository(tempDB)
	cp := NewConversationParser(repo)

	now := time.Now()

	// Create assistant message with Bash tool_use
	assistantMsg := ConversationMessage{
		Type:      "assistant",
		UUID:      "test-uuid-1",
		Timestamp: now.Format(time.RFC3339),
		CWD:       "/test/path",
		GitBranch: "main",
		SessionID: "session-123",
	}
	assistantMsg.Message.Role = "assistant"
	assistantMsg.Message.Content = make([]struct {
		Type      string                 `json:"type"`
		Text      string                 `json:"text,omitempty"`
		ID        string                 `json:"id,omitempty"`
		Name      string                 `json:"name,omitempty"`
		Input     map[string]interface{} `json:"input,omitempty"`
		ToolUseID string                 `json:"tool_use_id,omitempty"`
		Content   interface{}            `json:"content,omitempty"`
	}, 1)
	assistantMsg.Message.Content[0].Type = "tool_use"
	assistantMsg.Message.Content[0].ID = "bash-123"
	assistantMsg.Message.Content[0].Name = "Bash"
	assistantMsg.Message.Content[0].Input = map[string]interface{}{
		"command":     "ls -la",
		"description": "List files",
	}

	// Create user message with tool_result
	userMsg := ConversationMessage{
		Type:      "user",
		UUID:      "test-uuid-2",
		Timestamp: now.Add(1 * time.Second).Format(time.RFC3339),
		CWD:       "/test/path",
		GitBranch: "main",
		SessionID: "session-123",
	}
	userMsg.Message.Role = "user"
	userMsg.Message.Content = make([]struct {
		Type      string                 `json:"type"`
		Text      string                 `json:"text,omitempty"`
		ID        string                 `json:"id,omitempty"`
		Name      string                 `json:"name,omitempty"`
		Input     map[string]interface{} `json:"input,omitempty"`
		ToolUseID string                 `json:"tool_use_id,omitempty"`
		Content   interface{}            `json:"content,omitempty"`
	}, 1)
	userMsg.Message.Content[0].Type = "tool_result"
	userMsg.Message.Content[0].ToolUseID = "bash-123"
	userMsg.Message.Content[0].Content = "file1.txt\nfile2.txt"

	jsonData1, _ := json.Marshal(assistantMsg)
	jsonData2, _ := json.Marshal(userMsg)
	tempFile := createTempFile(t, string(jsonData1)+"\n"+string(jsonData2))
	defer os.Remove(tempFile)

	err := cp.ParseConversationFile(tempFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseConversationFileMalformedJSON(t *testing.T) {
	tempDB := setupTestDB(t)
	defer tempDB.Close()

	repo := database.NewRepository(tempDB)
	cp := NewConversationParser(repo)

	tempFile := createTempFile(t, "{invalid json}\n{\"type\": \"user\"}")
	defer os.Remove(tempFile)

	// Should not error, just skip malformed lines
	err := cp.ParseConversationFile(tempFile)
	if err != nil {
		t.Errorf("should skip malformed JSON gracefully, got error: %v", err)
	}
}

func TestToolExecutionStruct(t *testing.T) {
	now := time.Now()
	te := ToolExecution{
		ToolID:           "tool-123",
		ToolName:         "Read",
		Input:            map[string]interface{}{"file": "test.go"},
		Result:           "file contents",
		Success:          true,
		ConversationID:   "conv-123",
		WorkingDirectory: "/test/path",
		GitBranch:        "main",
		ExecutedAt:       now,
	}

	if te.ToolID != "tool-123" {
		t.Errorf("expected ToolID 'tool-123', got %q", te.ToolID)
	}

	if te.ToolName != "Read" {
		t.Errorf("expected ToolName 'Read', got %q", te.ToolName)
	}

	if !te.Success {
		t.Error("Success should be true")
	}

	if te.ConversationID != "conv-123" {
		t.Errorf("expected ConversationID 'conv-123', got %q", te.ConversationID)
	}
}

func TestConversationMessageStruct(t *testing.T) {
	now := time.Now()
	msg := ConversationMessage{
		Type:      "user",
		UUID:      "uuid-123",
		Timestamp: now.Format(time.RFC3339),
		CWD:       "/test",
		GitBranch: "main",
		SessionID: "session-123",
	}

	if msg.Type != "user" {
		t.Errorf("expected type 'user', got %q", msg.Type)
	}

	if msg.UUID != "uuid-123" {
		t.Errorf("expected UUID 'uuid-123', got %q", msg.UUID)
	}

	if msg.CWD != "/test" {
		t.Errorf("expected CWD '/test', got %q", msg.CWD)
	}

	if msg.GitBranch != "main" {
		t.Errorf("expected GitBranch 'main', got %q", msg.GitBranch)
	}

	if msg.SessionID != "session-123" {
		t.Errorf("expected SessionID 'session-123', got %q", msg.SessionID)
	}
}

// Helper functions

func setupTestDB(t *testing.T) *database.Database {
	t.Helper()

	// Reset singleton for test
	database.ResetInstance()

	// Create temp directory for test
	tempDir := t.TempDir()

	// Initialize database
	db, err := database.Initialize(tempDir)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return db
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jsonl")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return tmpFile
}
