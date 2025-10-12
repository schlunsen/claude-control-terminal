package analytics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/schlunsen/claude-control-terminal/internal/database"
)

// ConversationMessage represents a message in the conversation JSONL
type ConversationMessage struct {
	Type      string    `json:"type"`
	UUID      string    `json:"uuid"`
	Timestamp string    `json:"timestamp"`
	CWD       string    `json:"cwd"`
	GitBranch string    `json:"gitBranch"`
	SessionID string    `json:"sessionId"`
	Message   struct {
		Role    string `json:"role"`
		Content []struct {
			Type     string                 `json:"type"`
			Text     string                 `json:"text,omitempty"`
			ID       string                 `json:"id,omitempty"`
			Name     string                 `json:"name,omitempty"`
			Input    map[string]interface{} `json:"input,omitempty"`
			ToolUseID string                `json:"tool_use_id,omitempty"`
			Content  interface{}            `json:"content,omitempty"`
		} `json:"content"`
	} `json:"message"`
}

// ToolExecution represents a complete tool execution (use + result)
type ToolExecution struct {
	ToolID           string
	ToolName         string
	Input            map[string]interface{}
	Result           string
	Success          bool
	ConversationID   string
	WorkingDirectory string
	GitBranch        string
	ExecutedAt       time.Time
}

// ConversationParser parses Claude Code conversation files for tool usage
type ConversationParser struct {
	repo *database.Repository
}

// NewConversationParser creates a new conversation parser
func NewConversationParser(repo *database.Repository) *ConversationParser {
	return &ConversationParser{
		repo: repo,
	}
}

// ParseConversationFile parses a conversation file and records tool usage
func (cp *ConversationParser) ParseConversationFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	toolMap := make(map[string]*ToolExecution)
	scanner := bufio.NewScanner(file)

	// Increase scanner buffer size for large messages
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg ConversationMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // Skip malformed messages
		}

		// Process assistant messages for tool_use
		if msg.Type == "assistant" && msg.Message.Role == "assistant" {
			for _, content := range msg.Message.Content {
				if content.Type == "tool_use" && content.ID != "" && content.Name != "" {
					timestamp, _ := time.Parse(time.RFC3339, msg.Timestamp)
					toolMap[content.ID] = &ToolExecution{
						ToolID:           content.ID,
						ToolName:         content.Name,
						Input:            content.Input,
						ConversationID:   msg.SessionID,
						WorkingDirectory: msg.CWD,
						GitBranch:        msg.GitBranch,
						ExecutedAt:       timestamp,
					}
				}
			}
		}

		// Process user messages for tool_result
		if msg.Type == "user" && msg.Message.Role == "user" {
			for _, content := range msg.Message.Content {
				if content.Type == "tool_result" && content.ToolUseID != "" {
					if tool, exists := toolMap[content.ToolUseID]; exists {
						// Convert result content to string
						result := ""
						switch v := content.Content.(type) {
						case string:
							result = v
						case map[string]interface{}:
							if jsonBytes, err := json.Marshal(v); err == nil {
								result = string(jsonBytes)
							}
						}

						tool.Result = result
						tool.Success = !strings.Contains(strings.ToLower(result), "error") &&
							!strings.Contains(strings.ToLower(result), "failed")

						// Record the tool execution
						if err := cp.recordToolExecution(tool); err != nil {
							// Log error but continue processing
							fmt.Printf("Warning: failed to record tool execution: %v\n", err)
						}

						// Remove from map to free memory
						delete(toolMap, content.ToolUseID)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// recordToolExecution records a tool execution to the database
func (cp *ConversationParser) recordToolExecution(tool *ToolExecution) error {
	// Handle Bash commands specially
	if tool.ToolName == "Bash" {
		return cp.recordShellCommand(tool)
	}

	// Record other Claude commands
	return cp.recordClaudeCommand(tool)
}

// recordShellCommand records a bash command
func (cp *ConversationParser) recordShellCommand(tool *ToolExecution) error {
	command, _ := tool.Input["command"].(string)
	if command == "" {
		return nil // Skip if no command
	}

	description, _ := tool.Input["description"].(string)

	// Parse exit code from result if available
	var exitCode *int
	// TODO: Parse exit code from result if format is known

	// Estimate duration (not available in conversation format)
	var duration *int

	cmd := &database.ShellCommand{
		ConversationID:   tool.ConversationID,
		Command:          command,
		Description:      description,
		WorkingDirectory: tool.WorkingDirectory,
		GitBranch:        tool.GitBranch,
		ExitCode:         exitCode,
		Stdout:           tool.Result,
		Stderr:           "",
		DurationMs:       duration,
		ExecutedAt:       tool.ExecutedAt,
	}

	return cp.repo.RecordShellCommand(cmd)
}

// recordClaudeCommand records a Claude Code tool invocation
func (cp *ConversationParser) recordClaudeCommand(tool *ToolExecution) error {
	// Convert input to JSON string
	inputJSON, err := json.Marshal(tool.Input)
	if err != nil {
		inputJSON = []byte("{}")
	}

	// Estimate duration (not available)
	var duration *int

	cmd := &database.ClaudeCommand{
		ConversationID:   tool.ConversationID,
		ToolName:         tool.ToolName,
		Parameters:       string(inputJSON),
		Result:           tool.Result,
		WorkingDirectory: tool.WorkingDirectory,
		GitBranch:        tool.GitBranch,
		Success:          tool.Success,
		ErrorMessage:     "",
		DurationMs:       duration,
		ExecutedAt:       tool.ExecutedAt,
	}

	if !tool.Success {
		cmd.ErrorMessage = tool.Result
	}

	return cp.repo.RecordClaudeCommand(cmd)
}

// ParseAllConversations parses all conversation files in a directory
func (cp *ConversationParser) ParseAllConversations(claudeDir string) (int, error) {
	count := 0

	err := walkConversations(claudeDir, func(path string) error {
		if err := cp.ParseConversationFile(path); err != nil {
			// Log but don't fail on individual file errors
			fmt.Printf("Warning: failed to parse %s: %v\n", path, err)
		} else {
			count++
		}
		return nil
	})

	return count, err
}

// walkConversations walks all conversation files in the Claude directory
func walkConversations(claudeDir string, fn func(string) error) error {
	projectsDir := fmt.Sprintf("%s/projects", claudeDir)

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return fmt.Errorf("failed to read projects directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := fmt.Sprintf("%s/%s", projectsDir, entry.Name())
		files, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".jsonl") {
				filePath := fmt.Sprintf("%s/%s", projectPath, file.Name())
				if err := fn(filePath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
