package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry represents a cached component list
type CacheEntry struct {
	ComponentType string          `json:"component_type"`
	Components    []ComponentItem `json:"components"`
	Timestamp     time.Time       `json:"timestamp"`
}

// Cache handles local caching of component lists
type Cache struct {
	cacheDir string
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cache", "cct")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{cacheDir: cacheDir}, nil
}

// Get retrieves cached components for a type
func (c *Cache) Get(componentType string) ([]ComponentItem, bool, error) {
	cacheFile := filepath.Join(c.cacheDir, fmt.Sprintf("%s.json", componentType))

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil // Cache miss, not an error
		}
		return nil, false, fmt.Errorf("failed to read cache: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, fmt.Errorf("failed to parse cache: %w", err)
	}

	return entry.Components, true, nil
}

// Set stores components in the cache
func (c *Cache) Set(componentType string, components []ComponentItem) error {
	entry := CacheEntry{
		ComponentType: componentType,
		Components:    components,
		Timestamp:     time.Now(),
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	cacheFile := filepath.Join(c.cacheDir, fmt.Sprintf("%s.json", componentType))
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// GetAge returns the age of the cache for a component type
func (c *Cache) GetAge(componentType string) (time.Duration, error) {
	cacheFile := filepath.Join(c.cacheDir, fmt.Sprintf("%s.json", componentType))

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return 0, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return 0, err
	}

	return time.Since(entry.Timestamp), nil
}

// Clear removes all cached data
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(c.cacheDir, entry.Name())); err != nil {
				return fmt.Errorf("failed to remove cache file: %w", err)
			}
		}
	}

	return nil
}
