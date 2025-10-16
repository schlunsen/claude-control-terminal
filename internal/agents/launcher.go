package agents

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Launcher manages the agent server process
type Launcher struct {
	Config      *Config
	ServerDir   string
	Quiet       bool
	logFile     *os.File // Keep reference for cleanup
	cmd         *exec.Cmd // Keep reference to the process
	cancelFunc  context.CancelFunc // Context for cancellation
}

// NewLauncher creates a new launcher instance
func NewLauncher(config *Config, quiet bool) *Launcher {
	return &Launcher{
		Config:    config,
		ServerDir: config.ServerDir,
		Quiet:     quiet,
	}
}

// log prints a message unless in quiet mode
func (l *Launcher) log(format string, args ...interface{}) {
	if !l.Quiet {
		fmt.Printf(format+"\n", args...)
	}
}

// heartbeatUpdater updates the heartbeat file periodically
func (l *Launcher) heartbeatUpdater(heartbeatFile string) {
	ticker := time.NewTicker(500 * time.Millisecond) // Update every 500ms
	defer ticker.Stop()

	for range ticker.C {
		// Only update if the Go process is still running
		if isProcessRunning(os.Getpid()) {
			os.WriteFile(heartbeatFile, []byte("alive"), 0644)
		} else {
			// Parent process died, stop updating
			return
		}
	}
}

// venvPythonPath returns the path to the Python executable in the venv
func (l *Launcher) venvPythonPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(l.ServerDir, ".venv", "Scripts", "python.exe")
	}
	return filepath.Join(l.ServerDir, ".venv", "bin", "python")
}

// IsRunning checks if the agent server is currently running
func (l *Launcher) IsRunning() (bool, int, error) {
	// Read PID file
	pid, err := l.readPIDFile()
	if err != nil {
		return false, 0, nil // No PID file means not running
	}

	// Check if process is running
	if !isProcessRunning(pid) {
		// Stale PID file, remove it
		os.Remove(l.Config.PIDFile)
		return false, 0, nil
	}

	return true, pid, nil
}

// Start starts the agent server
func (l *Launcher) Start() error {
	// Check if already running
	running, pid, _ := l.IsRunning()
	if running {
		l.log("✓ Agent server already running (PID: %d)", pid)
		return nil
	}

	// Ensure server is installed
	installer := NewInstaller(l.ServerDir, l.Quiet)
	if err := installer.Install(); err != nil {
		return fmt.Errorf("failed to install agent server: %w", err)
	}

	// Check for updates
	if err := installer.Update(); err != nil {
		return fmt.Errorf("failed to update agent server: %w", err)
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	l.cancelFunc = cancel

	// Start the server
	l.log("→ Starting agent server on http://%s:%d", l.Config.Host, l.Config.Port)

	pythonPath := l.venvPythonPath()
	mainPath := filepath.Join(l.ServerDir, "main.py")

	cmd := exec.CommandContext(ctx, pythonPath, mainPath)
	cmd.Dir = l.ServerDir

	// Set up process group for proper cleanup
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true, // Create new process group
			Pgid:    0,   // Make this process the group leader
		}
	}

	// Set up environment variables
	heartbeatFile := filepath.Join(l.ServerDir, ".heartbeat")
	cmd.Env = append(l.Config.ToEnvVars(), "HEARTBEAT_FILE="+heartbeatFile)

	// Set up log file
	logFile, err := os.OpenFile(l.Config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close() // Always close the file handle

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Store reference to the command
	l.cmd = cmd

	// Start the process in the background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent server: %w", err)
	}

	// Write PID file
	if err := l.writePIDFile(cmd.Process.Pid); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Create initial heartbeat file
	if err := os.WriteFile(heartbeatFile, []byte("alive"), 0644); err != nil {
		l.log("Warning: failed to create heartbeat file: %v", err)
	}

	// Start heartbeat updater in background
	go l.heartbeatUpdater(heartbeatFile)

	// Wait a moment and check if it's still running
	time.Sleep(1 * time.Second)
	if !isProcessRunning(cmd.Process.Pid) {
		return fmt.Errorf("agent server failed to start. Check logs: %s", l.Config.LogFile)
	}

	l.log("✓ Agent server started (PID: %d)", cmd.Process.Pid)
	l.log("  Logs: %s", l.Config.LogFile)
	l.log("  Endpoint: http://%s:%d", l.Config.Host, l.Config.Port)

	return nil
}

// Cleanup gracefully shuts down the agent server
func (l *Launcher) Cleanup() {
	if l.cancelFunc != nil {
		l.cancelFunc()
	}

	// Wait a moment for graceful shutdown
	time.Sleep(500 * time.Millisecond)

	// If we have a stored command reference and it's still running, kill the process group
	if l.cmd != nil && l.cmd.Process != nil {
		if isProcessRunning(l.cmd.Process.Pid) {
			if runtime.GOOS != "windows" {
				// Kill the entire process group (-PID)
				syscall.Kill(-l.cmd.Process.Pid, syscall.SIGTERM)
				time.Sleep(1 * time.Second)

				// If still running, force kill
				if isProcessRunning(l.cmd.Process.Pid) {
					syscall.Kill(-l.cmd.Process.Pid, syscall.SIGKILL)
				}
			} else {
				// Windows fallback
				l.cmd.Process.Kill()
			}
		}
	}

	// Clean up PID file
	os.Remove(l.Config.PIDFile)
}

// Stop stops the agent server
func (l *Launcher) Stop() error {
	running, pid, _ := l.IsRunning()
	if !running {
		l.log("✓ Agent server is not running")
		return nil
	}

	l.log("→ Stopping agent server (PID: %d)...", pid)

	// Send SIGTERM to gracefully shut down
	if err := killProcess(pid); err != nil {
		return fmt.Errorf("failed to stop agent server: %w", err)
	}

	// Wait for process to stop (up to 10 seconds)
	for i := 0; i < 10; i++ {
		if !isProcessRunning(pid) {
			os.Remove(l.Config.PIDFile)
			l.log("✓ Agent server stopped")
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	// Force kill if it didn't stop
	l.log("⚠ Agent server did not stop gracefully, force killing...")
	if err := forceKillProcess(pid); err != nil {
		return fmt.Errorf("failed to force kill agent server: %w", err)
	}

	os.Remove(l.Config.PIDFile)
	l.log("✓ Agent server stopped (forced)")
	return nil
}

// Restart restarts the agent server
func (l *Launcher) Restart() error {
	l.log("→ Restarting agent server...")

	if err := l.Stop(); err != nil {
		return err
	}

	// Wait a moment before restarting
	time.Sleep(1 * time.Second)

	return l.Start()
}

// Status returns the status of the agent server
func (l *Launcher) Status() (string, error) {
	running, pid, _ := l.IsRunning()
	if !running {
		return "Agent server is not running", nil
	}

	return fmt.Sprintf("Agent server is running (PID: %d)\nEndpoint: http://%s:%d\nLogs: %s",
		pid, l.Config.Host, l.Config.Port, l.Config.LogFile), nil
}

// Logs tails the server logs
func (l *Launcher) Logs(lines int, follow bool) error {
	logFile := l.Config.LogFile

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return fmt.Errorf("log file not found: %s", logFile)
	}

	if follow {
		return l.tailLogs(logFile)
	}

	return l.showLastNLines(logFile, lines)
}

// showLastNLines shows the last N lines of the log file
func (l *Launcher) showLastNLines(logFile string, n int) error {
	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Print last N lines
	start := len(lines) - n
	if start < 0 {
		start = 0
	}

	for i := start; i < len(lines); i++ {
		fmt.Println(lines[i])
	}

	return nil
}

// tailLogs follows the log file and prints new lines
func (l *Launcher) tailLogs(logFile string) error {
	// Show last 10 lines first
	l.showLastNLines(logFile, 10)

	fmt.Println("--- Following logs (Ctrl+C to stop) ---")

	// Open file for reading
	file, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to end
	file.Seek(0, 2)

	scanner := bufio.NewScanner(file)
	for {
		if scanner.Scan() {
			fmt.Println(scanner.Text())
		} else {
			// No more lines, wait and retry
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// readPIDFile reads the PID from the PID file
func (l *Launcher) readPIDFile() (int, error) {
	data, err := os.ReadFile(l.Config.PIDFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in PID file: %w", err)
	}

	return pid, nil
}

// writePIDFile writes the PID to the PID file
func (l *Launcher) writePIDFile(pid int) error {
	return os.WriteFile(l.Config.PIDFile, []byte(strconv.Itoa(pid)), 0644)
}

// isProcessRunning checks if a process with the given PID is running
func isProcessRunning(pid int) bool {
	if runtime.GOOS == "windows" {
		// On Windows, use tasklist to check if process exists
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return false
		}
		return len(output) > 0 && !strings.Contains(string(output), "INFO: No tasks are running")
	}

	// On Unix, send signal 0 to check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// killProcess sends SIGTERM to a process
func killProcess(pid int) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid))
		return cmd.Run()
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(syscall.SIGTERM)
}

// forceKillProcess sends SIGKILL to a process
func forceKillProcess(pid int) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
		return cmd.Run()
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(syscall.SIGKILL)
}
