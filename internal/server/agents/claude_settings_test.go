package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDeduplicatePermissions tests the deduplicatePermissions function
func TestDeduplicatePermissions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"Bash(git:*)", "Bash(gh:*)", "Edit(**)"},
			expected: []string{"Bash(git:*)", "Bash(gh:*)", "Edit(**)"},
		},
		{
			name:     "with duplicates",
			input:    []string{"Bash(git:*)", "Bash(gh:*)", "Bash(git:*)", "Edit(**)", "Bash(gh:*)"},
			expected: []string{"Bash(git:*)", "Bash(gh:*)", "Edit(**)"},
		},
		{
			name:     "all duplicates",
			input:    []string{"Bash(git:*)", "Bash(git:*)", "Bash(git:*)"},
			expected: []string{"Bash(git:*)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicatePermissions(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("deduplicatePermissions() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("deduplicatePermissions()[%d] = %s, want %s", i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestAddPermissionNoDuplicates tests that AddPermission prevents duplicates
func TestAddPermissionNoDuplicates(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "cct-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create settings manager
	csm := NewClaudeSettingsManager(tempDir)

	// Add permission first time
	err = csm.AddPermission("Bash(git:*)")
	if err != nil {
		t.Fatalf("Failed to add permission: %v", err)
	}

	// Add same permission again
	err = csm.AddPermission("Bash(git:*)")
	if err != nil {
		t.Fatalf("Failed to add duplicate permission: %v", err)
	}

	// Load settings and verify only one permission exists
	settings, err := csm.LoadSettings()
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	if len(settings.Permissions.Allow) != 1 {
		t.Errorf("Expected 1 permission, got %d", len(settings.Permissions.Allow))
	}

	if settings.Permissions.Allow[0] != "Bash(git:*)" {
		t.Errorf("Expected 'Bash(git:*)', got '%s'", settings.Permissions.Allow[0])
	}
}

// TestLoadSettingsDeduplicates tests that loading settings with duplicates deduplicates them
func TestLoadSettingsDeduplicates(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "cct-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	// Manually create settings file with duplicates
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	settingsWithDuplicates := ClaudeSettings{
		Permissions: struct {
			Allow []string `json:"allow"`
		}{
			Allow: []string{"Bash(git:*)", "Bash(gh:*)", "Bash(git:*)", "Edit(**)", "Bash(gh:*)"},
		},
	}

	data, err := json.MarshalIndent(settingsWithDuplicates, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		t.Fatalf("Failed to write settings file: %v", err)
	}

	// Load settings (should deduplicate)
	csm := NewClaudeSettingsManager(tempDir)
	settings, err := csm.LoadSettings()
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Verify duplicates were removed
	expected := []string{"Bash(git:*)", "Bash(gh:*)", "Edit(**)"}
	if len(settings.Permissions.Allow) != len(expected) {
		t.Errorf("Expected %d permissions, got %d", len(expected), len(settings.Permissions.Allow))
	}

	for i, perm := range expected {
		if settings.Permissions.Allow[i] != perm {
			t.Errorf("Permission[%d] = %s, want %s", i, settings.Permissions.Allow[i], perm)
		}
	}
}

// TestSaveSettingsDeduplicates tests that saving settings deduplicates permissions
func TestSaveSettingsDeduplicates(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "cct-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create settings manager
	csm := NewClaudeSettingsManager(tempDir)

	// Create settings with duplicates
	settings := &ClaudeSettings{
		Permissions: struct {
			Allow []string `json:"allow"`
		}{
			Allow: []string{"Bash(git:*)", "Bash(gh:*)", "Bash(git:*)", "Edit(**)", "Bash(gh:*)"},
		},
	}

	// Save settings (should deduplicate)
	err = csm.SaveSettings(settings)
	if err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	// Load settings back
	loadedSettings, err := csm.LoadSettings()
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Verify duplicates were removed
	expected := []string{"Bash(git:*)", "Bash(gh:*)", "Edit(**)"}
	if len(loadedSettings.Permissions.Allow) != len(expected) {
		t.Errorf("Expected %d permissions, got %d", len(expected), len(loadedSettings.Permissions.Allow))
	}

	for i, perm := range expected {
		if loadedSettings.Permissions.Allow[i] != perm {
			t.Errorf("Permission[%d] = %s, want %s", i, loadedSettings.Permissions.Allow[i], perm)
		}
	}
}
