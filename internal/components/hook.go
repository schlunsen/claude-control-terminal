// Package components provides installers for Claude Code components including hooks.
// This file implements hook installation and management functionality.
package components

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HookInstaller handles installation of Claude Code hooks
type HookInstaller struct {
	claudeDir string
}

// ClaudeSettings represents the structure of settings.json
type ClaudeSettings struct {
	Hooks map[string]interface{} `json:"hooks,omitempty"`
	// Other settings fields are preserved as-is
	RawSettings map[string]interface{} `json:"-"`
}

// NewHookInstaller creates a new hook installer
func NewHookInstaller() *HookInstaller {
	homeDir, _ := os.UserHomeDir()
	claudeDir := filepath.Join(homeDir, ".claude")

	return &HookInstaller{
		claudeDir: claudeDir,
	}
}

// NewHookInstallerWithDir creates a hook installer with custom Claude directory
func NewHookInstallerWithDir(claudeDir string) *HookInstaller {
	return &HookInstaller{
		claudeDir: claudeDir,
	}
}

// InstallUserPromptLogger installs the user-prompt-logger hook for current project only
// Hooks are always installed in the project's .claude directory, never globally
func (hi *HookInstaller) InstallUserPromptLogger() error {
	// Get current working directory for project-based installation
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("📝 Installing User Prompt Logger Hook (project-only)...")

	// Project .claude directory
	settingsDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create project .claude directory: %w", err)
	}

	// Hooks subdirectory in PROJECT .claude dir (not global)
	hooksDir := filepath.Join(settingsDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create project hooks directory: %w", err)
	}

	// Copy hook script to PROJECT hooks directory
	hookName := "user-prompt-logger.sh"
	if err := hi.copyHookScript(hookName, hooksDir); err != nil {
		return fmt.Errorf("failed to copy hook script: %w", err)
	}

	// Update settings.json in project directory
	if err := hi.addHookToSettingsAtPath(settingsDir, hooksDir, hookName, "UserPromptSubmit"); err != nil {
		return fmt.Errorf("failed to update settings.json: %w", err)
	}

	fmt.Println("✅ User Prompt Logger Hook installed successfully!")
	fmt.Printf("   Project: %s\n", cwd)
	fmt.Printf("   Hook script: %s\n", filepath.Join(hooksDir, hookName))
	fmt.Printf("   Settings: %s\n", filepath.Join(settingsDir, "settings.local.json"))
	fmt.Println("\n💡 This hook will only capture prompts for this project")
	fmt.Println("   View analytics: cct --analytics")

	return nil
}

// copyHookScript copies a hook script from the embedded hooks directory to Claude's hooks directory
func (hi *HookInstaller) copyHookScript(hookName string, hooksDir string) error {
	// Find the source hook script
	// Try multiple locations: embedded in binary, current directory, or project root
	var sourceContent []byte
	var err error

	// Try 1: Current working directory (development)
	cwd, _ := os.Getwd()
	sourcePath := filepath.Join(cwd, "hooks", hookName)
	sourceContent, err = os.ReadFile(sourcePath)

	// Try 2: Relative to binary location
	if err != nil {
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		sourcePath = filepath.Join(execDir, "hooks", hookName)
		sourceContent, err = os.ReadFile(sourcePath)
	}

	// Try 3: Project root (go up from internal/components)
	if err != nil {
		// Assume we're in internal/components, go up to project root
		projectRoot := filepath.Join(cwd, "..", "..")
		sourcePath = filepath.Join(projectRoot, "hooks", hookName)
		sourceContent, err = os.ReadFile(sourcePath)
	}

	if err != nil {
		return fmt.Errorf("could not find hook script %s: %w", hookName, err)
	}

	// Write to destination
	destPath := filepath.Join(hooksDir, hookName)
	if err := os.WriteFile(destPath, sourceContent, 0755); err != nil {
		return fmt.Errorf("failed to write hook script: %w", err)
	}

	fmt.Printf("   ✓ Copied hook script to: %s\n", destPath)
	return nil
}

// addHookToSettingsAtPath adds a hook to settings.local.json at specified directory
func (hi *HookInstaller) addHookToSettingsAtPath(settingsDir string, hooksDir string, hookName string, eventName string) error {
	return hi.addHookToSettingsWithMatcher(settingsDir, hooksDir, hookName, eventName, "")
}

// addHookToSettingsWithMatcher adds a hook with optional matcher to settings.local.json
func (hi *HookInstaller) addHookToSettingsWithMatcher(settingsDir string, hooksDir string, hookName string, eventName string, matcher string) error {
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	// Read existing settings or create new one
	var rawSettings map[string]interface{}
	content, err := os.ReadFile(settingsPath)

	if err != nil {
		if os.IsNotExist(err) {
			// Create new settings
			rawSettings = make(map[string]interface{})
		} else {
			return fmt.Errorf("failed to read settings.json: %w", err)
		}
	} else {
		// Parse existing settings
		if err := json.Unmarshal(content, &rawSettings); err != nil {
			return fmt.Errorf("failed to parse settings.json: %w", err)
		}
	}

	// Get or create hooks section
	var hooks map[string]interface{}
	if hooksRaw, exists := rawSettings["hooks"]; exists {
		if h, ok := hooksRaw.(map[string]interface{}); ok {
			hooks = h
		} else {
			hooks = make(map[string]interface{})
		}
	} else {
		hooks = make(map[string]interface{})
	}

	// Add hook to the specified event
	// Use absolute path to PROJECT hooks directory
	hookScriptPath := filepath.Join(hooksDir, hookName)

	// Get or create event array
	var eventHooks []interface{}
	if eventRaw, exists := hooks[eventName]; exists {
		if arr, ok := eventRaw.([]interface{}); ok {
			eventHooks = arr
		} else {
			eventHooks = []interface{}{}
		}
	} else {
		eventHooks = []interface{}{}
	}

	// Check if hook already exists
	hookExists := false
	for _, entry := range eventHooks {
		if entryMap, ok := entry.(map[string]interface{}); ok {
			// Check if matcher matches (or both empty)
			entryMatcher, _ := entryMap["matcher"].(string)
			if entryMatcher == matcher {
				if hooksArr, ok := entryMap["hooks"].([]interface{}); ok {
					for _, h := range hooksArr {
						if hMap, ok := h.(map[string]interface{}); ok {
							if cmd, ok := hMap["command"].(string); ok && strings.Contains(cmd, hookName) {
								hookExists = true
								break
							}
						}
					}
				}
			}
		}
		if hookExists {
			break
		}
	}

	if !hookExists {
		// Add hook in the proper format
		hookEntry := map[string]interface{}{
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": hookScriptPath,
				},
			},
		}

		// Add matcher if specified (for PostToolUse, PreToolUse, etc.)
		if matcher != "" {
			hookEntry["matcher"] = matcher
		}

		eventHooks = append(eventHooks, hookEntry)
		hooks[eventName] = eventHooks

		if matcher != "" {
			fmt.Printf("   ✓ Added hook to %s event (matcher: %s)\n", eventName, matcher)
		} else {
			fmt.Printf("   ✓ Added hook to %s event\n", eventName)
		}
	} else {
		fmt.Printf("   ℹ Hook already exists in %s event\n", eventName)
	}

	rawSettings["hooks"] = hooks

	// Write back to file with pretty formatting
	output, err := json.MarshalIndent(rawSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	return nil
}

// UninstallUserPromptLogger removes the user-prompt-logger hook from current project
func (hi *HookInstaller) UninstallUserPromptLogger() error {
	fmt.Println("🗑️  Uninstalling User Prompt Logger Hook...")

	hookName := "user-prompt-logger.sh"

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Remove from project settings.json
	projectSettingsDir := filepath.Join(cwd, ".claude")
	if err := hi.removeHookFromSettingsAtPath(projectSettingsDir, hookName, "UserPromptSubmit"); err != nil {
		fmt.Printf("   ℹ Project settings: %v\n", err)
	}

	// Remove the hook script file from project directory
	hookScriptPath := filepath.Join(projectSettingsDir, "hooks", hookName)
	if err := os.Remove(hookScriptPath); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("   ⚠️  Failed to remove hook script: %v\n", err)
		}
	} else {
		fmt.Printf("   ✓ Removed hook script: %s\n", hookScriptPath)
	}

	fmt.Println("✅ User Prompt Logger Hook uninstalled successfully!")
	return nil
}

// removeHookFromSettingsAtPath removes a hook from settings.local.json at specified directory
func (hi *HookInstaller) removeHookFromSettingsAtPath(settingsDir string, hookName string, eventName string) error {
	settingsPath := filepath.Join(settingsDir, "settings.local.json")

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("settings.json not found")
	}

	var rawSettings map[string]interface{}
	if err := json.Unmarshal(content, &rawSettings); err != nil {
		return fmt.Errorf("failed to parse settings.json: %w", err)
	}

	// Get hooks section
	hooksRaw, exists := rawSettings["hooks"]
	if !exists {
		return nil // No hooks, nothing to remove
	}

	hooks, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	// Get event array
	eventRaw, exists := hooks[eventName]
	if !exists {
		return nil
	}

	eventHooks, ok := eventRaw.([]interface{})
	if !ok {
		return nil
	}

	// Filter out the hook
	var newEventHooks []interface{}
	removed := false
	for _, entry := range eventHooks {
		shouldKeep := true

		// Check for old format (simple string)
		if hStr, ok := entry.(string); ok && strings.Contains(hStr, hookName) {
			shouldKeep = false
			removed = true
		} else if entryMap, ok := entry.(map[string]interface{}); ok {
			// Check for new format (matcher -> hooks -> type/command)
			if hooksArr, ok := entryMap["hooks"].([]interface{}); ok {
				for _, h := range hooksArr {
					if hMap, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hMap["command"].(string); ok && strings.Contains(cmd, hookName) {
							shouldKeep = false
							removed = true
							break
						}
					}
				}
			}
		}

		if shouldKeep {
			newEventHooks = append(newEventHooks, entry)
		}
	}

	if removed {
		if len(newEventHooks) > 0 {
			hooks[eventName] = newEventHooks
		} else {
			delete(hooks, eventName)
		}
		rawSettings["hooks"] = hooks

		// Write back
		output, err := json.MarshalIndent(rawSettings, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal settings: %w", err)
		}

		if err := os.WriteFile(settingsPath, output, 0644); err != nil {
			return fmt.Errorf("failed to write settings.json: %w", err)
		}

		fmt.Printf("   ✓ Removed hook from %s event\n", eventName)
	}

	return nil
}

// CheckHookInstalled checks if the user-prompt-logger hook is installed in current project
// Only checks project-based installation, never global
func (hi *HookInstaller) CheckHookInstalled() (bool, error) {
	// Check project-based installation only
	cwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get current directory: %w", err)
	}

	projectSettingsPath := filepath.Join(cwd, ".claude", "settings.local.json")
	return hi.checkHookInSettingsFile(projectSettingsPath)
}

// checkHookInSettingsFile checks if hook is installed in specific settings file
func (hi *HookInstaller) checkHookInSettingsFile(settingsPath string) (bool, error) {

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var rawSettings map[string]interface{}
	if err := json.Unmarshal(content, &rawSettings); err != nil {
		return false, err
	}

	hooksRaw, exists := rawSettings["hooks"]
	if !exists {
		return false, nil
	}

	hooks, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return false, nil
	}

	eventRaw, exists := hooks["UserPromptSubmit"]
	if !exists {
		return false, nil
	}

	eventHooks, ok := eventRaw.([]interface{})
	if !ok {
		return false, nil
	}

	for _, entry := range eventHooks {
		// Check for old format (simple string)
		if hStr, ok := entry.(string); ok && strings.Contains(hStr, "user-prompt-logger") {
			return true, nil
		}
		// Check for new format (matcher -> hooks -> type/command)
		if entryMap, ok := entry.(map[string]interface{}); ok {
			if hooksArr, ok := entryMap["hooks"].([]interface{}); ok {
				for _, h := range hooksArr {
					if hMap, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hMap["command"].(string); ok && strings.Contains(cmd, "user-prompt-logger") {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}

// InstallToolLogger installs the tool-logger hook for current project only
// Hooks are always installed in the project's .claude directory, never globally
func (hi *HookInstaller) InstallToolLogger() error {
	// Get current working directory for project-based installation
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("🔧 Installing Tool Logger Hook (project-only)...")

	// Project .claude directory
	settingsDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create project .claude directory: %w", err)
	}

	// Hooks subdirectory in PROJECT .claude dir (not global)
	hooksDir := filepath.Join(settingsDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create project hooks directory: %w", err)
	}

	// Copy hook script to PROJECT hooks directory
	hookName := "tool-logger.sh"
	if err := hi.copyHookScript(hookName, hooksDir); err != nil {
		return fmt.Errorf("failed to copy hook script: %w", err)
	}

	// Update settings.json in project directory with PostToolUse hook and wildcard matcher
	if err := hi.addHookToSettingsWithMatcher(settingsDir, hooksDir, hookName, "PostToolUse", "*"); err != nil {
		return fmt.Errorf("failed to update settings.json: %w", err)
	}

	fmt.Println("✅ Tool Logger Hook installed successfully!")
	fmt.Printf("   Project: %s\n", cwd)
	fmt.Printf("   Hook script: %s\n", filepath.Join(hooksDir, hookName))
	fmt.Printf("   Settings: %s\n", filepath.Join(settingsDir, "settings.local.json"))
	fmt.Println("\n💡 This hook will capture all tool usage (Bash, Read, Edit, Write, etc.)")
	fmt.Println("   View analytics: cct --analytics")

	return nil
}

// UninstallToolLogger removes the tool-logger hook from current project
func (hi *HookInstaller) UninstallToolLogger() error {
	fmt.Println("🗑️  Uninstalling Tool Logger Hook...")

	hookName := "tool-logger.sh"

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Remove from project settings.json
	projectSettingsDir := filepath.Join(cwd, ".claude")
	if err := hi.removeHookFromSettingsAtPath(projectSettingsDir, hookName, "PostToolUse"); err != nil {
		fmt.Printf("   ℹ Project settings: %v\n", err)
	}

	// Remove the hook script file from project directory
	hookScriptPath := filepath.Join(projectSettingsDir, "hooks", hookName)
	if err := os.Remove(hookScriptPath); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("   ⚠️  Failed to remove hook script: %v\n", err)
		}
	} else {
		fmt.Printf("   ✓ Removed hook script: %s\n", hookScriptPath)
	}

	fmt.Println("✅ Tool Logger Hook uninstalled successfully!")
	return nil
}

// CheckToolLoggerInstalled checks if the tool-logger hook is installed in current project
// Only checks project-based installation, never global
func (hi *HookInstaller) CheckToolLoggerInstalled() (bool, error) {
	// Check project-based installation only
	cwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get current directory: %w", err)
	}

	projectSettingsPath := filepath.Join(cwd, ".claude", "settings.local.json")
	return hi.checkToolLoggerInSettingsFile(projectSettingsPath)
}

// checkToolLoggerInSettingsFile checks if tool-logger hook is installed in specific settings file
func (hi *HookInstaller) checkToolLoggerInSettingsFile(settingsPath string) (bool, error) {
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var rawSettings map[string]interface{}
	if err := json.Unmarshal(content, &rawSettings); err != nil {
		return false, err
	}

	hooksRaw, exists := rawSettings["hooks"]
	if !exists {
		return false, nil
	}

	hooks, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return false, nil
	}

	eventRaw, exists := hooks["PostToolUse"]
	if !exists {
		return false, nil
	}

	eventHooks, ok := eventRaw.([]interface{})
	if !ok {
		return false, nil
	}

	for _, entry := range eventHooks {
		if entryMap, ok := entry.(map[string]interface{}); ok {
			if hooksArr, ok := entryMap["hooks"].([]interface{}); ok {
				for _, h := range hooksArr {
					if hMap, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hMap["command"].(string); ok && strings.Contains(cmd, "tool-logger") {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}

// InstallAllHooks installs all hooks (user-prompt-logger, tool-logger, and notification-logger) for current project
// Always installs to project's .claude directory, never globally
func (hi *HookInstaller) InstallAllHooks() error {
	fmt.Println("📦 Installing All Hooks (project-only)...")
	fmt.Println()

	// Install user prompt logger
	if err := hi.InstallUserPromptLogger(); err != nil {
		return fmt.Errorf("failed to install user prompt logger: %w", err)
	}

	fmt.Println()

	// Install tool logger
	if err := hi.InstallToolLogger(); err != nil {
		return fmt.Errorf("failed to install tool logger: %w", err)
	}

	fmt.Println()

	// Install notification logger
	if err := hi.InstallNotificationLogger(); err != nil {
		return fmt.Errorf("failed to install notification logger: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ All hooks installed successfully!")
	fmt.Println("   All three hooks are project-specific and will only run in this directory")

	return nil
}

// UninstallAllHooks removes both hooks from current project
func (hi *HookInstaller) UninstallAllHooks() error {
	fmt.Println("🗑️  Uninstalling All Hooks...")
	fmt.Println()

	// Uninstall user prompt logger
	if err := hi.UninstallUserPromptLogger(); err != nil {
		fmt.Printf("   ⚠️  User prompt logger: %v\n", err)
	}

	fmt.Println()

	// Uninstall tool logger
	if err := hi.UninstallToolLogger(); err != nil {
		fmt.Printf("   ⚠️  Tool logger: %v\n", err)
	}

	fmt.Println()

	// Uninstall notification logger
	if err := hi.UninstallNotificationLogger(); err != nil {
		fmt.Printf("   ⚠️  Notification logger: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✅ All hooks uninstalled successfully!")

	return nil
}

// InstallNotificationLogger installs the notification-logger hook for current project only
// Hooks are always installed in the project's .claude directory, never globally
func (hi *HookInstaller) InstallNotificationLogger() error {
	// Get current working directory for project-based installation
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("🔔 Installing Notification Logger Hook (project-only)...")

	// Project .claude directory
	settingsDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create project .claude directory: %w", err)
	}

	// Hooks subdirectory in PROJECT .claude dir (not global)
	hooksDir := filepath.Join(settingsDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create project hooks directory: %w", err)
	}

	// Copy hook script to PROJECT hooks directory
	hookName := "notification-logger.sh"
	if err := hi.copyHookScript(hookName, hooksDir); err != nil {
		return fmt.Errorf("failed to copy hook script: %w", err)
	}

	// Update settings.json in project directory with Notification hook (no matcher needed)
	if err := hi.addHookToSettingsAtPath(settingsDir, hooksDir, hookName, "Notification"); err != nil {
		return fmt.Errorf("failed to update settings.json: %w", err)
	}

	fmt.Println("✅ Notification Logger Hook installed successfully!")
	fmt.Printf("   Project: %s\n", cwd)
	fmt.Printf("   Hook script: %s\n", filepath.Join(hooksDir, hookName))
	fmt.Printf("   Settings: %s\n", filepath.Join(settingsDir, "settings.local.json"))
	fmt.Println("\n💡 This hook will capture permission requests and idle alerts")
	fmt.Println("   View analytics: cct --analytics")

	return nil
}

// UninstallNotificationLogger removes the notification-logger hook from current project
func (hi *HookInstaller) UninstallNotificationLogger() error {
	fmt.Println("🗑️  Uninstalling Notification Logger Hook...")

	hookName := "notification-logger.sh"

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Remove from project settings.json
	projectSettingsDir := filepath.Join(cwd, ".claude")
	if err := hi.removeHookFromSettingsAtPath(projectSettingsDir, hookName, "Notification"); err != nil {
		fmt.Printf("   ℹ Project settings: %v\n", err)
	}

	// Remove the hook script file from project directory
	hookScriptPath := filepath.Join(projectSettingsDir, "hooks", hookName)
	if err := os.Remove(hookScriptPath); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("   ⚠️  Failed to remove hook script: %v\n", err)
		}
	} else {
		fmt.Printf("   ✓ Removed hook script: %s\n", hookScriptPath)
	}

	fmt.Println("✅ Notification Logger Hook uninstalled successfully!")
	return nil
}

// CheckNotificationLoggerInstalled checks if the notification-logger hook is installed in current project
// Only checks project-based installation, never global
func (hi *HookInstaller) CheckNotificationLoggerInstalled() (bool, error) {
	// Check project-based installation only
	cwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get current directory: %w", err)
	}

	projectSettingsPath := filepath.Join(cwd, ".claude", "settings.local.json")
	return hi.checkNotificationLoggerInSettingsFile(projectSettingsPath)
}

// checkNotificationLoggerInSettingsFile checks if notification-logger hook is installed in specific settings file
func (hi *HookInstaller) checkNotificationLoggerInSettingsFile(settingsPath string) (bool, error) {
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var rawSettings map[string]interface{}
	if err := json.Unmarshal(content, &rawSettings); err != nil {
		return false, err
	}

	hooksRaw, exists := rawSettings["hooks"]
	if !exists {
		return false, nil
	}

	hooks, ok := hooksRaw.(map[string]interface{})
	if !ok {
		return false, nil
	}

	eventRaw, exists := hooks["Notification"]
	if !exists {
		return false, nil
	}

	eventHooks, ok := eventRaw.([]interface{})
	if !ok {
		return false, nil
	}

	for _, entry := range eventHooks {
		if entryMap, ok := entry.(map[string]interface{}); ok {
			if hooksArr, ok := entryMap["hooks"].([]interface{}); ok {
				for _, h := range hooksArr {
					if hMap, ok := h.(map[string]interface{}); ok {
						if cmd, ok := hMap["command"].(string); ok && strings.Contains(cmd, "notification-logger") {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}
