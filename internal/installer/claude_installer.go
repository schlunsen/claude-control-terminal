// Package installer provides Claude CLI detection and installation functionality.
// It supports both native binary installation (preferred) and npm-based installation (fallback).
package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// InstallMethod represents the method used to install Claude CLI
type InstallMethod string

const (
	// InstallMethodNative uses the official install script (no Node.js required)
	InstallMethodNative InstallMethod = "native"
	// InstallMethodNPM uses npm to install Claude CLI (requires Node.js v18+)
	InstallMethodNPM InstallMethod = "npm"
)

// ClaudeInstaller handles Claude CLI detection and installation
type ClaudeInstaller struct {
	// Verbose enables detailed logging
	Verbose bool
	// Timeout for installation operations
	Timeout time.Duration
}

// NewClaudeInstaller creates a new installer with default settings
func NewClaudeInstaller() *ClaudeInstaller {
	return &ClaudeInstaller{
		Verbose: false,
		Timeout: 5 * time.Minute,
	}
}

// InstallResult contains information about the installation result
type InstallResult struct {
	Success       bool
	Method        InstallMethod
	ClaudePath    string
	Version       string
	Message       string
	Error         error
}

// IsClaudeInstalled checks if Claude CLI is available in PATH
func (ci *ClaudeInstaller) IsClaudeInstalled() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// GetClaudePath returns the full path to the Claude CLI binary
func (ci *ClaudeInstaller) GetClaudePath() (string, error) {
	path, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude CLI not found in PATH: %w", err)
	}
	return path, nil
}

// GetClaudeVersion runs 'claude --version' and returns the version string
func (ci *ClaudeInstaller) GetClaudeVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Claude version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// VerifyInstallation runs 'claude doctor' to verify the installation
func (ci *ClaudeInstaller) VerifyInstallation() error {
	if !ci.IsClaudeInstalled() {
		return fmt.Errorf("claude CLI not found in PATH")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "doctor")
	output, err := cmd.CombinedOutput()

	if ci.Verbose {
		fmt.Printf("Claude doctor output:\n%s\n", string(output))
	}

	if err != nil {
		return fmt.Errorf("claude doctor failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Install attempts to install Claude CLI using the specified method
func (ci *ClaudeInstaller) Install(method InstallMethod) *InstallResult {
	result := &InstallResult{
		Method: method,
	}

	// Check if already installed
	if ci.IsClaudeInstalled() {
		path, _ := ci.GetClaudePath()
		version, _ := ci.GetClaudeVersion()
		result.Success = true
		result.ClaudePath = path
		result.Version = version
		result.Message = "Claude CLI is already installed"
		return result
	}

	// Perform installation based on method
	switch method {
	case InstallMethodNative:
		return ci.installNative()
	case InstallMethodNPM:
		return ci.installNPM()
	default:
		result.Error = fmt.Errorf("unknown installation method: %s", method)
		return result
	}
}

// installNative installs Claude CLI using the official install script
func (ci *ClaudeInstaller) installNative() *InstallResult {
	result := &InstallResult{
		Method: InstallMethodNative,
	}

	if ci.Verbose {
		fmt.Println("Installing Claude CLI using native binary method...")
	}

	// Determine the install script URL based on OS
	var installScriptURL string
	var shellCommand string

	switch runtime.GOOS {
	case "darwin", "linux":
		installScriptURL = "https://claude.ai/install.sh"
		shellCommand = "bash"
	case "windows":
		// Windows uses PowerShell script
		installScriptURL = "https://claude.ai/install.ps1"
		shellCommand = "powershell"
	default:
		result.Error = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		return result
	}

	// Download the install script
	ctx, cancel := context.WithTimeout(context.Background(), ci.Timeout)
	defer cancel()

	if ci.Verbose {
		fmt.Printf("Downloading install script from %s...\n", installScriptURL)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", installScriptURL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to download install script: %w", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("failed to download install script: HTTP %d", resp.StatusCode)
		return result
	}

	// Save script to temporary file
	tmpFile, err := os.CreateTemp("", "claude-install-*")
	if err != nil {
		result.Error = fmt.Errorf("failed to create temporary file: %w", err)
		return result
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		result.Error = fmt.Errorf("failed to save install script: %w", err)
		return result
	}

	if err := tmpFile.Close(); err != nil {
		result.Error = fmt.Errorf("failed to close temporary file: %w", err)
		return result
	}

	// Make script executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			result.Error = fmt.Errorf("failed to make script executable: %w", err)
			return result
		}
	}

	// Execute the install script
	if ci.Verbose {
		fmt.Println("Running install script...")
	}

	installCtx, installCancel := context.WithTimeout(context.Background(), ci.Timeout)
	defer installCancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(installCtx, shellCommand, "-ExecutionPolicy", "Bypass", "-File", tmpFile.Name())
	} else {
		cmd = exec.CommandContext(installCtx, shellCommand, tmpFile.Name())
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		result.Error = fmt.Errorf("installation script failed: %w", err)
		return result
	}

	// Verify installation
	if !ci.IsClaudeInstalled() {
		result.Error = fmt.Errorf("installation completed but claude CLI not found in PATH")
		result.Message = "You may need to restart your terminal or source your shell profile"
		return result
	}

	// Get installed version and path
	path, _ := ci.GetClaudePath()
	version, _ := ci.GetClaudeVersion()

	result.Success = true
	result.ClaudePath = path
	result.Version = version
	result.Message = "Claude CLI installed successfully using native binary method"

	if ci.Verbose {
		fmt.Printf("Installation successful! Claude path: %s, Version: %s\n", path, version)
	}

	return result
}

// installNPM installs Claude CLI using npm
func (ci *ClaudeInstaller) installNPM() *InstallResult {
	result := &InstallResult{
		Method: InstallMethodNPM,
	}

	if ci.Verbose {
		fmt.Println("Installing Claude CLI using npm...")
	}

	// Check if Node.js and npm are available
	detector := NewNodeDetector()
	nodeInfo := detector.DetectNode()

	if !nodeInfo.Installed {
		result.Error = fmt.Errorf("node.js not found, please install Node.js v18+ or use native installation method")
		return result
	}

	if !nodeInfo.VersionOK {
		result.Error = fmt.Errorf("node.js version %s is too old, required: v18+", nodeInfo.Version)
		return result
	}

	if !nodeInfo.NPMAvailable {
		result.Error = fmt.Errorf("npm not found. Please ensure npm is installed")
		return result
	}

	// Install Claude CLI via npm
	ctx, cancel := context.WithTimeout(context.Background(), ci.Timeout)
	defer cancel()

	if ci.Verbose {
		fmt.Println("Running: npm install -g @anthropic-ai/claude-code")
	}

	cmd := exec.CommandContext(ctx, "npm", "install", "-g", "@anthropic-ai/claude-code")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		result.Error = fmt.Errorf("npm install failed: %w", err)
		result.Message = "Try running with sudo or use native installation method"
		return result
	}

	// Verify installation
	if !ci.IsClaudeInstalled() {
		result.Error = fmt.Errorf("npm install completed but claude CLI not found in PATH")
		result.Message = "You may need to restart your terminal or check npm global bin path"
		return result
	}

	// Get installed version and path
	path, _ := ci.GetClaudePath()
	version, _ := ci.GetClaudeVersion()

	result.Success = true
	result.ClaudePath = path
	result.Version = version
	result.Message = "Claude CLI installed successfully using npm"

	if ci.Verbose {
		fmt.Printf("Installation successful! Claude path: %s, Version: %s\n", path, version)
	}

	return result
}

// AutoInstall attempts to install Claude CLI using the best available method
func (ci *ClaudeInstaller) AutoInstall() *InstallResult {
	// Check if already installed
	if ci.IsClaudeInstalled() {
		path, _ := ci.GetClaudePath()
		version, _ := ci.GetClaudeVersion()
		return &InstallResult{
			Success:    true,
			ClaudePath: path,
			Version:    version,
			Message:    "Claude CLI is already installed",
		}
	}

	// Try native installation first (no Node.js dependency)
	if ci.Verbose {
		fmt.Println("Attempting native binary installation (preferred method)...")
	}
	result := ci.installNative()
	if result.Success {
		return result
	}

	// If native fails, try npm as fallback
	if ci.Verbose {
		fmt.Printf("Native installation failed: %v\n", result.Error)
		fmt.Println("Attempting npm installation as fallback...")
	}

	npmResult := ci.installNPM()
	if npmResult.Success {
		return npmResult
	}

	// Both methods failed
	return &InstallResult{
		Success: false,
		Error: fmt.Errorf("all installation methods failed. Native: %v, NPM: %v",
			result.Error, npmResult.Error),
		Message: "Please install Claude CLI manually: https://docs.claude.com/en/docs/claude-code/setup",
	}
}

// GetInstallInstructions returns platform-specific installation instructions
func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin", "linux":
		return `To install Claude CLI manually:

Native Binary (Recommended - no Node.js required):
  curl -fsSL https://claude.ai/install.sh | bash

NPM (Requires Node.js v18+):
  npm install -g @anthropic-ai/claude-code

After installation, restart your terminal or source your shell profile.`

	case "windows":
		return `To install Claude CLI manually:

Native Binary (PowerShell):
  irm https://claude.ai/install.ps1 | iex

NPM (Requires Node.js v18+):
  npm install -g @anthropic-ai/claude-code

After installation, restart your terminal.`

	default:
		return "Please visit https://docs.claude.com/en/docs/claude-code/setup for installation instructions."
	}
}

// GetRecommendedMethod returns the recommended installation method for the current platform
func GetRecommendedMethod() InstallMethod {
	// Native is preferred on all platforms as it doesn't require Node.js
	return InstallMethodNative
}

// CleanupOldInstallation removes old npm-based installation if exists
// Useful when migrating to native binary
func (ci *ClaudeInstaller) CleanupOldInstallation() error {
	// Check if npm-based installation exists
	detector := NewNodeDetector()
	nodeInfo := detector.DetectNode()

	if !nodeInfo.Installed || !nodeInfo.NPMAvailable {
		return nil // No npm, nothing to cleanup
	}

	// Check if Claude is installed via npm
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0", "@anthropic-ai/claude-code")
	output, err := cmd.Output()

	// If not found via npm, nothing to cleanup
	if err != nil || !strings.Contains(string(output), "@anthropic-ai/claude-code") {
		return nil
	}

	if ci.Verbose {
		fmt.Println("Found npm-based Claude installation, cleaning up...")
	}

	// Uninstall via npm
	uninstallCmd := exec.CommandContext(ctx, "npm", "uninstall", "-g", "@anthropic-ai/claude-code")
	uninstallCmd.Stdout = os.Stdout
	uninstallCmd.Stderr = os.Stderr

	if err := uninstallCmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall npm package: %w", err)
	}

	if ci.Verbose {
		fmt.Println("Old npm installation cleaned up successfully")
	}

	return nil
}

// CheckForUpdates checks if Claude CLI has updates available
func (ci *ClaudeInstaller) CheckForUpdates() (bool, string, error) {
	if !ci.IsClaudeInstalled() {
		return false, "", fmt.Errorf("claude CLI not installed")
	}

	// Claude CLI auto-updates itself, so we just check current version
	version, err := ci.GetClaudeVersion()
	if err != nil {
		return false, "", fmt.Errorf("failed to get version: %w", err)
	}

	// Note: Claude CLI handles updates automatically
	return false, version, nil
}

// GetInstallLocation returns the directory where Claude CLI is installed
func (ci *ClaudeInstaller) GetInstallLocation() (string, error) {
	claudePath, err := ci.GetClaudePath()
	if err != nil {
		return "", err
	}

	return filepath.Dir(claudePath), nil
}

// FindClaudePath finds the path to the actual claude executable.
// It checks PATH first, then common installation locations.
// This is useful when Claude is installed but not in PATH.
func FindClaudePath() (string, error) {
	// Check if claude is in PATH
	path, err := exec.LookPath("claude")
	if err == nil {
		// Resolve symlinks to get the actual binary
		realPath, err := filepath.EvalSymlinks(path)
		if err == nil {
			return realPath, nil
		}
		return path, nil
	}

	// Check common installation locations
	commonPaths := []string{
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude",
		filepath.Join(os.Getenv("HOME"), ".local/bin/claude"),
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("claude executable not found in PATH or common locations")
}
