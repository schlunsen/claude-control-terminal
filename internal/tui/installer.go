package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

// AgentInstallerForTUI handles agent installation in TUI mode
type AgentInstallerForTUI struct {
	config *fileops.GitHubConfig
}

// NewAgentInstallerForTUI creates a new agent installer for TUI
func NewAgentInstallerForTUI() *AgentInstallerForTUI {
	return &AgentInstallerForTUI{
		config: fileops.DefaultGitHubConfig(),
	}
}

// InstallAgent installs a specific agent component
func (ai *AgentInstallerForTUI) InstallAgent(agentName, category, targetDir string) error {
	var content string
	var err error
	var githubPath string

	// Try category path first if provided
	if category != "" && category != "root" {
		githubPath = fmt.Sprintf("components/agents/%s/%s.md", category, agentName)
		content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Try direct path
	githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
	content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
	if err != nil {
		return fmt.Errorf("failed to download agent: %w", err)
	}

Success:
	// Create .claude/agents directory
	agentsDir := filepath.Join(targetDir, ".claude", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Write the agent file
	targetFile := filepath.Join(agentsDir, agentName+".md")
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	return nil
}

// PreviewAgent previews a specific agent component
func (ai *AgentInstallerForTUI) PreviewAgent(agentName, category string) (string, error) {
	var content string
	var err error
	var githubPath string

	// Try category path first if provided
	if category != "" && category != "root" {
		githubPath = fmt.Sprintf("components/agents/%s/%s.md", category, agentName)
		content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
		if err == nil {
			return content, nil
		}
	}

	// Try direct path
	githubPath = fmt.Sprintf("components/agents/%s.md", agentName)
	content, err = fileops.DownloadFileFromGitHub(ai.config, githubPath, 0)
	if err != nil {
		return "", fmt.Errorf("failed to download agent: %w", err)
	}

	return content, nil
}

// CommandInstallerForTUI handles command installation in TUI mode
type CommandInstallerForTUI struct {
	config *fileops.GitHubConfig
}

// NewCommandInstallerForTUI creates a new command installer for TUI
func NewCommandInstallerForTUI() *CommandInstallerForTUI {
	return &CommandInstallerForTUI{
		config: fileops.DefaultGitHubConfig(),
	}
}

// InstallCommand installs a specific command component
func (ci *CommandInstallerForTUI) InstallCommand(commandName, category, targetDir string) error {
	var content string
	var err error
	var githubPath string

	// Try category path first if provided
	if category != "" && category != "root" {
		githubPath = fmt.Sprintf("components/commands/%s/%s.md", category, commandName)
		content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Try direct path
	githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
	content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
	if err != nil {
		return fmt.Errorf("failed to download command: %w", err)
	}

Success:
	// Create .claude/commands directory
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

	return nil
}

// PreviewCommand previews a specific command component
func (ci *CommandInstallerForTUI) PreviewCommand(commandName, category string) (string, error) {
	var content string
	var err error
	var githubPath string

	// Try category path first if provided
	if category != "" && category != "root" {
		githubPath = fmt.Sprintf("components/commands/%s/%s.md", category, commandName)
		content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
		if err == nil {
			return content, nil
		}
	}

	// Try direct path
	githubPath = fmt.Sprintf("components/commands/%s.md", commandName)
	content, err = fileops.DownloadFileFromGitHub(ci.config, githubPath, 0)
	if err != nil {
		return "", fmt.Errorf("failed to download command: %w", err)
	}

	return content, nil
}

// MCPInstallerForTUI handles MCP installation in TUI mode
type MCPInstallerForTUI struct {
	config *fileops.GitHubConfig
}

// NewMCPInstallerForTUI creates a new MCP installer for TUI
func NewMCPInstallerForTUI() *MCPInstallerForTUI {
	return &MCPInstallerForTUI{
		config: fileops.DefaultGitHubConfig(),
	}
}

// InstallMCP installs a specific MCP component
func (mi *MCPInstallerForTUI) InstallMCP(mcpName, category, targetDir string) error {
	var content string
	var err error
	var githubPath string

	// Try category path first if provided
	if category != "" && category != "root" {
		githubPath = fmt.Sprintf("components/mcps/%s/%s.json", category, mcpName)
		content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Try direct path
	githubPath = fmt.Sprintf("components/mcps/%s.json", mcpName)
	content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
	if err != nil {
		return fmt.Errorf("failed to download MCP: %w", err)
	}

Success:
	// Create .claude/mcp directory
	mcpDir := filepath.Join(targetDir, ".claude", "mcp")
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		return fmt.Errorf("failed to create MCP directory: %w", err)
	}

	// Write the MCP file
	targetFile := filepath.Join(mcpDir, mcpName+".json")
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write MCP file: %w", err)
	}

	// Register the MCP servers in .mcp.json
	serverKeys, err := fileops.MergeMCPServersFromJSON(fileops.MCPScopeProject, targetDir, content)
	if err != nil {
		return fmt.Errorf("failed to register MCP in .mcp.json: %w", err)
	}

	// Record installation metadata
	if err := fileops.AddMCPInstallation(fileops.MCPScopeProject, targetDir, mcpName, serverKeys, githubPath); err != nil {
		// Log warning but don't fail installation
		fmt.Printf("Warning: Failed to save installation metadata: %v\n", err)
	}

	return nil
}

// PreviewMCP previews a specific MCP component
func (mi *MCPInstallerForTUI) PreviewMCP(mcpName, category string) (string, error) {
	var content string
	var err error
	var githubPath string

	// Try category path first if provided
	if category != "" && category != "root" {
		githubPath = fmt.Sprintf("components/mcps/%s/%s.json", category, mcpName)
		content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
		if err == nil {
			return content, nil
		}
	}

	// Try direct path
	githubPath = fmt.Sprintf("components/mcps/%s.json", mcpName)
	content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
	if err != nil {
		return "", fmt.Errorf("failed to download MCP: %w", err)
	}

	return content, nil
}

// RemoveAgent removes an installed agent
// Returns nil if the agent is not installed (silently skips)
func (ai *AgentInstallerForTUI) RemoveAgent(agentName, targetDir string) error {
	// Check project installation
	projectFile := filepath.Join(targetDir, ".claude", "agents", agentName+".md")
	projectExists := false
	if _, err := os.Stat(projectFile); err == nil {
		projectExists = true
	}

	// Check global installation
	homeDir, _ := os.UserHomeDir()
	globalFile := filepath.Join(homeDir, ".claude", "agents", agentName+".md")
	globalExists := false
	if _, err := os.Stat(globalFile); err == nil {
		globalExists = true
	}

	// If not installed anywhere, silently skip (return nil)
	if !projectExists && !globalExists {
		return nil
	}

	// Remove project installation if exists
	if projectExists {
		if err := os.Remove(projectFile); err != nil {
			return fmt.Errorf("failed to remove project agent: %w", err)
		}
	}

	// Remove global installation if exists
	if globalExists {
		if err := os.Remove(globalFile); err != nil {
			return fmt.Errorf("failed to remove global agent: %w", err)
		}
	}

	return nil
}

// RemoveCommand removes an installed command
// Returns nil if the command is not installed (silently skips)
func (ci *CommandInstallerForTUI) RemoveCommand(commandName, targetDir string) error {
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

	// If not installed anywhere, silently skip (return nil)
	if !projectExists && !globalExists {
		return nil
	}

	// Remove project installation if exists
	if projectExists {
		if err := os.Remove(projectFile); err != nil {
			return fmt.Errorf("failed to remove project command: %w", err)
		}
	}

	// Remove global installation if exists
	if globalExists {
		if err := os.Remove(globalFile); err != nil {
			return fmt.Errorf("failed to remove global command: %w", err)
		}
	}

	return nil
}

// RemoveMCP removes an installed MCP
// Returns nil if the MCP is not installed (silently skips)
func (mi *MCPInstallerForTUI) RemoveMCP(mcpName, targetDir string) error {
	// Strategy 1: Try to use metadata for exact match
	installation, metadataErr := fileops.GetMCPInstallation(fileops.MCPScopeProject, targetDir, mcpName)
	if metadataErr == nil && installation != nil {
		// Found in metadata - remove exact server keys from .mcp.json
		configPath := fileops.GetMCPConfigPath(fileops.MCPScopeProject, targetDir)
		config, err := fileops.LoadMCPConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Remove each server key from the installation
		for _, serverKey := range installation.ServerKeys {
			delete(config.MCPServers, serverKey)
		}

		// Save updated config
		if err := fileops.SaveMCPConfig(configPath, config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// Remove metadata entry
		if err := fileops.RemoveMCPInstallation(fileops.MCPScopeProject, targetDir, mcpName); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: Failed to remove metadata: %v\n", err)
		}
	} else {
		// Strategy 2: Fallback to file-based removal (legacy support)
		// Check if MCP JSON file exists in .claude/mcp directory
		mcpFile := filepath.Join(targetDir, ".claude", "mcp", mcpName+".json")
		content, err := os.ReadFile(mcpFile)
		if err != nil {
			if os.IsNotExist(err) {
				// MCP not installed, silently skip
				return nil
			}
			return fmt.Errorf("failed to read MCP file: %w", err)
		}

		// Remove servers from .mcp.json by reading the exact servers from the MCP file
		removed, err := fileops.RemoveMCPServersByContent(fileops.MCPScopeProject, targetDir, string(content))
		if err != nil {
			return fmt.Errorf("failed to remove MCP from config: %w", err)
		}

		// If no servers were removed, the MCP might not be in .mcp.json (but file exists)
		// This is okay, we'll still remove the file
		_ = removed
	}

	// Remove the MCP JSON file from .claude/mcp directory (if it exists)
	mcpFile := filepath.Join(targetDir, ".claude", "mcp", mcpName+".json")
	if err := os.Remove(mcpFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove MCP file: %w", err)
	}

	return nil
}
