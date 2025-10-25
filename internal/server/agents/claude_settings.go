package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/schlunsen/claude-control-terminal/internal/logging"
)

// ClaudeSettings represents the structure of .claude/settings.local.json
type ClaudeSettings struct {
	Permissions struct {
		Allow []string `json:"allow"`
	} `json:"permissions"`
	EnableAllProjectMcpServers bool     `json:"enableAllProjectMcpServers,omitempty"`
	EnabledMcpjsonServers      []string `json:"enabledMcpjsonServers,omitempty"`
	Hooks                      any      `json:"hooks,omitempty"`
}

// ClaudeSettingsManager manages .claude/settings.local.json
type ClaudeSettingsManager struct {
	workingDir string
	mu         sync.RWMutex
}

// NewClaudeSettingsManager creates a new settings manager for the given working directory
func NewClaudeSettingsManager(workingDir string) *ClaudeSettingsManager {
	return &ClaudeSettingsManager{
		workingDir: workingDir,
	}
}

// getSettingsPath returns the path to settings.local.json
func (csm *ClaudeSettingsManager) getSettingsPath() string {
	return filepath.Join(csm.workingDir, ".claude", "settings.local.json")
}

// LoadSettings loads settings from .claude/settings.local.json
func (csm *ClaudeSettingsManager) LoadSettings() (*ClaudeSettings, error) {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	settingsPath := csm.getSettingsPath()

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// Return empty settings if file doesn't exist
		return &ClaudeSettings{}, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	return &settings, nil
}

// SaveSettings saves settings to .claude/settings.local.json
func (csm *ClaudeSettingsManager) SaveSettings(settings *ClaudeSettings) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	settingsPath := csm.getSettingsPath()

	// Ensure .claude directory exists
	claudeDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Marshal with pretty printing
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to file
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	logging.Info("✅ Saved settings to %s", settingsPath)
	return nil
}

// AddPermission adds a permission string to the allow list
func (csm *ClaudeSettingsManager) AddPermission(permission string) error {
	settings, err := csm.LoadSettings()
	if err != nil {
		return err
	}

	// Check if permission already exists
	for _, p := range settings.Permissions.Allow {
		if p == permission {
			logging.Info("Permission already exists: %s", permission)
			return nil
		}
	}

	// Add permission
	settings.Permissions.Allow = append(settings.Permissions.Allow, permission)

	return csm.SaveSettings(settings)
}

// RemovePermission removes a permission string from the allow list
func (csm *ClaudeSettingsManager) RemovePermission(permission string) error {
	settings, err := csm.LoadSettings()
	if err != nil {
		return err
	}

	// Filter out the permission
	newAllowList := []string{}
	found := false
	for _, p := range settings.Permissions.Allow {
		if p != permission {
			newAllowList = append(newAllowList, p)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("permission not found: %s", permission)
	}

	settings.Permissions.Allow = newAllowList

	return csm.SaveSettings(settings)
}

// GetAllowedPermissions returns the list of allowed permissions
func (csm *ClaudeSettingsManager) GetAllowedPermissions() ([]string, error) {
	settings, err := csm.LoadSettings()
	if err != nil {
		return nil, err
	}

	return settings.Permissions.Allow, nil
}

// FormatPermissionString formats a tool and pattern into Claude Desktop permission format
// Examples:
//   - "Bash", "*" -> "Bash(*)"
//   - "Write", "*" -> "Write(*)"
//   - "Read", "/path/to/dir/*" -> "Read(//path/to/dir/**)"
func FormatPermissionString(toolName string, pattern *RulePattern) string {
	if pattern == nil {
		// No pattern, allow all
		return fmt.Sprintf("%s(*)", toolName)
	}

	switch toolName {
	case "Bash":
		if pattern.CommandPrefix != nil && *pattern.CommandPrefix == "*" {
			return "Bash(*)"
		}
		if pattern.CommandPrefix != nil {
			return fmt.Sprintf("Bash(%s:*)", *pattern.CommandPrefix)
		}

	case "Read", "Write", "Edit":
		if pattern.DirectoryPath != nil {
			dirPath := *pattern.DirectoryPath
			// Handle wildcard patterns
			if dirPath == "*" {
				// Legacy single wildcard - convert to proper format
				return fmt.Sprintf("%s(/**)", toolName)
			}
			if dirPath == "/**" {
				// Recursive wildcard from root - this is the correct format
				return fmt.Sprintf("%s(/**)", toolName)
			}
			// Specific directory path - convert to absolute path with double slash prefix
			return fmt.Sprintf("%s(//%s/**)", toolName, dirPath)
		}

	case "Grep", "Glob":
		if pattern.PathPattern != nil && *pattern.PathPattern == "*" {
			return fmt.Sprintf("%s(*)", toolName)
		}
		if pattern.PathPattern != nil {
			return fmt.Sprintf("%s(//%s)", toolName, *pattern.PathPattern)
		}
	}

	// Fallback: allow all for this tool
	return fmt.Sprintf("%s(*)", toolName)
}

// ParsePermissionString parses a Claude Desktop permission string
// Examples:
//   - "Bash(*)" -> ("Bash", wildcard pattern)
//   - "Write(//path/**)" -> ("Write", directory pattern)
//   - "Bash(git:*)" -> ("Bash", command prefix pattern)
func ParsePermissionString(permStr string) (toolName string, pattern *RulePattern, err error) {
	// Simple parser for Claude Desktop format
	// Format: ToolName(pattern)

	if len(permStr) < 3 {
		return "", nil, fmt.Errorf("invalid permission string: %s", permStr)
	}

	// Find opening parenthesis
	openParen := -1
	for i, c := range permStr {
		if c == '(' {
			openParen = i
			break
		}
	}

	if openParen == -1 || permStr[len(permStr)-1] != ')' {
		return "", nil, fmt.Errorf("invalid permission format: %s", permStr)
	}

	toolName = permStr[:openParen]
	patternStr := permStr[openParen+1 : len(permStr)-1]

	pattern = &RulePattern{}

	// Handle wildcard
	if patternStr == "*" {
		switch toolName {
		case "Bash":
			pattern.CommandPrefix = stringPtr("*")
		case "Read", "Write", "Edit":
			pattern.DirectoryPath = stringPtr("*")
			pattern.FilePathPattern = stringPtr("*")
		case "Grep", "Glob":
			pattern.PathPattern = stringPtr("*")
		}
		return toolName, pattern, nil
	}

	// Handle command prefix (e.g., "git:*")
	if toolName == "Bash" && len(patternStr) > 2 && patternStr[len(patternStr)-2:] == ":*" {
		prefix := patternStr[:len(patternStr)-2]
		pattern.CommandPrefix = &prefix
		return toolName, pattern, nil
	}

	// Handle file paths (e.g., "//path/**")
	if len(patternStr) > 2 && patternStr[:2] == "//" {
		cleanPath := patternStr[2:] // Remove leading //
		// Remove trailing /** if present
		if len(cleanPath) > 3 && cleanPath[len(cleanPath)-3:] == "/**" {
			cleanPath = cleanPath[:len(cleanPath)-3]
		}

		switch toolName {
		case "Read", "Write", "Edit":
			pattern.DirectoryPath = &cleanPath
			pattern.FilePathPattern = stringPtr(cleanPath + "/*")
		case "Grep", "Glob":
			pattern.PathPattern = &cleanPath
		}
		return toolName, pattern, nil
	}

	return toolName, pattern, nil
}
