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
		if hookExists {
			break
		}
	}

	if !hookExists {
		// Add hook in the proper format: hooks -> type/command
		// No matcher needed for UserPromptSubmit - it should run on every prompt
		hookEntry := map[string]interface{}{
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": hookScriptPath,
				},
			},
		}
		eventHooks = append(eventHooks, hookEntry)
		hooks[eventName] = eventHooks
		fmt.Printf("   ✓ Added hook to %s event\n", eventName)
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
