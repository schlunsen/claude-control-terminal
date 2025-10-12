package tui

// ComponentItem represents a single component that can be installed
type ComponentItem struct {
	Name        string
	Category    string
	Description string
	Type        string // "agent", "command", "mcp"
	Selected    bool
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
