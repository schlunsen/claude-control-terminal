package tui

import (
	"fmt"
	"strings"

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

	switch msg.String() {
	case "up", "k":
		if m.permissionsCursor > 0 {
			m.permissionsCursor--
		}
	case "down", "j":
		if m.permissionsCursor < len(m.permissionItems)-1 {
			m.permissionsCursor++
		}
	case " ":
		// Toggle permission
		if m.permissionsCursor < len(m.permissionItems) {
			item := m.permissionItems[m.permissionsCursor]
			currentState := m.permissionStates[m.permissionsCursor]
			m.permissionStates[m.permissionsCursor] = !currentState

			// If enabling "Bypass All Permissions", disable other conflicting permissions
			if item.IsMode && item.ModeValue == fileops.PermissionModeBypassPermissions && !currentState {
				// Bypass mode is being enabled, nothing special to do
			}
		}
		return m, nil
	case "enter":
		// Save permissions and return to main
		m.permissionsSaving = true
		return m, savePermissionsCmd(m.permissionItems, m.permissionStates, m.targetDir)
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

// viewPermissionsScreen renders the permissions screen
func (m Model) viewPermissionsScreen() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚙️  Permissions Configuration") + "\n\n")
	b.WriteString(SubtitleStyle.Render("Configure Claude Code tool approvals") + "\n\n")

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

	// Show settings file location
	settingsPath := fileops.GetClaudeSettingsPath(m.targetDir)
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Settings: %s", settingsPath)) + "\n")

	// Show summary
	enabledCount := 0
	for _, enabled := range m.permissionStates {
		if enabled {
			enabledCount++
		}
	}
	b.WriteString(StatusInfoStyle.Render(fmt.Sprintf("Enabled: %d/%d", enabledCount, len(m.permissionItems))) + "\n\n")

	// Help text
	b.WriteString(HelpStyle.Render("↑/↓: Navigate • Space: Toggle • Enter: Save & Exit • R: Reset All • Esc: Cancel"))

	return BoxStyle.Width(m.width - 4).Render(b.String())
}

// Messages

type permissionsLoadedMsg struct {
	items  []fileops.PermissionItem
	states []bool
	err    error
}

type permissionsSavedMsg struct {
	err error
}

// Commands

func loadPermissionsCmd(targetDir string) tea.Cmd {
	return func() tea.Msg {
		settings, err := fileops.LoadClaudeSettings(targetDir)
		if err != nil {
			return permissionsLoadedMsg{err: err}
		}

		items := fileops.GetDefaultPermissionItems()
		states := make([]bool, len(items))

		for i, item := range items {
			states[i] = fileops.IsPermissionEnabled(settings, item)
		}

		return permissionsLoadedMsg{
			items:  items,
			states: states,
			err:    nil,
		}
	}
}

func savePermissionsCmd(items []fileops.PermissionItem, states []bool, targetDir string) tea.Cmd {
	return func() tea.Msg {
		settings, err := fileops.LoadClaudeSettings(targetDir)
		if err != nil {
			return permissionsSavedMsg{err: err}
		}

		// Apply all permission changes
		for i, item := range items {
			fileops.TogglePermission(settings, item, states[i])
		}

		if err := fileops.SaveClaudeSettings(settings, targetDir); err != nil {
			return permissionsSavedMsg{err: err}
		}

		return permissionsSavedMsg{err: nil}
	}
}
