package cmd

import (
	"github.com/pterm/pterm"
)

// ShowBanner displays the application banner with gradient colors
func ShowBanner() {
	// Clear screen
	pterm.Print("\033[H\033[2J")

	// Create gradient colors similar to Node.js version
	colors := []pterm.Color{
		pterm.FgLightRed,
		pterm.FgRed,
		pterm.FgYellow,
		pterm.FgLightYellow,
	}

	// Top border
	pterm.Println()
	topBorder := pterm.NewStyle(pterm.FgRed).Sprint("════════════════════════════════════════════════════════════════")
	pterm.Println(topBorder)
	pterm.Println()

	// Title with gradient effect
	title := "       🎮 Claude Control Terminal 🚀"
	coloredTitle := ""
	for i, char := range title {
		color := colors[i%len(colors)]
		coloredTitle += pterm.NewStyle(color).Sprint(string(char))
	}
	pterm.Println(coloredTitle)
	pterm.Println()

	// Subtitle
	subtitle := pterm.NewStyle(pterm.FgYellow).Sprint("       Your Command Center for Claude Code")
	pterm.Println(subtitle)
	pterm.Println()

	// Bottom border
	bottomBorder := pterm.NewStyle(pterm.FgRed).Sprint("════════════════════════════════════════════════════════════════")
	pterm.Println(bottomBorder)
	pterm.Println()

	// Info section
	setupMsg := pterm.NewStyle(pterm.FgLightYellow, pterm.Bold).Sprint("⚡ Supercharge your Claude Code workflow ⚡")
	pterm.Println(setupMsg)

	version := pterm.NewStyle(pterm.FgWhite).Sprintf("                             v%s (Go Edition)\n", Version)
	pterm.Println(version)

	// Links
	templatesLink := pterm.NewStyle(pterm.FgCyan).Sprint("🌐 Templates: ") +
		pterm.NewStyle(pterm.FgCyan, pterm.Underscore).Sprint("https://aitmpl.com")
	pterm.Println(templatesLink)

	docsLink := pterm.NewStyle(pterm.FgCyan).Sprint("📖 Documentation: ") +
		pterm.NewStyle(pterm.FgCyan, pterm.Underscore).Sprint("https://docs.aitmpl.com")
	pterm.Println(docsLink)

	pterm.Println()
}

// ShowSpinner creates and returns a spinner for long-running operations
func ShowSpinner(message string) *pterm.SpinnerPrinter {
	spinner, _ := pterm.DefaultSpinner.Start(message)
	return spinner
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	pterm.Success.Println(message)
}

// ShowError displays an error message
func ShowError(message string) {
	pterm.Error.Println(message)
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	pterm.Info.Println(message)
}

// ShowWarning displays a warning message
func ShowWarning(message string) {
	pterm.Warning.Println(message)
}

// ShowBox displays a message in a box
func ShowBox(title, content string) {
	pterm.DefaultBox.WithTitle(title).WithTitleTopCenter().Println(content)
}

// ShowProgress creates a progress bar
func ShowProgress(total int, message string) *pterm.ProgressbarPrinter {
	progressbar, _ := pterm.DefaultProgressbar.WithTotal(total).WithTitle(message).Start()
	return progressbar
}
