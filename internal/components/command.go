package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davila7/go-claude-templates/internal/fileops"
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

	// Support both category/command-name and direct command-name formats
	var githubPath string
	if strings.Contains(commandName, "/") {
		// Category/command format: security/vulnerability-scan
		githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
	} else {
		// Direct command format: check-file
		githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
	}

	if !silent {
		fmt.Println("📥 Downloading from GitHub (main branch)...")
	}

	// Download the command file
	content, err := fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return fmt.Errorf("command '%s' not found", commandName)
		}
		return fmt.Errorf("failed to download command: %w", err)
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
