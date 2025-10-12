package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

// ComponentLoader handles loading component lists from GitHub
type ComponentLoader struct {
	config *fileops.GitHubConfig
	cache  *Cache
}

// NewComponentLoader creates a new component loader
func NewComponentLoader() *ComponentLoader {
	cache, _ := NewCache()

	return &ComponentLoader{
		config: fileops.DefaultGitHubConfig(),
		cache:  cache,
	}
}

// LoadComponents loads all available components of a specific type from GitHub using the Git Tree API
func (cl *ComponentLoader) LoadComponents(componentType, targetDir string) ([]ComponentItem, error) {
	return cl.LoadComponentsWithCache(componentType, targetDir, false)
}

// LoadComponentsWithCache loads components with optional cache bypass
func (cl *ComponentLoader) LoadComponentsWithCache(componentType, targetDir string, forceRefresh bool) ([]ComponentItem, error) {
	metadata := GetComponentMetadata()
	meta, ok := metadata[componentType]
	if !ok {
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	// Try to load from cache first (unless force refresh)
	if !forceRefresh && cl.cache != nil {
		components, found, err := cl.cache.Get(componentType)
		if err == nil && found {
			// Update installation status for cached components
			for i := range components {
				installedGlobal, installedProject := CheckInstallationStatus(components[i].Name, componentType, targetDir)
				components[i].InstalledGlobal = installedGlobal
				components[i].InstalledProject = installedProject
			}
			return components, nil
		}
	}

	// Cache miss or force refresh - fetch from GitHub
	// Use Git Trees API to get all files in one request (avoids rate limiting)
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1",
		cl.config.Owner, cl.config.Repo, cl.config.Branch)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tree: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch tree (status: %d)", resp.StatusCode)
	}

	var tree struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, fmt.Errorf("failed to decode tree: %w", err)
	}

	// Parse the tree to find matching component files
	var allComponents []ComponentItem
	pathPrefix := cl.config.TemplatesPath + "/" + meta.Path + "/"

	for _, item := range tree.Tree {
		if item.Type != "blob" {
			continue
		}

		// Check if path starts with our component path
		if !strings.HasPrefix(item.Path, pathPrefix) {
			continue
		}

		// Check if it has the correct extension
		if !strings.HasSuffix(item.Path, meta.Extension) {
			continue
		}

		// Extract relative path
		relPath := strings.TrimPrefix(item.Path, pathPrefix)
		parts := strings.Split(relPath, "/")

		var name, category string
		if len(parts) == 1 {
			// File in root
			name = strings.TrimSuffix(parts[0], meta.Extension)
			category = "root"
		} else if len(parts) == 2 {
			// File in category
			category = parts[0]
			name = strings.TrimSuffix(parts[1], meta.Extension)
		} else {
			// Skip deeply nested files
			continue
		}

		// Check installation status
		installedGlobal, installedProject := CheckInstallationStatus(name, componentType, targetDir)

		allComponents = append(allComponents, ComponentItem{
			Name:             name,
			Category:         category,
			Description:      fmt.Sprintf("%s from %s", name, category),
			Type:             componentType,
			Selected:         false,
			InstalledGlobal:  installedGlobal,
			InstalledProject: installedProject,
		})
	}

	// Save to cache
	if cl.cache != nil {
		_ = cl.cache.Set(componentType, allComponents)
	}

	return allComponents, nil
}

