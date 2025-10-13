package fileops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_fileexists_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test file that exists
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !FileExists(testFile) {
		t.Error("FileExists returned false for existing file")
	}

	// Test file that doesn't exist
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	if FileExists(nonExistentFile) {
		t.Error("FileExists returned true for non-existent file")
	}
}

func TestDirExists(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_direxists_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test directory that exists
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if !DirExists(testDir) {
		t.Error("DirExists returned false for existing directory")
	}

	// Test directory that doesn't exist
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	if DirExists(nonExistentDir) {
		t.Error("DirExists returned true for non-existent directory")
	}

	// Test with a file (not a directory)
	testFile := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if DirExists(testFile) {
		t.Error("DirExists returned true for a file (not a directory)")
	}
}

func TestEnsureDir(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_ensuredir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating a single directory
	testDir := filepath.Join(tempDir, "newdir")
	if err := EnsureDir(testDir); err != nil {
		t.Errorf("EnsureDir failed: %v", err)
	}

	if !DirExists(testDir) {
		t.Error("Directory was not created")
	}

	// Test creating nested directories
	nestedDir := filepath.Join(tempDir, "a", "b", "c")
	if err := EnsureDir(nestedDir); err != nil {
		t.Errorf("EnsureDir failed for nested directories: %v", err)
	}

	if !DirExists(nestedDir) {
		t.Error("Nested directories were not created")
	}

	// Test with existing directory (should not error)
	if err := EnsureDir(testDir); err != nil {
		t.Errorf("EnsureDir failed for existing directory: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_copyfile_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	testContent := "This is test content"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test copying to same directory
	dstFile := filepath.Join(tempDir, "destination.txt")
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Errorf("CopyFile failed: %v", err)
	}

	// Verify destination file exists
	if !FileExists(dstFile) {
		t.Error("Destination file was not created")
	}

	// Verify content
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Content mismatch: expected %q, got %q", testContent, string(content))
	}

	// Test copying to nested directory (auto-create)
	nestedDst := filepath.Join(tempDir, "nested", "dir", "file.txt")
	if err := CopyFile(srcFile, nestedDst); err != nil {
		t.Errorf("CopyFile failed for nested destination: %v", err)
	}

	if !FileExists(nestedDst) {
		t.Error("Nested destination file was not created")
	}

	// Test copying non-existent file
	nonExistent := filepath.Join(tempDir, "nonexistent.txt")
	err = CopyFile(nonExistent, dstFile)
	if err == nil {
		t.Error("CopyFile should fail for non-existent source file")
	}
}

func TestCopyFilePermissions(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_copyfile_perms_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source file with specific permissions
	srcFile := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tempDir, "destination.txt")
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Check permissions
	srcInfo, _ := os.Stat(srcFile)
	dstInfo, _ := os.Stat(dstFile)

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permissions not preserved: source %v, destination %v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyDir(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_copydir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source directory structure
	srcDir := filepath.Join(tempDir, "source")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create files and subdirectories
	files := map[string]string{
		"file1.txt":            "content1",
		"file2.txt":            "content2",
		"subdir/file3.txt":     "content3",
		"subdir/deep/file4.txt": "content4",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(srcDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", relPath, err)
		}
	}

	// Copy directory
	dstDir := filepath.Join(tempDir, "destination")
	if err := CopyDir(srcDir, dstDir); err != nil {
		t.Errorf("CopyDir failed: %v", err)
	}

	// Verify destination directory exists
	if !DirExists(dstDir) {
		t.Fatal("Destination directory was not created")
	}

	// Verify all files were copied
	for relPath, expectedContent := range files {
		dstPath := filepath.Join(dstDir, relPath)

		if !FileExists(dstPath) {
			t.Errorf("File %s was not copied", relPath)
			continue
		}

		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", relPath, err)
			continue
		}

		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s: expected %q, got %q", relPath, expectedContent, string(content))
		}
	}

	// Verify subdirectories exist
	if !DirExists(filepath.Join(dstDir, "subdir")) {
		t.Error("Subdirectory 'subdir' was not copied")
	}

	if !DirExists(filepath.Join(dstDir, "subdir", "deep")) {
		t.Error("Nested subdirectory 'subdir/deep' was not copied")
	}
}

func TestCopyDirErrors(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_copydir_errors_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test copying non-existent directory
	nonExistent := filepath.Join(tempDir, "nonexistent")
	dstDir := filepath.Join(tempDir, "destination")
	err = CopyDir(nonExistent, dstDir)
	if err == nil {
		t.Error("CopyDir should fail for non-existent source directory")
	}
}

func TestCopyEmptyDir(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_copyemptydir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create empty source directory
	srcDir := filepath.Join(tempDir, "empty_source")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Copy empty directory
	dstDir := filepath.Join(tempDir, "empty_destination")
	if err := CopyDir(srcDir, dstDir); err != nil {
		t.Errorf("CopyDir failed for empty directory: %v", err)
	}

	// Verify destination exists and is empty
	if !DirExists(dstDir) {
		t.Error("Empty destination directory was not created")
	}

	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatalf("Failed to read destination directory: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected empty directory, got %d entries", len(entries))
	}
}

func TestCheckExistingFiles(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_checkexisting_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with no existing files
	existing, err := CheckExistingFiles(tempDir)
	if err != nil {
		t.Errorf("CheckExistingFiles failed: %v", err)
	}
	if len(existing) != 0 {
		t.Errorf("Expected no existing files, got %d", len(existing))
	}

	// Create CLAUDE.md
	claudeFile := filepath.Join(tempDir, "CLAUDE.md")
	if err := os.WriteFile(claudeFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	existing, err = CheckExistingFiles(tempDir)
	if err != nil {
		t.Errorf("CheckExistingFiles failed: %v", err)
	}
	if len(existing) != 1 || existing[0] != "CLAUDE.md" {
		t.Errorf("Expected [CLAUDE.md], got %v", existing)
	}

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.Mkdir(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	existing, err = CheckExistingFiles(tempDir)
	if err != nil {
		t.Errorf("CheckExistingFiles failed: %v", err)
	}
	if len(existing) != 2 {
		t.Errorf("Expected 2 existing files, got %d: %v", len(existing), existing)
	}

	// Create .mcp.json
	mcpFile := filepath.Join(tempDir, ".mcp.json")
	if err := os.WriteFile(mcpFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create .mcp.json: %v", err)
	}

	existing, err = CheckExistingFiles(tempDir)
	if err != nil {
		t.Errorf("CheckExistingFiles failed: %v", err)
	}
	if len(existing) != 3 {
		t.Errorf("Expected 3 existing files, got %d: %v", len(existing), existing)
	}
}

func TestCreateBackups(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_backups_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file in test dir: %v", err)
	}

	// Create backups
	existingFiles := []string{"test.txt", "testdir/"}
	err = CreateBackups(existingFiles, tempDir)
	if err != nil {
		t.Errorf("CreateBackups failed: %v", err)
	}

	// Verify backups exist
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	foundFileBackup := false
	foundDirBackup := false

	for _, entry := range entries {
		name := entry.Name()
		// Backup files have format: original-name.backup-timestamp
		if !entry.IsDir() && name != "test.txt" && filepath.Base(name) != "file.txt" {
			if strings.Contains(name, "backup") {
				foundFileBackup = true
			}
		}
		if entry.IsDir() && name != "testdir" {
			if strings.Contains(name, "backup") {
				foundDirBackup = true
			}
		}
	}

	if !foundFileBackup {
		t.Error("File backup was not created")
	}
	if !foundDirBackup {
		t.Error("Directory backup was not created")
	}
}

func TestProcessSettingsFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_settings_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test settings content
	settingsContent := `{
		"theme": "dark",
		"hooks": {
			"hook1": {"command": "test1"},
			"hook2": {"command": "test2"},
			"hook3": {"command": "test3"}
		}
	}`

	destPath := filepath.Join(tempDir, "settings.json")

	// Test with selected hooks
	selectedHooks := []string{"hook1", "hook3"}
	err = ProcessSettingsFile(settingsContent, destPath, selectedHooks)
	if err != nil {
		t.Errorf("ProcessSettingsFile failed: %v", err)
	}

	// Verify file was created
	if !FileExists(destPath) {
		t.Fatal("Settings file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to parse settings: %v", err)
	}

	// Verify theme is preserved
	if settings["theme"] != "dark" {
		t.Errorf("Expected theme 'dark', got %v", settings["theme"])
	}

	// Verify only selected hooks are present
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("Hooks not found in settings")
	}

	if len(hooks) != 2 {
		t.Errorf("Expected 2 hooks, got %d", len(hooks))
	}

	if _, ok := hooks["hook1"]; !ok {
		t.Error("hook1 not found")
	}
	if _, ok := hooks["hook3"]; !ok {
		t.Error("hook3 not found")
	}
	if _, ok := hooks["hook2"]; ok {
		t.Error("hook2 should not be present")
	}
}

func TestMergeSettingsFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_merge_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create existing settings file
	destPath := filepath.Join(tempDir, "settings.json")
	existingSettings := `{
		"theme": "light",
		"fontSize": 14,
		"hooks": {
			"existingHook": {"command": "existing"}
		}
	}`
	if err := os.WriteFile(destPath, []byte(existingSettings), 0644); err != nil {
		t.Fatalf("Failed to create existing settings: %v", err)
	}

	// New settings to merge
	newSettings := `{
		"theme": "dark",
		"lineNumbers": true,
		"hooks": {
			"newHook": {"command": "new"}
		}
	}`

	selectedHooks := []string{"newHook"}
	err = MergeSettingsFile(newSettings, destPath, selectedHooks)
	if err != nil {
		t.Errorf("MergeSettingsFile failed: %v", err)
	}

	// Read and verify merged content
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read merged settings: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Failed to parse merged settings: %v", err)
	}

	// Verify new theme overwrites old
	if settings["theme"] != "dark" {
		t.Errorf("Expected theme 'dark', got %v", settings["theme"])
	}

	// Verify old fontSize is preserved
	if settings["fontSize"] != float64(14) {
		t.Errorf("Expected fontSize 14, got %v", settings["fontSize"])
	}

	// Verify new setting is added
	if settings["lineNumbers"] != true {
		t.Errorf("Expected lineNumbers true, got %v", settings["lineNumbers"])
	}

	// Verify hooks are merged
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("Hooks not found in merged settings")
	}

	if _, ok := hooks["existingHook"]; !ok {
		t.Error("existingHook should be preserved")
	}
	if _, ok := hooks["newHook"]; !ok {
		t.Error("newHook should be added")
	}
}

func TestProcessMCPFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_mcp_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test MCP content
	mcpContent := `{
		"mcpServers": {
			"server1": {
				"command": "node",
				"args": ["server1.js"],
				"description": "Server 1 description"
			},
			"server2": {
				"command": "python",
				"args": ["server2.py"],
				"description": "Server 2 description"
			},
			"server3": {
				"command": "go",
				"args": ["run", "server3.go"]
			}
		}
	}`

	destPath := filepath.Join(tempDir, ".mcp.json")

	// Test with selected MCPs
	selectedMCPs := []string{"server1", "server3"}
	err = ProcessMCPFile(mcpContent, destPath, selectedMCPs)
	if err != nil {
		t.Errorf("ProcessMCPFile failed: %v", err)
	}

	// Verify file was created
	if !FileExists(destPath) {
		t.Fatal("MCP file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read MCP file: %v", err)
	}

	var mcpConfig map[string]interface{}
	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		t.Fatalf("Failed to parse MCP config: %v", err)
	}

	servers, ok := mcpConfig["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers not found in config")
	}

	// Verify only selected servers are present
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}

	if _, ok := servers["server1"]; !ok {
		t.Error("server1 not found")
	}
	if _, ok := servers["server3"]; !ok {
		t.Error("server3 not found")
	}
	if _, ok := servers["server2"]; ok {
		t.Error("server2 should not be present")
	}

	// Verify descriptions are removed
	server1, ok := servers["server1"].(map[string]interface{})
	if !ok {
		t.Fatal("server1 config not found")
	}
	if _, ok := server1["description"]; ok {
		t.Error("description should be removed from server1")
	}
	if server1["command"] != "node" {
		t.Errorf("Expected command 'node', got %v", server1["command"])
	}
}

func TestCheckWritePermissions(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "test_writeperms_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test writable directory
	if !CheckWritePermissions(tempDir) {
		t.Error("CheckWritePermissions returned false for writable directory")
	}

	// Test non-existent directory
	nonExistent := filepath.Join(tempDir, "nonexistent")
	if CheckWritePermissions(nonExistent) {
		t.Error("CheckWritePermissions should return false for non-existent directory")
	}
}

func TestMergeMaps(t *testing.T) {
	existing := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	new := map[string]interface{}{
		"key2": "newvalue2",
		"key4": "value4",
	}

	result := mergeMaps(existing, new)

	// Verify existing keys are preserved unless overwritten
	if result["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", result["key1"])
	}
	if result["key3"] != "value3" {
		t.Errorf("Expected key3='value3', got %v", result["key3"])
	}

	// Verify new values overwrite existing
	if result["key2"] != "newvalue2" {
		t.Errorf("Expected key2='newvalue2', got %v", result["key2"])
	}

	// Verify new keys are added
	if result["key4"] != "value4" {
		t.Errorf("Expected key4='value4', got %v", result["key4"])
	}

	// Verify result has correct number of keys
	if len(result) != 4 {
		t.Errorf("Expected 4 keys in result, got %d", len(result))
	}
}
