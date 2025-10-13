package providers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schlunsen/claude-control-terminal/internal/database"
)

func TestGetAvailableProviders(t *testing.T) {
	providers := GetAvailableProviders()

	if len(providers) == 0 {
		t.Error("expected at least one provider")
	}

	// Check that default Claude provider exists
	hasClaudeProvider := false
	for _, p := range providers {
		if p.ID == "claude" {
			hasClaudeProvider = true
			break
		}
	}

	if !hasClaudeProvider {
		t.Error("expected Claude provider to be in the list")
	}
}

func TestGetProviderByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		shouldFind bool
	}{
		{
			name:       "find claude provider",
			id:         "claude",
			shouldFind: true,
		},
		{
			name:       "non-existent provider",
			id:         "non-existent-provider-xyz",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := GetProviderByID(tt.id)

			if tt.shouldFind && provider == nil {
				t.Errorf("expected to find provider with ID %q", tt.id)
			}

			if !tt.shouldFind && provider != nil {
				t.Errorf("expected not to find provider with ID %q", tt.id)
			}

			if provider != nil && provider.ID != tt.id {
				t.Errorf("expected provider ID %q, got %q", tt.id, provider.ID)
			}
		})
	}
}

func TestGetEnvScriptPath(t *testing.T) {
	path := GetEnvScriptPath()

	if path == "" {
		t.Error("expected non-empty script path")
	}

	if !filepath.IsAbs(path) && !filepath.HasPrefix(path, ".claude") {
		t.Errorf("expected absolute path or .claude prefix, got %q", path)
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "short key",
			apiKey:   "short",
			expected: "****",
		},
		{
			name:     "normal key",
			apiKey:   "sk-ant-1234567890abcdef",
			expected: "cdef",
		},
		{
			name:     "long key",
			apiKey:   "very-long-api-key-with-many-characters",
			expected: "ters",
		},
		{
			name:     "empty key",
			apiKey:   "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskAPIKey(tt.apiKey)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateEnvScriptClaude(t *testing.T) {
	// Create temp directory for script
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	config := &database.ProviderConfig{
		ProviderID: "claude",
		APIKey:     "",
		IsCurrent:  true,
		UpdatedAt:  time.Now(),
	}

	err := GenerateEnvScript(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify script was created
	scriptPath := GetEnvScriptPath()
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("script file was not created")
	}

	// Read and verify script content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "unset ANTHROPIC_AUTH_TOKEN") {
		t.Error("expected script to unset ANTHROPIC_AUTH_TOKEN")
	}

	if !contains(contentStr, "unset ANTHROPIC_BASE_URL") {
		t.Error("expected script to unset ANTHROPIC_BASE_URL")
	}
}

func TestGenerateEnvScriptNilConfig(t *testing.T) {
	err := GenerateEnvScript(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestProviderStruct(t *testing.T) {
	p := Provider{
		Name:         "Test Provider",
		ID:           "test",
		BaseURL:      "https://api.test.com",
		Icon:         "🧪",
		Models:       []string{"model-1", "model-2"},
		DefaultModel: "model-1",
		Description:  "Test provider for testing",
	}

	if p.Name != "Test Provider" {
		t.Errorf("expected name 'Test Provider', got %q", p.Name)
	}

	if p.ID != "test" {
		t.Errorf("expected ID 'test', got %q", p.ID)
	}

	if len(p.Models) != 2 {
		t.Errorf("expected 2 models, got %d", len(p.Models))
	}

	if p.DefaultModel != "model-1" {
		t.Errorf("expected default model 'model-1', got %q", p.DefaultModel)
	}
}

func TestProvidersConfigStruct(t *testing.T) {
	config := ProvidersConfig{
		Providers: []Provider{
			{ID: "provider1"},
			{ID: "provider2"},
		},
	}

	if len(config.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(config.Providers))
	}
}

func TestLoadProviderConfig(t *testing.T) {
	database.ResetInstance()
	tempDir := t.TempDir()

	db, err := database.Initialize(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	repo := database.NewRepository(db)

	// Test loading when no config exists
	config, err := LoadProviderConfig(repo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// config might be nil if no current provider
	_ = config
}

func TestSaveProviderConfig(t *testing.T) {
	database.ResetInstance()
	tempDir := t.TempDir()

	db, err := database.Initialize(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	repo := database.NewRepository(db)

	config := &database.ProviderConfig{
		ProviderID: "test-provider",
		APIKey:     "test-key",
		IsCurrent:  true,
		UpdatedAt:  time.Now(),
	}

	err = SaveProviderConfig(repo, config)
	if err != nil {
		t.Errorf("unexpected error saving config: %v", err)
	}

	// Verify it was saved
	loaded, err := GetProviderConfig(repo, "test-provider")
	if err != nil {
		t.Errorf("unexpected error loading config: %v", err)
	}

	if loaded != nil && loaded.ProviderID != "test-provider" {
		t.Errorf("expected provider ID 'test-provider', got %q", loaded.ProviderID)
	}
}

func TestGetCurrentProviderInfo(t *testing.T) {
	database.ResetInstance()
	tempDir := t.TempDir()

	db, err := database.Initialize(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	repo := database.NewRepository(db)

	// Test with no provider configured
	name, configured, err := GetCurrentProviderInfo(repo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_ = name
	_ = configured
}

func TestDeleteProviderConfig(t *testing.T) {
	database.ResetInstance()
	tempDir := t.TempDir()

	db, err := database.Initialize(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	repo := database.NewRepository(db)

	// Save a config first
	config := &database.ProviderConfig{
		ProviderID: "test-provider",
		APIKey:     "test-key",
		IsCurrent:  true,
		UpdatedAt:  time.Now(),
	}

	err = SaveProviderConfig(repo, config)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Delete it
	err = DeleteProviderConfig(repo)
	if err != nil {
		t.Errorf("unexpected error deleting config: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
