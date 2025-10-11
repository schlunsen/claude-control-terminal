package fileops

import (
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
		TemplatesPath: "cli-tool/templates",
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

// DownloadFileFromGitHub downloads a single file from GitHub with retry logic
func DownloadFileFromGitHub(config *GitHubConfig, filePath string, retryCount int) (string, error) {
	// Check cache first
	if content, exists := downloadCache[filePath]; exists {
		return content, nil
	}

	maxRetries := 3
	baseDelay := 1 * time.Second
	retryDelay := baseDelay * time.Duration(1<<retryCount) // Exponential backoff: 1s, 2s, 4s

	githubURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/%s",
		config.Owner, config.Repo, config.Branch, config.TemplatesPath, filePath)

	resp, err := http.Get(githubURL)
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

// DownloadDirectoryFromGitHub downloads all files from a GitHub directory
func DownloadDirectoryFromGitHub(config *GitHubConfig, dirPath string, retryCount int) (map[string]string, error) {
	maxRetries := 5
	baseDelay := 2 * time.Second
	retryDelay := baseDelay * time.Duration(1<<retryCount) // Exponential backoff: 2s, 4s, 8s, 16s, 32s

	// GitHub API endpoint to get directory contents
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s/%s?ref=%s",
		config.Owner, config.Repo, config.TemplatesPath, dirPath, config.Branch)

	resp, err := http.Get(apiURL)
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

// ClearDownloadCache clears the download cache
func ClearDownloadCache() {
	downloadCache = make(map[string]string)
}
