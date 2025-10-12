package tui

import (
	"fmt"
	"os"
	"os/exec"
)

// LaunchClaudeInteractive suspends the TUI and launches Claude CLI interactively.
// When Claude exits, control returns to the caller.
func LaunchClaudeInteractive(workingDir string) error {
	// Find Claude binary in PATH
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	// Create command with full terminal control
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

// IsClaudeAvailable checks if Claude CLI is available in PATH
func IsClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// GetClaudePath returns the full path to the Claude CLI binary
func GetClaudePath() (string, error) {
	return exec.LookPath("claude")
}
