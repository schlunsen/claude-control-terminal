// Package analytics provides real-time monitoring and analysis of Claude Code conversations.
// It includes components for conversation state calculation, process detection, file watching,
// and conversation parsing from JSONL files.
package analytics

import (
	"sort"
	"strings"
	"time"
)

// Message represents a conversation message
type Message struct {
	Role      string                 `json:"role"`
	Timestamp time.Time              `json:"timestamp"`
	Content   interface{}            `json:"content"`
	ToolResults []interface{}        `json:"toolResults,omitempty"`
}

// isToolUse checks if a message is actually a tool use by Claude (not a real user message)
func (m *Message) isToolUse() bool {
	// Tool uses come through as "user" role messages but contain tool_result content
	if m.Role != "user" {
		return false
	}

	// Check if content is an array with tool_result type
	if contentArray, ok := m.Content.([]interface{}); ok {
		for _, item := range contentArray {
			if contentMap, ok := item.(map[string]interface{}); ok {
				if msgType, ok := contentMap["type"].(string); ok {
					if msgType == "tool_result" || msgType == "tool_use" {
						return true
					}
				}
			}
		}
	}

	return false
}

// RunningProcess represents an active Claude process
type RunningProcess struct {
	PID              string
	StartTime        time.Time
	WorkingDir       string
	HasActiveCommand bool
}

// StateCalculator handles conversation state determination logic
type StateCalculator struct {
	processCache map[string]interface{}
}

// NewStateCalculator creates a new StateCalculator
func NewStateCalculator() *StateCalculator {
	return &StateCalculator{
		processCache: make(map[string]interface{}),
	}
}

// DetermineConversationState determines the current state of a conversation
func (sc *StateCalculator) DetermineConversationState(messages []Message, lastModified time.Time, runningProcess *RunningProcess) string {
	now := time.Now()
	fileTimeDiff := now.Sub(lastModified)
	fileMinutesAgo := fileTimeDiff.Minutes()

	// Enhanced detection: Look for real Claude Code activity indicators
	claudeActivity := sc.detectRealClaudeActivity(messages, lastModified)
	if claudeActivity.IsActive {
		return claudeActivity.Status
	}

	// If there's very recent file activity (within 5 minutes), consider it active
	if fileMinutesAgo < 5 {
		return "Claude Code working..."
	}

	// If there's an active process, prioritize that
	if runningProcess != nil && runningProcess.HasActiveCommand {
		if len(messages) > 0 {
			sortedMessages := sortMessagesByTimestamp(messages)
			lastMessage := sortedMessages[len(sortedMessages)-1]
			lastMessageMinutesAgo := now.Sub(lastMessage.Timestamp).Minutes()

			if lastMessage.Role == "user" {
				// User sent message
				if lastMessageMinutesAgo < 3 {
					return "Claude Code working..."
				} else if lastMessageMinutesAgo < 10 {
					return "Awaiting response..."
				} else {
					return "Active session"
				}
			} else if lastMessage.Role == "assistant" {
				// Claude responded
				if lastMessageMinutesAgo < 10 {
					return "Awaiting user input..."
				} else {
					return "Active session"
				}
			}
		}

		return "Active session"
	}

	if len(messages) == 0 {
		if fileMinutesAgo < 5 {
			return "Waiting for input..."
		}
		return "Idle"
	}

	// Sort messages by timestamp
	sortedMessages := sortMessagesByTimestamp(messages)
	lastMessage := sortedMessages[len(sortedMessages)-1]
	lastMessageMinutesAgo := now.Sub(lastMessage.Timestamp).Minutes()

	// More generous logic for active conversations
	if lastMessage.Role == "user" {
		if lastMessageMinutesAgo < 3 {
			return "Claude Code working..."
		} else if lastMessageMinutesAgo < 10 {
			return "Awaiting response..."
		} else if lastMessageMinutesAgo < 30 {
			return "User typing..."
		} else {
			return "Recently active"
		}
	} else if lastMessage.Role == "assistant" {
		if lastMessageMinutesAgo < 10 {
			return "Awaiting user input..."
		} else if lastMessageMinutesAgo < 30 {
			return "User typing..."
		} else {
			return "Recently active"
		}
	}

	// Fallback states
	if fileMinutesAgo < 10 || lastMessageMinutesAgo < 30 {
		return "Recently active"
	}
	if fileMinutesAgo < 60 || lastMessageMinutesAgo < 120 {
		return "Idle"
	}
	return "Inactive"
}

// DetermineConversationStatus determines conversation status (active/recent/inactive)
func (sc *StateCalculator) DetermineConversationStatus(messages []Message, lastModified time.Time) string {
	now := time.Now()
	minutesAgo := now.Sub(lastModified).Minutes()

	if len(messages) == 0 {
		if minutesAgo < 5 {
			return "active"
		}
		return "inactive"
	}

	// Sort messages by timestamp
	sortedMessages := sortMessagesByTimestamp(messages)
	lastMessage := sortedMessages[len(sortedMessages)-1]
	lastMessageMinutesAgo := now.Sub(lastMessage.Timestamp).Minutes()

	// More balanced logic
	if lastMessage.Role == "user" && lastMessageMinutesAgo < 3 {
		return "active"
	} else if lastMessage.Role == "assistant" && lastMessageMinutesAgo < 5 {
		return "active"
	}

	// Use file modification time for recent activity
	if minutesAgo < 5 {
		return "active"
	}
	if minutesAgo < 30 {
		return "recent"
	}
	return "inactive"
}

// QuickStateCalculation provides fast state calculation without file I/O
func (sc *StateCalculator) QuickStateCalculation(lastModified time.Time, hasActiveProcess bool) string {
	if !hasActiveProcess {
		return ""
	}

	now := time.Now()
	timeDiff := now.Sub(lastModified).Seconds()

	// More stable state logic
	if timeDiff < 30 {
		return "Claude Code working..."
	} else if timeDiff < 300 { // 5 minutes
		return "Awaiting user input..."
	} else {
		return "User typing..."
	}
}

// ActivityDetection represents the result of activity detection
type ActivityDetection struct {
	IsActive bool
	Status   string
}

// detectRealClaudeActivity detects real Claude Code activity
func (sc *StateCalculator) detectRealClaudeActivity(messages []Message, lastModified time.Time) ActivityDetection {
	now := time.Now()
	fileMinutesAgo := now.Sub(lastModified).Minutes()

	if messages == nil || len(messages) == 0 {
		return ActivityDetection{IsActive: false, Status: "No messages"}
	}

	// Sort messages by timestamp
	sortedMessages := sortMessagesByTimestamp(messages)
	lastMessage := sortedMessages[len(sortedMessages)-1]
	messageMinutesAgo := now.Sub(lastMessage.Timestamp).Minutes()

	// 1. Very recent file modification
	if fileMinutesAgo < 1 {
		return ActivityDetection{IsActive: true, Status: "Claude Code working..."}
	}

	// 2. Recent user message with recent file activity
	if lastMessage.Role == "user" && messageMinutesAgo < 5 && fileMinutesAgo < 10 {
		return ActivityDetection{IsActive: true, Status: "Claude Code working..."}
	}

	// 3. Recent assistant message with very recent file activity
	if lastMessage.Role == "assistant" && messageMinutesAgo < 2 && fileMinutesAgo < 5 {
		return ActivityDetection{IsActive: true, Status: "Claude Code finishing..."}
	}

	// 4. Look for tool activity patterns
	recentMessages := sortedMessages
	if len(sortedMessages) > 3 {
		recentMessages = sortedMessages[len(sortedMessages)-3:]
	}

	hasRecentTools := false
	for _, msg := range recentMessages {
		if msg.ToolResults != nil && len(msg.ToolResults) > 0 {
			hasRecentTools = true
			break
		}
	}

	if hasRecentTools && messageMinutesAgo < 10 && fileMinutesAgo < 15 {
		return ActivityDetection{IsActive: true, Status: "Active session"}
	}

	// 5. Rapid message exchange pattern
	if len(sortedMessages) >= 2 {
		lastTwoMessages := sortedMessages[len(sortedMessages)-2:]
		timeBetween := lastTwoMessages[1].Timestamp.Sub(lastTwoMessages[0].Timestamp).Minutes()

		if timeBetween < 5 && messageMinutesAgo < 15 && fileMinutesAgo < 20 {
			return ActivityDetection{IsActive: true, Status: "Active conversation"}
		}
	}

	return ActivityDetection{IsActive: false, Status: ""}
}

// GetStateClass returns CSS class for state styling
func (sc *StateCalculator) GetStateClass(conversationState string) string {
	lower := strings.ToLower(conversationState)
	if strings.Contains(lower, "working") {
		return "working"
	}
	if strings.Contains(lower, "typing") {
		return "typing"
	}
	return ""
}

// ClearCache clears any cached state information
func (sc *StateCalculator) ClearCache() {
	sc.processCache = make(map[string]interface{})
}

// sortMessagesByTimestamp sorts messages by timestamp using stdlib sort for O(n log n) performance.
// It creates a copy to avoid modifying the original slice.
func sortMessagesByTimestamp(messages []Message) []Message {
	sorted := make([]Message, len(messages))
	copy(sorted, messages)

	// Use stdlib sort.Slice for O(n log n) performance instead of O(n²) bubble sort
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	return sorted
}
