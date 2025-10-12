package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Launch starts the TUI application
func Launch(targetDir string) error {
	for {
		// Create the model
		m := NewModel(targetDir)

		// Create the Bubble Tea program
		p := tea.NewProgram(m, tea.WithAltScreen())

		// Run the program
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}

		// Check if we should launch Claude
		if model, ok := finalModel.(Model); ok {
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
