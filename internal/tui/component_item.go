package tui

import (
	"os"
	"path/filepath"
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
	meta, ok := metadata[componentType]
	if !ok {
		return false, false
	}

	// Determine subdirectory based on component type
	var subDir string
	switch componentType {
	case "agent":
		subDir = "agents"
	case "command":
		subDir = "commands"
	case "mcp":
		subDir = "mcp"
	default:
		return false, false
	}

	fileName := componentName + meta.Extension

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
