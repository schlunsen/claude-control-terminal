package analytics

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Conversation represents a Claude Code conversation
type Conversation struct {
	ID               string    `json:"id"`
	Filename         string    `json:"filename"`
	FilePath         string    `json:"filePath"`
	MessageCount     int       `json:"messageCount"`
	FileSize         int64     `json:"fileSize"`
	LastModified     time.Time `json:"lastModified"`
	Created          time.Time `json:"created"`
	Tokens           int       `json:"tokens"`
	Project          string    `json:"project"`
	Status           string    `json:"status"`
	ConversationState string   `json:"conversationState"`
}

// ConversationAnalyzer handles conversation data loading and analysis
type ConversationAnalyzer struct {
	claudeDir string
}

// NewConversationAnalyzer creates a new ConversationAnalyzer
func NewConversationAnalyzer(claudeDir string) *ConversationAnalyzer {
	return &ConversationAnalyzer{
		claudeDir: claudeDir,
	}
}

// LoadConversations loads and parses all conversation files
func (ca *ConversationAnalyzer) LoadConversations(stateCalc *StateCalculator) ([]Conversation, error) {
	var conversations []Conversation

	// Find all .jsonl files recursively
	err := filepath.WalkDir(ca.claudeDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
			conv, err := ca.parseConversationFile(path, stateCalc)
			if err == nil {
				conversations = append(conversations, conv)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by last modified (newest first)
	for i := 0; i < len(conversations); i++ {
		for j := i + 1; j < len(conversations); j++ {
			if conversations[i].LastModified.Before(conversations[j].LastModified) {
				conversations[i], conversations[j] = conversations[j], conversations[i]
			}
		}
	}

	return conversations, nil
}

// parseConversationFile parses a single conversation file
func (ca *ConversationAnalyzer) parseConversationFile(filePath string, stateCalc *StateCalculator) (Conversation, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return Conversation{}, err
	}

	// Read and parse messages
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Conversation{}, err
	}

	messages, err := ca.parseMessages(string(content))
	if err != nil {
		messages = []Message{} // Continue with empty messages
	}

	// Extract project name
	project := ca.extractProjectFromPath(filePath)
	if project == "" {
		project = "Unknown"
	}

	// Calculate token estimate
	tokens := ca.estimateTokens(string(content))

	// Determine status and state
	status := stateCalc.DetermineConversationStatus(messages, info.ModTime())
	state := stateCalc.DetermineConversationState(messages, info.ModTime(), nil)

	conv := Conversation{
		ID:                filepath.Base(filePath[:len(filePath)-6]), // Remove .jsonl
		Filename:          filepath.Base(filePath),
		FilePath:          filePath,
		MessageCount:      len(messages),
		FileSize:          info.Size(),
		LastModified:      info.ModTime(),
		Created:           info.ModTime(), // Approximation
		Tokens:            tokens,
		Project:           project,
		Status:            status,
		ConversationState: state,
	}

	return conv, nil
}

// parseMessages parses JSONL messages
func (ca *ConversationAnalyzer) parseMessages(content string) ([]Message, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	messages := []Message{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		// Extract message data
		var msg Message
		if timestamp, ok := raw["timestamp"].(string); ok {
			msg.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		}

		if message, ok := raw["message"].(map[string]interface{}); ok {
			if role, ok := message["role"].(string); ok {
				msg.Role = role
			}
			msg.Content = message["content"]
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// extractProjectFromPath extracts project name from file path
func (ca *ConversationAnalyzer) extractProjectFromPath(filePath string) string {
	// Try to extract from path structure
	parts := strings.Split(filePath, string(filepath.Separator))

	// Look for "projects" directory
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	// Fallback: use parent directory name
	dir := filepath.Dir(filePath)
	return filepath.Base(dir)
}

// estimateTokens provides rough token estimation
func (ca *ConversationAnalyzer) estimateTokens(text string) int {
	// Rough estimation: 4 characters per token
	return len(text) / 4
}

// FormatBytes formats byte size for display
func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 Bytes"
	}

	k := int64(1024)
	sizes := []string{"Bytes", "KB", "MB", "GB"}

	i := 0
	b := float64(bytes)
	for b >= float64(k) && i < len(sizes)-1 {
		b /= float64(k)
		i++
	}

	return fmt.Sprintf("%.2f %s", b, sizes[i])
}
