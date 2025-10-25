package analytics

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 Bytes",
		},
		{
			name:     "bytes",
			bytes:    512,
			expected: "512.00 Bytes",
		},
		{
			name:     "kilobytes",
			bytes:    1024,
			expected: "1.00 KB",
		},
		{
			name:     "megabytes",
			bytes:    1024 * 1024,
			expected: "1.00 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1024 * 1024 * 1024,
			expected: "1.00 GB",
		},
		{
			name:     "mixed KB",
			bytes:    1536,
			expected: "1.50 KB",
		},
		{
			name:     "mixed MB",
			bytes:    1024*1024 + 512*1024,
			expected: "1.50 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestConversationAnalyzer_estimateTokens(t *testing.T) {
	ca := NewConversationAnalyzer("/tmp/test")

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "short text",
			text:     "Hello",
			expected: 1, // 5 / 4 = 1
		},
		{
			name:     "medium text",
			text:     "Hello, this is a test message",
			expected: 7, // 29 / 4 = 7
		},
		{
			name:     "100 characters",
			text:     "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			expected: 25, // 100 / 4 = 25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ca.estimateTokens(tt.text)
			if result != tt.expected {
				t.Errorf("estimateTokens(%q) = %d, want %d", tt.text, result, tt.expected)
			}
		})
	}
}

func TestConversationAnalyzer_extractProjectFromPath(t *testing.T) {
	ca := NewConversationAnalyzer("/tmp/test")

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "path with projects directory",
			filePath: filepath.Join("home", "user", "projects", "my-app", ".claude", "conversations", "test.jsonl"),
			expected: "my-app",
		},
		{
			name:     "path without projects directory",
			filePath: filepath.Join("home", "user", "code", "my-app", ".claude", "conversations", "test.jsonl"),
			expected: "conversations", // Uses parent directory
		},
		{
			name:     "root path",
			filePath: filepath.Join(string(filepath.Separator), "test.jsonl"),
			expected: string(filepath.Separator),
		},
		{
			name:     "multiple projects in path",
			filePath: filepath.Join("projects", "workspace", "projects", "my-app", ".claude", "test.jsonl"),
			expected: "workspace", // Returns first match
		},
		{
			name:     "projects at end",
			filePath: filepath.Join("home", "projects"),
			expected: "home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ca.extractProjectFromPath(tt.filePath)
			if result != tt.expected {
				t.Errorf("extractProjectFromPath(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestNewConversationAnalyzer(t *testing.T) {
	claudeDir := "/test/claude"
	ca := NewConversationAnalyzer(claudeDir)

	if ca == nil {
		t.Fatal("NewConversationAnalyzer() returned nil")
	}

	if ca.claudeDir != claudeDir {
		t.Errorf("NewConversationAnalyzer().claudeDir = %q, want %q", ca.claudeDir, claudeDir)
	}
}

func TestConversationAnalyzer_parseMessages(t *testing.T) {
	ca := NewConversationAnalyzer("/tmp/test")

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "empty content",
			content:       "",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "single valid message",
			content:       `{"timestamp":"2024-01-01T00:00:00Z","message":{"role":"user","content":"hello"}}`,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "multiple messages",
			content: `{"timestamp":"2024-01-01T00:00:00Z","message":{"role":"user","content":"hello"}}
{"timestamp":"2024-01-01T00:00:01Z","message":{"role":"assistant","content":"hi"}}`,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "invalid json line",
			content:       `{invalid json}`,
			expectedCount: 0,
			expectError:   false, // parseMessages ignores invalid lines
		},
		{
			name: "mixed valid and invalid",
			content: `{"timestamp":"2024-01-01T00:00:00Z","message":{"role":"user","content":"hello"}}
{invalid}
{"timestamp":"2024-01-01T00:00:02Z","message":{"role":"assistant","content":"hi"}}`,
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, err := ca.parseMessages(tt.content)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(messages) != tt.expectedCount {
				t.Errorf("parseMessages() returned %d messages, want %d", len(messages), tt.expectedCount)
			}
		})
	}
}
