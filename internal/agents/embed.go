package agents

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Embed the agents_server directory
//
//go:embed agents_server/*.py agents_server/*.toml agents_server/*.md agents_server/src/*.py
var agentsServerFS embed.FS

// Version of the embedded agent server (should match pyproject.toml)
const EmbeddedVersion = "0.1.0"

// ExtractAgentServer extracts the embedded agent server to the target directory.
// It creates the directory if it doesn't exist and overwrites existing files.
//
// Returns the path to the extracted server and any error encountered.
func ExtractAgentServer(targetDir string) (string, error) {
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	// Walk through the embedded filesystem
	err := fs.WalkDir(agentsServerFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root "." entry
		if path == "." {
			return nil
		}

		// Remove the "agents_server/" prefix from the path
		relPath := path
		if len(path) > 14 && path[:14] == "agents_server/" {
			relPath = path[14:]
		}

		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, 0755)
		}

		// Read file from embedded FS
		data, err := agentsServerFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
		}

		// Write file to target directory
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to extract agent server: %w", err)
	}

	return targetDir, nil
}

// WriteVersionFile writes the version file to track the installed version
func WriteVersionFile(targetDir string) error {
	versionPath := filepath.Join(targetDir, ".version")
	return os.WriteFile(versionPath, []byte(EmbeddedVersion), 0644)
}

// ReadVersionFile reads the installed version from the version file
func ReadVersionFile(targetDir string) (string, error) {
	versionPath := filepath.Join(targetDir, ".version")
	data, err := os.ReadFile(versionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No version file = not installed
		}
		return "", err
	}
	return string(data), nil
}

// NeedsUpdate checks if the installed version needs updating
func NeedsUpdate(targetDir string) (bool, error) {
	installedVersion, err := ReadVersionFile(targetDir)
	if err != nil {
		return false, err
	}

	// Empty version means not installed
	if installedVersion == "" {
		return true, nil
	}

	// Check if versions differ
	return installedVersion != EmbeddedVersion, nil
}
