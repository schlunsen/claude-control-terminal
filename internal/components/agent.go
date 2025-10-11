package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davila7/go-claude-templates/internal/fileops"
)

// AgentInstaller handles agent component installation
type AgentInstaller struct {
	config *fileops.GitHubConfig
}

// NewAgentInstaller creates a new agent installer
func NewAgentInstaller() *AgentInstaller {
	return &AgentInstaller{
		config: fileops.DefaultGitHubConfig(),
	}
}

// InstallAgent installs a specific agent component
func (ai *AgentInstaller) InstallAgent(agentName, targetDir string, silent bool) error {
	if !silent {
		fmt.Printf("🤖 Installing agent: %s\n", agentName)
	}

	// Support both category/agent-name and direct agent-name formats
	var githubPath string
	if strings.Contains(agentName, "/") {
		// Category/agent format: deep-research-team/academic-researcher
		githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
	} else {
		// Direct agent format: api-security-audit
		githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
	}

	if !silent {
		fmt.Println("📥 Downloading from GitHub (main branch)...")
	}

	// Download the agent file
	content, err := fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return fmt.Errorf("agent '%s' not found", agentName)
		}
		return fmt.Errorf("failed to download agent: %w", err)
	}

	// Create .claude/agents directory if it doesn't exist
	agentsDir := filepath.Join(targetDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Write the agent file - always to flat .claude/agents directory
	var fileName string
	if strings.Contains(agentName, "/") {
		parts := strings.Split(agentName, "/")
		fileName = parts[len(parts)-1] // Extract just the filename
	} else {
		fileName = agentName
	}

	targetFile := filepath.Join(agentsDir, fileName+".md")
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	if !silent {
		relPath, _ := filepath.Rel(targetDir, targetFile)
		fmt.Printf("✅ Agent '%s' installed successfully!\n", agentName)
		fmt.Printf("📁 Installed to: %s\n", relPath)
		fmt.Printf("📦 Downloaded from: https://raw.githubusercontent.com/%s/%s/%s/%s/%s\n",
			ai.config.Owner, ai.config.Repo, ai.config.Branch, ai.config.TemplatesPath, githubPath)
	}

	return nil
}

// InstallMultipleAgents installs multiple agents
func (ai *AgentInstaller) InstallMultipleAgents(agentNames []string, targetDir string, silent bool) error {
	successCount := 0
	failedCount := 0

	for _, agentName := range agentNames {
		if err := ai.InstallAgent(agentName, targetDir, silent); err != nil {
			fmt.Printf("❌ Failed to install agent '%s': %v\n", agentName, err)
			failedCount++
		} else {
			successCount++
		}
	}

	if !silent {
		fmt.Printf("\n📊 Installation Summary:\n")
		fmt.Printf("   ✅ Successful: %d\n", successCount)
		if failedCount > 0 {
			fmt.Printf("   ❌ Failed: %d\n", failedCount)
		}
	}

	if successCount == 0 {
		return fmt.Errorf("all agent installations failed")
	}

	return nil
}
