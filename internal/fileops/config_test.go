package fileops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPScopeConstants(t *testing.T) {
	if MCPScopeProject != "project" {
		t.Errorf("Expected MCPScopeProject to be 'project', got '%s'", MCPScopeProject)
	}

	if MCPScopeUser != "user" {
		t.Errorf("Expected MCPScopeUser to be 'user', got '%s'", MCPScopeUser)
	}
}

func TestLoadMCPConfig_NonExistent(t *testing.T) {
	// Test loading non-existent file
	config, err := LoadMCPConfig("/nonexistent/path/.mcp.json")
	if err != nil {
		t.Errorf("LoadMCPConfig should not error for non-existent file: %v", err)
	}

	if config == nil {
		t.Fatal("LoadMCPConfig should return empty config for non-existent file")
	}

	if config.MCPServers == nil {
		t.Error("MCPServers map should be initialized")
	}

	if len(config.MCPServers) != 0 {
		t.Errorf("MCPServers should be empty, got %d servers", len(config.MCPServers))
	}
}

func TestLoadMCPConfig_ValidFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_loadmcp_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, ".mcp.json")
	testConfig := `{
		"mcpServers": {
			"test-server": {
				"command": "node",
				"args": ["server.js"],
				"description": "Test server"
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	config, err := LoadMCPConfig(configPath)
	if err != nil {
		t.Errorf("LoadMCPConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("LoadMCPConfig returned nil")
	}

	if len(config.MCPServers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(config.MCPServers))
	}

	server, ok := config.MCPServers["test-server"]
	if !ok {
		t.Fatal("test-server not found in config")
	}

	if server.Command != "node" {
		t.Errorf("Expected command 'node', got '%s'", server.Command)
	}

	if server.Description != "Test server" {
		t.Errorf("Expected description 'Test server', got '%s'", server.Description)
	}
}

func TestLoadMCPConfig_InvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_invalidjson_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, ".mcp.json")
	invalidJSON := `{ invalid json }`

	if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err = LoadMCPConfig(configPath)
	if err == nil {
		t.Error("LoadMCPConfig should error for invalid JSON")
	}
}

func TestSaveMCPConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_savemcp_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, ".mcp.json")
	config := &ClaudeConfig{
		MCPServers: map[string]MCPServerConfig{
			"my-server": {
				Command:     "python",
				Args:        []string{"server.py"},
				Description: "My server",
			},
		},
	}

	err = SaveMCPConfig(configPath, config)
	if err != nil {
		t.Errorf("SaveMCPConfig failed: %v", err)
	}

	// Verify file was created
	if !FileExists(configPath) {
		t.Fatal("Config file was not created")
	}

	// Verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var loadedConfig ClaudeConfig
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("Failed to parse saved config: %v", err)
	}

	if len(loadedConfig.MCPServers) != 1 {
		t.Errorf("Expected 1 server in saved config, got %d", len(loadedConfig.MCPServers))
	}

	server, ok := loadedConfig.MCPServers["my-server"]
	if !ok {
		t.Fatal("my-server not found in saved config")
	}

	if server.Command != "python" {
		t.Errorf("Expected command 'python', got '%s'", server.Command)
	}
}

func TestSaveMCPConfig_CreatesDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_savemcp_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Path with non-existent subdirectory
	configPath := filepath.Join(tempDir, "subdir", ".mcp.json")
	config := &ClaudeConfig{
		MCPServers: make(map[string]MCPServerConfig),
	}

	err = SaveMCPConfig(configPath, config)
	if err != nil {
		t.Errorf("SaveMCPConfig should create directory: %v", err)
	}

	// Verify directory was created
	if !DirExists(filepath.Dir(configPath)) {
		t.Error("SaveMCPConfig did not create directory")
	}

	// Verify file was created
	if !FileExists(configPath) {
		t.Error("Config file was not created")
	}
}

func TestGetMCPConfigPath_Project(t *testing.T) {
	projectDir := "/test/project"
	path := GetMCPConfigPath(MCPScopeProject, projectDir)

	expected := filepath.Join(projectDir, ".mcp.json")
	if path != expected {
		t.Errorf("Expected path '%s', got '%s'", expected, path)
	}
}

func TestGetMCPConfigPath_User(t *testing.T) {
	projectDir := "/test/project"
	path := GetMCPConfigPath(MCPScopeUser, projectDir)

	// Should contain .claude/config.json
	expectedSuffix := filepath.Join(".claude", "config.json")
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("User config path should end with %s, got '%s'", expectedSuffix, path)
	}
}

func TestGetMCPConfigPath_Default(t *testing.T) {
	projectDir := "/test/project"
	// Pass invalid scope, should default to project
	path := GetMCPConfigPath(MCPScope("invalid"), projectDir)

	expected := filepath.Join(projectDir, ".mcp.json")
	if path != expected {
		t.Errorf("Invalid scope should default to project, expected '%s', got '%s'", expected, path)
	}
}

func TestAddMCPServerToConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_addserver_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	serverConfig := MCPServerConfig{
		Command:     "node",
		Args:        []string{"index.js"},
		Description: "Test MCP server",
	}

	err = AddMCPServerToConfig(MCPScopeProject, tempDir, "test-mcp", serverConfig)
	if err != nil {
		t.Errorf("AddMCPServerToConfig failed: %v", err)
	}

	// Verify server was added
	configPath := GetMCPConfigPath(MCPScopeProject, tempDir)
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	server, ok := config.MCPServers["test-mcp"]
	if !ok {
		t.Fatal("test-mcp not found in config")
	}

	if server.Command != "node" {
		t.Errorf("Expected command 'node', got '%s'", server.Command)
	}

	if server.Description != "Test MCP server" {
		t.Errorf("Expected description 'Test MCP server', got '%s'", server.Description)
	}
}

func TestAddMCPServerToConfig_Update(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_updateserver_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add initial server
	serverConfig := MCPServerConfig{
		Command: "node",
		Args:    []string{"v1.js"},
	}
	AddMCPServerToConfig(MCPScopeProject, tempDir, "my-server", serverConfig)

	// Update server
	updatedConfig := MCPServerConfig{
		Command: "node",
		Args:    []string{"v2.js"},
	}
	err = AddMCPServerToConfig(MCPScopeProject, tempDir, "my-server", updatedConfig)
	if err != nil {
		t.Errorf("AddMCPServerToConfig update failed: %v", err)
	}

	// Verify update
	configPath := GetMCPConfigPath(MCPScopeProject, tempDir)
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	server := config.MCPServers["my-server"]
	if len(server.Args) != 1 || server.Args[0] != "v2.js" {
		t.Errorf("Server was not updated correctly, got args: %v", server.Args)
	}
}

func TestRemoveMCPServerFromConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_removeserver_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add servers
	server1 := MCPServerConfig{Command: "node", Args: []string{"s1.js"}}
	server2 := MCPServerConfig{Command: "python", Args: []string{"s2.py"}}

	AddMCPServerToConfig(MCPScopeProject, tempDir, "server1", server1)
	AddMCPServerToConfig(MCPScopeProject, tempDir, "server2", server2)

	// Remove server1
	err = RemoveMCPServerFromConfig(MCPScopeProject, tempDir, "server1")
	if err != nil {
		t.Errorf("RemoveMCPServerFromConfig failed: %v", err)
	}

	// Verify removal
	configPath := GetMCPConfigPath(MCPScopeProject, tempDir)
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if _, ok := config.MCPServers["server1"]; ok {
		t.Error("server1 should have been removed")
	}

	if _, ok := config.MCPServers["server2"]; !ok {
		t.Error("server2 should still be present")
	}
}

func TestMergeMCPServersFromJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_merge_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add existing server
	existingServer := MCPServerConfig{Command: "existing"}
	AddMCPServerToConfig(MCPScopeProject, tempDir, "existing", existingServer)

	// Merge new servers
	mcpJSON := `{
		"mcpServers": {
			"new1": {
				"command": "node",
				"args": ["n1.js"]
			},
			"new2": {
				"command": "python",
				"args": ["n2.py"]
			}
		}
	}`

	added, err := MergeMCPServersFromJSON(MCPScopeProject, tempDir, mcpJSON)
	if err != nil {
		t.Errorf("MergeMCPServersFromJSON failed: %v", err)
	}

	if len(added) != 2 {
		t.Errorf("Expected 2 servers added, got %d", len(added))
	}

	// Verify merge
	configPath := GetMCPConfigPath(MCPScopeProject, tempDir)
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.MCPServers) != 3 {
		t.Errorf("Expected 3 servers total, got %d", len(config.MCPServers))
	}

	if _, ok := config.MCPServers["existing"]; !ok {
		t.Error("existing server should still be present")
	}
	if _, ok := config.MCPServers["new1"]; !ok {
		t.Error("new1 should be added")
	}
	if _, ok := config.MCPServers["new2"]; !ok {
		t.Error("new2 should be added")
	}
}

func TestMergeMCPServersFromJSON_InvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_merge_invalid_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidJSON := `{ invalid json }`

	_, err = MergeMCPServersFromJSON(MCPScopeProject, tempDir, invalidJSON)
	if err == nil {
		t.Error("MergeMCPServersFromJSON should error for invalid JSON")
	}
}

func TestRemoveMCPServers(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_removeservers_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Add servers
	AddMCPServerToConfig(MCPScopeProject, tempDir, "github", MCPServerConfig{Command: "node"})
	AddMCPServerToConfig(MCPScopeProject, tempDir, "github-pro", MCPServerConfig{Command: "node"})
	AddMCPServerToConfig(MCPScopeProject, tempDir, "other", MCPServerConfig{Command: "python"})

	// Remove servers matching "github"
	removed, err := RemoveMCPServers(MCPScopeProject, tempDir, "github")
	if err != nil {
		t.Errorf("RemoveMCPServers failed: %v", err)
	}

	if len(removed) == 0 {
		t.Error("Should have removed at least one server")
	}

	// Verify removal
	configPath := GetMCPConfigPath(MCPScopeProject, tempDir)
	config, err := LoadMCPConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if _, ok := config.MCPServers["github"]; ok {
		t.Error("github should have been removed")
	}

	if _, ok := config.MCPServers["other"]; !ok {
		t.Error("other should still be present")
	}
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"github", "github", true},
		{"github-pro", "github", true},
		{"my-github", "github", true},
		{"other", "github", false},
		{"", "test", false},
		{"test", "", true}, // Empty substring is contained in any string
	}

	for _, tt := range tests {
		t.Run(tt.s+"_contains_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestMCPServerConfig_AllFields(t *testing.T) {
	config := MCPServerConfig{
		Description: "Test description",
		Command:     "node",
		Args:        []string{"arg1", "arg2"},
		URL:         "https://example.com",
		Transport:   "stdio",
		Env:         map[string]string{"KEY": "value"},
	}

	if config.Description != "Test description" {
		t.Errorf("Expected Description 'Test description', got '%s'", config.Description)
	}
	if config.Command != "node" {
		t.Errorf("Expected Command 'node', got '%s'", config.Command)
	}
	if len(config.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(config.Args))
	}
	if config.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", config.URL)
	}
	if config.Transport != "stdio" {
		t.Errorf("Expected Transport 'stdio', got '%s'", config.Transport)
	}
	if config.Env["KEY"] != "value" {
		t.Errorf("Expected Env KEY='value', got '%s'", config.Env["KEY"])
	}
}
