package fileops

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MCPScope represents where the MCP configuration should be stored
type MCPScope string

const (
	// MCPScopeProject stores MCP in .mcp.json (default, project-specific)
	MCPScopeProject MCPScope = "project"
	// MCPScopeUser stores MCP in ~/.claude/config.json (global for user)
	MCPScopeUser MCPScope = "user"
)

// ClaudeConfig represents the structure of Claude Code's config files
type ClaudeConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers,omitempty"`
}

// MCPServerConfig represents a single MCP server configuration
type MCPServerConfig struct {
	Description string            `json:"description,omitempty"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	URL         string            `json:"url,omitempty"`
	Transport   string            `json:"transport,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

// LoadMCPConfig loads an MCP configuration file from the given path
func LoadMCPConfig(configPath string) (*ClaudeConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &ClaudeConfig{
				MCPServers: make(map[string]MCPServerConfig),
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize MCPServers map if it's nil
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServerConfig)
	}

	return &config, nil
}

// SaveMCPConfig saves an MCP configuration file to the given path
func SaveMCPConfig(configPath string, config *ClaudeConfig) error {
	// Ensure the directory exists first
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// If the config is empty (no MCP servers), handle specially
	if len(config.MCPServers) == 0 {
		// For .mcp.json files (project scope), delete the file if it exists
		if filepath.Base(configPath) == ".mcp.json" {
			if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove empty config file: %w", err)
			}
			return nil
		}
		// For other config files (like ~/.claude/config.json), ensure empty map exists
		// This maintains valid JSON structure
		if config.MCPServers == nil {
			config.MCPServers = make(map[string]MCPServerConfig)
		}
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetMCPConfigPath returns the path for MCP configuration based on scope
func GetMCPConfigPath(scope MCPScope, projectDir string) string {
	switch scope {
	case MCPScopeUser:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ".claude/config.json"
		}
		return filepath.Join(homeDir, ".claude", "config.json")
	case MCPScopeProject:
		fallthrough
	default:
		return filepath.Join(projectDir, ".mcp.json")
	}
}

// AddMCPServerToConfig adds an MCP server to the specified configuration file
func AddMCPServerToConfig(scope MCPScope, projectDir string, name string, serverConfig MCPServerConfig) error {
	configPath := GetMCPConfigPath(scope, projectDir)

	config, err := LoadMCPConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Add or update the MCP server
	config.MCPServers[name] = serverConfig

	if err := SaveMCPConfig(configPath, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RemoveMCPServerFromConfig removes an MCP server from the specified configuration file
func RemoveMCPServerFromConfig(scope MCPScope, projectDir string, name string) error {
	configPath := GetMCPConfigPath(scope, projectDir)

	config, err := LoadMCPConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	delete(config.MCPServers, name)

	if err := SaveMCPConfig(configPath, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// MergeMCPServersFromJSON parses an MCP JSON file and adds its servers to the config
func MergeMCPServersFromJSON(scope MCPScope, projectDir string, mcpJSONContent string) ([]string, error) {
	// Parse the MCP JSON file
	var mcpConfig ClaudeConfig
	if err := json.Unmarshal([]byte(mcpJSONContent), &mcpConfig); err != nil {
		return nil, fmt.Errorf("failed to parse MCP JSON: %w", err)
	}

	configPath := GetMCPConfigPath(scope, projectDir)

	// Load current config
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Track added server names
	addedServers := make([]string, 0, len(mcpConfig.MCPServers))

	// Merge all MCP servers from the downloaded config
	for name, server := range mcpConfig.MCPServers {
		config.MCPServers[name] = server
		addedServers = append(addedServers, name)
	}

	// Save updated config
	if err := SaveMCPConfig(configPath, config); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return addedServers, nil
}

// RemoveMCPServers removes MCP servers that match the given MCP name pattern
// It returns a list of removed server names
func RemoveMCPServers(scope MCPScope, projectDir string, mcpName string) ([]string, error) {
	configPath := GetMCPConfigPath(scope, projectDir)

	// Load current config
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Track removed server names
	removed := []string{}

	// Remove servers that match the MCP name using bidirectional matching
	mcpNameLower := strings.ToLower(mcpName)
	for serverName := range config.MCPServers {
		serverNameLower := strings.ToLower(serverName)

		// Bidirectional substring matching:
		// - "postgresql-integration" matches "postgresql"
		// - "postgresql" matches "postgresql-integration"
		// - "github" matches "GitHub"
		if mcpNamesMatch(mcpNameLower, serverNameLower) {
			delete(config.MCPServers, serverName)
			removed = append(removed, serverName)
		}
	}

	// Save updated config if any servers were removed
	if len(removed) > 0 {
		if err := SaveMCPConfig(configPath, config); err != nil {
			return removed, fmt.Errorf("failed to save config: %w", err)
		}
	}

	return removed, nil
}

// mcpNamesMatch checks if an MCP name matches a server name using bidirectional substring matching
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

// RemoveMCPServersByContent removes MCP servers by parsing the MCP JSON content
// and removing the exact servers defined in that content from the config
func RemoveMCPServersByContent(scope MCPScope, projectDir string, mcpJSONContent string) ([]string, error) {
	// Parse the MCP JSON file to get server names
	var mcpConfig ClaudeConfig
	if err := json.Unmarshal([]byte(mcpJSONContent), &mcpConfig); err != nil {
		return nil, fmt.Errorf("failed to parse MCP JSON: %w", err)
	}

	configPath := GetMCPConfigPath(scope, projectDir)

	// Load current config
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Track removed server names
	removed := []string{}

	// Remove exact servers from the MCP file
	for serverName := range mcpConfig.MCPServers {
		if _, exists := config.MCPServers[serverName]; exists {
			delete(config.MCPServers, serverName)
			removed = append(removed, serverName)
		}
	}

	// Save updated config if any servers were removed
	if len(removed) > 0 {
		if err := SaveMCPConfig(configPath, config); err != nil {
			return removed, fmt.Errorf("failed to save config: %w", err)
		}
	}

	return removed, nil
}
