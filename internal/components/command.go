// Package components provides slash command installation and management.
// This file handles command component installation, downloading slash commands from GitHub
// and installing them to the .claude/commands directory.
package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

// CommandInstaller handles command component installation
type CommandInstaller struct {
	config *fileops.GitHubConfig
}

// NewCommandInstaller creates a new command installer
func NewCommandInstaller() *CommandInstaller {
	return &CommandInstaller{
		config: fileops.DefaultGitHubConfig(),
	}
}

// InstallCommand installs a specific command component
func (ci *CommandInstaller) InstallCommand(commandName, targetDir string, silent bool) error {
	if !silent {
		fmt.Printf("⚡ Installing command: %s\n", commandName)
	}

	// Try multiple path formats to find the command
	var content string
	var err error
	var githubPath string

	// Format 1: Try with category (e.g., security/vulnerability-scan)
	if strings.Contains(commandName, "/") {
		githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
		if !silent {
			fmt.Printf("📥 Trying path: %s\n", githubPath)
		}
		content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Format 2: Try direct path
	githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
	if !silent {
		fmt.Printf("📥 Trying direct path: %s\n", githubPath)
	}
	content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
	if err == nil {
		goto Success
	}

	// Format 3: Search in common categories
	if !strings.Contains(commandName, "/") {
		categories := []string{
			"automation",
			"database",
			"deployment",
			"documentation",
			"game-development",
			"git",
			"git-workflow",
			"nextjs-vercel",
			"orchestration",
			"performance",
			"project-management",
			"security",
			"setup",
			"simulation",
			"svelte",
			"sync",
			"team",
			"testing",
			"utilities",
		}

		for _, category := range categories {
			githubPath = fmt.Sprintf("components/commands/%s/%s.md", category, commandName)
			if !silent {
				fmt.Printf("📥 Searching in %s category...\n", category)
			}
			content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
			if err == nil {
				goto Success
			}
		}
	}

	// All attempts failed
	return fmt.Errorf("command '%s' not found (tried multiple paths)", commandName)

Success:
	if !silent {
		fmt.Printf("✅ Found command at: %s\n", githubPath)
	}

	// Create .claude/commands directory if it doesn't exist
	commandsDir := filepath.Join(targetDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	// Write the command file
	var fileName string
	if strings.Contains(commandName, "/") {
		parts := strings.Split(commandName, "/")
		fileName = parts[len(parts)-1]
	} else {
		fileName = commandName
	}

	targetFile := filepath.Join(commandsDir, fileName+".md")
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write command file: %w", err)
	}

	if !silent {
		relPath, _ := filepath.Rel(targetDir, targetFile)
		fmt.Printf("✅ Command '%s' installed successfully!\n", commandName)
		fmt.Printf("📁 Installed to: %s\n", relPath)
	}

	return nil
}

// InstallMultipleCommands installs multiple commands
func (ci *CommandInstaller) InstallMultipleCommands(commandNames []string, targetDir string, silent bool) error {
	successCount := 0
	failedCount := 0

	for _, commandName := range commandNames {
		if err := ci.InstallCommand(commandName, targetDir, silent); err != nil {
			fmt.Printf("❌ Failed to install command '%s': %v\n", commandName, err)
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
		return fmt.Errorf("all command installations failed")
	}

	return nil
}

// PreviewCommand previews a specific command component without installing
func (ci *CommandInstaller) PreviewCommand(commandName string) error {
	// Try multiple path formats to find the command
	var content string
	var err error
	var githubPath string

	// Format 1: Try with category (e.g., security/vulnerability-scan)
	if strings.Contains(commandName, "/") {
		githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
		content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Format 2: Try direct path
	githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
	content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
	if err == nil {
		goto Success
	}

	// Format 3: Search in common categories
	if !strings.Contains(commandName, "/") {
		categories := []string{
			"automation",
			"database",
			"deployment",
			"documentation",
			"game-development",
			"git",
			"git-workflow",
			"nextjs-vercel",
			"orchestration",
			"performance",
			"project-management",
			"security",
			"setup",
			"simulation",
			"svelte",
			"sync",
			"team",
			"testing",
			"utilities",
		}

		for _, category := range categories {
			githubPath = fmt.Sprintf("components/commands/%s/%s.md", category, commandName)
			content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
			if err == nil {
				goto Success
			}
		}
	}

	// All attempts failed
	return fmt.Errorf("command '%s' not found (tried multiple paths)", commandName)

Success:
	// Display preview
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📄 Command: %s\n", commandName)
	fmt.Printf("🔗 Source: %s\n", githubPath)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Println(content)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return nil
}

// PreviewMultipleCommands previews multiple commands
func (ci *CommandInstaller) PreviewMultipleCommands(commandNames []string) error {
	successCount := 0
	failedCount := 0

	for _, commandName := range commandNames {
		if err := ci.PreviewCommand(commandName); err != nil {
			fmt.Printf("❌ Failed to preview command '%s': %v\n", commandName, err)
			failedCount++
		} else {
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("all command previews failed")
	}

	return nil
}

// RemoveCommand removes an installed command
func (ci *CommandInstaller) RemoveCommand(commandName, targetDir string, silent bool) error {
	if !silent {
		fmt.Printf("🗑️  Removing command: %s\n", commandName)
	}

	// Extract filename if category path provided
	var fileName string
	if strings.Contains(commandName, "/") {
		parts := strings.Split(commandName, "/")
		fileName = parts[len(parts)-1]
	} else {
		fileName = commandName
	}

	// Check project installation
	projectFile := filepath.Join(targetDir, ".claude", "commands", fileName+".md")
	projectExists := false
	if _, err := os.Stat(projectFile); err == nil {
		projectExists = true
	}

	// Check global installation
	homeDir, _ := os.UserHomeDir()
	globalFile := filepath.Join(homeDir, ".claude", "commands", fileName+".md")
	globalExists := false
	if _, err := os.Stat(globalFile); err == nil {
		globalExists = true
	}

	if !projectExists && !globalExists {
		return fmt.Errorf("command '%s' is not installed", commandName)
	}

	// Remove project installation if exists
	if projectExists {
		if err := os.Remove(projectFile); err != nil {
			return fmt.Errorf("failed to remove project command: %w", err)
		}
		if !silent {
			fmt.Printf("✅ Removed from project: %s\n", projectFile)
		}
	}

	// Remove global installation if exists
	if globalExists {
		if err := os.Remove(globalFile); err != nil {
			return fmt.Errorf("failed to remove global command: %w", err)
		}
		if !silent {
			fmt.Printf("✅ Removed from global: %s\n", globalFile)
		}
	}

	if !silent {
		fmt.Printf("✅ Command '%s' removed successfully!\n", commandName)
	}

	return nil
}
