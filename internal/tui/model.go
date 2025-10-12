package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents different views in the TUI
type Screen int

const (
	ScreenMain Screen = iota
	ScreenComponentList
	ScreenConfirm
	ScreenInstalling
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

	// UI state
	spinner     spinner.Model
	width       int
	height      int
	quitting    bool
	currentTheme int // 0=orange, 1=green, 2=cyan, 3=purple
}

// NewModel creates a new TUI model
func NewModel(targetDir string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.CharLimit = 50
	ti.Width = 40

	return Model{
		screen:         ScreenMain,
		componentTypes: []string{"Agents", "Commands", "MCPs"},
		selectedType:   0,
		targetDir:      targetDir,
		spinner:        s,
		searchInput:    ti,
		width:          80,
		height:         24,
		currentTheme:   GetCurrentThemeIndex(),
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
	case ScreenConfirm:
		return m.handleConfirmScreen(msg)
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
		// Load components for selected type
		m.screen = ScreenComponentList
		m.loading = true
		m.cursor = 0
		m.components = nil
		return m, loadComponentsCmd(m.getComponentType())
	case "esc":
		m.quitting = true
		return m, tea.Quit
	case "t", "T":
		// Cycle through themes
		m.currentTheme = (m.currentTheme + 1) % 4
		ApplyThemeByIndex(m.currentTheme)
		return m, nil
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
	case " ": // Space key returns a literal space, not "space"
		// Toggle selection
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			idx := m.filteredIndices[m.cursor]
			m.components[idx].Selected = !m.components[idx].Selected
		}
	case "a":
		// Select all filtered
		for _, idx := range m.filteredIndices {
			m.components[idx].Selected = true
		}
	case "A":
		// Deselect all
		for i := range m.components {
			m.components[i].Selected = false
		}
	case "/":
		// Activate search
		m.searchActive = true
		m.searchInput.Focus()
		return m, textinput.Blink
	case "enter":
		// Proceed to confirmation
		if m.getSelectedCount() > 0 {
			m.screen = ScreenConfirm
			return m, nil
		}
	case "r":
		// Refresh components from GitHub
		m.loading = true
		m.components = nil
		return m, loadComponentsCmd(m.getComponentType(), true)
	case "esc":
		// Go back to main screen
		m.screen = ScreenMain
		m.searchInput.SetValue("")
		m.updateFilteredIndices()
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

// handleCompleteScreen handles input on the complete screen
func (m Model) handleCompleteScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "q", "esc":
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
	case ScreenConfirm:
		return m.viewConfirmScreen()
	case ScreenInstalling:
		return m.viewInstallingScreen()
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
		if i == m.selectedType {
			cursor = SelectedItemStyle.Render("> ")
			b.WriteString(cursor + SelectedItemStyle.Render(icon+" "+componentType) + "\n")
		} else {
			b.WriteString(cursor + UnselectedItemStyle.Render(icon+" "+componentType) + "\n")
		}
	}

	b.WriteString("\n")

	// Show current theme
	themeName := GetThemeName(m.currentTheme)
	themeInfo := SubtitleStyle.Render(fmt.Sprintf("Theme: %s", themeName))
	b.WriteString(themeInfo + "\n\n")

	b.WriteString(HelpStyle.Render("↑/↓: Navigate • Enter: Select • T: Theme • Q/Esc: Quit"))

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
	} else {
		searchHint := InputStyle.Render("Press / to search")
		b.WriteString(searchHint + "\n\n")
	}

	// Component list
	if len(m.filteredIndices) == 0 {
		b.WriteString(StatusInfoStyle.Render("No components found") + "\n")
	} else {
		// Show a window of items
		start := m.cursor - 5
		if start < 0 {
			start = 0
		}
		end := start + 15
		if end > len(m.filteredIndices) {
			end = len(m.filteredIndices)
		}

		for i := start; i < end; i++ {
			idx := m.filteredIndices[i]
			component := m.components[idx]

			checkbox := "☐"
			if component.Selected {
				checkbox = "☑"
			}

			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			line := fmt.Sprintf("%s%s %s", cursor, checkbox, component.Name)
			if component.Category != "root" && component.Category != "" {
				line += CategoryStyle.Render(" ("+component.Category+")")
			}

			if i == m.cursor {
				b.WriteString(SelectedItemStyle.Render(line) + "\n")
			} else if component.Selected {
				b.WriteString(CheckedItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(UnselectedItemStyle.Render(line) + "\n")
			}
		}
	}

	// Status bar
	selectedCount := m.getSelectedCount()
	b.WriteString("\n")
	statusMsg := fmt.Sprintf("Selected: %d/%d", selectedCount, len(m.components))
	b.WriteString(StatusBarStyle.Render(statusMsg) + "\n\n")

	// Help
	b.WriteString(HelpStyle.Render(
		"↑/↓: Navigate • PgUp/PgDn: Page • Space: Toggle • A: Select All • a: Deselect All\n" +
			"/: Search • R: Refresh • Enter: Install • Esc: Back"))

	return BoxStyle.Width(m.width - 4).Render(b.String())
}

// viewConfirmScreen renders the confirmation screen
func (m Model) viewConfirmScreen() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Confirm Installation") + "\n\n")

	selected := m.getSelectedComponents()
	b.WriteString(fmt.Sprintf("You are about to install %s:\n\n",
		StatusInfoStyle.Render(fmt.Sprintf("%d component(s)", len(selected)))))

	for _, comp := range selected {
		icon := m.getIconForType(comp.Type + "s")
		b.WriteString(fmt.Sprintf("  %s %s", icon, comp.Name))
		if comp.Category != "root" && comp.Category != "" {
			b.WriteString(CategoryStyle.Render(" ("+comp.Category+")"))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Target directory: %s\n\n", StatusInfoStyle.Render(m.targetDir)))

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

	b.WriteString(HelpStyle.Render("R: Return to Main • Q/Enter: Quit"))

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
		"Agents":   "🤖",
		"Commands": "⚡",
		"MCPs":     "🔌",
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

// Commands

func loadComponentsCmd(componentType string, forceRefresh ...bool) tea.Cmd {
	return func() tea.Msg {
		loader := NewComponentLoader()

		refresh := false
		if len(forceRefresh) > 0 {
			refresh = forceRefresh[0]
		}

		components, err := loader.LoadComponentsWithCache(componentType, refresh)
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
