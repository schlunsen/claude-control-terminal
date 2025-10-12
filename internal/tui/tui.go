package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schlunsen/claude-control-terminal/internal/server"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	analyticsServer *server.Server
)

// Launch starts the TUI application with optional analytics server
func Launch(targetDir string) error {
	// Get Claude directory
	claudeDir := filepath.Join(os.Getenv("HOME"), ".claude")
	if targetDir != "." && targetDir != "" {
		claudeDir = filepath.Join(targetDir, ".claude")
	}

	// Start analytics server in background (enabled by default)
	// Use quiet mode to suppress output when running in TUI
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

	defer func() {
		// Cleanup analytics server on exit
		if analyticsServer != nil {
			analyticsServer.Shutdown()
		}
	}()

	for {
		// Create the model
		m := NewModel(targetDir)

		// Update analytics enabled state based on server status
		if analyticsServer == nil {
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
			// Handle analytics toggle
			if model.analyticsEnabled && analyticsServer == nil {
				// Start analytics server with quiet mode
				analyticsServer = server.NewServerWithOptions(claudeDir, 3333, true)
				if err := analyticsServer.Setup(); err == nil {
					go func() {
						if err := analyticsServer.Start(); err != nil {
							analyticsServer = nil
						}
					}()
				} else {
					analyticsServer = nil
				}
			} else if !model.analyticsEnabled && analyticsServer != nil {
				// Stop analytics server
				analyticsServer.Shutdown()
				analyticsServer = nil
			}

			if model.shouldLaunchClaude {
				// Launch Claude CLI
				if err := LaunchClaudeInteractive(targetDir); err != nil {
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
