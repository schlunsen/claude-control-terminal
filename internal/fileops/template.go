package fileops

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TemplateFile represents a file to be copied from template
type TemplateFile struct {
	Source      string
	Destination string
}

// TemplateConfig holds the configuration for template installation
type TemplateConfig struct {
	Language      string
	Framework     string
	Files         []TemplateFile
	SelectedHooks []string
	SelectedMCPs  []string
}

// CheckExistingFiles checks for existing Claude Code configuration
func CheckExistingFiles(targetDir string) ([]string, error) {
	var existingFiles []string

	// Check for existing CLAUDE.md
	claudeFile := filepath.Join(targetDir, "CLAUDE.md")
	if _, err := os.Stat(claudeFile); err == nil {
		existingFiles = append(existingFiles, "CLAUDE.md")
	}

	// Check for existing .claude directory
	claudeDir := filepath.Join(targetDir, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		existingFiles = append(existingFiles, ".claude/")
	}

	// Check for existing .mcp.json
	mcpFile := filepath.Join(targetDir, ".mcp.json")
	if _, err := os.Stat(mcpFile); err == nil {
		existingFiles = append(existingFiles, ".mcp.json")
	}

	return existingFiles, nil
}

// CreateBackups creates timestamped backups of existing files
func CreateBackups(existingFiles []string, targetDir string) error {
	timestamp := time.Now().Format("2006-01-02T15-04-05")

	for _, file := range existingFiles {
		sourcePath := filepath.Join(targetDir, file)
		backupName := strings.ReplaceAll(file, "/", "") + ".backup-" + timestamp
		backupPath := filepath.Join(targetDir, backupName)

		// Use appropriate copy method based on file type
		info, err := os.Stat(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", file, err)
		}

		if info.IsDir() {
			if err := CopyDir(sourcePath, backupPath); err != nil {
				return fmt.Errorf("failed to backup directory %s: %w", file, err)
			}
		} else {
			if err := CopyFile(sourcePath, backupPath); err != nil {
				return fmt.Errorf("failed to backup file %s: %w", file, err)
			}
		}
	}

	return nil
}

// ProcessSettingsFile processes settings.json with selected hooks
func ProcessSettingsFile(settingsContent string, destPath string, selectedHooks []string) error {
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(settingsContent), &settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Filter hooks based on selection
	if len(selectedHooks) > 0 {
		if hooks, ok := settings["hooks"].(map[string]interface{}); ok {
			filteredHooks := make(map[string]interface{})
			for _, hookID := range selectedHooks {
				if hookValue, exists := hooks[hookID]; exists {
					filteredHooks[hookID] = hookValue
				}
			}
			settings["hooks"] = filteredHooks
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write settings file
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// MergeSettingsFile merges new settings with existing settings
func MergeSettingsFile(settingsContent string, destPath string, selectedHooks []string) error {
	var newSettings map[string]interface{}
	if err := json.Unmarshal([]byte(settingsContent), &newSettings); err != nil {
		return fmt.Errorf("failed to parse new settings: %w", err)
	}

	// Read existing settings if file exists
	existingSettings := make(map[string]interface{})
	if data, err := os.ReadFile(destPath); err == nil {
		json.Unmarshal(data, &existingSettings)
	}

	// Filter hooks in new settings
	if len(selectedHooks) > 0 {
		if hooks, ok := newSettings["hooks"].(map[string]interface{}); ok {
			filteredHooks := make(map[string]interface{})
			for _, hookID := range selectedHooks {
				if hookValue, exists := hooks[hookID]; exists {
					filteredHooks[hookID] = hookValue
				}
			}
			newSettings["hooks"] = filteredHooks
		}
	}

	// Merge settings
	mergedSettings := mergeMaps(existingSettings, newSettings)

	// Merge hooks specifically
	if existingHooks, ok := existingSettings["hooks"].(map[string]interface{}); ok {
		if newHooks, ok := newSettings["hooks"].(map[string]interface{}); ok {
			mergedHooks := mergeMaps(existingHooks, newHooks)
			mergedSettings["hooks"] = mergedHooks
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write merged settings
	data, err := json.MarshalIndent(mergedSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal merged settings: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// ProcessMCPFile processes MCP configuration with selected MCPs
func ProcessMCPFile(mcpContent string, destPath string, selectedMCPs []string) error {
	var mcpConfig map[string]interface{}
	if err := json.Unmarshal([]byte(mcpContent), &mcpConfig); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	// Clean and prepare MCP config (remove descriptions)
	cleanMcpConfig := map[string]interface{}{
		"mcpServers": make(map[string]interface{}),
	}

	if servers, ok := mcpConfig["mcpServers"].(map[string]interface{}); ok {
		cleanServers := make(map[string]interface{})

		for serverName, serverConfig := range servers {
			if serverMap, ok := serverConfig.(map[string]interface{}); ok {
				// Copy server config without description
				cleanServerConfig := make(map[string]interface{})
				for key, value := range serverMap {
					if key != "description" {
						cleanServerConfig[key] = value
					}
				}
				cleanServers[serverName] = cleanServerConfig
			}
		}

		cleanMcpConfig["mcpServers"] = cleanServers
	}

	// Filter MCPs based on selection
	if len(selectedMCPs) > 0 {
		if servers, ok := cleanMcpConfig["mcpServers"].(map[string]interface{}); ok {
			filteredServers := make(map[string]interface{})
			for _, mcpID := range selectedMCPs {
				if serverValue, exists := servers[mcpID]; exists {
					filteredServers[mcpID] = serverValue
				}
			}
			cleanMcpConfig["mcpServers"] = filteredServers
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write MCP config
	data, err := json.MarshalIndent(cleanMcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write MCP file: %w", err)
	}

	return nil
}

// Helper function to merge two maps
func mergeMaps(existing, new map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing
	for k, v := range existing {
		result[k] = v
	}

	// Overwrite with new
	for k, v := range new {
		result[k] = v
	}

	return result
}

// CheckWritePermissions checks if we can write to the target directory
func CheckWritePermissions(targetDir string) bool {
	testFile := filepath.Join(targetDir, ".claude-test-write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return false
	}
	os.Remove(testFile)
	return true
}
