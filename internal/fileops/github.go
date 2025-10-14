package fileops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubConfig holds configuration for downloading templates from GitHub
type GitHubConfig struct {
	Owner         string
	Repo          string
	Branch        string
	TemplatesPath string
}

// DefaultGitHubConfig returns the default GitHub configuration
func DefaultGitHubConfig() *GitHubConfig {
	return &GitHubConfig{
		Owner:         "davila7",
		Repo:          "claude-code-templates",
		Branch:        "main",
		TemplatesPath: "cli-tool",
	}
}

// GitHubFile represents a file from GitHub API
type GitHubFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// DownloadCache stores downloaded files to avoid repeated downloads
var downloadCache = make(map[string]string)

// defaultTimeout is the default timeout for HTTP requests
var defaultTimeout = 30 * time.Second

// httpClient is a shared HTTP client with timeout configuration
var httpClient = &http.Client{
	Timeout: defaultTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

// SetHTTPTimeout allows changing the HTTP client timeout (useful for testing)
func SetHTTPTimeout(timeout time.Duration) {
	httpClient.Timeout = timeout
	if httpClient.Transport != nil {
		if transport, ok := httpClient.Transport.(*http.Transport); ok {
			transport.TLSHandshakeTimeout = timeout / 3
		}
	}
}

// downloadFunc is the actual download function, can be overridden for testing
var downloadFunc func(config *GitHubConfig, filePath string, retryCount int) (string, error)

// DownloadFileFromGitHub downloads a single file from GitHub with retry logic and timeout.
// It uses exponential backoff for retries and caches successful downloads.
func DownloadFileFromGitHub(config *GitHubConfig, filePath string, retryCount int) (string, error) {
	if downloadFunc != nil {
		return downloadFunc(config, filePath, retryCount)
	}
	return defaultDownloadFileFromGitHub(config, filePath, retryCount)
}

// defaultDownloadFileFromGitHub is the actual implementation
func defaultDownloadFileFromGitHub(config *GitHubConfig, filePath string, retryCount int) (string, error) {
	// Check cache first
	if content, exists := downloadCache[filePath]; exists {
		return content, nil
	}

	maxRetries := 3
	baseDelay := 1 * time.Second
	retryDelay := baseDelay * time.Duration(1<<retryCount) // Exponential backoff: 1s, 2s, 4s

	githubURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/%s",
		config.Owner, config.Repo, config.Branch, config.TemplatesPath, filePath)

	// Create request with context and timeout (use the client's timeout)
	ctx, cancel := context.WithTimeout(context.Background(), httpClient.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", githubURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for %s: %w", filePath, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		// Network errors - retry if possible
		if retryCount < maxRetries {
			time.Sleep(retryDelay)
			return DownloadFileFromGitHub(config, filePath, retryCount+1)
		}
		return "", fmt.Errorf("network error downloading %s: %w", filePath, err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == 403 && retryCount < maxRetries {
		time.Sleep(retryDelay)
		return DownloadFileFromGitHub(config, filePath, retryCount+1)
	}

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("file not found: %s (404)", filePath)
	}

	if resp.StatusCode != 200 {
		if retryCount < maxRetries {
			time.Sleep(retryDelay)
			return DownloadFileFromGitHub(config, filePath, retryCount+1)
		}
		return "", fmt.Errorf("failed to download %s: status %d", filePath, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response for %s: %w", filePath, err)
	}

	content := string(body)
	downloadCache[filePath] = content
	return content, nil
}

// DownloadDirectoryFromGitHub downloads all files from a GitHub directory with timeout.
// It handles rate limiting and uses exponential backoff for retries.
func DownloadDirectoryFromGitHub(config *GitHubConfig, dirPath string, retryCount int) (map[string]string, error) {
	maxRetries := 5
	baseDelay := 2 * time.Second
	retryDelay := baseDelay * time.Duration(1<<retryCount) // Exponential backoff: 2s, 4s, 8s, 16s, 32s

	// GitHub API endpoint to get directory contents
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s/%s?ref=%s",
		config.Owner, config.Repo, config.TemplatesPath, dirPath, config.Branch)

	// Create request with context and timeout (use the client's timeout)
	ctx, cancel := context.WithTimeout(context.Background(), httpClient.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		if retryCount < maxRetries {
			time.Sleep(retryDelay)
			return DownloadDirectoryFromGitHub(config, dirPath, retryCount+1)
		}
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == 403 {
		rateLimitRemaining := resp.Header.Get("x-ratelimit-remaining")
		isRateLimit := rateLimitRemaining == "0" || strings.Contains(strings.ToLower(resp.Status), "rate limit")

		if isRateLimit && retryCount < maxRetries {
			waitTime := retryDelay

			// If we have reset time, calculate exact wait time
			if rateLimitReset := resp.Header.Get("x-ratelimit-reset"); rateLimitReset != "" {
				// Parse and use reset time (simplified version)
				waitTime = retryDelay
				if waitTime > 60*time.Second {
					waitTime = 60 * time.Second // Cap at 60 seconds
				}
			}

			time.Sleep(waitTime)
			return DownloadDirectoryFromGitHub(config, dirPath, retryCount+1)
		}

		if isRateLimit {
			return make(map[string]string), nil // Return empty map instead of error
		}

		return make(map[string]string), nil // Different 403 error
	}

	// 404 is ok for some directories
	if resp.StatusCode == 404 {
		return make(map[string]string), nil
	}

	if resp.StatusCode != 200 {
		if retryCount < maxRetries {
			time.Sleep(retryDelay)
			return DownloadDirectoryFromGitHub(config, dirPath, retryCount+1)
		}
		return nil, fmt.Errorf("failed to get directory listing: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var items []GitHubFile
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("error parsing directory listing: %w", err)
	}

	files := make(map[string]string)
	successCount := 0
	skipCount := 0

	for _, item := range items {
		if item.Type == "file" {
			// Extract relative path
			relativePath := strings.TrimPrefix(item.Path, config.TemplatesPath+"/")

			content, err := DownloadFileFromGitHub(config, relativePath, 0)
			if err != nil {
				skipCount++
				continue
			}

			files[item.Name] = content
			successCount++
		}
	}

	return files, nil
}

// ClearDownloadCache clears the download cache to force fresh downloads.
func ClearDownloadCache() {
	downloadCache = make(map[string]string)
}

// MockDownloadFunc allows tests to override the download function
func MockDownloadFunc(mockFunc func(config *GitHubConfig, filePath string, retryCount int) (string, error)) {
	downloadFunc = mockFunc
}
