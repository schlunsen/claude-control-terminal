package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ResetPoint represents a point in time when analytics were "reset"
type ResetPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	TokenDelta     int       `json:"tokenDelta"`
	ConversationDelta int    `json:"conversationDelta"`
	Reason         string    `json:"reason"`
}

// ResetTracker manages soft resets with delta tracking
type ResetTracker struct {
	claudeDir   string
	resetFile   string
	resetPoint  *ResetPoint
	mutex       sync.RWMutex
}

// NewResetTracker creates a new ResetTracker
func NewResetTracker(claudeDir string) *ResetTracker {
	resetFile := filepath.Join(claudeDir, ".analytics_reset")
	rt := &ResetTracker{
		claudeDir: claudeDir,
		resetFile: resetFile,
	}
	rt.loadResetPoint()
	return rt
}

// loadResetPoint loads the reset point from disk
func (rt *ResetTracker) loadResetPoint() {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	data, err := os.ReadFile(rt.resetFile)
	if err != nil {
		// No reset point exists yet
		return
	}

	var point ResetPoint
	if err := json.Unmarshal(data, &point); err != nil {
		fmt.Printf("Warning: Failed to parse reset point: %v\n", err)
		return
	}

	rt.resetPoint = &point
}

// saveResetPoint saves the reset point to disk
// NOTE: Caller must hold the mutex lock
func (rt *ResetTracker) saveResetPoint() error {
	if rt.resetPoint == nil {
		return nil
	}

	data, err := json.MarshalIndent(rt.resetPoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reset point: %w", err)
	}

	if err := os.WriteFile(rt.resetFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write reset file: %w", err)
	}

	return nil
}

// SetResetPoint creates a new reset point with current totals
func (rt *ResetTracker) SetResetPoint(totalTokens, totalConversations int, reason string) error {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	rt.resetPoint = &ResetPoint{
		Timestamp:         time.Now(),
		TokenDelta:        totalTokens,
		ConversationDelta: totalConversations,
		Reason:            reason,
	}

	if err := rt.saveResetPoint(); err != nil {
		return err
	}

	return nil
}

// GetResetPoint returns the current reset point
func (rt *ResetTracker) GetResetPoint() *ResetPoint {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	if rt.resetPoint == nil {
		return nil
	}

	// Return a copy
	return &ResetPoint{
		Timestamp:         rt.resetPoint.Timestamp,
		TokenDelta:        rt.resetPoint.TokenDelta,
		ConversationDelta: rt.resetPoint.ConversationDelta,
		Reason:            rt.resetPoint.Reason,
	}
}

// ApplyDelta applies the reset delta to raw counts
func (rt *ResetTracker) ApplyDelta(rawTokens, rawConversations int) (adjustedTokens, adjustedConversations int) {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	if rt.resetPoint == nil {
		return rawTokens, rawConversations
	}

	adjustedTokens = rawTokens - rt.resetPoint.TokenDelta
	adjustedConversations = rawConversations - rt.resetPoint.ConversationDelta

	// Ensure non-negative
	if adjustedTokens < 0 {
		adjustedTokens = 0
	}
	if adjustedConversations < 0 {
		adjustedConversations = 0
	}

	return adjustedTokens, adjustedConversations
}

// ClearResetPoint removes the reset point
func (rt *ResetTracker) ClearResetPoint() error {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	rt.resetPoint = nil

	// Remove file if it exists
	if err := os.Remove(rt.resetFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove reset file: %w", err)
	}

	return nil
}

// HasResetPoint returns true if a reset point is set
func (rt *ResetTracker) HasResetPoint() bool {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	return rt.resetPoint != nil
}
