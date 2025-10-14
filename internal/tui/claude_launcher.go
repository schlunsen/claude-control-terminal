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
	// Find Claude binary (checks PATH and common locations like ~/.local/bin)
	claudePath, err := installer.FindClaudePath()
	if err != nil {
		return fmt.Errorf("claude CLI not found. Please install Claude CLI first.\nRun 'cct --install-claude' or use the installer from the main menu.\n\nError: %w", err)
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
	// Find Claude binary (checks PATH and common locations like ~/.local/bin)
	claudePath, err := installer.FindClaudePath()
	if err != nil {
		return fmt.Errorf("claude CLI not found. Please install Claude CLI first.\nRun 'cct --install-claude' or use the installer from the main menu.\n\nError: %w", err)
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

// IsClaudeAvailable checks if Claude CLI is available (in PATH or common locations)
func IsClaudeAvailable() bool {
	_, err := installer.FindClaudePath()
	return err == nil
}

// GetClaudePath returns the full path to the Claude CLI binary
func GetClaudePath() (string, error) {
	return installer.FindClaudePath()
}

// InstallClaude attempts to install Claude CLI with user interaction
func InstallClaude() error {
	ci := installer.NewClaudeInstaller()
	ci.Verbose = true

	fmt.Println("\n🚀 Claude CLI Installer")
	fmt.Println("========================")

	// Check if already installed (checks PATH and common locations)
	claudePath, err := installer.FindClaudePath()
	if err == nil {
		// Already installed
		fmt.Printf("✓ Claude CLI is already installed\n")
		fmt.Printf("  Location: %s\n", claudePath)

		// Try to get version
		version, vErr := ci.GetClaudeVersion()
		if vErr == nil {
			fmt.Printf("  Version: %s\n", version)
		}
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

	// Installation might have succeeded but not in PATH yet
	// Check common locations as a fallback
	claudePath, pathErr := installer.FindClaudePath()
	if pathErr == nil {
		fmt.Printf("\n✓ Claude CLI installed successfully!\n")
		fmt.Printf("  Location: %s\n", claudePath)
		fmt.Println("\nNote: Claude is installed but not in your PATH.")
		fmt.Println("The installer will still find it, but you may want to add ~/.local/bin to your PATH.")
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
