package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

// ComponentItem represents a single component that can be installed
type ComponentItem struct {
	Name             string
	Category         string
	Description      string
	Type             string // "agent", "command", "mcp"
	Selected         bool
	InstalledGlobal  bool // Installed in ~/.claude/
	InstalledProject bool // Installed in project .claude/
}

// ComponentMetadata holds metadata for a component type
type ComponentMetadata struct {
	Type       string
	Icon       string
	Path       string
	Extension  string
	Categories []string
}

// GetComponentMetadata returns metadata for each component type
func GetComponentMetadata() map[string]ComponentMetadata {
	return map[string]ComponentMetadata{
		"agent": {
			Type:      "agent",
			Icon:      "🤖",
			Path:      "components/agents",
			Extension: ".md",
			Categories: []string{
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
			},
		},
		"command": {
			Type:      "command",
			Icon:      "⚡",
			Path:      "components/commands",
			Extension: ".md",
			Categories: []string{
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
			},
		},
		"mcp": {
			Type:      "mcp",
			Icon:      "🔌",
			Path:      "components/mcps",
			Extension: ".json",
			Categories: []string{
				"browser_automation",
				"database",
				"deepgraph",
				"devtools",
				"filesystem",
				"integration",
				"marketing",
				"productivity",
				"web",
			},
		},
	}
}

// CheckInstallationStatus checks if a component is installed globally and/or in project
func CheckInstallationStatus(componentName, componentType, projectDir string) (global bool, project bool) {
	metadata := GetComponentMetadata()
	_, ok := metadata[componentType]
	if !ok {
		return false, false
	}

	// MCPs use a different detection method - check config files
	if componentType == "mcp" {
		return checkMCPInstallation(componentName, projectDir)
	}

	// For agents and commands, check for individual files
	var subDir string
	switch componentType {
	case "agent":
		subDir = "agents"
	case "command":
		subDir = "commands"
	default:
		return false, false
	}

	fileName := componentName + metadata[componentType].Extension

	// Check global installation (~/.claude/)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(homeDir, ".claude", subDir, fileName)
		if _, err := os.Stat(globalPath); err == nil {
			global = true
		}
	}

	// Check project installation (projectDir/.claude/)
	projectPath := filepath.Join(projectDir, ".claude", subDir, fileName)
	if _, err := os.Stat(projectPath); err == nil {
		project = true
	}

	return global, project
}

// checkMCPInstallation checks if an MCP is installed by looking for server entries in config files
func checkMCPInstallation(mcpName, projectDir string) (global bool, project bool) {
	// Strategy 1: Check metadata first (most accurate)
	// Check global metadata
	if installation, err := fileops.GetMCPInstallation(fileops.MCPScopeUser, projectDir, mcpName); err == nil && installation != nil {
		global = true
	}

	// Check project metadata
	if installation, err := fileops.GetMCPInstallation(fileops.MCPScopeProject, projectDir, mcpName); err == nil && installation != nil {
		project = true
	}

	// Strategy 2: Fallback to config file scanning (for legacy installs without metadata)
	if !global && !project {
		mcpNameLower := strings.ToLower(mcpName)

		// Check global installation (~/.claude/config.json)
		homeDir, err := os.UserHomeDir()
		if err == nil {
			globalConfigPath := filepath.Join(homeDir, ".claude", "config.json")
			if config, err := loadMCPConfigSafe(globalConfigPath); err == nil {
				for serverName := range config.MCPServers {
					if mcpNamesMatch(mcpNameLower, strings.ToLower(serverName)) {
						global = true
						break
					}
				}
			}
		}

		// Check project installation (<projectDir>/.mcp.json)
		projectConfigPath := filepath.Join(projectDir, ".mcp.json")
		if config, err := loadMCPConfigSafe(projectConfigPath); err == nil {
			for serverName := range config.MCPServers {
				if mcpNamesMatch(mcpNameLower, strings.ToLower(serverName)) {
					project = true
					break
				}
			}
		}
	}

	return global, project
}

// mcpNamesMatch checks if an MCP name matches a server name using bidirectional substring matching
// This handles cases like:
// - "postgresql-integration" should match "postgresql"
// - "postgresql" should match "postgresql-integration"
// - "github" should match "GitHub"
func mcpNamesMatch(mcpName, serverName string) bool {
	// Exact match
	if mcpName == serverName {
		return true
	}

	// Bidirectional substring matching
	if strings.Contains(serverName, mcpName) || strings.Contains(mcpName, serverName) {
		return true
	}

	return false
}

// loadMCPConfigSafe loads an MCP config file, returning empty config on error
func loadMCPConfigSafe(configPath string) (*fileops.ClaudeConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &fileops.ClaudeConfig{
			MCPServers: make(map[string]fileops.MCPServerConfig),
		}, err
	}

	var config fileops.ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return &fileops.ClaudeConfig{
			MCPServers: make(map[string]fileops.MCPServerConfig),
		}, err
	}

	if config.MCPServers == nil {
		config.MCPServers = make(map[string]fileops.MCPServerConfig)
	}

	return &config, nil
}
