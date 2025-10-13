package fileops

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PermissionMode represents the default permission behavior
type PermissionMode string

const (
	PermissionModeDefault           PermissionMode = "default"
	PermissionModeAcceptEdits       PermissionMode = "acceptEdits"
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
	PermissionModePlan              PermissionMode = "plan"
)

// PermissionsConfig represents the permissions object in Claude Code settings
type PermissionsConfig struct {
	Allow       []string       `json:"allow,omitempty"`
	Ask         []string       `json:"ask,omitempty"`
	Deny        []string       `json:"deny,omitempty"`
	DefaultMode PermissionMode `json:"defaultMode,omitempty"`
}

// ClaudeSettings represents the complete Claude Code settings.json structure
type ClaudeSettings struct {
	Schema                    string            `json:"$schema,omitempty"`
	AlwaysThinkingEnabled     bool              `json:"alwaysThinkingEnabled,omitempty"`
	FeedbackSurveyState       interface{}       `json:"feedbackSurveyState,omitempty"`
	Permissions               *PermissionsConfig `json:"permissions,omitempty"`
	EnableAllProjectMcpServers bool              `json:"enableAllProjectMcpServers,omitempty"`
	EnabledMcpjsonServers     []string          `json:"enabledMcpjsonServers,omitempty"`
	DisabledMcpjsonServers    []string          `json:"disabledMcpjsonServers,omitempty"`
}

// PermissionItem represents a toggleable permission in the UI
type PermissionItem struct {
	Name        string
	Description string
	Pattern     string   // The permission rule pattern (e.g., "Bash(git *)")
	Patterns    []string // Multiple patterns for a single item
	IsMode      bool     // If true, this sets defaultMode instead of adding to allow list
	ModeValue   PermissionMode
	Category    string
}

// GetDefaultPermissionItems returns the list of configurable permissions
func GetDefaultPermissionItems() []PermissionItem {
	return []PermissionItem{
		{
			Name:        "Git Commands",
			Description: "Allow all git commands",
			Patterns:    []string{"Bash(git *)", "Bash(git add:*)", "Bash(git commit:*)", "Bash(git push:*)"},
			Category:    "bash",
		},
		{
			Name:        "Just Commands",
			Description: "Allow all just task runner commands",
			Patterns:    []string{"Bash(just *)"},
			Category:    "bash",
		},
		{
			Name:        "Make Commands",
			Description: "Allow all make build commands",
			Patterns:    []string{"Bash(make *)"},
			Category:    "bash",
		},
		{
			Name:        "Docker Commands",
			Description: "Allow all docker and docker-compose commands",
			Patterns:    []string{"Bash(docker *)", "Bash(docker-compose *)"},
			Category:    "bash",
		},
		{
			Name:        "NPM/Yarn Commands",
			Description: "Allow all npm and yarn package manager commands",
			Patterns:    []string{"Bash(npm *)", "Bash(yarn *)"},
			Category:    "bash",
		},
		{
			Name:        "All Bash Commands",
			Description: "Allow all bash/shell commands (use with caution)",
			Patterns:    []string{"Bash(*)"},
			Category:    "bash",
		},
		{
			Name:        "Web Fetch",
			Description: "Allow fetching content from any website",
			Patterns:    []string{"WebFetch(*)"},
			Category:    "tools",
		},
		{
			Name:        "File Read",
			Description: "Allow reading any file",
			Patterns:    []string{"Read(*)"},
			Category:    "tools",
		},
		{
			Name:        "File Edit",
			Description: "Allow editing any file",
			Patterns:    []string{"Edit(*)"},
			Category:    "tools",
		},
		{
			Name:        "File Write",
			Description: "Allow writing new files",
			Patterns:    []string{"Write(*)"},
			Category:    "tools",
		},
		{
			Name:        "Bypass All Permissions",
			Description: "Disable all permission prompts (full trust mode)",
			IsMode:      true,
			ModeValue:   PermissionModeBypassPermissions,
			Category:    "mode",
		},
	}
}

// GetClaudeSettingsPath returns the path to Claude Code local settings file
// Always uses the project's .claude/settings.local.json relative to the project directory
func GetClaudeSettingsPath(projectDir ...string) string {
	// Use provided project directory, or default to current directory
	dir := "."
	if len(projectDir) > 0 && projectDir[0] != "" {
		dir = projectDir[0]
	}

	return filepath.Join(dir, ".claude", "settings.local.json")
}

// LoadClaudeSettings loads the Claude Code local settings file
// If projectDir is provided, it loads from the project's .claude/settings.local.json
func LoadClaudeSettings(projectDir ...string) (*ClaudeSettings, error) {
	settingsPath := GetClaudeSettingsPath(projectDir...)

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default settings if file doesn't exist
			return &ClaudeSettings{
				Schema: "https://json.schemastore.org/claude-code-settings.json",
			}, nil
		}
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	return &settings, nil
}

// SaveClaudeSettings saves the Claude Code local settings file
// If projectDir is provided, it saves to the project's .claude/settings.local.json
func SaveClaudeSettings(settings *ClaudeSettings, projectDir ...string) error {
	settingsPath := GetClaudeSettingsPath(projectDir...)

	// Ensure the directory exists
	settingsDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Clean up empty permissions object
	if settings.Permissions != nil {
		// Check if permissions object is effectively empty
		isEmpty := len(settings.Permissions.Allow) == 0 &&
			len(settings.Permissions.Ask) == 0 &&
			len(settings.Permissions.Deny) == 0 &&
			(settings.Permissions.DefaultMode == "" || settings.Permissions.DefaultMode == PermissionModeDefault)

		if isEmpty {
			settings.Permissions = nil
		}
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// IsPermissionEnabled checks if a permission item is currently enabled
func IsPermissionEnabled(settings *ClaudeSettings, item PermissionItem) bool {
	if settings.Permissions == nil {
		return false
	}

	// Check if this is a mode setting
	if item.IsMode {
		return settings.Permissions.DefaultMode == item.ModeValue
	}

	// Check if all patterns are in the allow list
	if len(item.Patterns) > 0 {
		for _, pattern := range item.Patterns {
			found := false
			for _, allowed := range settings.Permissions.Allow {
				if allowed == pattern {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	// Single pattern check
	if item.Pattern != "" {
		for _, allowed := range settings.Permissions.Allow {
			if allowed == item.Pattern {
				return true
			}
		}
	}

	return false
}

// TogglePermission toggles a permission on or off
func TogglePermission(settings *ClaudeSettings, item PermissionItem, enabled bool) {
	// Initialize permissions if nil
	if settings.Permissions == nil {
		settings.Permissions = &PermissionsConfig{
			Allow: []string{},
			Ask:   []string{},
			Deny:  []string{},
		}
	}

	// Handle mode setting
	if item.IsMode {
		if enabled {
			settings.Permissions.DefaultMode = item.ModeValue
		} else {
			settings.Permissions.DefaultMode = PermissionModeDefault
		}
		return
	}

	// Handle pattern-based permissions
	patterns := item.Patterns
	if len(patterns) == 0 && item.Pattern != "" {
		patterns = []string{item.Pattern}
	}

	if enabled {
		// Add patterns to allow list (avoid duplicates)
		for _, pattern := range patterns {
			found := false
			for _, existing := range settings.Permissions.Allow {
				if existing == pattern {
					found = true
					break
				}
			}
			if !found {
				settings.Permissions.Allow = append(settings.Permissions.Allow, pattern)
			}
		}
	} else {
		// Remove patterns from allow list
		newAllow := []string{}
		for _, existing := range settings.Permissions.Allow {
			shouldKeep := true
			for _, pattern := range patterns {
				if existing == pattern {
					shouldKeep = false
					break
				}
			}
			if shouldKeep {
				newAllow = append(newAllow, existing)
			}
		}
		settings.Permissions.Allow = newAllow
	}
}

// GetPermissionSummary returns a summary of enabled permissions
func GetPermissionSummary(settings *ClaudeSettings) string {
	if settings.Permissions == nil {
		return "No permissions configured"
	}

	if settings.Permissions.DefaultMode == PermissionModeBypassPermissions {
		return "Bypass all permissions (full trust)"
	}

	allowCount := len(settings.Permissions.Allow)
	if allowCount == 0 {
		return "No permissions allowed (safe mode)"
	}

	return fmt.Sprintf("%d permission rule(s) active", allowCount)
}
