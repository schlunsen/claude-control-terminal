package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/schlunsen/claude-control-terminal/internal/fileops"
)

// handlePermissionsScreen handles input on the permissions screen
func (m Model) handlePermissionsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.permissionsLoading {
		if msg.String() == "esc" {
			m.screen = ScreenMain
			return m, nil
		}
		return m, nil
	}

	// Handle custom permission input mode
	if m.permissionsAddingCustom {
		switch msg.String() {
		case "esc":
			m.permissionsAddingCustom = false
			m.permissionsCustomInput.Blur()
			m.permissionsCustomInput.SetValue("")
			return m, nil
		case "enter":
			// Add custom permission
			customPattern := strings.TrimSpace(m.permissionsCustomInput.Value())
			if customPattern != "" {
				// Get current settings
				var settings *fileops.ClaudeSettings
				switch m.permissionsCurrentTab {
				case fileops.SettingsSourceGlobal:
					settings = m.permissionsMultiSource.Global
				case fileops.SettingsSourceProject:
					settings = m.permissionsMultiSource.Project
				case fileops.SettingsSourceLocal:
					settings = m.permissionsMultiSource.Local
				}

				// Initialize permissions if needed
				if settings.Permissions == nil {
					settings.Permissions = &fileops.PermissionsConfig{}
				}

				// Add to allow list if not already present
				found := false
				for _, existing := range settings.Permissions.Allow {
					if existing == customPattern {
						found = true
						break
					}
				}
				if !found {
					settings.Permissions.Allow = append(settings.Permissions.Allow, customPattern)
				}
			}
			m.permissionsAddingCustom = false
			m.permissionsCustomInput.Blur()
			m.permissionsCustomInput.SetValue("")
			return m, nil
		default:
			var cmd tea.Cmd
			m.permissionsCustomInput, cmd = m.permissionsCustomInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "up", "k":
		if m.permissionsCursor > 0 {
			m.permissionsCursor--
		}
	case "down", "j":
		if m.permissionsCursor < len(m.permissionItems)-1 {
			m.permissionsCursor++
		}
	case "tab", "1", "2", "3":
		// Save current tab's state before switching
		m.saveCurrentTabStates()

		// Switch tabs
		switch msg.String() {
		case "1":
			m.permissionsCurrentTab = fileops.SettingsSourceGlobal
		case "2":
			m.permissionsCurrentTab = fileops.SettingsSourceProject
		case "3":
			m.permissionsCurrentTab = fileops.SettingsSourceLocal
		case "tab":
			// Cycle forwards through tabs
			switch m.permissionsCurrentTab {
			case fileops.SettingsSourceGlobal:
				m.permissionsCurrentTab = fileops.SettingsSourceProject
			case fileops.SettingsSourceProject:
				m.permissionsCurrentTab = fileops.SettingsSourceLocal
			case fileops.SettingsSourceLocal:
				m.permissionsCurrentTab = fileops.SettingsSourceGlobal
			}
		}

		// Load the new tab's states
		m.loadCurrentTabStates()
		return m, nil

	case " ":
		// Toggle permission
		if m.permissionsCursor < len(m.permissionItems) {
			item := m.permissionItems[m.permissionsCursor]
			currentState := m.permissionStates[m.permissionsCursor]
			m.permissionStates[m.permissionsCursor] = !currentState

			// If enabling "Bypass All Permissions", note that it will override other permissions
			// This is handled during save, so no action needed here
			_ = item.ModeValue == fileops.PermissionModeBypassPermissions
		}
		return m, nil
	case "a":
		// Activate custom permission input mode
		m.permissionsAddingCustom = true
		m.permissionsCustomInput.Focus()
		return m, textinput.Blink
	case "d", "x":
		// Delete a custom permission from the raw list (not from predefined items)
		// This will be shown in the view - for now just skip
		return m, nil
	case "enter":
		// Save current tab's state before saving
		m.saveCurrentTabStates()

		// Get the current settings object (with any custom permissions added)
		var currentSettings *fileops.ClaudeSettings
		switch m.permissionsCurrentTab {
		case fileops.SettingsSourceGlobal:
			currentSettings = m.permissionsMultiSource.Global
		case fileops.SettingsSourceProject:
			currentSettings = m.permissionsMultiSource.Project
		case fileops.SettingsSourceLocal:
			currentSettings = m.permissionsMultiSource.Local
		}

		// Save current tab's permissions and return to main
		m.permissionsSaving = true
		return m, savePermissionsCmd(m.permissionItems, m.permissionStates, m.permissionsCurrentTab, currentSettings, m.targetDir)
	case "r":
		// Reset all permissions to safe defaults
		for i := range m.permissionStates {
			m.permissionStates[i] = false
		}
		return m, nil
	case "esc":
		// Go back to main without saving
		m.screen = ScreenMain
		return m, nil
	}

	return m, nil
}

// saveCurrentTabStates saves the current permissionStates to the appropriate tab's state array
func (m *Model) saveCurrentTabStates() {
	switch m.permissionsCurrentTab {
	case fileops.SettingsSourceGlobal:
		m.permissionsGlobalStates = m.permissionStates
	case fileops.SettingsSourceProject:
		m.permissionsProjectStates = m.permissionStates
	case fileops.SettingsSourceLocal:
		m.permissionsLocalStates = m.permissionStates
	}
}

// loadCurrentTabStates loads the appropriate tab's state into permissionStates
func (m *Model) loadCurrentTabStates() {
	switch m.permissionsCurrentTab {
	case fileops.SettingsSourceGlobal:
		m.permissionStates = m.permissionsGlobalStates
	case fileops.SettingsSourceProject:
		m.permissionStates = m.permissionsProjectStates
	case fileops.SettingsSourceLocal:
		m.permissionStates = m.permissionsLocalStates
	}
}

// viewPermissionsScreen renders the permissions screen
func (m Model) viewPermissionsScreen() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚙️  Permissions Configuration") + "\n\n")

	if m.permissionsLoading {
		b.WriteString(m.spinner.View() + " Loading permissions...\n")
		return BoxStyle.Render(b.String())
	}

	if m.permissionsSaving {
		b.WriteString(m.spinner.View() + " Saving permissions...\n")
		return BoxStyle.Render(b.String())
	}

	if m.permissionsError != nil {
		b.WriteString(StatusErrorStyle.Render("Error: ") + m.permissionsError.Error() + "\n\n")
		b.WriteString(HelpStyle.Render("Esc: Go Back"))
		return BoxStyle.Render(b.String())
	}

	// Render tabs
	tabs := []struct {
		name   string
		source fileops.SettingsSource
	}{
		{"1: Global", fileops.SettingsSourceGlobal},
		{"2: Project", fileops.SettingsSourceProject},
		{"3: Local", fileops.SettingsSourceLocal},
	}

	var tabsStr string
	for _, tab := range tabs {
		if tab.source == m.permissionsCurrentTab {
			tabsStr += SelectedItemStyle.Render("["+tab.name+"]") + " "
		} else {
			tabsStr += HelpStyle.Render(" "+tab.name+" ") + " "
		}
	}
	b.WriteString(tabsStr + "\n\n")

	// Show current settings file path and existence
	var settingsPath string
	var fileExists bool
	switch m.permissionsCurrentTab {
	case fileops.SettingsSourceGlobal:
		settingsPath = fileops.GetGlobalSettingsPath()
		fileExists = fileops.SettingsFileExists(fileops.SettingsSourceGlobal, m.targetDir)
	case fileops.SettingsSourceProject:
		settingsPath = fileops.GetProjectSettingsPath(m.targetDir)
		fileExists = fileops.SettingsFileExists(fileops.SettingsSourceProject, m.targetDir)
	case fileops.SettingsSourceLocal:
		settingsPath = fileops.GetLocalSettingsPath(m.targetDir)
		fileExists = fileops.SettingsFileExists(fileops.SettingsSourceLocal, m.targetDir)
	}

	existsIndicator := ""
	if fileExists {
		existsIndicator = StatusSuccessStyle.Render(" ✓ exists")
	} else {
		existsIndicator = StatusWarningStyle.Render(" ⚠ will be created")
	}
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("File: %s", settingsPath)) + existsIndicator + "\n\n")

	// Group permissions by category
	categories := map[string][]int{
		"bash":  {},
		"tools": {},
		"mode":  {},
	}

	for i, item := range m.permissionItems {
		categories[item.Category] = append(categories[item.Category], i)
	}

	// Display categories
	categoryOrder := []string{"bash", "tools", "mode"}
	categoryNames := map[string]string{
		"bash":  "Shell Commands",
		"tools": "Tool Permissions",
		"mode":  "Permission Modes",
	}

	for _, cat := range categoryOrder {
		indices := categories[cat]
		if len(indices) == 0 {
			continue
		}

		b.WriteString(CategoryStyle.Render(categoryNames[cat]) + "\n")

		for _, i := range indices {
			item := m.permissionItems[i]
			enabled := m.permissionStates[i]

			cursor := "  "
			checkbox := "[ ]"

			if i == m.permissionsCursor {
				cursor = "> "
			}

			if enabled {
				checkbox = "[✓]"
			}

			// Build the line
			line := fmt.Sprintf("%s%s %s", cursor, checkbox, item.Name)

			// Add description for selected item
			if i == m.permissionsCursor {
				b.WriteString(SelectedItemStyle.Render(line) + "\n")
				b.WriteString("  " + SubtitleStyle.Render(item.Description) + "\n")
			} else {
				if enabled {
					b.WriteString(StatusSuccessStyle.Render(line) + "\n")
				} else {
					b.WriteString(line + "\n")
				}
			}
		}
		b.WriteString("\n")
	}

	// Show summary
	enabledCount := 0
	for _, enabled := range m.permissionStates {
		if enabled {
			enabledCount++
		}
	}
	b.WriteString(StatusInfoStyle.Render(fmt.Sprintf("Enabled: %d/%d", enabledCount, len(m.permissionItems))) + "\n\n")

	// Show raw permission patterns currently in the settings
	var currentSettings *fileops.ClaudeSettings
	switch m.permissionsCurrentTab {
	case fileops.SettingsSourceGlobal:
		currentSettings = m.permissionsMultiSource.Global
	case fileops.SettingsSourceProject:
		currentSettings = m.permissionsMultiSource.Project
	case fileops.SettingsSourceLocal:
		currentSettings = m.permissionsMultiSource.Local
	}

	b.WriteString(CategoryStyle.Render("Raw Permission Patterns") + "\n")
	if currentSettings != nil && currentSettings.Permissions != nil && len(currentSettings.Permissions.Allow) > 0 {
		for _, pattern := range currentSettings.Permissions.Allow {
			b.WriteString(StatusSuccessStyle.Render("  ✓ ") + pattern + "\n")
		}
	} else {
		b.WriteString(HelpStyle.Render("  (no custom permissions)") + "\n")
	}
	b.WriteString("\n")

	// Show custom permission input if active
	if m.permissionsAddingCustom {
		b.WriteString(CategoryStyle.Render("Add Custom Permission") + "\n")
		b.WriteString(InputFocusedStyle.Render(m.permissionsCustomInput.View()) + "\n")
		b.WriteString(HelpStyle.Render("Enter: Add • Esc: Cancel") + "\n\n")
	}

	// Help text - show different help based on screen size
	if m.height < 30 {
		// Compact help for smaller screens
		b.WriteString(HelpStyle.Render("↑/↓: Nav • Space: Toggle • Tab: Switch Tab • A: Add Custom\n"))
		b.WriteString(HelpStyle.Render("Enter: Save • R: Reset • Esc: Cancel"))
	} else {
		// Full help for larger screens
		b.WriteString(HelpStyle.Render("↑/↓: Navigate • Space: Toggle • Tab or 1/2/3: Switch Tab\n"))
		b.WriteString(HelpStyle.Render("A: Add Custom Permission • Enter: Save & Exit • R: Reset All • Esc: Cancel"))
	}

	return BoxStyle.Width(m.width - 4).Render(b.String())
}

// Messages

type permissionsLoadedMsg struct {
	items         []fileops.PermissionItem
	multiSource   *fileops.MultiSourceSettings
	globalStates  []bool
	projectStates []bool
	localStates   []bool
	err           error
}

type permissionsSavedMsg struct {
	err error
}

// Commands

func loadPermissionsCmd(targetDir string) tea.Cmd {
	return func() tea.Msg {
		// Load settings from all three sources
		multiSource, err := fileops.LoadAllSettings(targetDir)
		if err != nil {
			return permissionsLoadedMsg{err: err}
		}

		items := fileops.GetDefaultPermissionItems()

		// Calculate states for each source
		globalStates := make([]bool, len(items))
		projectStates := make([]bool, len(items))
		localStates := make([]bool, len(items))

		for i, item := range items {
			globalStates[i] = fileops.IsPermissionEnabled(multiSource.Global, item)
			projectStates[i] = fileops.IsPermissionEnabled(multiSource.Project, item)
			localStates[i] = fileops.IsPermissionEnabled(multiSource.Local, item)
		}

		return permissionsLoadedMsg{
			items:         items,
			multiSource:   multiSource,
			globalStates:  globalStates,
			projectStates: projectStates,
			localStates:   localStates,
			err:           nil,
		}
	}
}

func savePermissionsCmd(items []fileops.PermissionItem, states []bool, source fileops.SettingsSource, settings *fileops.ClaudeSettings, targetDir string) tea.Cmd {
	return func() tea.Msg {
		// Apply all permission changes from toggled states
		for i, item := range items {
			fileops.TogglePermission(settings, item, states[i])
		}

		// Save to the appropriate file
		var err error
		switch source {
		case fileops.SettingsSourceGlobal:
			err = fileops.SaveGlobalSettings(settings)
		case fileops.SettingsSourceProject:
			err = fileops.SaveProjectSettings(settings, targetDir)
		case fileops.SettingsSourceLocal:
			err = fileops.SaveLocalSettings(settings, targetDir)
		default:
			return permissionsSavedMsg{err: fmt.Errorf("invalid settings source")}
		}

		if err != nil {
			return permissionsSavedMsg{err: err}
		}

		return permissionsSavedMsg{err: nil}
	}
}
