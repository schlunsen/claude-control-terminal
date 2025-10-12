package analytics

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// Process represents a detected Claude process with metadata about its execution.
type Process struct {
	PID        string
	Command    string
	WorkingDir string
	StartTime  time.Time
	Status     string
	User       string
}

// ProcessCache holds cached process data with TTL for performance.
type ProcessCache struct {
	Data      []Process
	Timestamp time.Time
	TTL       time.Duration
}

// ProcessDetector handles Claude CLI process detection using system commands.
// It caches results to reduce system overhead.
type ProcessDetector struct {
	cache ProcessCache
}

// NewProcessDetector creates a new ProcessDetector with default cache TTL.
func NewProcessDetector() *ProcessDetector {
	return &ProcessDetector{
		cache: ProcessCache{
			Data:      nil,
			Timestamp: time.Time{},
			TTL:       500 * time.Millisecond,
		},
	}
}

// DetectRunningClaudeProcesses detects running Claude CLI processes using ps command.
// It returns a list of active processes matching the Claude CLI pattern.
func (pd *ProcessDetector) DetectRunningClaudeProcesses() ([]Process, error) {
	// Check cache first
	now := time.Now()
	if pd.cache.Data != nil && now.Sub(pd.cache.Timestamp) < pd.cache.TTL {
		return pd.cache.Data, nil
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use exec.CommandContext with separate arguments (no shell injection)
	// Build the command safely without shell
	cmd := exec.CommandContext(ctx, "ps", "aux")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		// ps command error, but no processes is not an error
		if ctx.Err() == context.DeadlineExceeded {
			return []Process{}, nil
		}
		return []Process{}, nil
	}

	// Filter output for Claude processes
	processes := pd.parseProcessOutput(out.String())

	// Cache the result
	pd.cache = ProcessCache{
		Data:      processes,
		Timestamp: now,
		TTL:       500 * time.Millisecond,
	}

	return processes, nil
}

// parseProcessOutput parses ps aux output and filters for Claude CLI processes.
// It avoids false positives from Chrome, analytics tools, and mobile apps.
func (pd *ProcessDetector) parseProcessOutput(output string) []Process {
	lines := strings.Split(output, "\n")
	processes := []Process{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse ps output
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		// Get full command (everything after the 10th field)
		fullCommand := strings.Join(fields[10:], " ")

		// Filter for Claude CLI processes
		isClaudeProcess := strings.Contains(fullCommand, "claude") &&
			!strings.Contains(fullCommand, "chrome_crashpad_handler") &&
			!strings.Contains(fullCommand, "create-claude-config") &&
			!strings.Contains(fullCommand, "chats-mobile") &&
			!strings.Contains(fullCommand, "analytics") &&
			!strings.Contains(fullCommand, "/Applications/Claude.app") &&
			!strings.Contains(fullCommand, "npm start") &&
			!strings.Contains(fullCommand, "grep") &&
			(strings.TrimSpace(fullCommand) == "claude" ||
				strings.Contains(fullCommand, "claude --") ||
				strings.Contains(fullCommand, "claude ") ||
				strings.Contains(fullCommand, "/claude") ||
				strings.Contains(fullCommand, "bin/claude"))

		if !isClaudeProcess {
			continue
		}

		// Extract working directory if available
		workingDir := "unknown"
		if strings.Contains(fullCommand, "--cwd") {
			parts := strings.Split(fullCommand, "--cwd")
			if len(parts) > 1 {
				cwdPart := strings.TrimSpace(parts[1])
				cwdFields := strings.Fields(cwdPart)
				if len(cwdFields) > 0 {
					workingDir = strings.Trim(cwdFields[0], "=")
				}
			}
		}

		process := Process{
			PID:        fields[1],
			Command:    fullCommand,
			WorkingDir: workingDir,
			StartTime:  time.Now(), // Approximation
			Status:     "running",
			User:       fields[0],
		}

		processes = append(processes, process)
	}

	return processes
}

// GetCachedProcesses returns cached process data if available and fresh.
func (pd *ProcessDetector) GetCachedProcesses() []Process {
	now := time.Now()
	if pd.cache.Data != nil && now.Sub(pd.cache.Timestamp) < pd.cache.TTL {
		return pd.cache.Data
	}
	return []Process{}
}

// ClearCache clears the process cache, forcing a fresh scan on next detection.
func (pd *ProcessDetector) ClearCache() {
	pd.cache = ProcessCache{
		Data:      nil,
		Timestamp: time.Time{},
		TTL:       500 * time.Millisecond,
	}
}

// HasActiveProcesses checks if there are any active Claude processes.
func (pd *ProcessDetector) HasActiveProcesses() (bool, error) {
	processes, err := pd.DetectRunningClaudeProcesses()
	if err != nil {
		return false, err
	}
	return len(processes) > 0, nil
}

// ProcessStats holds process statistics for reporting.
type ProcessStats struct {
	Total                  int
	WithKnownWorkingDir    int
	WithUnknownWorkingDir  int
	Processes              []Process
}

// GetProcessStats returns statistics about detected processes.
func (pd *ProcessDetector) GetProcessStats() (*ProcessStats, error) {
	processes, err := pd.DetectRunningClaudeProcesses()
	if err != nil {
		return nil, err
	}

	knownCount := 0
	unknownCount := 0

	for _, p := range processes {
		if p.WorkingDir != "unknown" {
			knownCount++
		} else {
			unknownCount++
		}
	}

	return &ProcessStats{
		Total:                 len(processes),
		WithKnownWorkingDir:   knownCount,
		WithUnknownWorkingDir: unknownCount,
		Processes:             processes,
	}, nil
}
