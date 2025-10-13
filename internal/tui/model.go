package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/schlunsen/claude-control-terminal/internal/server"
)

// Screen represents different views in the TUI
type Screen int

const (
	ScreenMain Screen = iota
	ScreenComponentList
	ScreenPreview
	ScreenConfirm
	ScreenConfirmRemove
	ScreenInstalling
	ScreenRemoving
	ScreenComplete
)

// Model represents the application state
type Model struct {
	// Current screen
	screen Screen

	// Component type selection
	componentTypes []string
	selectedType   int

	// Component list
	components      []ComponentItem
	cursor          int
	filteredIndices []int
	loading         bool
	loadError       error

	// Search
	searchInput  textinput.Model
	searchActive bool

	// Installation
	targetDir      string
	installing     bool
	installError   error
	installSuccess []string
	installFailed  []string

	// Preview
	previewContent  string
	previewLoading  bool
	previewError    error
	previewScroll   int
	previewComponent ComponentItem

	// UI state
	spinner            spinner.Model
	width              int
	height             int
	quitting           bool
	currentTheme       int  // 0=orange, 1=green, 2=cyan, 3=purple
	shouldLaunchClaude bool // Signal to launch Claude after TUI exits
	launchLastSession  bool // Signal to launch Claude with -c parameter

	// Analytics state
	analyticsEnabled bool            // Whether analytics server is running
	analyticsServer  *server.Server  // Reference to analytics server
	claudeDir        string          // Claude directory for analytics
}

// NewModel creates a new TUI model
func NewModel(targetDir string) Model {
	return NewModelWithServer(targetDir, "", nil)
}

// NewModelWithServer creates a new TUI model with analytics server reference
func NewModelWithServer(targetDir, claudeDir string, analyticsServer *server.Server) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.CharLimit = 50
	ti.Width = 40

	componentTypes := []string{"Agents", "Commands", "MCPs"}

	// Add "Launch Claude" options if Claude is available
	if IsClaudeAvailable() {
		componentTypes = append(componentTypes, "Launch last Claude session")
		componentTypes = append(componentTypes, "Launch Claude")
	}

	analyticsEnabled := analyticsServer != nil

	return Model{
		screen:           ScreenMain,
		componentTypes:   componentTypes,
		selectedType:     0,
		targetDir:        targetDir,
		spinner:          s,
		searchInput:      ti,
		width:            80,
		height:           24,
		currentTheme:     GetCurrentThemeIndex(),
		analyticsEnabled: analyticsEnabled,
		analyticsServer:  analyticsServer,
		claudeDir:        claudeDir,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case componentsLoadedMsg:
		m.loading = false
		m.components = msg.components
		m.loadError = msg.err
		m.updateFilteredIndices()
		return m, nil

	case installCompleteMsg:
		m.installing = false
		m.installSuccess = msg.success
		m.installFailed = msg.failed
		m.installError = msg.err
		m.screen = ScreenComplete
		return m, nil

	case removeCompleteMsg:
		m.installing = false
		m.installSuccess = msg.success
		m.installFailed = msg.failed
		m.installError = msg.err
		m.screen = ScreenComplete
		return m, nil

	case previewLoadedMsg:
		m.previewLoading = false
		m.previewContent = msg.content
		m.previewError = msg.err
		return m, nil

	case toggleAnalyticsMsg:
		// Handle immediate analytics server toggle
		if msg.enabled && m.analyticsServer == nil {
			// Start analytics server with quiet mode
			m.analyticsServer = server.NewServerWithOptions(msg.targetDir, 3333, true)
			if err := m.analyticsServer.Setup(); err == nil {
				go func() {
					if err := m.analyticsServer.Start(); err != nil {
						// Server failed to start
						m.analyticsServer = nil
					}
				}()
				m.analyticsEnabled = true
			} else {
				m.analyticsServer = nil
				m.analyticsEnabled = false
			}
		} else if !msg.enabled && m.analyticsServer != nil {
			// Stop analytics server immediately
			m.analyticsServer.Shutdown()
			m.analyticsServer = nil
			m.analyticsEnabled = false
		}
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if m.screen == ScreenMain || m.screen == ScreenComplete {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Screen-specific keys
	switch m.screen {
	case ScreenMain:
		return m.handleMainScreen(msg)
	case ScreenComponentList:
		return m.handleComponentListScreen(msg)
	case ScreenPreview:
		return m.handlePreviewScreen(msg)
	case ScreenConfirm:
		return m.handleConfirmScreen(msg)
	case ScreenConfirmRemove:
		return m.handleConfirmRemoveScreen(msg)
	case ScreenComplete:
		return m.handleCompleteScreen(msg)
	}

	return m, nil
}

// handleMainScreen handles input on the main screen
func (m Model) handleMainScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedType > 0 {
			m.selectedType--
		}
	case "down", "j":
		if m.selectedType < len(m.componentTypes)-1 {
			m.selectedType++
		}
	case "enter":
		// Check if "Launch last Claude session" was selected
		if m.componentTypes[m.selectedType] == "Launch last Claude session" {
			m.shouldLaunchClaude = true
			m.launchLastSession = true
			m.quitting = true
			return m, tea.Quit
		}
		// Check if "Launch Claude" was selected
		if m.componentTypes[m.selectedType] == "Launch Claude" {
			m.shouldLaunchClaude = true
			m.launchLastSession = false
			m.quitting = true
			return m, tea.Quit
		}

		// Load components for selected type
		m.screen = ScreenComponentList
		m.loading = true
		m.cursor = 0
		m.components = nil
		return m, loadComponentsCmd(m.getComponentType(), m.targetDir)
	case "esc":
		m.quitting = true
		return m, tea.Quit
	case "t", "T":
		// Cycle through themes
		m.currentTheme = (m.currentTheme + 1) % 4
		ApplyThemeByIndex(m.currentTheme)
		return m, nil
	case "a", "A":
		// Toggle analytics on/off
		m.analyticsEnabled = !m.analyticsEnabled
		return m, toggleAnalyticsCmd(m.analyticsEnabled, m.claudeDir)
	}
	return m, nil
}

// handleComponentListScreen handles input on the component list screen
func (m Model) handleComponentListScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}

	if m.searchActive {
		switch msg.String() {
		case "esc":
			m.searchActive = false
			m.searchInput.Blur()
			m.searchInput.SetValue("")
			m.updateFilteredIndices()
			return m, nil
		case "enter":
			m.searchActive = false
			m.searchInput.Blur()
			return m, nil
		default:
			// Pass all other keys to the search input
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.updateFilteredIndices()
			return m, cmd
		}
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filteredIndices)-1 {
			m.cursor++
		}
	case "pgup":
		// Page up - jump 10 items up
		m.cursor -= 10
		if m.cursor < 0 {
			m.cursor = 0
		}
	case "pgdown":
		// Page down - jump 10 items down
		m.cursor += 10
		if m.cursor >= len(m.filteredIndices) {
			m.cursor = len(m.filteredIndices) - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
		}
	case "/":
		// Activate search
		m.searchActive = true
		m.searchInput.Focus()
		return m, textinput.Blink
	case "enter":
		// Proceed to confirmation with current component
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			idx := m.filteredIndices[m.cursor]
			// Mark only this component as selected
			for i := range m.components {
				m.components[i].Selected = false
			}
			m.components[idx].Selected = true
			m.screen = ScreenConfirm
			return m, nil
		}
	case "p":
		// Preview selected component
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			idx := m.filteredIndices[m.cursor]
			m.previewComponent = m.components[idx]
			m.screen = ScreenPreview
			m.previewLoading = true
			m.previewScroll = 0
			m.previewContent = ""
			m.previewError = nil
			return m, loadPreviewCmd(m.previewComponent)
		}
	case "d", "x":
		// Remove selected component (only if installed)
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			idx := m.filteredIndices[m.cursor]
			comp := m.components[idx]
			// Only allow removal if component is installed
			if comp.InstalledProject || comp.InstalledGlobal {
				// Mark for removal and go to confirm screen
				for i := range m.components {
					m.components[i].Selected = false
				}
				m.components[idx].Selected = true
				m.screen = ScreenConfirmRemove
				return m, nil
			}
		}
	case "r":
		// Refresh components from GitHub
		m.loading = true
		m.components = nil
		return m, loadComponentsCmd(m.getComponentType(), m.targetDir, true)
	case "esc":
		// If there's an active filter, clear it first
		if m.searchInput.Value() != "" {
			m.searchInput.SetValue("")
			m.updateFilteredIndices()
			return m, nil
		}
		// Otherwise go back to main screen
		m.screen = ScreenMain
		return m, nil
	}

	return m, nil
}

// handlePreviewScreen handles input on the preview screen
func (m Model) handlePreviewScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.previewLoading {
		if msg.String() == "esc" {
			m.screen = ScreenComponentList
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.previewScroll > 0 {
			m.previewScroll--
		}
	case "down", "j":
		m.previewScroll++
	case "pgup":
		m.previewScroll -= 10
		if m.previewScroll < 0 {
			m.previewScroll = 0
		}
	case "pgdown":
		m.previewScroll += 10
	case "g":
		// Go to top
		m.previewScroll = 0
	case "G":
		// Go to bottom
		m.previewScroll = 9999 // Will be clamped in view
	case "i":
		// Install this component
		m.previewComponent.Selected = true
		idx := m.filteredIndices[m.cursor]
		m.components[idx].Selected = true
		m.screen = ScreenConfirm
		return m, nil
	case "esc", "q":
		// Go back to component list
		m.screen = ScreenComponentList
		return m, nil
	}

	return m, nil
}

// handleConfirmScreen handles input on the confirmation screen
func (m Model) handleConfirmScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		// Start installation
		m.screen = ScreenInstalling
		m.installing = true
		return m, installComponentsCmd(m.getSelectedComponents(), m.targetDir)
	case "n", "esc":
		// Go back to component list
		m.screen = ScreenComponentList
		return m, nil
	}
	return m, nil
}

// handleConfirmRemoveScreen handles input on the removal confirmation screen
func (m Model) handleConfirmRemoveScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		// Start removal
		m.screen = ScreenRemoving
		m.installing = true
		return m, removeComponentsCmd(m.getSelectedComponents(), m.targetDir)
	case "n", "esc":
		// Go back to component list
		m.screen = ScreenComponentList
		return m, nil
	}
	return m, nil
}

// handleCompleteScreen handles input on the complete screen
func (m Model) handleCompleteScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "c":
		// Launch Claude
		if !IsClaudeAvailable() {
			return m, nil
		}
		m.shouldLaunchClaude = true
		m.quitting = true
		return m, tea.Quit
	case "r":
		// Return to main screen
		m.screen = ScreenMain
		m.components = nil
		m.installSuccess = nil
		m.installFailed = nil
		m.installError = nil
		return m, nil
	case "esc", "enter":
		// Return to component list
		m.screen = ScreenComponentList
		m.installSuccess = nil
		m.installFailed = nil
		m.installError = nil
		return m, nil
	}
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case ScreenMain:
		return m.viewMainScreen()
	case ScreenComponentList:
		return m.viewComponentListScreen()
	case ScreenPreview:
		return m.viewPreviewScreen()
	case ScreenConfirm:
		return m.viewConfirmScreen()
	case ScreenConfirmRemove:
		return m.viewConfirmRemoveScreen()
	case ScreenInstalling:
		return m.viewInstallingScreen()
	case ScreenRemoving:
		return m.viewRemovingScreen()
	case ScreenComplete:
		return m.viewCompleteScreen()
	}

	return ""
}

// viewMainScreen renders the main screen
func (m Model) viewMainScreen() string {
	var b strings.Builder

	b.WriteString(GetBannerStyled())
	b.WriteString("\n")

	b.WriteString(TitleStyle.Render("Select Component Type") + "\n\n")

	for i, componentType := range m.componentTypes {
		icon := m.getIconForType(componentType)
		cursor := "  "

		// Special styling for "Launch Claude" options
		isLaunchClaude := componentType == "Launch Claude" || componentType == "Launch last Claude session"

		if i == m.selectedType {
			cursor = SelectedItemStyle.Render("> ")
			if isLaunchClaude {
				// Make Launch Claude options stand out more when selected
				b.WriteString(cursor + StatusSuccessStyle.Render(icon+" "+componentType) + "\n")
			} else {
				b.WriteString(cursor + SelectedItemStyle.Render(icon+" "+componentType) + "\n")
			}
		} else {
			if isLaunchClaude {
				// Make Launch Claude options stand out even when not selected
				b.WriteString(cursor + StatusInfoStyle.Render(icon+" "+componentType) + "\n")
			} else {
				b.WriteString(cursor + UnselectedItemStyle.Render(icon+" "+componentType) + "\n")
			}
		}
	}

	b.WriteString("\n")

	// Show current theme and analytics status
	themeName := GetThemeName(m.currentTheme)
	themeInfo := SubtitleStyle.Render(fmt.Sprintf("Theme: %s", themeName))
	b.WriteString(themeInfo + "\n")

	// Analytics status
	analyticsStatus := "OFF"
	analyticsStyle := StatusErrorStyle
	if m.analyticsEnabled {
		analyticsStatus = "ON"
		analyticsStyle = StatusSuccessStyle
	}
	b.WriteString(SubtitleStyle.Render("Analytics: ") + analyticsStyle.Render(analyticsStatus))
	b.WriteString(SubtitleStyle.Render(" (http://localhost:3333)") + "\n\n")

	b.WriteString(HelpStyle.Render("↑/↓: Navigate • Enter: Select • T: Theme • A: Toggle Analytics • Q/Esc: Quit"))

	return BoxStyle.Render(b.String())
}

// viewComponentListScreen renders the component list screen
func (m Model) viewComponentListScreen() string {
	var b strings.Builder

	_ = m.getComponentType() // typeStr declared but not used, keeping call for side effects
	icon := m.getIconForType(m.componentTypes[m.selectedType])

	b.WriteString(TitleStyle.Render(fmt.Sprintf("%s Browse %s", icon, m.componentTypes[m.selectedType])) + "\n\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Loading components from GitHub...\n")
		return BoxStyle.Render(b.String())
	}

	if m.loadError != nil {
		b.WriteString(StatusErrorStyle.Render("Error loading components: ") + m.loadError.Error() + "\n\n")
		b.WriteString(HelpStyle.Render("Esc: Go Back"))
		return BoxStyle.Render(b.String())
	}

	// Search bar
	if m.searchActive {
		b.WriteString(InputFocusedStyle.Render("Search: "+m.searchInput.View()) + "\n\n")
	} else if m.searchInput.Value() != "" {
		// Show active filter even when search is not focused
		b.WriteString(InputStyle.Render("Filter: "+m.searchInput.Value()) + " " +
			HelpStyle.Render("(/ to edit, Esc to clear)") + "\n\n")
	} else {
		searchHint := InputStyle.Render("Press / to search")
		b.WriteString(searchHint + "\n\n")
	}

	// Component list
	if len(m.filteredIndices) == 0 {
		b.WriteString(StatusInfoStyle.Render("No components found") + "\n")
	} else {
		// Calculate how many items we can show based on terminal height
		// Reserve space for: title (2 lines), search (2 lines), status (2 lines), help (3 lines), padding (3 lines)
		reservedLines := 12
		availableLines := m.height - reservedLines
		if availableLines < 5 {
			availableLines = 5 // Minimum 5 items
		}
		if availableLines > 20 {
			availableLines = 20 // Maximum 20 items for performance
		}

		// Center the cursor in the viewport
		halfView := availableLines / 2
		start := m.cursor - halfView
		if start < 0 {
			start = 0
		}
		end := start + availableLines
		if end > len(m.filteredIndices) {
			end = len(m.filteredIndices)
			// Adjust start to show full viewport if possible
			start = end - availableLines
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			idx := m.filteredIndices[i]
			component := m.components[idx]

			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			// Determine style based on installation status
			// Green for project-installed, Yellow for global-installed
			var nameStyle lipgloss.Style
			if component.InstalledProject {
				nameStyle = StatusSuccessStyle // Green
			} else if component.InstalledGlobal {
				nameStyle = StatusWarningStyle // Yellow
			} else {
				nameStyle = lipgloss.NewStyle() // Default (no special color)
			}

			// Build the line with styled name
			line := cursor + nameStyle.Render(component.Name)
			if component.Category != "root" && component.Category != "" {
				line += CategoryStyle.Render(" ("+component.Category+")")
			}

			// Add installation indicators
			var indicators string
			if component.InstalledGlobal {
				indicators += InstalledIndicatorStyle.Render(" [G]")
			}
			if component.InstalledProject {
				indicators += InstalledIndicatorStyle.Render(" [P]")
			}
			line += indicators

			if i == m.cursor {
				b.WriteString(SelectedItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		}
	}

	b.WriteString("\n")

	// Help - keep compact for small terminals
	if m.height < 20 {
		// Compact help for small terminals
		b.WriteString(HelpStyle.Render("↑/↓: Navigate • P: Preview • Enter: Install • D: Remove • Esc: Back\n"))
		b.WriteString(StatusSuccessStyle.Render("[P]=Project  "))
		b.WriteString(StatusWarningStyle.Render("[G]=Global"))
	} else {
		// Full help for larger terminals
		b.WriteString(HelpStyle.Render("↑/↓: Navigate • PgUp/PgDn: Page • /: Search • P: Preview • R: Refresh\n"))
		b.WriteString(HelpStyle.Render("Enter: Install • D: Remove • Esc: Back • "))
		b.WriteString(StatusSuccessStyle.Render("[P]=Project  "))
		b.WriteString(StatusWarningStyle.Render("[G]=Global"))
	}

	return BoxStyle.Width(m.width - 4).Render(b.String())
}

// viewPreviewScreen renders the preview screen
func (m Model) viewPreviewScreen() string {
	var b strings.Builder

	icon := m.getIconForType(m.previewComponent.Type + "s")
	b.WriteString(TitleStyle.Render(fmt.Sprintf("%s Preview: %s", icon, m.previewComponent.Name)) + "\n")

	if m.previewComponent.Category != "root" && m.previewComponent.Category != "" {
		b.WriteString(CategoryStyle.Render("Category: "+m.previewComponent.Category) + "\n")
	}
	b.WriteString("\n")

	if m.previewLoading {
		b.WriteString(m.spinner.View() + " Loading preview...\n")
		return BoxStyle.Render(b.String())
	}

	if m.previewError != nil {
		b.WriteString(StatusErrorStyle.Render("Error loading preview: ") + m.previewError.Error() + "\n\n")
		b.WriteString(HelpStyle.Render("Esc: Go Back"))
		return BoxStyle.Render(b.String())
	}

	// Display content with scrolling
	lines := strings.Split(m.previewContent, "\n")

	// Calculate viewport
	reservedLines := 10 // title, category, help, padding
	availableLines := m.height - reservedLines
	if availableLines < 10 {
		availableLines = 10
	}

	// Clamp scroll
	maxScroll := len(lines) - availableLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.previewScroll > maxScroll {
		m.previewScroll = maxScroll
	}

	// Display visible lines
	start := m.previewScroll
	end := start + availableLines
	if end > len(lines) {
		end = len(lines)
	}

	for i := start; i < end; i++ {
		b.WriteString(lines[i] + "\n")
	}

	// Scroll indicator
	if len(lines) > availableLines {
		scrollInfo := fmt.Sprintf("\n[Showing lines %d-%d of %d]", start+1, end, len(lines))
		b.WriteString(SubtitleStyle.Render(scrollInfo) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: Scroll • PgUp/PgDn: Page • g: Top • G: Bottom • I: Install • Esc/Q: Back"))

	return BoxStyle.Width(m.width - 4).Render(b.String())
}

// viewConfirmScreen renders the confirmation screen
func (m Model) viewConfirmScreen() string {
	var b strings.Builder

	selected := m.getSelectedComponents()
	if len(selected) == 0 {
		b.WriteString(StatusErrorStyle.Render("No component selected") + "\n")
		return BoxStyle.Render(b.String())
	}

	comp := selected[0] // Single component
	icon := m.getIconForType(comp.Type + "s")

	b.WriteString(TitleStyle.Render("Confirm Installation") + "\n\n")

	b.WriteString(fmt.Sprintf("Are you sure you want to install:\n\n"))
	b.WriteString(fmt.Sprintf("  %s %s", icon, StatusInfoStyle.Render(comp.Name)))
	if comp.Category != "root" && comp.Category != "" {
		b.WriteString(CategoryStyle.Render(" ("+comp.Category+")"))
	}
	b.WriteString("\n")

	// Show current installation status
	if comp.InstalledGlobal || comp.InstalledProject {
		b.WriteString("\n")
		b.WriteString(StatusWarningStyle.Render("Already installed:"))
		if comp.InstalledGlobal {
			b.WriteString(" [Global]")
		}
		if comp.InstalledProject {
			b.WriteString(" [Project]")
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Target: %s\n\n", SubtitleStyle.Render(m.targetDir)))

	b.WriteString(HelpStyle.Render("Y/Enter: Install • N/Esc: Cancel"))

	return BoxStyle.Render(b.String())
}

// viewInstallingScreen renders the installing screen
func (m Model) viewInstallingScreen() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Installing Components") + "\n\n")
	b.WriteString(m.spinner.View() + " Installing components, please wait...\n")

	return BoxStyle.Render(b.String())
}

// viewConfirmRemoveScreen renders the removal confirmation screen
func (m Model) viewConfirmRemoveScreen() string {
	var b strings.Builder

	selected := m.getSelectedComponents()
	if len(selected) == 0 {
		b.WriteString(StatusErrorStyle.Render("No component selected") + "\n")
		return BoxStyle.Render(b.String())
	}

	comp := selected[0] // Single component
	icon := m.getIconForType(comp.Type + "s")

	b.WriteString(TitleStyle.Render("Confirm Removal") + "\n\n")

	b.WriteString(fmt.Sprintf("Are you sure you want to remove:\n\n"))
	b.WriteString(fmt.Sprintf("  %s %s", icon, StatusWarningStyle.Render(comp.Name)))
	if comp.Category != "root" && comp.Category != "" {
		b.WriteString(CategoryStyle.Render(" ("+comp.Category+")"))
	}
	b.WriteString("\n")

	// Show installation locations
	b.WriteString("\n")
	b.WriteString(StatusInfoStyle.Render("Installed in:"))
	if comp.InstalledGlobal {
		b.WriteString(" [Global]")
	}
	if comp.InstalledProject {
		b.WriteString(" [Project]")
	}
	b.WriteString("\n\n")

	b.WriteString(StatusWarningStyle.Render("⚠️  This will remove the component from your system.") + "\n\n")

	b.WriteString(HelpStyle.Render("Y/Enter: Remove • N/Esc: Cancel"))

	return BoxStyle.Render(b.String())
}

// viewRemovingScreen renders the removing screen
func (m Model) viewRemovingScreen() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Removing Components") + "\n\n")
	b.WriteString(m.spinner.View() + " Removing components, please wait...\n")

	return BoxStyle.Render(b.String())
}

// viewCompleteScreen renders the completion screen
func (m Model) viewCompleteScreen() string {
	var b strings.Builder

	if m.installError != nil {
		b.WriteString(StatusErrorStyle.Render("Installation Error") + "\n\n")
		b.WriteString(m.installError.Error() + "\n")
	} else if len(m.installFailed) > 0 {
		b.WriteString(StatusSuccessStyle.Render("Partial Installation Complete") + "\n\n")
	} else {
		b.WriteString(StatusSuccessStyle.Render("Installation Complete!") + "\n\n")
	}

	if len(m.installSuccess) > 0 {
		b.WriteString(StatusSuccessStyle.Render(fmt.Sprintf("Successfully installed %d component(s):", len(m.installSuccess))) + "\n")
		for _, name := range m.installSuccess {
			b.WriteString(fmt.Sprintf("  ✓ %s\n", name))
		}
		b.WriteString("\n")
	}

	if len(m.installFailed) > 0 {
		b.WriteString(StatusErrorStyle.Render(fmt.Sprintf("Failed to install %d component(s):", len(m.installFailed))) + "\n")
		for _, name := range m.installFailed {
			b.WriteString(fmt.Sprintf("  ✗ %s\n", name))
		}
		b.WriteString("\n")
	}

	// Add Launch Claude option if available
	if IsClaudeAvailable() {
		b.WriteString(HelpStyle.Render("Enter/Esc: Back to List • C: Launch Claude • R: Main Menu • Q: Quit"))
	} else {
		b.WriteString(HelpStyle.Render("Enter/Esc: Back to List • R: Main Menu • Q: Quit"))
	}

	return BoxStyle.Render(b.String())
}

// Helper methods

func (m Model) getComponentType() string {
	types := map[string]string{
		"Agents":   "agent",
		"Commands": "command",
		"MCPs":     "mcp",
	}
	return types[m.componentTypes[m.selectedType]]
}

func (m Model) getIconForType(typeName string) string {
	icons := map[string]string{
		"Agents":                     "🤖",
		"Commands":                   "⚡",
		"MCPs":                       "🔌",
		"Launch last Claude session": "🔄",
		"Launch Claude":              "🚀",
	}
	if icon, ok := icons[typeName]; ok {
		return icon
	}
	return "📦"
}

func (m Model) getSelectedCount() int {
	count := 0
	for _, comp := range m.components {
		if comp.Selected {
			count++
		}
	}
	return count
}

func (m Model) getSelectedComponents() []ComponentItem {
	var selected []ComponentItem
	for _, comp := range m.components {
		if comp.Selected {
			selected = append(selected, comp)
		}
	}
	return selected
}

func (m *Model) updateFilteredIndices() {
	searchTerm := strings.ToLower(m.searchInput.Value())
	m.filteredIndices = nil

	for i, comp := range m.components {
		if searchTerm == "" ||
			strings.Contains(strings.ToLower(comp.Name), searchTerm) ||
			strings.Contains(strings.ToLower(comp.Category), searchTerm) {
			m.filteredIndices = append(m.filteredIndices, i)
		}
	}

	// Reset cursor if needed
	if m.cursor >= len(m.filteredIndices) {
		m.cursor = len(m.filteredIndices) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
}

// Messages

type componentsLoadedMsg struct {
	components []ComponentItem
	err        error
}

type installCompleteMsg struct {
	success []string
	failed  []string
	err     error
}

type removeCompleteMsg struct {
	success []string
	failed  []string
	err     error
}

type previewLoadedMsg struct {
	content string
	err     error
}

// Commands

func loadComponentsCmd(componentType, targetDir string, forceRefresh ...bool) tea.Cmd {
	return func() tea.Msg {
		loader := NewComponentLoader()

		refresh := false
		if len(forceRefresh) > 0 {
			refresh = forceRefresh[0]
		}

		components, err := loader.LoadComponentsWithCache(componentType, targetDir, refresh)
		return componentsLoadedMsg{
			components: components,
			err:        err,
		}
	}
}

func installComponentsCmd(components []ComponentItem, targetDir string) tea.Cmd {
	return func() tea.Msg {
		var success []string
		var failed []string

		for _, comp := range components {
			var err error
			switch comp.Type {
			case "agent":
				installer := NewAgentInstallerForTUI()
				err = installer.InstallAgent(comp.Name, comp.Category, targetDir)
			case "command":
				installer := NewCommandInstallerForTUI()
				err = installer.InstallCommand(comp.Name, comp.Category, targetDir)
			case "mcp":
				installer := NewMCPInstallerForTUI()
				err = installer.InstallMCP(comp.Name, comp.Category, targetDir)
			}

			if err != nil {
				failed = append(failed, comp.Name)
			} else {
				success = append(success, comp.Name)
			}
		}

		return installCompleteMsg{
			success: success,
			failed:  failed,
			err:     nil,
		}
	}
}

func removeComponentsCmd(components []ComponentItem, targetDir string) tea.Cmd {
	return func() tea.Msg {
		var success []string
		var failed []string

		for _, comp := range components {
			var err error
			switch comp.Type {
			case "agent":
				installer := NewAgentInstallerForTUI()
				err = installer.RemoveAgent(comp.Name, targetDir)
			case "command":
				installer := NewCommandInstallerForTUI()
				err = installer.RemoveCommand(comp.Name, targetDir)
			case "mcp":
				installer := NewMCPInstallerForTUI()
				err = installer.RemoveMCP(comp.Name, targetDir)
			}

			if err != nil {
				failed = append(failed, comp.Name)
			} else {
				success = append(success, comp.Name)
			}
		}

		return removeCompleteMsg{
			success: success,
			failed:  failed,
			err:     nil,
		}
	}
}

func loadPreviewCmd(component ComponentItem) tea.Cmd {
	return func() tea.Msg {
		var content string
		var err error

		switch component.Type {
		case "agent":
			installer := NewAgentInstallerForTUI()
			content, err = installer.PreviewAgent(component.Name, component.Category)
		case "command":
			installer := NewCommandInstallerForTUI()
			content, err = installer.PreviewCommand(component.Name, component.Category)
		case "mcp":
			installer := NewMCPInstallerForTUI()
			content, err = installer.PreviewMCP(component.Name, component.Category)
		}

		return previewLoadedMsg{
			content: content,
			err:     err,
		}
	}
}

type toggleAnalyticsMsg struct {
	enabled   bool
	targetDir string
}

func toggleAnalyticsCmd(enabled bool, targetDir string) tea.Cmd {
	return func() tea.Msg {
		return toggleAnalyticsMsg{
			enabled:   enabled,
			targetDir: targetDir,
		}
	}
}
