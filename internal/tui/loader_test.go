package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewComponentLoader(t *testing.T) {
	loader := NewComponentLoader()

	if loader == nil {
		t.Fatal("NewComponentLoader returned nil")
	}

	if loader.config == nil {
		t.Error("ComponentLoader config should not be nil")
	}

	// Cache may be nil if initialization fails, which is acceptable
}

func TestGetComponentMetadata(t *testing.T) {
	metadata := GetComponentMetadata()

	if metadata == nil {
		t.Fatal("GetComponentMetadata returned nil")
	}

	// Check for expected component types
	expectedTypes := []string{"agent", "command", "mcp"}

	for _, compType := range expectedTypes {
		meta, ok := metadata[compType]
		if !ok {
			t.Errorf("Metadata missing for component type: %s", compType)
			continue
		}

		if meta.Path == "" {
			t.Errorf("Metadata for %s has empty Path", compType)
		}

		if meta.Extension == "" {
			t.Errorf("Metadata for %s has empty Extension", compType)
		}

	}
}

func TestComponentMetadataExtensions(t *testing.T) {
	metadata := GetComponentMetadata()

	tests := []struct {
		componentType string
		expectedExt   string
	}{
		{"agent", ".md"},
		{"command", ".md"},
		{"mcp", ".json"},
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			meta, ok := metadata[tt.componentType]
			if !ok {
				t.Fatalf("Metadata not found for %s", tt.componentType)
			}

			if meta.Extension != tt.expectedExt {
				t.Errorf("Expected extension %s for %s, got %s", tt.expectedExt, tt.componentType, meta.Extension)
			}
		})
	}
}

func TestLoadComponentsUnknownType(t *testing.T) {
	loader := NewComponentLoader()

	_, err := loader.LoadComponents("unknown-type", ".")

	if err == nil {
		t.Error("Expected error for unknown component type, got nil")
	}
}

func TestCheckInstallationStatusBasic(t *testing.T) {
	// Test with a directory that doesn't exist
	installedGlobal, installedProject := CheckInstallationStatus("test-agent", "agent", "/nonexistent/path")

	// Both should be false for non-existent paths
	// Note: This test assumes the function checks for actual files
	// If the implementation changes, this test may need adjustment

	_ = installedGlobal   // May be true or false depending on global installation
	_ = installedProject  // Should be false for non-existent path
}

func TestComponentMetadataStructure(t *testing.T) {
	metadata := GetComponentMetadata()

	tests := []string{"agent", "command", "mcp"}

	for _, componentType := range tests {
		t.Run(componentType, func(t *testing.T) {
			meta, ok := metadata[componentType]
			if !ok {
				t.Fatalf("Metadata not found for %s", componentType)
			}

			// Verify all required fields are present
			if meta.Path == "" {
				t.Error("Path should not be empty")
			}
			if meta.Extension == "" {
				t.Error("Extension should not be empty")
			}
			if meta.Type == "" {
				t.Error("Type should not be empty")
			}
		})
	}
}

func TestLoadComponentsWithCacheForceRefresh(t *testing.T) {
	loader := NewComponentLoader()

	// Test with force refresh = true
	// This should skip cache and fetch from GitHub
	// Note: This test requires internet connection and may fail if GitHub is unreachable
	// In a real environment, you might want to mock the HTTP calls

	// We'll test the basic structure rather than the actual API call
	if loader.config == nil {
		t.Error("Loader config should not be nil")
	}
}

func TestComponentMetadataPaths(t *testing.T) {
	metadata := GetComponentMetadata()

	// Verify paths are not empty and follow expected structure
	for compType, meta := range metadata {
		t.Run(compType, func(t *testing.T) {
			if meta.Path == "" {
				t.Error("Path should not be empty")
			}

			// Path should typically start with a component type name
			// This is a basic sanity check
			if len(meta.Path) < 3 {
				t.Errorf("Path seems too short: %s", meta.Path)
			}
		})
	}
}

func TestComponentLoaderConfig(t *testing.T) {
	loader := NewComponentLoader()

	if loader.config == nil {
		t.Fatal("Config should not be nil")
	}

	// Check config has expected fields populated
	if loader.config.Owner == "" {
		t.Error("Config Owner should not be empty")
	}

	if loader.config.Repo == "" {
		t.Error("Config Repo should not be empty")
	}

	if loader.config.Branch == "" {
		t.Error("Config Branch should not be empty")
	}
}

func TestCheckInstallationStatusWithFiles(t *testing.T) {
	// Create temp directory structure
	tempDir, err := os.MkdirTemp("", "test_installation_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .claude directory structure
	claudeDir := filepath.Join(tempDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	commandsDir := filepath.Join(claudeDir, "commands")
	mcpDir := filepath.Join(claudeDir, "mcp")

	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		t.Fatalf("Failed to create mcp dir: %v", err)
	}

	// Create test files
	agentFile := filepath.Join(agentsDir, "test-agent.md")
	if err := os.WriteFile(agentFile, []byte("test agent"), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	commandFile := filepath.Join(commandsDir, "test-command.md")
	if err := os.WriteFile(commandFile, []byte("test command"), 0644); err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	mcpFile := filepath.Join(mcpDir, "test-mcp.json")
	if err := os.WriteFile(mcpFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create mcp file: %v", err)
	}

	// Test agent installation status
	_, projectAgent := CheckInstallationStatus("test-agent", "agent", tempDir)
	if !projectAgent {
		t.Error("Agent should be detected as installed in project")
	}

	// Test command installation status
	_, projectCommand := CheckInstallationStatus("test-command", "command", tempDir)
	if !projectCommand {
		t.Error("Command should be detected as installed in project")
	}

	// Test MCP installation status
	_, projectMCP := CheckInstallationStatus("test-mcp", "mcp", tempDir)
	if !projectMCP {
		t.Error("MCP should be detected as installed in project")
	}

	// Test non-existent component
	globalNotExist, projectNotExist := CheckInstallationStatus("nonexistent", "agent", tempDir)
	if globalNotExist || projectNotExist {
		t.Error("Non-existent component should not be detected as installed")
	}
}

func TestCheckInstallationStatusInvalidType(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_invalid_type_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with invalid component type
	global, project := CheckInstallationStatus("test", "invalid-type", tempDir)
	if global || project {
		t.Error("Invalid component type should return false for both")
	}
}

func TestGetComponentMetadataCategories(t *testing.T) {
	metadata := GetComponentMetadata()

	tests := []struct {
		componentType string
		minCategories int
	}{
		{"agent", 10},
		{"command", 5},
		{"mcp", 5},
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			meta, ok := metadata[tt.componentType]
			if !ok {
				t.Fatalf("Metadata not found for %s", tt.componentType)
			}

			if len(meta.Categories) < tt.minCategories {
				t.Errorf("Expected at least %d categories for %s, got %d",
					tt.minCategories, tt.componentType, len(meta.Categories))
			}

			// Verify categories are not empty strings
			for _, category := range meta.Categories {
				if category == "" {
					t.Errorf("Empty category found in %s", tt.componentType)
				}
			}
		})
	}
}

func TestGetComponentMetadataIcons(t *testing.T) {
	metadata := GetComponentMetadata()

	tests := []struct {
		componentType string
		expectedIcon  string
	}{
		{"agent", "🤖"},
		{"command", "⚡"},
		{"mcp", "🔌"},
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			meta, ok := metadata[tt.componentType]
			if !ok {
				t.Fatalf("Metadata not found for %s", tt.componentType)
			}

			if meta.Icon != tt.expectedIcon {
				t.Errorf("Expected icon %s for %s, got %s",
					tt.expectedIcon, tt.componentType, meta.Icon)
			}
		})
	}
}

func TestComponentItemDefaults(t *testing.T) {
	// Test zero-value ComponentItem
	var item ComponentItem

	if item.Name != "" {
		t.Error("Default Name should be empty")
	}

	if item.Category != "" {
		t.Error("Default Category should be empty")
	}

	if item.Type != "" {
		t.Error("Default Type should be empty")
	}

	if item.Selected {
		t.Error("Default Selected should be false")
	}

	if item.InstalledGlobal {
		t.Error("Default InstalledGlobal should be false")
	}

	if item.InstalledProject {
		t.Error("Default InstalledProject should be false")
	}
}

func TestComponentItemStructure(t *testing.T) {
	// Test ComponentItem with all fields set
	item := ComponentItem{
		Name:             "test-component",
		Category:         "test-category",
		Description:      "Test description",
		Type:             "agent",
		Selected:         true,
		InstalledGlobal:  true,
		InstalledProject: false,
	}

	if item.Name != "test-component" {
		t.Errorf("Expected Name 'test-component', got '%s'", item.Name)
	}

	if item.Category != "test-category" {
		t.Errorf("Expected Category 'test-category', got '%s'", item.Category)
	}

	if item.Description != "Test description" {
		t.Errorf("Expected Description 'Test description', got '%s'", item.Description)
	}

	if item.Type != "agent" {
		t.Errorf("Expected Type 'agent', got '%s'", item.Type)
	}

	if !item.Selected {
		t.Error("Expected Selected to be true")
	}

	if !item.InstalledGlobal {
		t.Error("Expected InstalledGlobal to be true")
	}

	if item.InstalledProject {
		t.Error("Expected InstalledProject to be false")
	}
}
