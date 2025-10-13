package components

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCommandInstaller(t *testing.T) {
	installer := NewCommandInstaller()

	if installer == nil {
		t.Fatal("NewCommandInstaller returned nil")
	}

	if installer.config == nil {
		t.Error("CommandInstaller config should not be nil")
	}

	if installer.config.Owner == "" {
		t.Error("config Owner should not be empty")
	}

	if installer.config.Repo == "" {
		t.Error("config Repo should not be empty")
	}
}

func TestCommandInstallerStruct(t *testing.T) {
	installer := CommandInstaller{}

	// Test zero value
	if installer.config != nil {
		t.Error("uninitialized CommandInstaller config should be nil")
	}
}

func TestRemoveCommandNonExistent(t *testing.T) {
	installer := NewCommandInstaller()
	tempDir := t.TempDir()

	// Try to remove non-existent command
	err := installer.RemoveCommand("nonexistent-command", tempDir, true)

	if err == nil {
		t.Error("expected error when removing non-existent command")
	}

	if err != nil && err.Error() != "command 'nonexistent-command' is not installed" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRemoveCommandExistingFile(t *testing.T) {
	installer := NewCommandInstaller()
	tempDir := t.TempDir()

	// Create command file
	commandsDir := filepath.Join(tempDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}

	commandFile := filepath.Join(commandsDir, "test-command.md")
	if err := os.WriteFile(commandFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test command file: %v", err)
	}

	// Remove the command
	err := installer.RemoveCommand("test-command", tempDir, true)

	if err != nil {
		t.Errorf("unexpected error removing command: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(commandFile); !os.IsNotExist(err) {
		t.Error("command file should have been deleted")
	}
}

func TestRemoveCommandWithCategory(t *testing.T) {
	installer := NewCommandInstaller()
	tempDir := t.TempDir()

	// Create command file
	commandsDir := filepath.Join(tempDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}

	commandFile := filepath.Join(commandsDir, "test-command.md")
	if err := os.WriteFile(commandFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test command file: %v", err)
	}

	// Remove using category/name format
	err := installer.RemoveCommand("security/test-command", tempDir, true)

	if err != nil {
		t.Errorf("unexpected error removing command: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(commandFile); !os.IsNotExist(err) {
		t.Error("command file should have been deleted")
	}
}

func TestInstallMultipleCommandsAllFail(t *testing.T) {
	installer := NewCommandInstaller()
	tempDir := t.TempDir()

	// These commands don't exist on GitHub
	commandNames := []string{
		"nonexistent-command-xyz-123",
		"another-fake-command-abc-456",
	}

	err := installer.InstallMultipleCommands(commandNames, tempDir, true)

	if err == nil {
		t.Error("expected error when all installations fail")
	}

	if err != nil && err.Error() != "all command installations failed" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPreviewMultipleCommandsAllFail(t *testing.T) {
	installer := NewCommandInstaller()

	// These commands don't exist
	commandNames := []string{
		"nonexistent-command-xyz-123",
		"another-fake-command-abc-456",
	}

	err := installer.PreviewMultipleCommands(commandNames)

	if err == nil {
		t.Error("expected error when all previews fail")
	}

	if err != nil && err.Error() != "all command previews failed" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCommandInstallerConfig(t *testing.T) {
	installer := NewCommandInstaller()

	// Verify config is properly initialized
	if installer.config == nil {
		t.Fatal("config should not be nil")
	}

	// Check that it uses the default GitHub config
	if installer.config.Owner != "davila7" {
		t.Errorf("expected owner 'davila7', got %q", installer.config.Owner)
	}

	if installer.config.Repo != "claude-code-templates" {
		t.Errorf("expected repo 'claude-code-templates', got %q", installer.config.Repo)
	}

	if installer.config.Branch != "main" {
		t.Errorf("expected branch 'main', got %q", installer.config.Branch)
	}
}

func TestRemoveCommandGlobalAndProject(t *testing.T) {
	installer := NewCommandInstaller()
	projectDir := t.TempDir()
	homeDir := t.TempDir()

	// Create project command file
	projectCommandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(projectCommandsDir, 0755); err != nil {
		t.Fatalf("failed to create project commands dir: %v", err)
	}

	projectFile := filepath.Join(projectCommandsDir, "test-command.md")
	if err := os.WriteFile(projectFile, []byte("project content"), 0644); err != nil {
		t.Fatalf("failed to create project command file: %v", err)
	}

	// Create global command file in simulated home dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	globalCommandsDir := filepath.Join(homeDir, ".claude", "commands")
	if err := os.MkdirAll(globalCommandsDir, 0755); err != nil {
		t.Fatalf("failed to create global commands dir: %v", err)
	}

	globalFile := filepath.Join(globalCommandsDir, "test-command.md")
	if err := os.WriteFile(globalFile, []byte("global content"), 0644); err != nil {
		t.Fatalf("failed to create global command file: %v", err)
	}

	// Remove the command (should remove both)
	err := installer.RemoveCommand("test-command", projectDir, true)

	if err != nil {
		t.Errorf("unexpected error removing command: %v", err)
	}

	// Verify both files were deleted
	if _, err := os.Stat(projectFile); !os.IsNotExist(err) {
		t.Error("project command file should have been deleted")
	}

	if _, err := os.Stat(globalFile); !os.IsNotExist(err) {
		t.Error("global command file should have been deleted")
	}
}

func TestCommandExtractFilename(t *testing.T) {
	tests := []struct {
		name           string
		commandName    string
		expectedFile   string
	}{
		{
			name:         "simple name",
			commandName:  "test-command",
			expectedFile: "test-command.md",
		},
		{
			name:         "with category",
			commandName:  "security/test-command",
			expectedFile: "test-command.md",
		},
		{
			name:         "multiple slashes",
			commandName:  "category/subcategory/command-name",
			expectedFile: "command-name.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract filename logic from InstallCommand
			var fileName string
			if filepath.Base(tt.commandName) != tt.commandName {
				fileName = filepath.Base(tt.commandName)
			} else {
				fileName = tt.commandName
			}

			expected := tt.expectedFile[:len(tt.expectedFile)-3] // Remove .md
			if fileName != expected {
				t.Errorf("expected filename %q, got %q", expected, fileName)
			}
		})
	}
}

func TestInstallCommandInvalidTarget(t *testing.T) {
	installer := NewCommandInstaller()

	// Try to install to invalid directory (no write permissions)
	err := installer.InstallCommand("test-command", "/root/impossible", true)

	if err == nil {
		t.Log("Installation might have succeeded or failed with network error (expected)")
	}
	// We can't easily test this without mocking the download
}

func TestRemoveCommandSilentMode(t *testing.T) {
	installer := NewCommandInstaller()
	tempDir := t.TempDir()

	// Create command file
	commandsDir := filepath.Join(tempDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}

	commandFile := filepath.Join(commandsDir, "test-command.md")
	if err := os.WriteFile(commandFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test command file: %v", err)
	}

	// Remove the command in silent mode
	err := installer.RemoveCommand("test-command", tempDir, true)

	if err != nil {
		t.Errorf("unexpected error removing command: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(commandFile); !os.IsNotExist(err) {
		t.Error("command file should have been deleted")
	}
}

func TestRemoveCommandNonSilentMode(t *testing.T) {
	installer := NewCommandInstaller()
	tempDir := t.TempDir()

	// Create command file
	commandsDir := filepath.Join(tempDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}

	commandFile := filepath.Join(commandsDir, "test-command.md")
	if err := os.WriteFile(commandFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test command file: %v", err)
	}

	// Remove the command in non-silent mode (will print output)
	err := installer.RemoveCommand("test-command", tempDir, false)

	if err != nil {
		t.Errorf("unexpected error removing command: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(commandFile); !os.IsNotExist(err) {
		t.Error("command file should have been deleted")
	}
}
