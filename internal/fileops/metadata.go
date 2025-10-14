package fileops

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MCPInstallation represents metadata about an installed MCP
type MCPInstallation struct {
	// InstallName is the name used during installation (e.g., "postgresql-integration")
	InstallName string `json:"installName"`
	// ServerKeys are the actual server keys added to .mcp.json (e.g., ["postgresql"])
	ServerKeys []string `json:"serverKeys"`
	// SourcePath is the GitHub path where the MCP was downloaded from
	SourcePath string `json:"sourcePath"`
	// InstalledAt is the timestamp when the MCP was installed
	InstalledAt time.Time `json:"installedAt"`
	// Scope indicates if it was installed as "project" or "user"
	Scope MCPScope `json:"scope"`
}

// MCPMetadata tracks all MCP installations
type MCPMetadata struct {
	// Installations maps install names to their metadata
	Installations map[string]MCPInstallation `json:"installations"`
}

// GetMCPMetadataPath returns the path for MCP metadata based on scope
func GetMCPMetadataPath(scope MCPScope, projectDir string) string {
	switch scope {
	case MCPScopeUser:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ".claude/.mcp-metadata.json"
		}
		return filepath.Join(homeDir, ".claude", ".mcp-metadata.json")
	case MCPScopeProject:
		fallthrough
	default:
		return filepath.Join(projectDir, ".mcp-metadata.json")
	}
}

// LoadMCPMetadata loads MCP metadata from the appropriate file
func LoadMCPMetadata(scope MCPScope, projectDir string) (*MCPMetadata, error) {
	metadataPath := GetMCPMetadataPath(scope, projectDir)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata if file doesn't exist
			return &MCPMetadata{
				Installations: make(map[string]MCPInstallation),
			}, nil
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata MCPMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	// Initialize Installations map if it's nil
	if metadata.Installations == nil {
		metadata.Installations = make(map[string]MCPInstallation)
	}

	return &metadata, nil
}

// SaveMCPMetadata saves MCP metadata to the appropriate file
func SaveMCPMetadata(scope MCPScope, projectDir string, metadata *MCPMetadata) error {
	metadataPath := GetMCPMetadataPath(scope, projectDir)

	// Ensure the directory exists first
	metadataDir := filepath.Dir(metadataPath)
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// If the metadata is empty, delete the file
	if metadata.Installations == nil || len(metadata.Installations) == 0 {
		if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove empty metadata file: %w", err)
		}
		return nil
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// AddMCPInstallation adds metadata about a newly installed MCP
func AddMCPInstallation(scope MCPScope, projectDir string, installName string, serverKeys []string, sourcePath string) error {
	metadata, err := LoadMCPMetadata(scope, projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Create installation record
	installation := MCPInstallation{
		InstallName: installName,
		ServerKeys:  serverKeys,
		SourcePath:  sourcePath,
		InstalledAt: time.Now(),
		Scope:       scope,
	}

	metadata.Installations[installName] = installation

	if err := SaveMCPMetadata(scope, projectDir, metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// RemoveMCPInstallation removes metadata about an uninstalled MCP
func RemoveMCPInstallation(scope MCPScope, projectDir string, installName string) error {
	metadata, err := LoadMCPMetadata(scope, projectDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	delete(metadata.Installations, installName)

	if err := SaveMCPMetadata(scope, projectDir, metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// GetMCPInstallation retrieves metadata for a specific MCP installation
func GetMCPInstallation(scope MCPScope, projectDir string, installName string) (*MCPInstallation, error) {
	metadata, err := LoadMCPMetadata(scope, projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	installation, ok := metadata.Installations[installName]
	if !ok {
		return nil, fmt.Errorf("MCP '%s' not found in metadata", installName)
	}

	return &installation, nil
}

// GetInstalledMCPs returns a list of all installed MCPs
func GetInstalledMCPs(scope MCPScope, projectDir string) ([]MCPInstallation, error) {
	metadata, err := LoadMCPMetadata(scope, projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	installations := make([]MCPInstallation, 0, len(metadata.Installations))
	for _, installation := range metadata.Installations {
		installations = append(installations, installation)
	}

	return installations, nil
}

// FindMCPByServerKey searches for an MCP installation that contains the given server key
func FindMCPByServerKey(scope MCPScope, projectDir string, serverKey string) (*MCPInstallation, error) {
	metadata, err := LoadMCPMetadata(scope, projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	for _, installation := range metadata.Installations {
		for _, key := range installation.ServerKeys {
			if key == serverKey {
				return &installation, nil
			}
		}
	}

	return nil, fmt.Errorf("no MCP found with server key '%s'", serverKey)
}
