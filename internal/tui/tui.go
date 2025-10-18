// Package tui provides the terminal user interface for CCT using Bubble Tea.
// This file handles launching the TUI application and managing the analytics server lifecycle.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	agentspkg "github.com/schlunsen/claude-control-terminal/internal/agents"
	"github.com/schlunsen/claude-control-terminal/internal/server"
	tea "github.com/charmbracelet/bubbletea"
)

// Removed global analyticsServer - now managed in Model

// shutdownMsg is sent when we receive a shutdown signal
type shutdownMsg struct {
	signal os.Signal
}

// Launch starts the TUI application with optional analytics server
func Launch(targetDir string) error {
	// Get Claude directory
	claudeDir := filepath.Join(os.Getenv("HOME"), ".claude")
	if targetDir != "." && targetDir != "" {
		claudeDir = filepath.Join(targetDir, ".claude")
	}

	// Start analytics server in background (enabled by default)
	// Use quiet mode to suppress output when running in TUI
	var analyticsServer *server.Server
	analyticsServer = server.NewServerWithOptions(claudeDir, 3333, true)
	if err := analyticsServer.Setup(); err == nil {
		// Start server in background goroutine
		go func() {
			if err := analyticsServer.Start(); err != nil {
				// Server failed to start, but don't block TUI
				analyticsServer = nil
			}
		}()
	} else {
		analyticsServer = nil
	}

	// Start agent server in background (enabled by default)
	var agentServerPID int
	var agentServerEnabled bool
	agentConfig := agentspkg.DefaultConfig()
	agentLauncher := agentspkg.NewLauncher(agentConfig, true, true) // quiet mode, background mode

	// Check if already running
	running, pid, _ := agentLauncher.IsRunning()
	if running {
		agentServerEnabled = true
		agentServerPID = pid
	} else {
		// Try to start it (non-blocking in background mode)
		if err := agentLauncher.Start(); err == nil {
			// Wait briefly for server to start
			time.Sleep(500 * time.Millisecond)

			// Get the PID
			running, pid, _ := agentLauncher.IsRunning()
			if running {
				agentServerEnabled = true
				agentServerPID = pid
			}
		}
	}

	defer func() {
		// Cleanup analytics server on exit
		if analyticsServer != nil {
			analyticsServer.Shutdown()
		}
		// Clean up agent server if we started it in this session
		agentLauncher.Cleanup()
	}()

	for {
		// Create the model with analytics and agent server references
		m := NewModelWithServers(targetDir, claudeDir, analyticsServer, agentServerEnabled, agentServerPID)

		// Update analytics enabled state based on server status
		if m.analyticsServer == nil {
			m.analyticsEnabled = false
		}

		// Create the Bubble Tea program
		p := tea.NewProgram(m, tea.WithAltScreen())

		// Run the program
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}

		// Check if we should launch Claude
		if model, ok := finalModel.(Model); ok {
			// Sync analytics server reference from model
			analyticsServer = model.analyticsServer

			// Sync agent server state from model
			agentServerEnabled = model.agentServerEnabled
			agentServerPID = model.agentServerPID

			if model.shouldLaunchClaude {
				// Launch Claude CLI with appropriate flags
				var err error
				if model.launchLastSession {
					// Launch with -c flag to continue last session
					err = LaunchClaudeWithLastSession(targetDir)
				} else {
					// Launch normally
					err = LaunchClaudeInteractive(targetDir)
				}

				if err != nil {
					// Show error but continue back to TUI
					fmt.Printf("Error launching Claude: %v\n", err)
					fmt.Println("Press Enter to continue...")
					fmt.Scanln()
				}
				// Loop back to restart TUI
				continue
			}

			// Normal quit
			return nil
		}

		// If we can't cast, just quit
		return nil
	}
}
