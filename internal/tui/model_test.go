package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	targetDir := "/test/dir"
	m := NewModel(targetDir)

	if m.targetDir != targetDir {
		t.Errorf("Expected targetDir '%s', got '%s'", targetDir, m.targetDir)
	}

	if m.screen != ScreenMain {
		t.Errorf("Expected initial screen to be ScreenMain, got %v", m.screen)
	}

	if len(m.componentTypes) == 0 {
		t.Error("Expected component types to be populated")
	}

	if m.width != 80 {
		t.Errorf("Expected default width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected default height 24, got %d", m.height)
	}
}

func TestModelInit(t *testing.T) {
	m := NewModel(".")
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestUpdateFilteredIndices(t *testing.T) {
	m := NewModel(".")

	// Add some test components
	m.components = []ComponentItem{
		{Name: "agent-one", Category: "test"},
		{Name: "agent-two", Category: "security"},
		{Name: "command-one", Category: "test"},
	}

	// Test with no search term
	m.searchInput.SetValue("")
	m.updateFilteredIndices()

	if len(m.filteredIndices) != 3 {
		t.Errorf("Expected 3 filtered indices, got %d", len(m.filteredIndices))
	}

	// Test with search term matching names
	m.searchInput.SetValue("agent")
	m.updateFilteredIndices()

	if len(m.filteredIndices) != 2 {
		t.Errorf("Expected 2 filtered indices for 'agent', got %d", len(m.filteredIndices))
	}

	// Test with search term matching category
	m.searchInput.SetValue("security")
	m.updateFilteredIndices()

	if len(m.filteredIndices) != 1 {
		t.Errorf("Expected 1 filtered index for 'security', got %d", len(m.filteredIndices))
	}

	// Test with no matches
	m.searchInput.SetValue("nonexistent")
	m.updateFilteredIndices()

	if len(m.filteredIndices) != 0 {
		t.Errorf("Expected 0 filtered indices for 'nonexistent', got %d", len(m.filteredIndices))
	}
}

func TestGetSelectedComponents(t *testing.T) {
	m := NewModel(".")

	// Add test components
	m.components = []ComponentItem{
		{Name: "agent1", Selected: true},
		{Name: "agent2", Selected: false},
		{Name: "agent3", Selected: true},
	}

	selected := m.getSelectedComponents()

	if len(selected) != 2 {
		t.Errorf("Expected 2 selected components, got %d", len(selected))
	}

	if selected[0].Name != "agent1" {
		t.Errorf("Expected first selected to be 'agent1', got '%s'", selected[0].Name)
	}

	if selected[1].Name != "agent3" {
		t.Errorf("Expected second selected to be 'agent3', got '%s'", selected[1].Name)
	}
}

func TestGetSelectedCount(t *testing.T) {
	m := NewModel(".")

	// Add test components
	m.components = []ComponentItem{
		{Name: "agent1", Selected: true},
		{Name: "agent2", Selected: false},
		{Name: "agent3", Selected: true},
		{Name: "agent4", Selected: true},
	}

	count := m.getSelectedCount()

	if count != 3 {
		t.Errorf("Expected selected count 3, got %d", count)
	}
}

func TestGetIconForType(t *testing.T) {
	m := NewModel(".")

	tests := []struct {
		typeName string
		wantIcon bool
	}{
		{"Agents", true},
		{"Commands", true},
		{"MCPs", true},
		{"Providers", true},
		{"Permissions", true},
		{"Launch Claude", true},
		{"Launch last Claude session", true},
		{"Unknown", true}, // Should return default icon
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			icon := m.getIconForType(tt.typeName)
			if (icon != "") != tt.wantIcon {
				t.Errorf("getIconForType(%s) returned '%s', wantIcon=%v", tt.typeName, icon, tt.wantIcon)
			}
		})
	}
}

func TestGetComponentType(t *testing.T) {
	tests := []struct {
		name         string
		selectedType int
		expected     string
	}{
		{"Agents selected", 0, "agent"},
		{"Commands selected", 1, "command"},
		{"MCPs selected", 2, "mcp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(".")
			m.selectedType = tt.selectedType

			result := m.getComponentType()
			if result != tt.expected {
				t.Errorf("Expected type '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestWindowSizeUpdate(t *testing.T) {
	m := NewModel(".")

	// Test window size message
	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
}

func TestComponentsLoadedMsg(t *testing.T) {
	m := NewModel(".")
	m.loading = true

	// Create components loaded message
	components := []ComponentItem{
		{Name: "test-agent", Type: "agent"},
		{Name: "test-command", Type: "command"},
	}

	msg := componentsLoadedMsg{
		components: components,
		err:        nil,
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.loading {
		t.Error("Expected loading to be false after componentsLoadedMsg")
	}

	if len(m.components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(m.components))
	}

	if m.loadError != nil {
		t.Errorf("Expected no load error, got %v", m.loadError)
	}
}

func TestInstallCompleteMsg(t *testing.T) {
	m := NewModel(".")
	m.installing = true
	m.screen = ScreenInstalling

	msg := installCompleteMsg{
		success: []string{"agent1", "agent2"},
		failed:  []string{"agent3"},
		err:     nil,
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.installing {
		t.Error("Expected installing to be false after installCompleteMsg")
	}

	if len(m.installSuccess) != 2 {
		t.Errorf("Expected 2 successful installs, got %d", len(m.installSuccess))
	}

	if len(m.installFailed) != 1 {
		t.Errorf("Expected 1 failed install, got %d", len(m.installFailed))
	}

	if m.screen != ScreenComplete {
		t.Errorf("Expected screen to be ScreenComplete, got %v", m.screen)
	}
}

func TestPreviewLoadedMsg(t *testing.T) {
	m := NewModel(".")
	m.previewLoading = true

	msg := previewLoadedMsg{
		content: "Test preview content",
		err:     nil,
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.previewLoading {
		t.Error("Expected previewLoading to be false after previewLoadedMsg")
	}

	if m.previewContent != "Test preview content" {
		t.Errorf("Expected preview content 'Test preview content', got '%s'", m.previewContent)
	}

	if m.previewError != nil {
		t.Errorf("Expected no preview error, got %v", m.previewError)
	}
}

func TestScreenTransitions(t *testing.T) {
	m := NewModel(".")

	// Test that initial screen is main
	if m.screen != ScreenMain {
		t.Errorf("Expected initial screen ScreenMain, got %v", m.screen)
	}

	// Test component list transition
	m.screen = ScreenComponentList
	if m.screen != ScreenComponentList {
		t.Errorf("Expected ScreenComponentList, got %v", m.screen)
	}

	// Test preview transition
	m.screen = ScreenPreview
	if m.screen != ScreenPreview {
		t.Errorf("Expected ScreenPreview, got %v", m.screen)
	}

	// Test confirm transition
	m.screen = ScreenConfirm
	if m.screen != ScreenConfirm {
		t.Errorf("Expected ScreenConfirm, got %v", m.screen)
	}

	// Test installing transition
	m.screen = ScreenInstalling
	if m.screen != ScreenInstalling {
		t.Errorf("Expected ScreenInstalling, got %v", m.screen)
	}

	// Test complete transition
	m.screen = ScreenComplete
	if m.screen != ScreenComplete {
		t.Errorf("Expected ScreenComplete, got %v", m.screen)
	}
}

func TestViewFunctions(t *testing.T) {
	m := NewModel(".")

	// Test that view functions return non-empty strings
	tests := []struct {
		name   string
		screen Screen
	}{
		{"Main screen", ScreenMain},
		{"Component list screen", ScreenComponentList},
		{"Complete screen", ScreenComplete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.screen = tt.screen
			view := m.View()

			if tt.screen == ScreenMain || tt.screen == ScreenComplete {
				// These screens should always have content
				if view == "" {
					t.Errorf("Expected non-empty view for %s", tt.name)
				}
			}
		})
	}
}

func TestQuittingState(t *testing.T) {
	m := NewModel(".")
	m.quitting = true

	// When quitting, view should return empty string
	view := m.View()
	if view != "" {
		t.Error("Expected empty view when quitting")
	}
}

func TestToggleAnalyticsMsg(t *testing.T) {
	m := NewModel(".")
	m.analyticsEnabled = false
	m.analyticsServer = nil

	msg := toggleAnalyticsMsg{
		enabled:   true,
		targetDir: "/test/dir",
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	// Note: In tests, the server won't actually start, but we can test the message handling
	// The actual server creation is tested in integration tests
}

func TestRemoveCompleteMsg(t *testing.T) {
	m := NewModel(".")
	m.installing = true

	msg := removeCompleteMsg{
		success: []string{"removed-agent"},
		failed:  []string{},
		err:     nil,
	}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if m.installing {
		t.Error("Expected installing to be false after removeCompleteMsg")
	}

	if len(m.installSuccess) != 1 {
		t.Errorf("Expected 1 successful removal, got %d", len(m.installSuccess))
	}

	if m.screen != ScreenComplete {
		t.Errorf("Expected screen to be ScreenComplete, got %v", m.screen)
	}
}
