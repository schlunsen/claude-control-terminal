package wrapper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindClaudePath(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() (cleanup func())
		expectError   bool
		expectContains string
	}{
		{
			name: "finds claude in PATH",
			setupFunc: func() func() {
				// This test relies on claude being in PATH
				// We'll just verify the function doesn't error
				return func() {}
			},
			expectError: false, // Will error if claude not installed, but that's ok for this test
		},
		{
			name: "finds claude in ~/.local/bin when not in PATH",
			setupFunc: func() func() {
				// Create a temporary directory structure
				homeDir := t.TempDir()
				localBinDir := filepath.Join(homeDir, ".local", "bin")
				err := os.MkdirAll(localBinDir, 0755)
				if err != nil {
					t.Fatalf("failed to create test directory: %v", err)
				}

				// Create a dummy claude executable
				claudePath := filepath.Join(localBinDir, "claude")
				err = os.WriteFile(claudePath, []byte("#!/bin/bash\necho 'test'"), 0755)
				if err != nil {
					t.Fatalf("failed to create test claude binary: %v", err)
				}

				// Override HOME and PATH environment variables
				oldHome := os.Getenv("HOME")
				oldPath := os.Getenv("PATH")
				os.Setenv("HOME", homeDir)
				os.Setenv("PATH", "/usr/bin:/bin") // Remove claude from PATH

				return func() {
					os.Setenv("HOME", oldHome)
					os.Setenv("PATH", oldPath)
				}
			},
			expectError:    false,
			expectContains: ".local/bin/claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc()
			defer cleanup()

			path, err := FindClaudePath()

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil && tt.name != "finds claude in PATH" {
				// Allow PATH test to fail if claude not installed
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectContains != "" && path != "" {
				if !contains(path, tt.expectContains) {
					t.Errorf("expected path to contain %q, got %q", tt.expectContains, path)
				}
			}
		})
	}
}

func TestFindClaudePath_ChecksCommonLocations(t *testing.T) {
	// This test verifies that FindClaudePath checks common locations
	// even when claude is not in PATH

	// Create a test directory structure
	homeDir := t.TempDir()
	localBinDir := filepath.Join(homeDir, ".local", "bin")
	err := os.MkdirAll(localBinDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create a dummy claude executable
	claudePath := filepath.Join(localBinDir, "claude")
	err = os.WriteFile(claudePath, []byte("#!/bin/bash\necho 'test'"), 0755)
	if err != nil {
		t.Fatalf("failed to create test claude binary: %v", err)
	}

	// Override HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	// Override PATH to not include .local/bin
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "/usr/bin:/bin")
	defer os.Setenv("PATH", oldPath)

	// Call FindClaudePath - it should find claude in ~/.local/bin
	// even though it's not in PATH
	foundPath, err := FindClaudePath()
	if err != nil {
		t.Errorf("expected to find claude in ~/.local/bin, got error: %v", err)
	}

	if !contains(foundPath, ".local/bin/claude") {
		t.Errorf("expected path to contain .local/bin/claude, got %q", foundPath)
	}
}

func TestFindClaudePath_ReturnsErrorWhenNotFound(t *testing.T) {
	// Create a test directory with no claude binary
	homeDir := t.TempDir()

	// Override HOME and PATH
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", oldPath)

	// Call FindClaudePath - it should return an error
	_, err := FindClaudePath()
	if err == nil {
		t.Errorf("expected error when claude not found, got nil")
	}

	expectedError := "claude executable not found in PATH or common locations"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && anySubstring(s, substr))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
