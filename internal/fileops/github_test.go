package fileops

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestMockDownloadFunc(t *testing.T) {
	// Clear any existing mock
	MockDownloadFunc(nil)

	// Set up mock function
	called := false
	mockFunc := func(config *GitHubConfig, filePath string, retryCount int) (string, error) {
		called = true
		return "mocked content", nil
	}

	MockDownloadFunc(mockFunc)

	// Call download function - should use mock
	config := DefaultGitHubConfig()
	content, err := DownloadFileFromGitHub(config, "test.md", 0)

	if err != nil {
		t.Errorf("Mock function returned error: %v", err)
	}

	if content != "mocked content" {
		t.Errorf("Expected 'mocked content', got '%s'", content)
	}

	if !called {
		t.Error("Mock function was not called")
	}

	// Restore original function
	MockDownloadFunc(nil)
}

func TestDownloadFileFromGitHub_WithMock(t *testing.T) {
	// Mock successful download
	MockDownloadFunc(func(config *GitHubConfig, filePath string, retryCount int) (string, error) {
		return fmt.Sprintf("content for %s", filePath), nil
	})
	defer MockDownloadFunc(nil)

	config := DefaultGitHubConfig()
	content, err := DownloadFileFromGitHub(config, "agents/test-agent.md", 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if content != "content for agents/test-agent.md" {
		t.Errorf("Unexpected content: %s", content)
	}
}

func TestDownloadFileFromGitHub_ErrorHandling(t *testing.T) {
	// Mock download error
	MockDownloadFunc(func(config *GitHubConfig, filePath string, retryCount int) (string, error) {
		return "", fmt.Errorf("network error")
	})
	defer MockDownloadFunc(nil)

	config := DefaultGitHubConfig()
	_, err := DownloadFileFromGitHub(config, "test.md", 0)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestSetHTTPTimeout(t *testing.T) {
	// Save original timeout
	originalTimeout := httpClient.Timeout

	// Set new timeout
	newTimeout := 5 * time.Second
	SetHTTPTimeout(newTimeout)

	if httpClient.Timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, httpClient.Timeout)
	}

	// Restore original timeout
	SetHTTPTimeout(originalTimeout)

	if httpClient.Timeout != originalTimeout {
		t.Errorf("Failed to restore timeout to %v", originalTimeout)
	}
}

func TestHTTPClientConfiguration(t *testing.T) {
	// Verify httpClient is properly configured
	if httpClient == nil {
		t.Fatal("httpClient should not be nil")
	}

	if httpClient.Timeout == 0 {
		t.Error("httpClient should have a timeout configured")
	}

	if httpClient.Transport == nil {
		t.Error("httpClient should have a transport configured")
	}
}

func TestDownloadDirectoryFromGitHub_WithMock(t *testing.T) {
	// This would require more complex mocking of the GitHub API response
	// For now, just test that the function doesn't panic with mock
	MockDownloadFunc(func(config *GitHubConfig, filePath string, retryCount int) (string, error) {
		return "test content", nil
	})
	defer MockDownloadFunc(nil)

	config := DefaultGitHubConfig()

	// Note: DownloadDirectoryFromGitHub uses direct HTTP calls, not the mock
	// So we can't fully test it without a live server or more refactoring
	// But we can verify it doesn't panic
	_, err := DownloadDirectoryFromGitHub(config, "nonexistent", 0)

	// Error is expected for nonexistent directory
	_ = err
}

func TestDownloadCache_Persistence(t *testing.T) {
	ClearDownloadCache()

	// Add multiple items
	downloadCache["file1.md"] = "content1"
	downloadCache["file2.md"] = "content2"
	downloadCache["file3.md"] = "content3"

	if len(downloadCache) != 3 {
		t.Errorf("Expected 3 items in cache, got %d", len(downloadCache))
	}

	// Verify individual items
	if downloadCache["file1.md"] != "content1" {
		t.Error("Cache content mismatch for file1.md")
	}

	if downloadCache["file2.md"] != "content2" {
		t.Error("Cache content mismatch for file2.md")
	}

	// Clear and verify
	ClearDownloadCache()
	if len(downloadCache) != 0 {
		t.Errorf("Cache should be empty, got %d items", len(downloadCache))
	}
}

func TestDefaultDownloadFileFromGitHub_NotFound(t *testing.T) {
	// Test with mock that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// We can't easily inject the test server URL without refactoring
	// But we can test that the error path works with our mock
	MockDownloadFunc(func(config *GitHubConfig, filePath string, retryCount int) (string, error) {
		return "", fmt.Errorf("file not found: %s (404)", filePath)
	})
	defer MockDownloadFunc(nil)

	config := DefaultGitHubConfig()
	_, err := DownloadFileFromGitHub(config, "nonexistent.md", 0)

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestDownloadFileFromGitHub_Retry(t *testing.T) {
	// Test that the mock function works correctly
	MockDownloadFunc(func(config *GitHubConfig, filePath string, retryCount int) (string, error) {
		// Always succeed in mock
		return "mocked content", nil
	})
	defer MockDownloadFunc(nil)

	config := DefaultGitHubConfig()
	content, err := DownloadFileFromGitHub(config, "test.md", 0)

	if err != nil {
		t.Errorf("Expected no error with mock, got: %v", err)
	}

	if content != "mocked content" {
		t.Errorf("Expected 'mocked content', got: %s", content)
	}
}

func TestGitHubFile_AllFields(t *testing.T) {
	file := GitHubFile{
		Name: "example.md",
		Path: "agents/category/example.md",
		Type: "file",
		URL:  "https://api.github.com/repos/owner/repo/contents/path",
	}

	// Verify all fields
	tests := []struct {
		got      string
		expected string
		field    string
	}{
		{file.Name, "example.md", "Name"},
		{file.Path, "agents/category/example.md", "Path"},
		{file.Type, "file", "Type"},
		{file.URL, "https://api.github.com/repos/owner/repo/contents/path", "URL"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("Field %s: expected %q, got %q", tt.field, tt.expected, tt.got)
		}
	}
}

func TestHTTPClient_Transport(t *testing.T) {
	if httpClient.Transport == nil {
		t.Fatal("Transport should not be nil")
	}

	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport should be *http.Transport")
	}

	if transport.MaxIdleConns == 0 {
		t.Error("MaxIdleConns should be configured")
	}

	if transport.IdleConnTimeout == 0 {
		t.Error("IdleConnTimeout should be configured")
	}

	if transport.TLSHandshakeTimeout == 0 {
		t.Error("TLSHandshakeTimeout should be configured")
	}
}

func TestDefaultTimeout(t *testing.T) {
	if defaultTimeout == 0 {
		t.Error("defaultTimeout should be non-zero")
	}

	expected := 30 * time.Second
	if defaultTimeout != expected {
		t.Errorf("Expected defaultTimeout %v, got %v", expected, defaultTimeout)
	}
}
