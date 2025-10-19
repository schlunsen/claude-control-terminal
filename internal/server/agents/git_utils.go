package agents

import (
	"os/exec"
	"strings"
)

// GetGitBranch returns the current git branch for the given directory.
// Returns empty string if not a git repository or if there's an error.
func GetGitBranch(workingDir string) string {
	if workingDir == "" {
		return ""
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		// Not a git repository or error occurred
		return ""
	}

	branch := strings.TrimSpace(string(output))
	return branch
}

// IsGitRepository checks if the given directory is a git repository
func IsGitRepository(workingDir string) bool {
	if workingDir == "" {
		return false
	}

	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = workingDir

	return cmd.Run() == nil
}
