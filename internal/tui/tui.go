package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Launch starts the TUI application
func Launch(targetDir string) error {
	// Create the model
	m := NewModel(targetDir)

	// Create the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
