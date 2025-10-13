package fileops

import (
	"testing"
)

func TestDefaultGitHubConfig(t *testing.T) {
	config := DefaultGitHubConfig()

	if config == nil {
		t.Fatal("DefaultGitHubConfig returned nil")
	}

	if config.Owner == "" {
		t.Error("Owner should not be empty")
	}

	if config.Repo == "" {
		t.Error("Repo should not be empty")
	}

	if config.Branch == "" {
		t.Error("Branch should not be empty")
	}

	if config.TemplatesPath == "" {
		t.Error("TemplatesPath should not be empty")
	}

	// Verify expected values
	if config.Owner != "davila7" {
		t.Errorf("Expected Owner 'davila7', got '%s'", config.Owner)
	}

	if config.Repo != "claude-code-templates" {
		t.Errorf("Expected Repo 'claude-code-templates', got '%s'", config.Repo)
	}

	if config.Branch != "main" {
		t.Errorf("Expected Branch 'main', got '%s'", config.Branch)
	}

	if config.TemplatesPath != "cli-tool" {
		t.Errorf("Expected TemplatesPath 'cli-tool', got '%s'", config.TemplatesPath)
	}
}

func TestGitHubFileStructure(t *testing.T) {
	// Test GitHubFile structure
	file := GitHubFile{
		Name: "test.md",
		Path: "agents/test.md",
		Type: "file",
		URL:  "https://api.github.com/repos/test/test/contents/agents/test.md",
	}

	if file.Name != "test.md" {
		t.Errorf("Expected Name 'test.md', got '%s'", file.Name)
	}

	if file.Type != "file" {
		t.Errorf("Expected Type 'file', got '%s'", file.Type)
	}

	if file.Path != "agents/test.md" {
		t.Errorf("Expected Path 'agents/test.md', got '%s'", file.Path)
	}
}

func TestClearDownloadCache(t *testing.T) {
	// Add something to cache
	downloadCache["test-key"] = "test-value"

	// Verify it's in cache
	if _, exists := downloadCache["test-key"]; !exists {
		t.Fatal("Test data not in cache")
	}

	// Clear cache
	ClearDownloadCache()

	// Verify cache is empty
	if len(downloadCache) != 0 {
		t.Errorf("Cache should be empty after clear, got %d items", len(downloadCache))
	}

	// Verify specific key is gone
	if _, exists := downloadCache["test-key"]; exists {
		t.Error("Cache should not contain test-key after clear")
	}
}

func TestDownloadFileCaching(t *testing.T) {
	// Clear cache first
	ClearDownloadCache()

	// Test that cache is initially empty
	if len(downloadCache) != 0 {
		t.Errorf("Cache should start empty, got %d items", len(downloadCache))
	}

	// Note: We can't test actual downloads without mocking HTTP
	// This test just verifies the cache structure exists and works
	testKey := "test/path.md"
	testValue := "test content"

	// Manually add to cache (simulating what DownloadFileFromGitHub does)
	downloadCache[testKey] = testValue

	// Verify it's cached
	if cached, exists := downloadCache[testKey]; !exists {
		t.Error("Value should be in cache")
	} else if cached != testValue {
		t.Errorf("Expected cached value '%s', got '%s'", testValue, cached)
	}

	// Clear and verify
	ClearDownloadCache()
	if _, exists := downloadCache[testKey]; exists {
		t.Error("Cache should be cleared")
	}
}

func TestGitHubConfigStructure(t *testing.T) {
	config := &GitHubConfig{
		Owner:         "test-owner",
		Repo:          "test-repo",
		Branch:        "test-branch",
		TemplatesPath: "test-path",
	}

	if config.Owner != "test-owner" {
		t.Errorf("Expected Owner 'test-owner', got '%s'", config.Owner)
	}

	if config.Repo != "test-repo" {
		t.Errorf("Expected Repo 'test-repo', got '%s'", config.Repo)
	}

	if config.Branch != "test-branch" {
		t.Errorf("Expected Branch 'test-branch', got '%s'", config.Branch)
	}

	if config.TemplatesPath != "test-path" {
		t.Errorf("Expected TemplatesPath 'test-path', got '%s'", config.TemplatesPath)
	}
}
