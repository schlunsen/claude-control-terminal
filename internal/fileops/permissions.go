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
			Description: "Allow all git and gh (GitHub CLI) commands",
			Patterns:    []string{"Bash(git:*)", "Bash(gh:*)"},
			Category:    "bash",
		},
		{
			Name:        "Just Commands",
			Description: "Allow all just task runner commands",
			Patterns:    []string{"Bash(just:*)"},
			Category:    "bash",
		},
		{
			Name:        "Make Commands",
			Description: "Allow all make build commands",
			Patterns:    []string{"Bash(make:*)"},
			Category:    "bash",
		},
		{
			Name:        "Docker Commands",
			Description: "Allow all docker and docker-compose commands",
			Patterns:    []string{"Bash(docker:*)", "Bash(docker-compose:*)"},
			Category:    "bash",
		},
		{
			Name:        "NPM/Yarn Commands",
			Description: "Allow all npm and yarn package manager commands",
			Patterns:    []string{"Bash(npm:*)", "Bash(yarn:*)"},
			Category:    "bash",
		},
		{
			Name:        "All Bash Commands",
			Description: "Allow all bash/shell commands (use with caution)",
			Patterns:    []string{"Bash(*)"},
			Category:    "bash",
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

// SettingsSource represents where settings are loaded from
type SettingsSource string

const (
	SettingsSourceGlobal  SettingsSource = "global"  // ~/.claude/settings.json
	SettingsSourceProject SettingsSource = "project" // .claude/settings.json
	SettingsSourceLocal   SettingsSource = "local"   // .claude/settings.local.json
)

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

// GetGlobalSettingsPath returns the path to global Claude Code settings
func GetGlobalSettingsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude", "settings.json")
}

// GetProjectSettingsPath returns the path to project Claude Code settings
func GetProjectSettingsPath(projectDir ...string) string {
	dir := "."
	if len(projectDir) > 0 && projectDir[0] != "" {
		dir = projectDir[0]
	}
	return filepath.Join(dir, ".claude", "settings.json")
}

// GetLocalSettingsPath returns the path to local Claude Code settings (gitignored)
func GetLocalSettingsPath(projectDir ...string) string {
	return GetClaudeSettingsPath(projectDir...)
}

// LoadClaudeSettings loads the Claude Code local settings file
// If projectDir is provided, it loads from the project's .claude/settings.local.json
func LoadClaudeSettings(projectDir ...string) (*ClaudeSettings, error) {
	settingsPath := GetClaudeSettingsPath(projectDir...)
	return loadSettingsFromPath(settingsPath)
}

// loadSettingsFromPath loads settings from a specific file path
func loadSettingsFromPath(settingsPath string) (*ClaudeSettings, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty settings if file doesn't exist
			return &ClaudeSettings{}, nil
		}
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Deduplicate permissions on load to clean up any existing duplicates
	if settings.Permissions != nil {
		settings.Permissions.Allow = deduplicateStringSlice(settings.Permissions.Allow)
		settings.Permissions.Ask = deduplicateStringSlice(settings.Permissions.Ask)
		settings.Permissions.Deny = deduplicateStringSlice(settings.Permissions.Deny)
	}

	return &settings, nil
}

// LoadGlobalSettings loads the global Claude Code settings
func LoadGlobalSettings() (*ClaudeSettings, error) {
	settingsPath := GetGlobalSettingsPath()
	if settingsPath == "" {
		return &ClaudeSettings{}, nil
	}
	return loadSettingsFromPath(settingsPath)
}

// LoadProjectSettings loads the project Claude Code settings
func LoadProjectSettings(projectDir ...string) (*ClaudeSettings, error) {
	settingsPath := GetProjectSettingsPath(projectDir...)
	return loadSettingsFromPath(settingsPath)
}

// LoadLocalSettings loads the local Claude Code settings (gitignored)
func LoadLocalSettings(projectDir ...string) (*ClaudeSettings, error) {
	settingsPath := GetLocalSettingsPath(projectDir...)
	return loadSettingsFromPath(settingsPath)
}

// MultiSourceSettings holds settings from all three sources
type MultiSourceSettings struct {
	Global  *ClaudeSettings
	Project *ClaudeSettings
	Local   *ClaudeSettings
}

// LoadAllSettings loads settings from all three sources (global, project, local)
func LoadAllSettings(projectDir ...string) (*MultiSourceSettings, error) {
	global, err := LoadGlobalSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load global settings: %w", err)
	}

	project, err := LoadProjectSettings(projectDir...)
	if err != nil {
		return nil, fmt.Errorf("failed to load project settings: %w", err)
	}

	local, err := LoadLocalSettings(projectDir...)
	if err != nil {
		return nil, fmt.Errorf("failed to load local settings: %w", err)
	}

	return &MultiSourceSettings{
		Global:  global,
		Project: project,
		Local:   local,
	}, nil
}

// GetEffectiveSettings returns the effective settings based on priority
// Priority: Local > Project > Global
func (m *MultiSourceSettings) GetEffectiveSettings() *ClaudeSettings {
	effective := &ClaudeSettings{}

	// Start with global, then override with project, then local
	sources := []*ClaudeSettings{m.Global, m.Project, m.Local}

	for _, settings := range sources {
		if settings == nil {
			continue
		}

		// Merge permissions
		if settings.Permissions != nil {
			if effective.Permissions == nil {
				effective.Permissions = &PermissionsConfig{}
			}

			// Override allow list
			if len(settings.Permissions.Allow) > 0 {
				effective.Permissions.Allow = settings.Permissions.Allow
			}

			// Override ask list
			if len(settings.Permissions.Ask) > 0 {
				effective.Permissions.Ask = settings.Permissions.Ask
			}

			// Override deny list
			if len(settings.Permissions.Deny) > 0 {
				effective.Permissions.Deny = settings.Permissions.Deny
			}

			// Override default mode
			if settings.Permissions.DefaultMode != "" {
				effective.Permissions.DefaultMode = settings.Permissions.DefaultMode
			}
		}
	}

	return effective
}

// SaveClaudeSettings saves the Claude Code local settings file
// If projectDir is provided, it saves to the project's .claude/settings.local.json
func SaveClaudeSettings(settings *ClaudeSettings, projectDir ...string) error {
	settingsPath := GetClaudeSettingsPath(projectDir...)
	return saveSettingsToPath(settings, settingsPath)
}

// SaveGlobalSettings saves settings to the global settings file
func SaveGlobalSettings(settings *ClaudeSettings) error {
	settingsPath := GetGlobalSettingsPath()
	if settingsPath == "" {
		return fmt.Errorf("could not determine home directory")
	}
	return saveSettingsToPath(settings, settingsPath)
}

// SaveProjectSettings saves settings to the project settings file
func SaveProjectSettings(settings *ClaudeSettings, projectDir ...string) error {
	settingsPath := GetProjectSettingsPath(projectDir...)
	return saveSettingsToPath(settings, settingsPath)
}

// SaveLocalSettings saves settings to the local settings file (gitignored)
func SaveLocalSettings(settings *ClaudeSettings, projectDir ...string) error {
	settingsPath := GetLocalSettingsPath(projectDir...)
	return saveSettingsToPath(settings, settingsPath)
}

// saveSettingsToPath saves settings to a specific path
func saveSettingsToPath(settings *ClaudeSettings, settingsPath string) error {
	// Ensure the directory exists
	settingsDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Clean up empty permissions object and default values
	if settings.Permissions != nil {
		// Deduplicate all permission lists before saving
		settings.Permissions.Allow = deduplicateStringSlice(settings.Permissions.Allow)
		settings.Permissions.Ask = deduplicateStringSlice(settings.Permissions.Ask)
		settings.Permissions.Deny = deduplicateStringSlice(settings.Permissions.Deny)

		// Remove defaultMode if it's the default value
		if settings.Permissions.DefaultMode == PermissionModeDefault {
			settings.Permissions.DefaultMode = ""
		}

		// Check if permissions object is effectively empty
		isEmpty := len(settings.Permissions.Allow) == 0 &&
			len(settings.Permissions.Ask) == 0 &&
			len(settings.Permissions.Deny) == 0 &&
			settings.Permissions.DefaultMode == ""

		if isEmpty {
			settings.Permissions = nil
		}
	}

	// Don't set schema - leave it as is from the loaded settings
	// Users may or may not want the schema in their settings files

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

// SettingsFileExists checks if a settings file exists at the given path
func SettingsFileExists(source SettingsSource, projectDir ...string) bool {
	var path string
	switch source {
	case SettingsSourceGlobal:
		path = GetGlobalSettingsPath()
	case SettingsSourceProject:
		path = GetProjectSettingsPath(projectDir...)
	case SettingsSourceLocal:
		path = GetLocalSettingsPath(projectDir...)
	default:
		return false
	}

	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return err == nil
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
		settings.Permissions = &PermissionsConfig{}
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

// deduplicateStringSlice removes duplicate strings from a slice
// while preserving the order of first occurrence
func deduplicateStringSlice(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
