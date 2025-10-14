package fileops

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetMCPMetadataPath_Project(t *testing.T) {
	projectDir := "/test/project"
	path := GetMCPMetadataPath(MCPScopeProject, projectDir)

	expected := filepath.Join(projectDir, ".mcp-metadata.json")
	if path != expected {
		t.Errorf("Expected path '%s', got '%s'", expected, path)
	}
}

func TestGetMCPMetadataPath_User(t *testing.T) {
	projectDir := "/test/project"
	path := GetMCPMetadataPath(MCPScopeUser, projectDir)

	// Should contain .claude/.mcp-metadata.json
	expectedSuffix := filepath.Join(".claude", ".mcp-metadata.json")
	if len(path) < len(expectedSuffix) {
		t.Errorf("Path too short: %s", path)
	}
}

func TestLoadMCPMetadata_NonExistent(t *testing.T) {
	// Test loading non-existent file
	metadata, err := LoadMCPMetadata(MCPScopeProject, "/nonexistent/path")
	if err != nil {
		t.Errorf("LoadMCPMetadata should not error for non-existent file: %v", err)
	}

	if metadata == nil {
		t.Fatal("LoadMCPMetadata should return empty metadata for non-existent file")
	}

	if metadata.Installations == nil {
		t.Error("Installations map should be initialized")
	}

	if len(metadata.Installations) != 0 {
		t.Errorf("Installations should be empty, got %d", len(metadata.Installations))
	}
}

func TestLoadMCPMetadata_ValidFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_metadata_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	metadataPath := filepath.Join(tempDir, ".mcp-metadata.json")
	testMetadata := `{
		"installations": {
			"test-mcp": {
				"installName": "test-mcp",
				"serverKeys": ["test-server"],
				"sourcePath": "components/mcps/test-mcp.json",
				"installedAt": "2024-01-01T00:00:00Z",
				"scope": "project"
			}
		}
	}`

	if err := os.WriteFile(metadataPath, []byte(testMetadata), 0644); err != nil {
		t.Fatalf("Failed to create metadata file: %v", err)
	}

	metadata, err := LoadMCPMetadata(MCPScopeProject, tempDir)
	if err != nil {
		t.Errorf("LoadMCPMetadata failed: %v", err)
	}

	if metadata == nil {
		t.Fatal("LoadMCPMetadata returned nil")
	}

	if len(metadata.Installations) != 1 {
		t.Errorf("Expected 1 installation, got %d", len(metadata.Installations))
	}

	installation, ok := metadata.Installations["test-mcp"]
	if !ok {
		t.Fatal("test-mcp not found in metadata")
	}

	if installation.InstallName != "test-mcp" {
		t.Errorf("Expected InstallName 'test-mcp', got '%s'", installation.InstallName)
	}

	if len(installation.ServerKeys) != 1 || installation.ServerKeys[0] != "test-server" {
		t.Errorf("Expected ServerKeys ['test-server'], got %v", installation.ServerKeys)
	}
}

func TestSaveMCPMetadata(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_save_metadata_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	metadata := &MCPMetadata{
		Installations: map[string]MCPInstallation{
			"my-mcp": {
				InstallName: "my-mcp",
				ServerKeys:  []string{"server1", "server2"},
				SourcePath:  "components/mcps/my-mcp.json",
				InstalledAt: time.Now(),
				Scope:       MCPScopeProject,
			},
		},
	}

	err = SaveMCPMetadata(MCPScopeProject, tempDir, metadata)
	if err != nil {
		t.Errorf("SaveMCPMetadata failed: %v", err)
	}

	// Verify file was created
	metadataPath := GetMCPMetadataPath(MCPScopeProject, tempDir)
	if !FileExists(metadataPath) {
		t.Fatal("Metadata file was not created")
	}

	// Verify content by loading it back
	loadedMetadata, err := LoadMCPMetadata(MCPScopeProject, tempDir)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if len(loadedMetadata.Installations) != 1 {
		t.Errorf("Expected 1 installation, got %d", len(loadedMetadata.Installations))
	}

	installation, ok := loadedMetadata.Installations["my-mcp"]
	if !ok {
		t.Fatal("my-mcp not found in loaded metadata")
	}

	if len(installation.ServerKeys) != 2 {
		t.Errorf("Expected 2 server keys, got %d", len(installation.ServerKeys))
	}
}

func TestSaveMCPMetadata_EmptyDeletes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_empty_metadata_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a metadata file first
	metadata := &MCPMetadata{
		Installations: map[string]MCPInstallation{
			"test": {
				InstallName: "test",
				ServerKeys:  []string{"test"},
				SourcePath:  "test",
				InstalledAt: time.Now(),
				Scope:       MCPScopeProject,
			},
		},
	}
	SaveMCPMetadata(MCPScopeProject, tempDir, metadata)

	// Now save empty metadata
	emptyMetadata := &MCPMetadata{
		Installations: make(map[string]MCPInstallation),
	}
	err = SaveMCPMetadata(MCPScopeProject, tempDir, emptyMetadata)
	if err != nil {
		t.Errorf("SaveMCPMetadata failed for empty metadata: %v", err)
	}

	// Verify file was deleted
	metadataPath := GetMCPMetadataPath(MCPScopeProject, tempDir)
	if FileExists(metadataPath) {
		t.Error("Empty metadata file should be deleted")
	}
}

func TestAddMCPInstallation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_add_installation_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = AddMCPInstallation(
		MCPScopeProject,
		tempDir,
		"postgresql-integration",
		[]string{"postgresql", "postgres"},
		"components/mcps/database/postgresql-integration.json",
	)
	if err != nil {
		t.Errorf("AddMCPInstallation failed: %v", err)
	}

	// Verify installation was added
	metadata, err := LoadMCPMetadata(MCPScopeProject, tempDir)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	installation, ok := metadata.Installations["postgresql-integration"]
	if !ok {
		t.Fatal("postgresql-integration not found in metadata")
	}

	if installation.InstallName != "postgresql-integration" {
		t.Errorf("Expected InstallName 'postgresql-integration', got '%s'", installation.InstallName)
	}

	if len(installation.ServerKeys) != 2 {
		t.Errorf("Expected 2 server keys, got %d", len(installation.ServerKeys))
	}

	if installation.SourcePath != "components/mcps/database/postgresql-integration.json" {
		t.Errorf("Unexpected SourcePath: %s", installation.SourcePath)
	}
}

func TestRemoveMCPInstallation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_remove_installation_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add installations
	AddMCPInstallation(MCPScopeProject, tempDir, "mcp1", []string{"s1"}, "path1")
	AddMCPInstallation(MCPScopeProject, tempDir, "mcp2", []string{"s2"}, "path2")

	// Remove mcp1
	err = RemoveMCPInstallation(MCPScopeProject, tempDir, "mcp1")
	if err != nil {
		t.Errorf("RemoveMCPInstallation failed: %v", err)
	}

	// Verify removal
	metadata, err := LoadMCPMetadata(MCPScopeProject, tempDir)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if _, ok := metadata.Installations["mcp1"]; ok {
		t.Error("mcp1 should have been removed")
	}

	if _, ok := metadata.Installations["mcp2"]; !ok {
		t.Error("mcp2 should still be present")
	}
}

func TestGetMCPInstallation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_get_installation_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add installation
	AddMCPInstallation(
		MCPScopeProject,
		tempDir,
		"test-mcp",
		[]string{"server1", "server2"},
		"components/mcps/test-mcp.json",
	)

	// Retrieve installation
	installation, err := GetMCPInstallation(MCPScopeProject, tempDir, "test-mcp")
	if err != nil {
		t.Errorf("GetMCPInstallation failed: %v", err)
	}

	if installation == nil {
		t.Fatal("GetMCPInstallation returned nil")
	}

	if installation.InstallName != "test-mcp" {
		t.Errorf("Expected InstallName 'test-mcp', got '%s'", installation.InstallName)
	}

	if len(installation.ServerKeys) != 2 {
		t.Errorf("Expected 2 server keys, got %d", len(installation.ServerKeys))
	}
}

func TestGetMCPInstallation_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_get_notfound_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Try to get non-existent installation
	_, err = GetMCPInstallation(MCPScopeProject, tempDir, "nonexistent")
	if err == nil {
		t.Error("GetMCPInstallation should error for non-existent installation")
	}
}

func TestGetInstalledMCPs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_list_installations_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add multiple installations
	AddMCPInstallation(MCPScopeProject, tempDir, "mcp1", []string{"s1"}, "path1")
	AddMCPInstallation(MCPScopeProject, tempDir, "mcp2", []string{"s2"}, "path2")
	AddMCPInstallation(MCPScopeProject, tempDir, "mcp3", []string{"s3"}, "path3")

	// Get all installations
	installations, err := GetInstalledMCPs(MCPScopeProject, tempDir)
	if err != nil {
		t.Errorf("GetInstalledMCPs failed: %v", err)
	}

	if len(installations) != 3 {
		t.Errorf("Expected 3 installations, got %d", len(installations))
	}

	// Verify all installations are present
	names := make(map[string]bool)
	for _, installation := range installations {
		names[installation.InstallName] = true
	}

	expectedNames := []string{"mcp1", "mcp2", "mcp3"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("Installation '%s' not found in list", name)
		}
	}
}

func TestGetInstalledMCPs_Empty(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_list_empty_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Get installations from empty metadata
	installations, err := GetInstalledMCPs(MCPScopeProject, tempDir)
	if err != nil {
		t.Errorf("GetInstalledMCPs failed: %v", err)
	}

	if len(installations) != 0 {
		t.Errorf("Expected 0 installations, got %d", len(installations))
	}
}

func TestFindMCPByServerKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_find_by_key_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add installation with multiple server keys
	AddMCPInstallation(
		MCPScopeProject,
		tempDir,
		"postgresql-integration",
		[]string{"postgresql", "postgres", "pg"},
		"components/mcps/database/postgresql-integration.json",
	)

	// Test finding by each server key
	tests := []string{"postgresql", "postgres", "pg"}
	for _, serverKey := range tests {
		t.Run(serverKey, func(t *testing.T) {
			installation, err := FindMCPByServerKey(MCPScopeProject, tempDir, serverKey)
			if err != nil {
				t.Errorf("FindMCPByServerKey failed for '%s': %v", serverKey, err)
			}

			if installation == nil {
				t.Fatal("FindMCPByServerKey returned nil")
			}

			if installation.InstallName != "postgresql-integration" {
				t.Errorf("Expected InstallName 'postgresql-integration', got '%s'", installation.InstallName)
			}
		})
	}
}

func TestFindMCPByServerKey_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_find_notfound_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Try to find by non-existent server key
	_, err = FindMCPByServerKey(MCPScopeProject, tempDir, "nonexistent")
	if err == nil {
		t.Error("FindMCPByServerKey should error for non-existent server key")
	}
}

func TestMCPInstallation_AllFields(t *testing.T) {
	now := time.Now()
	installation := MCPInstallation{
		InstallName: "test-mcp",
		ServerKeys:  []string{"key1", "key2"},
		SourcePath:  "components/mcps/test.json",
		InstalledAt: now,
		Scope:       MCPScopeProject,
	}

	if installation.InstallName != "test-mcp" {
		t.Errorf("Expected InstallName 'test-mcp', got '%s'", installation.InstallName)
	}

	if len(installation.ServerKeys) != 2 {
		t.Errorf("Expected 2 server keys, got %d", len(installation.ServerKeys))
	}

	if installation.SourcePath != "components/mcps/test.json" {
		t.Errorf("Unexpected SourcePath: %s", installation.SourcePath)
	}

	if installation.InstalledAt != now {
		t.Error("InstalledAt timestamp mismatch")
	}

	if installation.Scope != MCPScopeProject {
		t.Errorf("Expected Scope 'project', got '%s'", installation.Scope)
	}
}
