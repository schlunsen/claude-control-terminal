package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davila7/go-claude-templates/internal/fileops"
)

// MCPInstaller handles MCP component installation
type MCPInstaller struct {
	config *fileops.GitHubConfig
}

// NewMCPInstaller creates a new MCP installer
func NewMCPInstaller() *MCPInstaller {
	return &MCPInstaller{
		config: fileops.DefaultGitHubConfig(),
	}
}

// InstallMCP installs a specific MCP component
func (mi *MCPInstaller) InstallMCP(mcpName, targetDir string, silent bool) error {
	if !silent {
		fmt.Printf("🔌 Installing MCP: %s\n", mcpName)
	}

	// Try multiple path formats to find the MCP
	var content string
	var err error
	var githubPath string

	// Format 1: Try with category if provided
	if strings.Contains(mcpName, "/") {
		githubPath = fmt.Sprintf("components/mcps/%s.json", mcpName)
		if !silent {
			fmt.Printf("📥 Trying path: %s\n", githubPath)
		}
		content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Format 2: Try direct path (most common)
	githubPath = fmt.Sprintf("components/mcps/%s.json", mcpName)
	if !silent {
		fmt.Printf("📥 Trying direct path: %s\n", githubPath)
	}
	content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
	if err == nil {
		goto Success
	}

	// Format 3: Search in common categories
	if !strings.Contains(mcpName, "/") {
		categories := []string{
			"browser_automation",
			"database",
			"deepgraph",
			"devtools",
			"filesystem",
			"integration",
			"marketing",
			"productivity",
			"web",
		}

		for _, category := range categories {
			githubPath = fmt.Sprintf("components/mcps/%s/%s.json", category, mcpName)
			if !silent {
				fmt.Printf("📥 Searching in %s category...\n", category)
			}
			content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
			if err == nil {
				goto Success
			}
		}
	}

	// All attempts failed
	return fmt.Errorf("MCP '%s' not found (tried multiple paths)", mcpName)

Success:
	if !silent {
		fmt.Printf("✅ Found MCP at: %s\n", githubPath)
	}

	// MCP files go to .claude/mcp/ directory
	mcpDir := filepath.Join(targetDir, ".claude", "mcp")
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		return fmt.Errorf("failed to create MCP directory: %w", err)
	}

	targetFile := filepath.Join(mcpDir, mcpName+".json")
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write MCP file: %w", err)
	}

	if !silent {
		relPath, _ := filepath.Rel(targetDir, targetFile)
		fmt.Printf("✅ MCP '%s' installed successfully!\n", mcpName)
		fmt.Printf("📁 Installed to: %s\n", relPath)
	}

	return nil
}

// InstallMultipleMCPs installs multiple MCPs
func (mi *MCPInstaller) InstallMultipleMCPs(mcpNames []string, targetDir string, silent bool) error {
	successCount := 0
	failedCount := 0

	for _, mcpName := range mcpNames {
		if err := mi.InstallMCP(mcpName, targetDir, silent); err != nil {
			fmt.Printf("❌ Failed to install MCP '%s': %v\n", mcpName, err)
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
		return fmt.Errorf("all MCP installations failed")
	}

	return nil
}
