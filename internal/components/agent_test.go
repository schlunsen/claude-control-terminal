package components

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

func TestNewAgentInstaller(t *testing.T) {
	installer := NewAgentInstaller()

	if installer == nil {
		t.Fatal("NewAgentInstaller returned nil")
	}

	if installer.config == nil {
		t.Error("AgentInstaller config should not be nil")
	}

	if installer.config.Owner == "" {
		t.Error("config Owner should not be empty")
	}

	if installer.config.Repo == "" {
		t.Error("config Repo should not be empty")
	}
}

func TestAgentInstallerStruct(t *testing.T) {
	installer := AgentInstaller{}

	// Test zero value
	if installer.config != nil {
		t.Error("uninitialized AgentInstaller config should be nil")
	}
}

func TestInstallAgentExtractsFilename(t *testing.T) {
	tests := []struct {
		name           string
		agentName      string
		expectedFile   string
	}{
		{
			name:         "simple name",
			agentName:    "test-agent",
			expectedFile: "test-agent.md",
		},
		{
			name:         "with category",
			agentName:    "security/test-agent",
			expectedFile: "test-agent.md",
		},
		{
			name:         "multiple slashes",
			agentName:    "category/subcategory/agent-name",
			expectedFile: "agent-name.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract filename logic from InstallAgent
			var fileName string
			if filepath.Base(tt.agentName) != tt.agentName {
				fileName = filepath.Base(tt.agentName)
			} else {
				fileName = tt.agentName
			}

			expected := tt.expectedFile[:len(tt.expectedFile)-3] // Remove .md
			if fileName != expected {
				t.Errorf("expected filename %q, got %q", expected, fileName)
			}
		})
	}
}

func TestRemoveAgentNonExistent(t *testing.T) {
	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	// Try to remove non-existent agent
	err := installer.RemoveAgent("nonexistent-agent", tempDir, true)

	if err == nil {
		t.Error("expected error when removing non-existent agent")
	}

	if err != nil && err.Error() != "agent 'nonexistent-agent' is not installed" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRemoveAgentExistingFile(t *testing.T) {
	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	// Create agent file
	agentsDir := filepath.Join(tempDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	agentFile := filepath.Join(agentsDir, "test-agent.md")
	if err := os.WriteFile(agentFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test agent file: %v", err)
	}

	// Remove the agent
	err := installer.RemoveAgent("test-agent", tempDir, true)

	if err != nil {
		t.Errorf("unexpected error removing agent: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		t.Error("agent file should have been deleted")
	}
}

func TestRemoveAgentWithCategory(t *testing.T) {
	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	// Create agent file
	agentsDir := filepath.Join(tempDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	agentFile := filepath.Join(agentsDir, "test-agent.md")
	if err := os.WriteFile(agentFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test agent file: %v", err)
	}

	// Remove using category/name format
	err := installer.RemoveAgent("security/test-agent", tempDir, true)

	if err != nil {
		t.Errorf("unexpected error removing agent: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		t.Error("agent file should have been deleted")
	}
}

func TestInstallMultipleAgentsAllFail(t *testing.T) {
	// Mock the download function to simulate failures without making real HTTP calls
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return "", fmt.Errorf("file not found: %s (404)", filePath)
	})
	defer fileops.MockDownloadFunc(nil) // Restore default

	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	// These agents don't exist on GitHub
	agentNames := []string{
		"nonexistent-agent-xyz-123",
		"another-fake-agent-abc-456",
	}

	err := installer.InstallMultipleAgents(agentNames, tempDir, true)

	if err == nil {
		t.Error("expected error when all installations fail")
	}

	if err != nil && err.Error() != "all agent installations failed" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPreviewMultipleAgentsAllFail(t *testing.T) {
	// Mock the download function to simulate failures without making real HTTP calls
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return "", fmt.Errorf("file not found: %s (404)", filePath)
	})
	defer fileops.MockDownloadFunc(nil) // Restore default

	installer := NewAgentInstaller()

	// These agents don't exist
	agentNames := []string{
		"nonexistent-agent-xyz-123",
		"another-fake-agent-abc-456",
	}

	err := installer.PreviewMultipleAgents(agentNames)

	if err == nil {
		t.Error("expected error when all previews fail")
	}

	if err != nil && err.Error() != "all agent previews failed" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAgentInstallerConfig(t *testing.T) {
	installer := NewAgentInstaller()

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

func TestRemoveAgentGlobalAndProject(t *testing.T) {
	installer := NewAgentInstaller()
	projectDir := t.TempDir()
	homeDir := t.TempDir()

	// Create project agent file
	projectAgentsDir := filepath.Join(projectDir, ".claude", "agents")
	if err := os.MkdirAll(projectAgentsDir, 0755); err != nil {
		t.Fatalf("failed to create project agents dir: %v", err)
	}

	projectFile := filepath.Join(projectAgentsDir, "test-agent.md")
	if err := os.WriteFile(projectFile, []byte("project content"), 0644); err != nil {
		t.Fatalf("failed to create project agent file: %v", err)
	}

	// Create global agent file in simulated home dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", oldHome)

	globalAgentsDir := filepath.Join(homeDir, ".claude", "agents")
	if err := os.MkdirAll(globalAgentsDir, 0755); err != nil {
		t.Fatalf("failed to create global agents dir: %v", err)
	}

	globalFile := filepath.Join(globalAgentsDir, "test-agent.md")
	if err := os.WriteFile(globalFile, []byte("global content"), 0644); err != nil {
		t.Fatalf("failed to create global agent file: %v", err)
	}

	// Remove the agent (should remove both)
	err := installer.RemoveAgent("test-agent", projectDir, true)

	if err != nil {
		t.Errorf("unexpected error removing agent: %v", err)
	}

	// Verify both files were deleted
	if _, err := os.Stat(projectFile); !os.IsNotExist(err) {
		t.Error("project agent file should have been deleted")
	}

	if _, err := os.Stat(globalFile); !os.IsNotExist(err) {
		t.Error("global agent file should have been deleted")
	}
}

func TestInstallAgentInvalidTarget(t *testing.T) {
	// Mock the download function to simulate failures without making real HTTP calls
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return "", fmt.Errorf("file not found: %s (404)", filePath)
	})
	defer fileops.MockDownloadFunc(nil) // Restore default

	installer := NewAgentInstaller()

	// Try to install to invalid directory - should fail with file not found
	err := installer.InstallAgent("test-agent", "/root/impossible", true)

	if err == nil {
		t.Error("expected error when agent not found")
	}
}

func TestInstallAgentSuccess(t *testing.T) {
	// Mock successful download
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return "# Test Agent\nThis is a test agent content", nil
	})
	defer fileops.MockDownloadFunc(nil)

	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	err := installer.InstallAgent("test-agent", tempDir, true)
	if err != nil {
		t.Errorf("InstallAgent failed: %v", err)
	}

	// Verify file was created
	agentFile := filepath.Join(tempDir, ".claude", "agents", "test-agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Error("agent file should have been created")
	}

	// Verify content
	content, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("failed to read agent file: %v", err)
	}

	if string(content) != "# Test Agent\nThis is a test agent content" {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestInstallAgentWithCategory(t *testing.T) {
	// Mock successful download
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return "# Security Agent\nContent", nil
	})
	defer fileops.MockDownloadFunc(nil)

	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	// Install with category path
	err := installer.InstallAgent("security/security-agent", tempDir, true)
	if err != nil {
		t.Errorf("InstallAgent with category failed: %v", err)
	}

	// Verify file was created (should use base name)
	agentFile := filepath.Join(tempDir, ".claude", "agents", "security-agent.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Error("agent file should have been created")
	}
}

func TestPreviewAgentSuccess(t *testing.T) {
	// Mock successful download
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return "# Preview Agent\nThis is preview content", nil
	})
	defer fileops.MockDownloadFunc(nil)

	installer := NewAgentInstaller()

	err := installer.PreviewAgent("preview-agent")
	if err != nil {
		t.Errorf("PreviewAgent failed: %v", err)
	}
}

func TestPreviewAgentNotFound(t *testing.T) {
	// Mock download failure
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return "", fmt.Errorf("file not found: %s (404)", filePath)
	})
	defer fileops.MockDownloadFunc(nil)

	installer := NewAgentInstaller()

	err := installer.PreviewAgent("nonexistent-agent")
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
}

func TestInstallMultipleAgentsPartialSuccess(t *testing.T) {
	// Mock download function that succeeds for some, fails for others
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		if filepath.Base(filePath) == "good-agent.md" {
			return "# Good Agent", nil
		}
		return "", fmt.Errorf("file not found: %s (404)", filePath)
	})
	defer fileops.MockDownloadFunc(nil)

	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	agentNames := []string{
		"good-agent",
		"bad-agent",
	}

	err := installer.InstallMultipleAgents(agentNames, tempDir, true)

	// Should not return error if at least one succeeds
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify good agent was installed
	goodFile := filepath.Join(tempDir, ".claude", "agents", "good-agent.md")
	if _, err := os.Stat(goodFile); os.IsNotExist(err) {
		t.Error("good agent file should have been created")
	}
}

func TestInstallMultipleAgentsAllSuccess(t *testing.T) {
	// Mock successful downloads for all agents
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return fmt.Sprintf("# Agent content for %s", filepath.Base(filePath)), nil
	})
	defer fileops.MockDownloadFunc(nil)

	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	agentNames := []string{
		"agent-one",
		"agent-two",
		"agent-three",
	}

	err := installer.InstallMultipleAgents(agentNames, tempDir, true)
	if err != nil {
		t.Errorf("InstallMultipleAgents failed: %v", err)
	}

	// Verify all agents were installed
	for _, name := range agentNames {
		agentFile := filepath.Join(tempDir, ".claude", "agents", name+".md")
		if _, err := os.Stat(agentFile); os.IsNotExist(err) {
			t.Errorf("agent file %s should have been created", name)
		}
	}
}

func TestPreviewMultipleAgentsAllSuccess(t *testing.T) {
	// Mock successful downloads
	fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
		return fmt.Sprintf("# Content for %s", filepath.Base(filePath)), nil
	})
	defer fileops.MockDownloadFunc(nil)

	installer := NewAgentInstaller()

	agentNames := []string{
		"agent-one",
		"agent-two",
	}

	err := installer.PreviewMultipleAgents(agentNames)
	if err != nil {
		t.Errorf("PreviewMultipleAgents failed: %v", err)
	}
}

func TestAgentInstallerEdgeCases(t *testing.T) {
	installer := NewAgentInstaller()
	tempDir := t.TempDir()

	t.Run("empty agent name", func(t *testing.T) {
		fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
			return "", fmt.Errorf("empty name")
		})
		defer fileops.MockDownloadFunc(nil)

		err := installer.InstallAgent("", tempDir, true)
		if err == nil {
			t.Error("expected error for empty agent name")
		}
	})

	t.Run("agent with spaces", func(t *testing.T) {
		fileops.MockDownloadFunc(func(config *fileops.GitHubConfig, filePath string, retryCount int) (string, error) {
			return "# Agent Content", nil
		})
		defer fileops.MockDownloadFunc(nil)

		err := installer.InstallAgent("agent with spaces", tempDir, true)
		// Should handle spaces appropriately
		_ = err
	})
}
