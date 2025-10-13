package cmd

import (
	"testing"
)

func TestParseComponentList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single component",
			input:    "agent1",
			expected: []string{"agent1"},
		},
		{
			name:     "Multiple components",
			input:    "agent1,agent2,agent3",
			expected: []string{"agent1", "agent2", "agent3"},
		},
		{
			name:     "Components with spaces",
			input:    "agent1, agent2 , agent3",
			expected: []string{"agent1", "agent2", "agent3"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "Only commas",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "Mixed spaces and commas",
			input:    " agent1 ,  , agent2,  agent3  ",
			expected: []string{"agent1", "agent2", "agent3"},
		},
		{
			name:     "Single component with trailing comma",
			input:    "agent1,",
			expected: []string{"agent1"},
		},
		{
			name:     "Hyphenated names",
			input:    "api-tester,code-reviewer,debug-assistant",
			expected: []string{"api-tester", "code-reviewer", "debug-assistant"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseComponentList(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d components, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected component[%d] = %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestVersionInfo(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	if Name == "" {
		t.Error("Name should not be empty")
	}

	if Name != "claude-control-terminal" {
		t.Errorf("Expected name 'claude-control-terminal', got '%s'", Name)
	}
}

func TestRootCommandInitialization(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	if rootCmd.Use != "cct" {
		t.Errorf("Expected command use 'cct', got '%s'", rootCmd.Use)
	}

	if rootCmd.Version != Version {
		t.Errorf("Expected command version '%s', got '%s'", Version, rootCmd.Version)
	}

	// Verify root command has expected flags
	flags := rootCmd.Flags()

	// Check for key flags
	expectedFlags := []string{
		"agent",
		"command",
		"mcp",
		"analytics",
		"docker-init",
		"docker-build",
		"preview",
		"dry-run",
	}

	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("Expected flag '%s' to be defined", flagName)
		}
	}
}

func TestRootCommandPersistentFlags(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	persistentFlags := rootCmd.PersistentFlags()

	// Check for persistent flags
	expectedPersistentFlags := []string{
		"verbose",
		"directory",
		"yes",
	}

	for _, flagName := range expectedPersistentFlags {
		if persistentFlags.Lookup(flagName) == nil {
			t.Errorf("Expected persistent flag '%s' to be defined", flagName)
		}
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute function exists
	// We don't actually execute it in tests to avoid side effects
	// Just verify it's callable by checking rootCmd exists
	if rootCmd == nil {
		t.Error("rootCmd should be defined for Execute to work")
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are defined
	if Version == "" {
		t.Error("Version constant should not be empty")
	}

	if Name == "" {
		t.Error("Name constant should not be empty")
	}

	// Test expected values
	expectedVersion := "0.2.15" // Update this when version changes
	if Version != expectedVersion {
		t.Logf("Version changed from %s to %s - update test if intentional", expectedVersion, Version)
	}

	if Name != "claude-control-terminal" {
		t.Errorf("Expected Name 'claude-control-terminal', got '%s'", Name)
	}
}
