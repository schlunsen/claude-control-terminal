// Package database provides git utility functions for extracting repository metadata.
// This file contains helpers for getting the current git branch from a working directory.
package database

import (
	"os/exec"
	"strings"
)

// GetCurrentGitBranch returns the current git branch for the given directory
// Returns empty string if not in a git repository or if git is not available
func GetCurrentGitBranch(workingDir string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	branch := strings.TrimSpace(string(output))
	return branch
}
