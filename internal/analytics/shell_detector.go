package analytics

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// BackgroundShell represents a background bash shell process
type BackgroundShell struct {
	ShellID    string    `json:"shell_id"`
	PID        string    `json:"pid"`
	Command    string    `json:"command"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"start_time"`
	WorkingDir string    `json:"working_dir"`
}

// ShellDetector handles background bash shell detection
type ShellDetector struct {
	cache struct {
		data      []BackgroundShell
		timestamp time.Time
		ttl       time.Duration
	}
}

// NewShellDetector creates a new ShellDetector
func NewShellDetector() *ShellDetector {
	sd := &ShellDetector{}
	sd.cache.ttl = 500 * time.Millisecond
	return sd
}

// DetectBackgroundShells detects running background bash shells from Claude Code
func (sd *ShellDetector) DetectBackgroundShells() ([]BackgroundShell, error) {
	// Check cache first
	now := time.Now()
	if sd.cache.data != nil && now.Sub(sd.cache.timestamp) < sd.cache.ttl {
		return sd.cache.data, nil
	}

	shells := []BackgroundShell{}

	// Try to detect background shells by looking for bash processes
	// that might be spawned by Claude Code
	cmd := exec.Command("sh", "-c", "ps aux | grep -E 'bash|sh' | grep -v grep | grep -v '/bin/bash' | grep -v '/bin/sh'")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		// No processes found is not an error
		return shells, nil
	}

	lines := strings.Split(out.String(), "\n")
	shellIDPattern := regexp.MustCompile(`shell-(\d+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		fullCommand := strings.Join(fields[10:], " ")

		// Look for shell IDs in command or related processes
		matches := shellIDPattern.FindStringSubmatch(fullCommand)
		shellID := ""
		if len(matches) > 1 {
			shellID = matches[1]
		}

		// Check if this looks like a background shell
		// (running commands, not interactive shells)
		isBackgroundShell := strings.Contains(fullCommand, "bash -c") ||
			strings.Contains(fullCommand, "sh -c") ||
			(shellID != "" && !strings.Contains(fullCommand, "grep"))

		if !isBackgroundShell {
			continue
		}

		// Extract PID
		pid := fields[1]

		// Determine status (running, idle, etc.)
		status := "running"
		cpuUsage := fields[2]
		if cpuUsage == "0.0" {
			status = "idle"
		}

		shell := BackgroundShell{
			ShellID:    shellID,
			PID:        pid,
			Command:    fullCommand,
			Status:     status,
			StartTime:  time.Now(), // Approximation
			WorkingDir: "unknown",
		}

		shells = append(shells, shell)
	}

	// Cache the result
	sd.cache.data = shells
	sd.cache.timestamp = now

	return shells, nil
}

// GetCachedShells returns cached shell data
func (sd *ShellDetector) GetCachedShells() []BackgroundShell {
	now := time.Now()
	if sd.cache.data != nil && now.Sub(sd.cache.timestamp) < sd.cache.ttl {
		return sd.cache.data
	}
	return []BackgroundShell{}
}

// ClearCache clears the shell cache
func (sd *ShellDetector) ClearCache() {
	sd.cache.data = nil
	sd.cache.timestamp = time.Time{}
}

// ShellStats holds shell statistics
type ShellStats struct {
	Total       int               `json:"total"`
	Running     int               `json:"running"`
	Idle        int               `json:"idle"`
	Shells      []BackgroundShell `json:"shells"`
}

// GetShellStats returns statistics about detected shells
func (sd *ShellDetector) GetShellStats() (*ShellStats, error) {
	shells, err := sd.DetectBackgroundShells()
	if err != nil {
		return nil, err
	}

	runningCount := 0
	idleCount := 0

	for _, shell := range shells {
		switch shell.Status {
		case "running":
			runningCount++
		case "idle":
			idleCount++
		}
	}

	return &ShellStats{
		Total:   len(shells),
		Running: runningCount,
		Idle:    idleCount,
		Shells:  shells,
	}, nil
}
