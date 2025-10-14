// Package components provides MCP (Model Context Protocol) server installation.
// This file handles MCP component installation, downloading MCP configurations from GitHub
// and registering them in claude_desktop_config.json for both project and user scopes.
package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

// ParseMCPScope converts a string scope to an MCPScope type
func ParseMCPScope(scopeStr string) fileops.MCPScope {
	switch strings.ToLower(strings.TrimSpace(scopeStr)) {
	case "user", "global":
		return fileops.MCPScopeUser
	case "project", "local", "":
		return fileops.MCPScopeProject
	default:
		// Default to project scope
		return fileops.MCPScopeProject
	}
}

// MCPInstaller handles MCP component installation
type MCPInstaller struct {
	config *fileops.GitHubConfig
	scope  fileops.MCPScope
}

// NewMCPInstaller creates a new MCP installer
func NewMCPInstaller(scope fileops.MCPScope) *MCPInstaller {
	return &MCPInstaller{
		config: fileops.DefaultGitHubConfig(),
		scope:  scope,
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

	// Parse the MCP JSON and merge into the appropriate config file
	serverNames, err := fileops.MergeMCPServersFromJSON(mi.scope, targetDir, content)
	if err != nil {
		return fmt.Errorf("failed to register MCP: %w", err)
	}

	// Record installation metadata
	if err := fileops.AddMCPInstallation(mi.scope, targetDir, mcpName, serverNames, githubPath); err != nil {
		// Log warning but don't fail installation
		if !silent {
			fmt.Printf("⚠️  Warning: Failed to save installation metadata: %v\n", err)
		}
	}

	if !silent {
		configPath := fileops.GetMCPConfigPath(mi.scope, targetDir)
		relPath, _ := filepath.Rel(targetDir, configPath)

		fmt.Printf("✅ MCP '%s' installed successfully!\n", mcpName)
		for _, serverName := range serverNames {
			fmt.Printf("   🔌 Registered server: %s\n", serverName)
		}

		scopeName := "project"
		if mi.scope == fileops.MCPScopeUser {
			scopeName = "user (global)"
		}
		fmt.Printf("📁 Scope: %s\n", scopeName)
		fmt.Printf("📄 Config: %s\n", relPath)
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

// PreviewMCP previews a specific MCP component without installing
func (mi *MCPInstaller) PreviewMCP(mcpName string) error {
	// Try multiple path formats to find the MCP
	var content string
	var err error
	var githubPath string

	// Format 1: Try with category if provided
	if strings.Contains(mcpName, "/") {
		githubPath = fmt.Sprintf("components/mcps/%s.json", mcpName)
		content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
		if err == nil {
			goto Success
		}
	}

	// Format 2: Try direct path (most common)
	githubPath = fmt.Sprintf("components/mcps/%s.json", mcpName)
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
			content, err = fileops.DownloadFileFromGitHub(mi.config, githubPath, 0)
			if err == nil {
				goto Success
			}
		}
	}

	// All attempts failed
	return fmt.Errorf("MCP '%s' not found (tried multiple paths)", mcpName)

Success:
	// Display preview
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📄 MCP: %s\n", mcpName)
	fmt.Printf("🔗 Source: %s\n", githubPath)

	scopeName := "project"
	if mi.scope == fileops.MCPScopeUser {
		scopeName = "user (global)"
	}
	fmt.Printf("📁 Scope: %s\n", scopeName)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Println(content)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return nil
}

// PreviewMultipleMCPs previews multiple MCPs
func (mi *MCPInstaller) PreviewMultipleMCPs(mcpNames []string) error {
	successCount := 0
	failedCount := 0

	for _, mcpName := range mcpNames {
		if err := mi.PreviewMCP(mcpName); err != nil {
			fmt.Printf("❌ Failed to preview MCP '%s': %v\n", mcpName, err)
			failedCount++
		} else {
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("all MCP previews failed")
	}

	return nil
}

// RemoveMCP removes an installed MCP by removing its server entries from the config
func (mi *MCPInstaller) RemoveMCP(mcpName, targetDir string, silent bool) error {
	if !silent {
		fmt.Printf("🗑️  Removing MCP: %s\n", mcpName)
	}

	var removed []string
	var err error

	// Strategy 1: Try to use metadata for exact match
	installation, metadataErr := fileops.GetMCPInstallation(mi.scope, targetDir, mcpName)
	if metadataErr == nil && installation != nil {
		// Found in metadata - remove exact server keys
		if !silent {
			fmt.Printf("📋 Found in metadata: %d server(s) registered\n", len(installation.ServerKeys))
		}

		configPath := fileops.GetMCPConfigPath(mi.scope, targetDir)
		config, err := fileops.LoadMCPConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Remove each server key from the installation
		for _, serverKey := range installation.ServerKeys {
			if _, exists := config.MCPServers[serverKey]; exists {
				delete(config.MCPServers, serverKey)
				removed = append(removed, serverKey)
			}
		}

		// Save updated config
		if len(removed) > 0 {
			if err := fileops.SaveMCPConfig(configPath, config); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
		}

		// Remove metadata entry
		if err := fileops.RemoveMCPInstallation(mi.scope, targetDir, mcpName); err != nil {
			// Log warning but don't fail
			if !silent {
				fmt.Printf("⚠️  Warning: Failed to remove metadata: %v\n", err)
			}
		}
	} else {
		// Strategy 2: Fallback to substring matching (for legacy installs or if metadata missing)
		if !silent {
			fmt.Printf("📋 Metadata not found, using pattern matching fallback\n")
		}

		removed, err = fileops.RemoveMCPServers(mi.scope, targetDir, mcpName)
		if err != nil {
			return fmt.Errorf("failed to remove MCP: %w", err)
		}
	}

	if len(removed) == 0 {
		return fmt.Errorf("MCP '%s' is not installed or no matching servers found", mcpName)
	}

	if !silent {
		configPath := fileops.GetMCPConfigPath(mi.scope, targetDir)
		fmt.Printf("✅ Removed %d server(s) from config:\n", len(removed))
		for _, serverName := range removed {
			fmt.Printf("   🔌 %s\n", serverName)
		}

		scopeName := "project"
		if mi.scope == fileops.MCPScopeUser {
			scopeName = "user (global)"
		}
		fmt.Printf("📁 Scope: %s\n", scopeName)
		fmt.Printf("📄 Config: %s\n", configPath)
		fmt.Printf("✅ MCP '%s' removed successfully!\n", mcpName)
	}

	return nil
}
