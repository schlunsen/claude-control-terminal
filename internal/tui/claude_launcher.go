package tui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/schlunsen/claude-control-terminal/internal/installer"
	"github.com/schlunsen/claude-control-terminal/internal/providers"
)

// LaunchClaudeInteractive suspends the TUI and launches Claude CLI interactively.
// When Claude exits, control returns to the caller.
func LaunchClaudeInteractive(workingDir string) error {
	// Check if Claude is installed, offer to install if not
	ci := installer.NewClaudeInstaller()
	if !ci.IsClaudeInstalled() {
		return fmt.Errorf("claude CLI not found in PATH. Please install Claude CLI first.\nRun 'cct --install-claude' or use the installer from the main menu")
	}

	// Find Claude binary in PATH
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	// Check if provider is configured
	providerScriptPath := providers.GetEnvScriptPath()
	if _, err := os.Stat(providerScriptPath); err == nil {
		// Provider script exists, source it before launching Claude
		// Use bash to source the script and then run Claude
		shellCommand := fmt.Sprintf("source %s && %s", providerScriptPath, claudePath)
		cmd := exec.Command("bash", "-c", shellCommand)
		cmd.Dir = workingDir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running Claude CLI: %w", err)
		}
		return nil
	}

	// No provider configured, launch Claude directly
	cmd := exec.Command(claudePath)
	cmd.Dir = workingDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run Claude and wait for it to exit
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running Claude CLI: %w", err)
	}

	return nil
}

// LaunchClaudeWithLastSession suspends the TUI and launches Claude CLI with the -c parameter
// to continue the last conversation. When Claude exits, control returns to the caller.
func LaunchClaudeWithLastSession(workingDir string) error {
	// Check if Claude is installed, offer to install if not
	ci := installer.NewClaudeInstaller()
	if !ci.IsClaudeInstalled() {
		return fmt.Errorf("claude CLI not found in PATH. Please install Claude CLI first.\nRun 'cct --install-claude' or use the installer from the main menu")
	}

	// Find Claude binary in PATH
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	// Check if provider is configured
	providerScriptPath := providers.GetEnvScriptPath()
	if _, err := os.Stat(providerScriptPath); err == nil {
		// Provider script exists, source it before launching Claude
		// Use bash to source the script and then run Claude with -c
		shellCommand := fmt.Sprintf("source %s && %s -c", providerScriptPath, claudePath)
		cmd := exec.Command("bash", "-c", shellCommand)
		cmd.Dir = workingDir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running Claude CLI with -c: %w", err)
		}
		return nil
	}

	// No provider configured, launch Claude directly
	cmd := exec.Command(claudePath, "-c")
	cmd.Dir = workingDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run Claude and wait for it to exit
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running Claude CLI with -c: %w", err)
	}

	return nil
}

// IsClaudeAvailable checks if Claude CLI is available in PATH
func IsClaudeAvailable() bool {
	ci := installer.NewClaudeInstaller()
	return ci.IsClaudeInstalled()
}

// GetClaudePath returns the full path to the Claude CLI binary
func GetClaudePath() (string, error) {
	ci := installer.NewClaudeInstaller()
	return ci.GetClaudePath()
}

// InstallClaude attempts to install Claude CLI with user interaction
func InstallClaude() error {
	ci := installer.NewClaudeInstaller()
	ci.Verbose = true

	fmt.Println("\n🚀 Claude CLI Installer")
	fmt.Println("========================")

	// Check if already installed
	if ci.IsClaudeInstalled() {
		version, _ := ci.GetClaudeVersion()
		path, _ := ci.GetClaudePath()
		fmt.Printf("✓ Claude CLI is already installed\n")
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Location: %s\n", path)
		return nil
	}

	// Show detection info
	nd := installer.NewNodeDetector()
	fmt.Println(nd.FormatNodeInfo())
	fmt.Println()

	// Recommend installation method
	fmt.Println("Recommended: Native binary installation (no Node.js required)")
	fmt.Println()
	fmt.Print("Proceed with automatic installation? (y/n): ")

	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("\nInstallation cancelled.")
		fmt.Println(installer.GetInstallInstructions())
		return fmt.Errorf("installation cancelled by user")
	}

	// Attempt auto-installation
	fmt.Println("\nInstalling Claude CLI...")
	result := ci.AutoInstall()

	if result.Success {
		fmt.Printf("\n✓ %s\n", result.Message)
		fmt.Printf("  Version: %s\n", result.Version)
		fmt.Printf("  Location: %s\n", result.ClaudePath)
		fmt.Println("\nRun 'claude doctor' to verify your installation.")
		return nil
	}

	// Installation failed
	fmt.Printf("\n✗ Installation failed: %v\n", result.Error)
	if result.Message != "" {
		fmt.Printf("  %s\n", result.Message)
	}
	fmt.Println("\nManual installation instructions:")
	fmt.Println(installer.GetInstallInstructions())

	return result.Error
}
