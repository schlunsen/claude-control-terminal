package wrapper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/schlunsen/claude-control-terminal/internal/database"
)

// ClaudeWrapper wraps the claude command to intercept user input
type ClaudeWrapper struct {
	claudePath string
	repo       *database.Repository
}

// NewClaudeWrapper creates a new wrapper
func NewClaudeWrapper(claudePath string, repo *database.Repository) *ClaudeWrapper {
	return &ClaudeWrapper{
		claudePath: claudePath,
		repo:       repo,
	}
}

// Execute runs claude with the given arguments and records user input
func (w *ClaudeWrapper) Execute(args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}

	// Get git branch
	gitBranch := w.getGitBranch(cwd)

	// Check if there's a message argument (for non-interactive mode)
	message := w.extractMessage(args)
	if message != "" {
		// Record the message
		if err := w.recordMessage(message, cwd, gitBranch); err != nil {
			fmt.Printf("Warning: failed to record message: %v\n", err)
		}
	}

	// Execute the actual claude command
	cmd := exec.Command(w.claudePath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// extractMessage extracts the message from command arguments
func (w *ClaudeWrapper) extractMessage(args []string) string {
	// Look for message after certain flags
	for i, arg := range args {
		if (arg == "-m" || arg == "--message") && i+1 < len(args) {
			return args[i+1]
		}
	}

	// Check if first non-flag argument is the message
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}

	return ""
}

// recordMessage records a user message to the database
func (w *ClaudeWrapper) recordMessage(message, cwd, gitBranch string) error {
	msg := &database.UserMessage{
		Message:          message,
		WorkingDirectory: cwd,
		GitBranch:        gitBranch,
		MessageLength:    len(message),
		SubmittedAt:      time.Now(),
	}

	return w.repo.RecordUserMessage(msg)
}

// getGitBranch gets the current git branch
func (w *ClaudeWrapper) getGitBranch(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// FindClaudePath finds the path to the actual claude executable
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
