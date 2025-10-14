package analytics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewResetTracker(t *testing.T) {
	tempDir := t.TempDir()

	rt := NewResetTracker(tempDir)

	if rt == nil {
		t.Fatal("NewResetTracker returned nil")
	}

	if rt.claudeDir != tempDir {
		t.Errorf("expected claudeDir %q, got %q", tempDir, rt.claudeDir)
	}

	expectedResetFile := filepath.Join(tempDir, ".analytics_reset")
	if rt.resetFile != expectedResetFile {
		t.Errorf("expected resetFile %q, got %q", expectedResetFile, rt.resetFile)
	}

	if rt.resetPoint != nil {
		t.Error("expected resetPoint to be nil for new tracker")
	}
}

func TestResetTracker_SetResetPoint(t *testing.T) {
	tempDir := t.TempDir()
	rt := NewResetTracker(tempDir)

	totalTokens := 1000
	totalConversations := 5
	reason := "Test reset"

	err := rt.SetResetPoint(totalTokens, totalConversations, reason)
	if err != nil {
		t.Fatalf("SetResetPoint failed: %v", err)
	}

	// Verify reset point was set
	if !rt.HasResetPoint() {
		t.Error("reset point should be set")
	}

	point := rt.GetResetPoint()
	if point == nil {
		t.Fatal("GetResetPoint returned nil")
	}

	if point.TokenDelta != totalTokens {
		t.Errorf("expected TokenDelta %d, got %d", totalTokens, point.TokenDelta)
	}

	if point.ConversationDelta != totalConversations {
		t.Errorf("expected ConversationDelta %d, got %d", totalConversations, point.ConversationDelta)
	}

	if point.Reason != reason {
		t.Errorf("expected Reason %q, got %q", reason, point.Reason)
	}

	// Verify timestamp is recent
	if time.Since(point.Timestamp) > 5*time.Second {
		t.Error("timestamp should be recent")
	}

	// Verify file was created
	resetFile := filepath.Join(tempDir, ".analytics_reset")
	if _, err := os.Stat(resetFile); os.IsNotExist(err) {
		t.Error("reset file was not created")
	}
}

func TestResetTracker_GetResetPoint(t *testing.T) {
	tempDir := t.TempDir()
	rt := NewResetTracker(tempDir)

	// Should return nil when no reset point
	point := rt.GetResetPoint()
	if point != nil {
		t.Error("expected nil when no reset point set")
	}

	// Set a reset point
	err := rt.SetResetPoint(100, 2, "test")
	if err != nil {
		t.Fatalf("SetResetPoint failed: %v", err)
	}

	// Should return a copy of the reset point
	point = rt.GetResetPoint()
	if point == nil {
		t.Fatal("GetResetPoint returned nil after setting")
	}

	// Modify the returned point - should not affect internal state
	originalTokenDelta := point.TokenDelta
	point.TokenDelta = 999
	point.ConversationDelta = 999

	// Get again and verify it wasn't modified
	point2 := rt.GetResetPoint()
	if point2.TokenDelta != originalTokenDelta {
		t.Error("returned reset point should be a copy, not a reference")
	}
}

func TestResetTracker_ApplyDelta(t *testing.T) {
	tempDir := t.TempDir()
	rt := NewResetTracker(tempDir)

	tests := []struct {
		name                  string
		hasResetPoint         bool
		resetTokens           int
		resetConversations    int
		rawTokens             int
		rawConversations      int
		expectedTokens        int
		expectedConversations int
	}{
		{
			name:                  "no reset point",
			hasResetPoint:         false,
			rawTokens:             1000,
			rawConversations:      5,
			expectedTokens:        1000,
			expectedConversations: 5,
		},
		{
			name:                  "with reset point",
			hasResetPoint:         true,
			resetTokens:           500,
			resetConversations:    2,
			rawTokens:             1500,
			rawConversations:      7,
			expectedTokens:        1000,
			expectedConversations: 5,
		},
		{
			name:                  "negative result clamped to zero",
			hasResetPoint:         true,
			resetTokens:           1000,
			resetConversations:    5,
			rawTokens:             500,
			rawConversations:      3,
			expectedTokens:        0,
			expectedConversations: 0,
		},
		{
			name:                  "exact match results in zero",
			hasResetPoint:         true,
			resetTokens:           1000,
			resetConversations:    5,
			rawTokens:             1000,
			rawConversations:      5,
			expectedTokens:        0,
			expectedConversations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			rt.ClearResetPoint()

			if tt.hasResetPoint {
				err := rt.SetResetPoint(tt.resetTokens, tt.resetConversations, "test")
				if err != nil {
					t.Fatalf("SetResetPoint failed: %v", err)
				}
			}

			adjustedTokens, adjustedConversations := rt.ApplyDelta(tt.rawTokens, tt.rawConversations)

			if adjustedTokens != tt.expectedTokens {
				t.Errorf("expected adjustedTokens %d, got %d", tt.expectedTokens, adjustedTokens)
			}

			if adjustedConversations != tt.expectedConversations {
				t.Errorf("expected adjustedConversations %d, got %d", tt.expectedConversations, adjustedConversations)
			}
		})
	}
}

func TestResetTracker_ClearResetPoint(t *testing.T) {
	tempDir := t.TempDir()
	rt := NewResetTracker(tempDir)

	// Set a reset point
	err := rt.SetResetPoint(100, 5, "test")
	if err != nil {
		t.Fatalf("SetResetPoint failed: %v", err)
	}

	if !rt.HasResetPoint() {
		t.Fatal("reset point should be set")
	}

	resetFile := filepath.Join(tempDir, ".analytics_reset")
	if _, err := os.Stat(resetFile); os.IsNotExist(err) {
		t.Error("reset file should exist")
	}

	// Clear the reset point
	err = rt.ClearResetPoint()
	if err != nil {
		t.Fatalf("ClearResetPoint failed: %v", err)
	}

	if rt.HasResetPoint() {
		t.Error("reset point should be cleared")
	}

	if _, err := os.Stat(resetFile); !os.IsNotExist(err) {
		t.Error("reset file should be deleted")
	}

	// Clearing again should not error
	err = rt.ClearResetPoint()
	if err != nil {
		t.Errorf("clearing non-existent reset point should not error: %v", err)
	}
}

func TestResetTracker_HasResetPoint(t *testing.T) {
	tempDir := t.TempDir()
	rt := NewResetTracker(tempDir)

	if rt.HasResetPoint() {
		t.Error("new tracker should not have reset point")
	}

	err := rt.SetResetPoint(100, 5, "test")
	if err != nil {
		t.Fatalf("SetResetPoint failed: %v", err)
	}

	if !rt.HasResetPoint() {
		t.Error("should have reset point after setting")
	}

	err = rt.ClearResetPoint()
	if err != nil {
		t.Fatalf("ClearResetPoint failed: %v", err)
	}

	if rt.HasResetPoint() {
		t.Error("should not have reset point after clearing")
	}
}

func TestResetTracker_LoadExistingResetPoint(t *testing.T) {
	tempDir := t.TempDir()

	// Create a reset point file manually
	resetPoint := ResetPoint{
		Timestamp:         time.Now().Add(-1 * time.Hour),
		TokenDelta:        500,
		ConversationDelta: 3,
		Reason:            "Previous reset",
	}

	data, err := json.MarshalIndent(resetPoint, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal reset point: %v", err)
	}

	resetFile := filepath.Join(tempDir, ".analytics_reset")
	err = os.WriteFile(resetFile, data, 0644)
	if err != nil {
		t.Fatalf("failed to write reset file: %v", err)
	}

	// Create tracker - should load the existing reset point
	rt := NewResetTracker(tempDir)

	if !rt.HasResetPoint() {
		t.Fatal("should have loaded existing reset point")
	}

	loadedPoint := rt.GetResetPoint()
	if loadedPoint == nil {
		t.Fatal("GetResetPoint returned nil")
	}

	if loadedPoint.TokenDelta != resetPoint.TokenDelta {
		t.Errorf("expected TokenDelta %d, got %d", resetPoint.TokenDelta, loadedPoint.TokenDelta)
	}

	if loadedPoint.ConversationDelta != resetPoint.ConversationDelta {
		t.Errorf("expected ConversationDelta %d, got %d", resetPoint.ConversationDelta, loadedPoint.ConversationDelta)
	}

	if loadedPoint.Reason != resetPoint.Reason {
		t.Errorf("expected Reason %q, got %q", resetPoint.Reason, loadedPoint.Reason)
	}
}

func TestResetTracker_LoadInvalidResetPoint(t *testing.T) {
	tempDir := t.TempDir()

	// Create an invalid reset point file
	resetFile := filepath.Join(tempDir, ".analytics_reset")
	err := os.WriteFile(resetFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("failed to write reset file: %v", err)
	}

	// Create tracker - should handle invalid file gracefully
	rt := NewResetTracker(tempDir)

	// Should not have a reset point due to parse error
	if rt.HasResetPoint() {
		t.Error("should not have reset point when file is invalid")
	}
}

func TestResetTracker_PersistenceAcrossInstances(t *testing.T) {
	tempDir := t.TempDir()

	// Create first tracker and set reset point
	rt1 := NewResetTracker(tempDir)
	err := rt1.SetResetPoint(750, 4, "persistence test")
	if err != nil {
		t.Fatalf("SetResetPoint failed: %v", err)
	}

	// Create second tracker - should load the same reset point
	rt2 := NewResetTracker(tempDir)

	if !rt2.HasResetPoint() {
		t.Fatal("second tracker should have loaded reset point")
	}

	point := rt2.GetResetPoint()
	if point.TokenDelta != 750 {
		t.Errorf("expected TokenDelta 750, got %d", point.TokenDelta)
	}

	if point.ConversationDelta != 4 {
		t.Errorf("expected ConversationDelta 4, got %d", point.ConversationDelta)
	}

	if point.Reason != "persistence test" {
		t.Errorf("expected Reason %q, got %q", "persistence test", point.Reason)
	}
}

func TestResetPoint_JSONSerialization(t *testing.T) {
	now := time.Now()
	original := ResetPoint{
		Timestamp:         now,
		TokenDelta:        1234,
		ConversationDelta: 10,
		Reason:            "JSON test",
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded ResetPoint
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Compare
	if decoded.TokenDelta != original.TokenDelta {
		t.Errorf("TokenDelta mismatch: expected %d, got %d", original.TokenDelta, decoded.TokenDelta)
	}

	if decoded.ConversationDelta != original.ConversationDelta {
		t.Errorf("ConversationDelta mismatch: expected %d, got %d", original.ConversationDelta, decoded.ConversationDelta)
	}

	if decoded.Reason != original.Reason {
		t.Errorf("Reason mismatch: expected %q, got %q", original.Reason, decoded.Reason)
	}

	// Timestamps should be close (JSON loses some precision)
	if decoded.Timestamp.Unix() != original.Timestamp.Unix() {
		t.Errorf("Timestamp mismatch: expected %v, got %v", original.Timestamp, decoded.Timestamp)
	}
}
