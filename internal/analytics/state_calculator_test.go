package analytics

import (
	"testing"
	"time"
)

func TestNewStateCalculator(t *testing.T) {
	sc := NewStateCalculator()

	if sc == nil {
		t.Fatal("NewStateCalculator returned nil")
	}

	if sc.processCache == nil {
		t.Error("processCache should be initialized")
	}
}

func TestDetermineConversationState(t *testing.T) {
	sc := NewStateCalculator()
	now := time.Now()

	tests := []struct {
		name            string
		messages        []Message
		lastModified    time.Time
		runningProcess  *RunningProcess
		expectedState   string
	}{
		{
			name:          "no messages, recent file modification",
			messages:      []Message{},
			lastModified:  now.Add(-2 * time.Minute),
			runningProcess: nil,
			expectedState: "Claude Code working...",
		},
		{
			name:          "no messages, old file modification",
			messages:      []Message{},
			lastModified:  now.Add(-10 * time.Minute),
			runningProcess: nil,
			expectedState: "Idle",
		},
		{
			name: "recent user message",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-1 * time.Minute),
					Content:   "test message",
				},
			},
			lastModified:  now.Add(-1 * time.Minute),
			runningProcess: nil,
			expectedState: "Claude Code working...",
		},
		{
			name: "recent assistant message",
			messages: []Message{
				{
					Role:      "assistant",
					Timestamp: now.Add(-5 * time.Minute),
					Content:   "test response",
				},
			},
			lastModified:  now.Add(-5 * time.Minute),
			runningProcess: nil,
			expectedState: "Awaiting user input...",
		},
		{
			name: "user message with active process",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-2 * time.Minute),
					Content:   "test message",
				},
			},
			lastModified: now.Add(-2 * time.Minute),
			runningProcess: &RunningProcess{
				PID:              "12345",
				StartTime:        now.Add(-10 * time.Minute),
				WorkingDir:       "/test",
				HasActiveCommand: true,
			},
			expectedState: "Claude Code working...",
		},
		{
			name: "old user message",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-15 * time.Minute),
					Content:   "test message",
				},
			},
			lastModified:  now.Add(-15 * time.Minute),
			runningProcess: nil,
			expectedState: "User typing...",
		},
		{
			name: "very old message",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-45 * time.Minute),
					Content:   "test message",
				},
			},
			lastModified:  now.Add(-45 * time.Minute),
			runningProcess: nil,
			expectedState: "Recently active",
		},
		{
			name: "very old conversation",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-150 * time.Minute),
					Content:   "test message",
				},
			},
			lastModified:  now.Add(-150 * time.Minute),
			runningProcess: nil,
			expectedState: "Recently active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := sc.DetermineConversationState(tt.messages, tt.lastModified, tt.runningProcess)
			if state != tt.expectedState {
				t.Errorf("expected state %q, got %q", tt.expectedState, state)
			}
		})
	}
}

func TestDetermineConversationStatus(t *testing.T) {
	sc := NewStateCalculator()
	now := time.Now()

	tests := []struct {
		name           string
		messages       []Message
		lastModified   time.Time
		expectedStatus string
	}{
		{
			name:           "no messages, recent file",
			messages:       []Message{},
			lastModified:   now.Add(-2 * time.Minute),
			expectedStatus: "active",
		},
		{
			name:           "no messages, old file",
			messages:       []Message{},
			lastModified:   now.Add(-10 * time.Minute),
			expectedStatus: "inactive",
		},
		{
			name: "recent user message",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-1 * time.Minute),
					Content:   "test",
				},
			},
			lastModified:   now.Add(-1 * time.Minute),
			expectedStatus: "active",
		},
		{
			name: "recent assistant message",
			messages: []Message{
				{
					Role:      "assistant",
					Timestamp: now.Add(-3 * time.Minute),
					Content:   "test",
				},
			},
			lastModified:   now.Add(-3 * time.Minute),
			expectedStatus: "active",
		},
		{
			name: "recent file modification",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-10 * time.Minute),
					Content:   "test",
				},
			},
			lastModified:   now.Add(-3 * time.Minute),
			expectedStatus: "active",
		},
		{
			name: "recent conversation",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-20 * time.Minute),
					Content:   "test",
				},
			},
			lastModified:   now.Add(-20 * time.Minute),
			expectedStatus: "recent",
		},
		{
			name: "inactive conversation",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-60 * time.Minute),
					Content:   "test",
				},
			},
			lastModified:   now.Add(-60 * time.Minute),
			expectedStatus: "inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := sc.DetermineConversationStatus(tt.messages, tt.lastModified)
			if status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, status)
			}
		})
	}
}

func TestQuickStateCalculation(t *testing.T) {
	sc := NewStateCalculator()
	now := time.Now()

	tests := []struct {
		name             string
		lastModified     time.Time
		hasActiveProcess bool
		expectedState    string
	}{
		{
			name:             "no active process",
			lastModified:     now,
			hasActiveProcess: false,
			expectedState:    "",
		},
		{
			name:             "very recent activity",
			lastModified:     now.Add(-10 * time.Second),
			hasActiveProcess: true,
			expectedState:    "Claude Code working...",
		},
		{
			name:             "recent activity",
			lastModified:     now.Add(-2 * time.Minute),
			hasActiveProcess: true,
			expectedState:    "Awaiting user input...",
		},
		{
			name:             "older activity",
			lastModified:     now.Add(-10 * time.Minute),
			hasActiveProcess: true,
			expectedState:    "User typing...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := sc.QuickStateCalculation(tt.lastModified, tt.hasActiveProcess)
			if state != tt.expectedState {
				t.Errorf("expected state %q, got %q", tt.expectedState, state)
			}
		})
	}
}

func TestDetectRealClaudeActivity(t *testing.T) {
	sc := NewStateCalculator()
	now := time.Now()

	tests := []struct {
		name           string
		messages       []Message
		lastModified   time.Time
		expectActive   bool
	}{
		{
			name:         "no messages",
			messages:     nil,
			lastModified: now,
			expectActive: false,
		},
		{
			name:         "empty messages",
			messages:     []Message{},
			lastModified: now,
			expectActive: false,
		},
		{
			name:         "very recent file modification",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-2 * time.Minute),
					Content:   "test",
				},
			},
			lastModified: now.Add(-30 * time.Second),
			expectActive: true,
		},
		{
			name: "recent user message with recent file",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-3 * time.Minute),
					Content:   "test",
				},
			},
			lastModified: now.Add(-5 * time.Minute),
			expectActive: true,
		},
		{
			name: "messages with tool results",
			messages: []Message{
				{
					Role:      "assistant",
					Timestamp: now.Add(-5 * time.Minute),
					Content:   "test",
					ToolResults: []interface{}{
						map[string]interface{}{"result": "success"},
					},
				},
			},
			lastModified: now.Add(-8 * time.Minute),
			expectActive: true,
		},
		{
			name: "rapid message exchange",
			messages: []Message{
				{
					Role:      "user",
					Timestamp: now.Add(-15 * time.Minute),
					Content:   "test",
				},
				{
					Role:      "assistant",
					Timestamp: now.Add(-12 * time.Minute),
					Content:   "response",
				},
			},
			lastModified: now.Add(-11 * time.Minute),
			expectActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sc.detectRealClaudeActivity(tt.messages, tt.lastModified)
			if result.IsActive != tt.expectActive {
				t.Errorf("expected IsActive=%v, got %v (status: %s)", tt.expectActive, result.IsActive, result.Status)
			}
		})
	}
}

func TestGetStateClass(t *testing.T) {
	sc := NewStateCalculator()

	tests := []struct {
		state         string
		expectedClass string
	}{
		{"Claude Code working...", "working"},
		{"User typing...", "typing"},
		{"Awaiting response...", ""},
		{"Active session", ""},
		{"Idle", ""},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			class := sc.GetStateClass(tt.state)
			if class != tt.expectedClass {
				t.Errorf("expected class %q, got %q", tt.expectedClass, class)
			}
		})
	}
}

func TestClearCache(t *testing.T) {
	sc := NewStateCalculator()

	// Add some cache data
	sc.processCache["test"] = "value"

	// Clear cache
	sc.ClearCache()

	// Verify cache is empty
	if len(sc.processCache) != 0 {
		t.Errorf("expected empty cache, got %d items", len(sc.processCache))
	}
}

func TestSortMessagesByTimestamp(t *testing.T) {
	now := time.Now()
	messages := []Message{
		{
			Role:      "assistant",
			Timestamp: now.Add(-1 * time.Hour),
			Content:   "third",
		},
		{
			Role:      "user",
			Timestamp: now,
			Content:   "first",
		},
		{
			Role:      "assistant",
			Timestamp: now.Add(-30 * time.Minute),
			Content:   "second",
		},
	}

	sorted := sortMessagesByTimestamp(messages)

	// Check that messages are sorted
	if len(sorted) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(sorted))
	}

	if sorted[0].Content != "third" {
		t.Errorf("expected first message to be 'third', got %v", sorted[0].Content)
	}

	if sorted[1].Content != "second" {
		t.Errorf("expected second message to be 'second', got %v", sorted[1].Content)
	}

	if sorted[2].Content != "first" {
		t.Errorf("expected third message to be 'first', got %v", sorted[2].Content)
	}

	// Verify original slice is not modified
	if messages[0].Content != "third" {
		t.Error("original slice was modified")
	}
}

func TestSortMessagesByTimestampEmptySlice(t *testing.T) {
	messages := []Message{}
	sorted := sortMessagesByTimestamp(messages)

	if len(sorted) != 0 {
		t.Errorf("expected empty slice, got %d messages", len(sorted))
	}
}

func TestActivityDetectionStruct(t *testing.T) {
	ad := ActivityDetection{
		IsActive: true,
		Status:   "Working",
	}

	if !ad.IsActive {
		t.Error("IsActive should be true")
	}

	if ad.Status != "Working" {
		t.Errorf("expected status 'Working', got %q", ad.Status)
	}
}

func TestRunningProcessStruct(t *testing.T) {
	now := time.Now()
	rp := RunningProcess{
		PID:              "12345",
		StartTime:        now,
		WorkingDir:       "/test/dir",
		HasActiveCommand: true,
	}

	if rp.PID != "12345" {
		t.Errorf("expected PID '12345', got %q", rp.PID)
	}

	if !rp.HasActiveCommand {
		t.Error("HasActiveCommand should be true")
	}

	if rp.WorkingDir != "/test/dir" {
		t.Errorf("expected WorkingDir '/test/dir', got %q", rp.WorkingDir)
	}
}

func TestMessageStruct(t *testing.T) {
	now := time.Now()
	msg := Message{
		Role:      "user",
		Timestamp: now,
		Content:   "test content",
		ToolResults: []interface{}{
			map[string]interface{}{"key": "value"},
		},
	}

	if msg.Role != "user" {
		t.Errorf("expected role 'user', got %q", msg.Role)
	}

	if msg.Content != "test content" {
		t.Errorf("expected content 'test content', got %v", msg.Content)
	}

	if len(msg.ToolResults) != 1 {
		t.Errorf("expected 1 tool result, got %d", len(msg.ToolResults))
	}
}
