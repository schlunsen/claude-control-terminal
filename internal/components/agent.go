// Package components provides installation and management of Claude Code components.
// This file handles agent component installation, downloading agents from GitHub
// and installing them to the .claude/agents directory.
package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
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

	// Try multiple path formats to find the agent
	var content string
	var err error
	var githubPath string

	// Format 1: Try with category (e.g., ai-specialists/data-scientist)
	if strings.Contains(agentName, "/") {
		githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
		if !silent {
			fmt.Printf("📥 Trying path: %s\n", githubPath)
		}
		content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Format 2: Try direct path (e.g., api-security-audit)
	githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
	if !silent {
		fmt.Printf("📥 Trying direct path: %s\n", githubPath)
	}
	content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
	if err == nil {
		goto Success
	}

	// Format 3: Search in common categories if simple name provided
	if !strings.Contains(agentName, "/") {
		categories := []string{
			"ai-specialists",
			"api-graphql",
			"blockchain-web3",
			"business-marketing",
			"data-ai",
			"database",
			"deep-research-team",
			"development-team",
			"development-tools",
			"devops-infrastructure",
			"documentation",
			"expert-advisors",
			"ffmpeg-clip-team",
			"game-development",
			"git",
			"mcp-dev-team",
			"modernization",
			"obsidian-ops-team",
			"ocr-extraction-team",
			"performance-testing",
			"podcast-creator-team",
			"programming-languages",
			"realtime",
			"security",
			"web-tools",
		}

		for _, category := range categories {
			githubPath = fmt.Sprintf("components/agents/%s/%s.md", category, agentName)
			if !silent {
				fmt.Printf("📥 Searching in %s category...\n", category)
			}
			content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
			if err == nil {
				goto Success
			}
		}
	}

	// All attempts failed
	return fmt.Errorf("agent '%s' not found (tried multiple paths)", agentName)

Success:
	if !silent {
		fmt.Printf("✅ Found agent at: %s\n", githubPath)
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

// PreviewAgent previews a specific agent component without installing
func (ai *AgentInstaller) PreviewAgent(agentName string) error {
	// Try multiple path formats to find the agent
	var content string
	var err error
	var githubPath string

	// Format 1: Try with category (e.g., ai-specialists/data-scientist)
	if strings.Contains(agentName, "/") {
		githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
		content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Format 2: Try direct path (e.g., api-security-audit)
	githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
	content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
	if err == nil {
		goto Success
	}

	// Format 3: Search in common categories if simple name provided
	if !strings.Contains(agentName, "/") {
		categories := []string{
			"ai-specialists",
			"api-graphql",
			"blockchain-web3",
			"business-marketing",
			"data-ai",
			"database",
			"deep-research-team",
			"development-team",
			"development-tools",
			"devops-infrastructure",
			"documentation",
			"expert-advisors",
			"ffmpeg-clip-team",
			"game-development",
			"git",
			"mcp-dev-team",
			"modernization",
			"obsidian-ops-team",
			"ocr-extraction-team",
			"performance-testing",
			"podcast-creator-team",
			"programming-languages",
			"realtime",
			"security",
			"web-tools",
		}

		for _, category := range categories {
			githubPath = fmt.Sprintf("components/agents/%s/%s.md", category, agentName)
			content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
			if err == nil {
				goto Success
			}
		}
	}

	// All attempts failed
	return fmt.Errorf("agent '%s' not found (tried multiple paths)", agentName)

Success:
	// Display preview
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📄 Agent: %s\n", agentName)
	fmt.Printf("🔗 Source: %s\n", githubPath)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Println(content)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return nil
}

// PreviewMultipleAgents previews multiple agents
func (ai *AgentInstaller) PreviewMultipleAgents(agentNames []string) error {
	successCount := 0
	failedCount := 0

	for _, agentName := range agentNames {
		if err := ai.PreviewAgent(agentName); err != nil {
			fmt.Printf("❌ Failed to preview agent '%s': %v\n", agentName, err)
			failedCount++
		} else {
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("all agent previews failed")
	}

	return nil
}

// RemoveAgent removes an installed agent
func (ai *AgentInstaller) RemoveAgent(agentName, targetDir string, silent bool) error {
	if !silent {
		fmt.Printf("🗑️  Removing agent: %s\n", agentName)
	}

	// Extract filename if category path provided
	var fileName string
	if strings.Contains(agentName, "/") {
		parts := strings.Split(agentName, "/")
		fileName = parts[len(parts)-1]
	} else {
		fileName = agentName
	}

	// Check project installation
	projectFile := filepath.Join(targetDir, ".claude", "agents", fileName+".md")
	projectExists := false
	if _, err := os.Stat(projectFile); err == nil {
		projectExists = true
	}

	// Check global installation
	homeDir, _ := os.UserHomeDir()
	globalFile := filepath.Join(homeDir, ".claude", "agents", fileName+".md")
	globalExists := false
	if _, err := os.Stat(globalFile); err == nil {
		globalExists = true
	}

	if !projectExists && !globalExists {
		return fmt.Errorf("agent '%s' is not installed", agentName)
	}

	// Remove project installation if exists
	if projectExists {
		if err := os.Remove(projectFile); err != nil {
			return fmt.Errorf("failed to remove project agent: %w", err)
		}
		if !silent {
			fmt.Printf("✅ Removed from project: %s\n", projectFile)
		}
	}

	// Remove global installation if exists
	if globalExists {
		if err := os.Remove(globalFile); err != nil {
			return fmt.Errorf("failed to remove global agent: %w", err)
		}
		if !silent {
			fmt.Printf("✅ Removed from global: %s\n", globalFile)
		}
	}

	if !silent {
		fmt.Printf("✅ Agent '%s' removed successfully!\n", agentName)
	}

	return nil
}
