// Package installer provides Node.js detection functionality for npm-based Claude CLI installation.
package installer

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NodeInfo contains information about the Node.js installation
type NodeInfo struct {
	// Installed indicates if Node.js is found in PATH
	Installed bool
	// Version is the Node.js version string (e.g., "v18.17.0")
	Version string
	// VersionMajor is the major version number (e.g., 18)
	VersionMajor int
	// VersionMinor is the minor version number (e.g., 17)
	VersionMinor int
	// VersionPatch is the patch version number (e.g., 0)
	VersionPatch int
	// VersionOK indicates if version meets minimum requirements (v18+)
	VersionOK bool
	// NodePath is the full path to the node executable
	NodePath string
	// NPMAvailable indicates if npm is available
	NPMAvailable bool
	// NPMVersion is the npm version string
	NPMVersion string
	// NPMPath is the full path to the npm executable
	NPMPath string
}

// NodeDetector handles Node.js detection and version checking
type NodeDetector struct {
	// MinMajorVersion is the minimum required major version (default: 18)
	MinMajorVersion int
	// Timeout for detection operations
	Timeout time.Duration
}

// NewNodeDetector creates a new Node.js detector with default settings
func NewNodeDetector() *NodeDetector {
	return &NodeDetector{
		MinMajorVersion: 18, // Claude CLI requires Node.js v18+
		Timeout:         10 * time.Second,
	}
}

// IsNodeInstalled checks if Node.js is available in PATH
func (nd *NodeDetector) IsNodeInstalled() bool {
	_, err := exec.LookPath("node")
	return err == nil
}

// IsNPMInstalled checks if npm is available in PATH
func (nd *NodeDetector) IsNPMInstalled() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

// GetNodePath returns the full path to the node executable
func (nd *NodeDetector) GetNodePath() (string, error) {
	path, err := exec.LookPath("node")
	if err != nil {
		return "", fmt.Errorf("node not found in PATH: %w", err)
	}
	return path, nil
}

// GetNPMPath returns the full path to the npm executable
func (nd *NodeDetector) GetNPMPath() (string, error) {
	path, err := exec.LookPath("npm")
	if err != nil {
		return "", fmt.Errorf("npm not found in PATH: %w", err)
	}
	return path, nil
}

// GetNodeVersion runs 'node --version' and returns the version string
func (nd *NodeDetector) GetNodeVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nd.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Node.js version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetNPMVersion runs 'npm --version' and returns the version string
func (nd *NodeDetector) GetNPMVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), nd.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npm", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get npm version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ParseNodeVersion parses a Node.js version string (e.g., "v18.17.0") into major, minor, patch
func (nd *NodeDetector) ParseNodeVersion(version string) (major, minor, patch int, err error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Match version pattern (e.g., "18.17.0")
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)

	if len(matches) < 4 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err = strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
	}

	return major, minor, patch, nil
}

// CheckVersionRequirement checks if the given version meets minimum requirements
func (nd *NodeDetector) CheckVersionRequirement(major int) bool {
	return major >= nd.MinMajorVersion
}

// DetectNode performs comprehensive Node.js detection and returns NodeInfo
func (nd *NodeDetector) DetectNode() *NodeInfo {
	info := &NodeInfo{
		Installed:    false,
		VersionOK:    false,
		NPMAvailable: false,
	}

	// Check if Node.js is installed
	if !nd.IsNodeInstalled() {
		return info
	}

	info.Installed = true

	// Get Node.js path
	nodePath, err := nd.GetNodePath()
	if err == nil {
		info.NodePath = nodePath
	}

	// Get Node.js version
	version, err := nd.GetNodeVersion()
	if err != nil {
		// Node is installed but can't get version - something is wrong
		return info
	}

	info.Version = version

	// Parse version
	major, minor, patch, err := nd.ParseNodeVersion(version)
	if err != nil {
		// Can't parse version - assume incompatible
		return info
	}

	info.VersionMajor = major
	info.VersionMinor = minor
	info.VersionPatch = patch

	// Check if version meets requirements
	info.VersionOK = nd.CheckVersionRequirement(major)

	// Check npm availability
	info.NPMAvailable = nd.IsNPMInstalled()
	if info.NPMAvailable {
		npmPath, err := nd.GetNPMPath()
		if err == nil {
			info.NPMPath = npmPath
		}

		npmVersion, err := nd.GetNPMVersion()
		if err == nil {
			info.NPMVersion = npmVersion
		}
	}

	return info
}

// GetNodeInstallInstructions returns platform-specific instructions for installing Node.js
func GetNodeInstallInstructions() string {
	return `Node.js v18+ is required for npm-based installation.

To install Node.js:

macOS:
  # Using Homebrew
  brew install node

  # Using nvm (Node Version Manager)
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
  nvm install 18
  nvm use 18

Linux:
  # Using nvm (recommended)
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
  nvm install 18
  nvm use 18

  # Using apt (Ubuntu/Debian)
  curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
  sudo apt-get install -y nodejs

  # Using dnf (Fedora)
  sudo dnf install nodejs

Windows:
  # Download installer from nodejs.org
  https://nodejs.org/en/download/

  # Or using Chocolatey
  choco install nodejs

After installation, restart your terminal and run 'node --version' to verify.`
}

// GetRecommendation returns a recommendation based on Node.js availability
func (nd *NodeDetector) GetRecommendation() string {
	info := nd.DetectNode()

	if !info.Installed {
		return "Node.js not found. Recommend using native binary installation (no Node.js required)."
	}

	if !info.VersionOK {
		return fmt.Sprintf("Node.js %s is too old (v%d+ required). Recommend using native binary installation or upgrading Node.js.",
			info.Version, nd.MinMajorVersion)
	}

	if !info.NPMAvailable {
		return "Node.js found but npm is missing. Recommend using native binary installation or installing npm."
	}

	return fmt.Sprintf("Node.js %s with npm %s found. npm-based installation is available.",
		info.Version, info.NPMVersion)
}

// FormatNodeInfo returns a human-readable summary of Node.js detection
func (nd *NodeDetector) FormatNodeInfo() string {
	info := nd.DetectNode()

	if !info.Installed {
		return "Node.js: Not installed"
	}

	status := "✗ Too old"
	if info.VersionOK {
		status = "✓ Compatible"
	}

	npmStatus := "✗ Not found"
	if info.NPMAvailable {
		npmStatus = fmt.Sprintf("✓ %s", info.NPMVersion)
	}

	return fmt.Sprintf(`Node.js: %s %s
  Path: %s
  npm: %s`,
		info.Version, status, info.NodePath, npmStatus)
}
