package components

import (
	"testing"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

func TestParseMCPScope(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected fileops.MCPScope
	}{
		{
			name:     "user scope",
			input:    "user",
			expected: fileops.MCPScopeUser,
		},
		{
			name:     "global scope",
			input:    "global",
			expected: fileops.MCPScopeUser,
		},
		{
			name:     "project scope",
			input:    "project",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "local scope",
			input:    "local",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "empty string defaults to project",
			input:    "",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "unknown value defaults to project",
			input:    "unknown",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "uppercase user",
			input:    "USER",
			expected: fileops.MCPScopeUser,
		},
		{
			name:     "mixed case global",
			input:    "Global",
			expected: fileops.MCPScopeUser,
		},
		{
			name:     "uppercase project",
			input:    "PROJECT",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "user with whitespace",
			input:    "  user  ",
			expected: fileops.MCPScopeUser,
		},
		{
			name:     "project with whitespace",
			input:    "  project  ",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "tabs and spaces",
			input:    "\t global \t",
			expected: fileops.MCPScopeUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMCPScope(tt.input)
			if result != tt.expected {
				t.Errorf("ParseMCPScope(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewMCPInstaller(t *testing.T) {
	tests := []struct {
		name  string
		scope fileops.MCPScope
	}{
		{
			name:  "project scope",
			scope: fileops.MCPScopeProject,
		},
		{
			name:  "user scope",
			scope: fileops.MCPScopeUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installer := NewMCPInstaller(tt.scope)

			if installer == nil {
				t.Fatal("NewMCPInstaller returned nil")
			}

			if installer.scope != tt.scope {
				t.Errorf("Expected scope %v, got %v", tt.scope, installer.scope)
			}

			if installer.config == nil {
				t.Error("Config should not be nil")
			}
		})
	}
}

func TestMCPInstallerStructure(t *testing.T) {
	// Test that MCPInstaller has expected fields
	installer := NewMCPInstaller(fileops.MCPScopeProject)

	if installer.config == nil {
		t.Error("Config should be initialized")
	}

	if installer.scope != fileops.MCPScopeProject {
		t.Error("Scope should be set correctly")
	}

	// Test that config has expected structure
	if installer.config.Owner == "" {
		t.Error("Config Owner should not be empty")
	}

	if installer.config.Repo == "" {
		t.Error("Config Repo should not be empty")
	}

	if installer.config.Branch == "" {
		t.Error("Config Branch should not be empty")
	}
}

func TestMCPScopeConsistency(t *testing.T) {
	// Test that multiple calls with same input give same result
	input := "user"

	result1 := ParseMCPScope(input)
	result2 := ParseMCPScope(input)

	if result1 != result2 {
		t.Error("ParseMCPScope should be deterministic")
	}
}

func TestParseMCPScopeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected fileops.MCPScope
	}{
		{
			name:     "special characters",
			input:    "user!@#",
			expected: fileops.MCPScopeProject, // Should default to project for invalid input
		},
		{
			name:     "numbers",
			input:    "123",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "mixed valid and invalid",
			input:    "userproject",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: fileops.MCPScopeProject,
		},
		{
			name:     "newline characters",
			input:    "\n\r",
			expected: fileops.MCPScopeProject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMCPScope(tt.input)
			if result != tt.expected {
				t.Errorf("ParseMCPScope(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMCPScopeValues(t *testing.T) {
	// Test that the scope constants have distinct values
	if fileops.MCPScopeProject == fileops.MCPScopeUser {
		t.Error("MCPScopeProject and MCPScopeUser should have different values")
	}
}
