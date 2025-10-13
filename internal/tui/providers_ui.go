package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/schlunsen/claude-control-terminal/internal/database"
	"github.com/schlunsen/claude-control-terminal/internal/providers"
)

// Provider messages

type providerSavedMsg struct {
	success bool
	err     error
}

// Provider commands

func saveProviderCmd(repo *database.Repository, config *database.ProviderConfig) tea.Cmd {
	return func() tea.Msg {
		// Save the provider configuration to database
		if err := providers.SaveProviderConfig(repo, config); err != nil {
			return providerSavedMsg{success: false, err: err}
		}

		// Generate the environment script
		if err := providers.GenerateEnvScript(config); err != nil {
			return providerSavedMsg{success: false, err: err}
		}

		return providerSavedMsg{success: true, err: nil}
	}
}

// Handler: Providers list screen
func (m Model) handleProvidersListScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	providersList := providers.GetAvailableProviders()

	// Load current provider config to highlight it
	if m.selectedProviderID == "" && m.hasProviderConfig && m.dbRepo != nil {
		if config, err := providers.LoadProviderConfig(m.dbRepo); err == nil && config != nil {
			m.selectedProviderID = config.ProviderID
		}
	}

	switch msg.String() {
	case "up", "k":
		if m.providersCursor > 0 {
			m.providersCursor--
		}
	case "down", "j":
		if m.providersCursor < len(providersList)-1 {
			m.providersCursor++
		}
	case "enter":
		// Select provider
		selectedProvider := providersList[m.providersCursor]
		m.selectedProviderID = selectedProvider.ID

		// Special handling for Claude (default) - no API key needed
		if selectedProvider.ID == "claude" {
			// Create a minimal configuration
			config := &database.ProviderConfig{
				ProviderID: "claude",
				APIKey:     "", // No API key needed for default
				CustomURL:  "",
			}

			// Save and generate script immediately
			m.screen = ScreenProviderSaving
			m.providerSaving = true
			return m, saveProviderCmd(m.dbRepo, config)
		}

		// For other providers, move to input screen
		m.screen = ScreenProviderInput

		// Check if this is custom provider
		isCustomProvider := selectedProvider.ID == "custom"

		// Try to load existing configuration for this specific provider
		if m.dbRepo != nil {
			existingConfig, err := providers.GetProviderConfig(m.dbRepo, selectedProvider.ID)
			if err == nil && existingConfig != nil {
				// Pre-fill with saved values
				m.providerAPIKeyInput.SetValue(existingConfig.APIKey)
				m.providerCustomURL.SetValue(existingConfig.CustomURL)
				m.providerModelInput.SetValue(existingConfig.ModelName)

				// Set model cursor to saved model if found (for non-custom providers)
				if !isCustomProvider && existingConfig.ModelName != "" {
					// Find the saved model in the list
					// Add 1 to account for "No model" option at position 0
					found := false
					for i, model := range selectedProvider.Models {
						if model == existingConfig.ModelName {
							m.providerModelCursor = i + 1
							found = true
							break
						}
					}
					if !found {
						m.providerModelCursor = 0 // Default to "No model"
					}
				} else {
					m.providerModelCursor = 0 // Default to "No model (use provider default)"
				}
			} else {
				// Reset input fields for new provider
				m.providerAPIKeyInput.SetValue("")
				m.providerCustomURL.SetValue("")
				m.providerModelInput.SetValue("")
				m.providerModelCursor = 0
			}
		}

		m.providerAPIKeyInput.Focus()
		m.providerError = nil

		return m, textinput.Blink
	case "d", "x":
		// Delete current provider configuration
		if m.hasProviderConfig && m.dbRepo != nil {
			if err := providers.DeleteProviderConfig(m.dbRepo); err != nil {
				m.providerError = err
				return m, nil
			}
			// Update state to reflect deletion
			m.hasProviderConfig = false
			m.currentProviderName = ""
			m.providerSuccessMsg = "Provider configuration deleted"
		}
		return m, nil
	case "esc":
		// Go back to main screen
		m.screen = ScreenMain
		return m, nil
	}

	return m, nil
}

// Handler: Provider input screen
func (m Model) handleProviderInputScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	provider := providers.GetProviderByID(m.selectedProviderID)
	if provider == nil {
		m.providerError = fmt.Errorf("provider not found")
		m.screen = ScreenProvidersList
		return m, nil
	}

	// Check if we're in custom URL mode (only for Custom provider)
	isCustomProvider := provider.ID == "custom"
	apiKeyFocused := m.providerAPIKeyInput.Focused()
	customURLFocused := m.providerCustomURL.Focused()
	modelInputFocused := m.providerModelInput.Focused()
	hasModels := len(provider.Models) > 0

	switch msg.String() {
	case "esc":
		// Go back to providers list
		m.providerAPIKeyInput.Blur()
		m.providerCustomURL.Blur()
		m.providerModelInput.Blur()
		m.screen = ScreenProvidersList
		return m, nil
	case "up", "k":
		// Navigate model list (when not in input field)
		if !apiKeyFocused && !customURLFocused && !modelInputFocused && hasModels {
			if m.providerModelCursor > 0 {
				m.providerModelCursor--
			}
			return m, nil
		}
	case "down", "j":
		// Navigate model list (when not in input field)
		// Account for the extra "No model" option at position 0
		if !apiKeyFocused && !customURLFocused && !modelInputFocused && hasModels {
			if m.providerModelCursor < len(provider.Models) {
				m.providerModelCursor++
			}
			return m, nil
		}
	case "tab", "shift+tab":
		// For custom provider: cycle through API key -> Custom URL -> Model Name
		if isCustomProvider {
			if apiKeyFocused {
				m.providerAPIKeyInput.Blur()
				m.providerCustomURL.Focus()
			} else if customURLFocused {
				m.providerCustomURL.Blur()
				m.providerModelInput.Focus()
			} else {
				m.providerModelInput.Blur()
				m.providerAPIKeyInput.Focus()
			}
			return m, textinput.Blink
		}

		// For providers with models: toggle between input field and model selection
		if hasModels {
			if apiKeyFocused {
				m.providerAPIKeyInput.Blur()
			} else {
				m.providerAPIKeyInput.Focus()
			}
			return m, textinput.Blink
		}
	case "enter":
		// Save provider configuration
		apiKey := strings.TrimSpace(m.providerAPIKeyInput.Value())
		if apiKey == "" {
			m.providerError = fmt.Errorf("API key is required")
			return m, nil
		}

		// For custom provider, validate custom URL
		customURL := ""
		if isCustomProvider {
			customURL = strings.TrimSpace(m.providerCustomURL.Value())
			if customURL == "" {
				m.providerError = fmt.Errorf("custom URL is required for Custom provider")
				return m, nil
			}
		}

		// Get selected model
		selectedModel := ""
		if isCustomProvider {
			// For custom provider, get model from input field
			selectedModel = strings.TrimSpace(m.providerModelInput.Value())
		} else if hasModels && m.providerModelCursor > 0 && m.providerModelCursor <= len(provider.Models) {
			// For other providers with model lists, get from cursor position
			// If cursor is at 0, it means "No model (use provider default)"
			selectedModel = provider.Models[m.providerModelCursor-1]
		}

		// Create configuration
		config := &database.ProviderConfig{
			ProviderID: provider.ID,
			APIKey:     apiKey,
			CustomURL:  customURL,
			ModelName:  selectedModel,
		}

		// Save configuration
		m.screen = ScreenProviderSaving
		m.providerSaving = true
		m.providerAPIKeyInput.Blur()
		m.providerCustomURL.Blur()
		m.providerModelInput.Blur()

		return m, saveProviderCmd(m.dbRepo, config)
	}

	// Update the focused input field
	var cmd tea.Cmd
	if apiKeyFocused {
		m.providerAPIKeyInput, cmd = m.providerAPIKeyInput.Update(msg)
	} else if customURLFocused {
		m.providerCustomURL, cmd = m.providerCustomURL.Update(msg)
	} else if modelInputFocused {
		m.providerModelInput, cmd = m.providerModelInput.Update(msg)
	}

	return m, cmd
}

// Handler: Provider complete screen
func (m Model) handleProviderCompleteScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		// Go back to main screen
		m.screen = ScreenMain

		// Reload provider info
		if m.dbRepo != nil {
			currentProviderName, hasProviderConfig, _ := providers.GetCurrentProviderInfo(m.dbRepo)
			m.currentProviderName = currentProviderName
			m.hasProviderConfig = hasProviderConfig
		}

		return m, nil
	}
	return m, nil
}

// Update function integration - add this case to Update() in model.go
func (m Model) handleProviderSavedMsg(msg providerSavedMsg) (Model, tea.Cmd) {
	m.providerSaving = false

	if msg.err != nil {
		m.providerError = msg.err
		m.screen = ScreenProviderInput
		m.providerAPIKeyInput.Focus()
		return m, textinput.Blink
	}

	// Success - move to complete screen
	m.screen = ScreenProviderComplete
	provider := providers.GetProviderByID(m.selectedProviderID)
	if provider != nil {
		m.providerSuccessMsg = fmt.Sprintf("✓ %s configured successfully!", provider.Name)
	} else {
		m.providerSuccessMsg = "✓ Provider configured successfully!"
	}

	return m, nil
}

// View: Providers list screen
func (m Model) viewProvidersListScreen() string {
	var b strings.Builder

	providersList := providers.GetAvailableProviders()

	b.WriteString(TitleStyle.Render("🔑 Configure AI Provider") + "\n\n")
	b.WriteString(SubtitleStyle.Render("Select a provider to configure:") + "\n\n")

	// Load current provider config to show which is active
	var currentProviderID string
	if m.dbRepo != nil {
		if config, err := providers.LoadProviderConfig(m.dbRepo); err == nil && config != nil {
			currentProviderID = config.ProviderID
		}
	}

	// Display provider list
	for i, provider := range providersList {
		cursor := "  "
		if i == m.providersCursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%s %s", cursor, provider.Icon, provider.Name)

		// Highlight current provider with a check mark
		if currentProviderID != "" && provider.ID == currentProviderID {
			line += StatusSuccessStyle.Render(" ✓")
		}

		if i == m.providersCursor {
			b.WriteString(SelectedItemStyle.Render(line) + "\n")
		} else {
			b.WriteString(UnselectedItemStyle.Render(line) + "\n")
		}
	}

	b.WriteString("\n")

	// Show current provider info
	if m.hasProviderConfig {
		b.WriteString(StatusInfoStyle.Render(fmt.Sprintf("Current provider: %s", m.currentProviderName)) + "\n")
		b.WriteString(SubtitleStyle.Render("To switch providers, select a different one above") + "\n\n")
	} else {
		b.WriteString(StatusWarningStyle.Render("No provider configured") + "\n\n")
	}

	// Show success message if present
	if m.providerSuccessMsg != "" {
		b.WriteString(StatusSuccessStyle.Render(m.providerSuccessMsg) + "\n\n")
	}

	// Show error if present
	if m.providerError != nil {
		b.WriteString(StatusErrorStyle.Render("Error: "+m.providerError.Error()) + "\n\n")
	}

	b.WriteString(HelpStyle.Render("↑/↓: Navigate • Enter: Select • D: Delete Config • Esc: Back"))

	return BoxStyle.Render(b.String())
}

// View: Provider input screen
func (m Model) viewProviderInputScreen() string {
	var b strings.Builder

	provider := providers.GetProviderByID(m.selectedProviderID)
	if provider == nil {
		b.WriteString(StatusErrorStyle.Render("Provider not found") + "\n")
		return BoxStyle.Render(b.String())
	}

	// Check if we need compact mode (small terminal)
	isCompactMode := m.height < 25
	isCustomProvider := provider.ID == "custom"

	b.WriteString(TitleStyle.Render(fmt.Sprintf("%s Configure %s", provider.Icon, provider.Name)) + "\n\n")

	// In compact mode, show ONLY the currently focused/active field
	if isCompactMode {
		// Show focused field with position indicator
		if m.providerAPIKeyInput.Focused() {
			b.WriteString(SubtitleStyle.Render("API Key:") + StatusInfoStyle.Render(" (Field 1/"+m.getProviderFieldCount(provider)+")") + "\n")
			b.WriteString(InputFocusedStyle.Render(m.providerAPIKeyInput.View()) + "\n\n")
		} else if m.providerCustomURL.Focused() {
			b.WriteString(SubtitleStyle.Render("Base URL:") + StatusInfoStyle.Render(" (Field 2/"+m.getProviderFieldCount(provider)+")") + "\n")
			b.WriteString(InputFocusedStyle.Render(m.providerCustomURL.View()) + "\n\n")
		} else if m.providerModelInput.Focused() {
			b.WriteString(SubtitleStyle.Render("Model Name (optional):") + StatusInfoStyle.Render(" (Field 3/"+m.getProviderFieldCount(provider)+")") + "\n")
			b.WriteString(InputFocusedStyle.Render(m.providerModelInput.View()) + "\n\n")
		} else if !isCustomProvider && len(provider.Models) > 0 {
			// Model selection is active
			totalModels := len(provider.Models) + 1
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Select Model (%d/%d):", m.providerModelCursor+1, totalModels)) + StatusSuccessStyle.Render(" ↑/↓") + "\n\n")

			// Show only 3 models around the cursor for very compact view
			maxVisible := 3
			start := m.providerModelCursor - 1
			if start < 0 {
				start = 0
			}
			end := start + maxVisible
			if end > totalModels {
				end = totalModels
				start = end - maxVisible
				if start < 0 {
					start = 0
				}
			}

			// Show indicator if there are more above
			if start > 0 {
				b.WriteString(SubtitleStyle.Render("  ⬆ "+fmt.Sprintf("%d more above", start)+" ⬆\n"))
			}

			// Show "No model" option if in range
			if start == 0 {
				cursor := "  "
				if m.providerModelCursor == 0 {
					cursor = "> "
				}
				line := cursor + "No model (default)"
				if m.providerModelCursor == 0 {
					b.WriteString(SelectedItemStyle.Render(line) + "\n")
				} else {
					b.WriteString(UnselectedItemStyle.Render(line) + "\n")
				}
				start = 1
			}

			// Show visible models
			for i := start - 1; i < end - 1 && i < len(provider.Models); i++ {
				cursor := "  "
				if i+1 == m.providerModelCursor {
					cursor = "> "
				}
				line := cursor + provider.Models[i]
				if i+1 == m.providerModelCursor {
					b.WriteString(SelectedItemStyle.Render(line) + "\n")
				} else {
					b.WriteString(UnselectedItemStyle.Render(line) + "\n")
				}
			}

			// Show indicator if there are more below
			if end < totalModels {
				remaining := totalModels - end
				b.WriteString(SubtitleStyle.Render("  ⬇ "+fmt.Sprintf("%d more below", remaining)+" ⬇\n"))
			}
			b.WriteString("\n")
		}
	} else {
		// Full mode - show all fields
		// API Key input
		b.WriteString(SubtitleStyle.Render("API Key:") + "\n")
		if m.providerAPIKeyInput.Focused() {
			b.WriteString(InputFocusedStyle.Render(m.providerAPIKeyInput.View()) + "\n\n")
		} else {
			b.WriteString(InputStyle.Render(m.providerAPIKeyInput.View()) + "\n\n")
		}

		// Custom URL input (only for Custom provider)
		if isCustomProvider {
			b.WriteString(SubtitleStyle.Render("Base URL:") + "\n")
			if m.providerCustomURL.Focused() {
				b.WriteString(InputFocusedStyle.Render(m.providerCustomURL.View()) + "\n\n")
			} else {
				b.WriteString(InputStyle.Render(m.providerCustomURL.View()) + "\n\n")
			}

			// Model name input (for Custom provider)
			b.WriteString(SubtitleStyle.Render("Model Name (optional):") + "\n")
			if m.providerModelInput.Focused() {
				b.WriteString(InputFocusedStyle.Render(m.providerModelInput.View()) + "\n\n")
			} else {
				b.WriteString(InputStyle.Render(m.providerModelInput.View()) + "\n\n")
			}
		} else {
			// Show the base URL for non-custom providers
			b.WriteString(SubtitleStyle.Render("Base URL: ") + CategoryStyle.Render(provider.BaseURL) + "\n\n")
		}

		// Model selection (if models are available)
		if !isCustomProvider && len(provider.Models) > 0 {
			// Show if model selection is active (input not focused)
			modelSelectionActive := !m.providerAPIKeyInput.Focused() && !m.providerCustomURL.Focused()

			if modelSelectionActive {
				b.WriteString(SubtitleStyle.Render("Model: ") + StatusSuccessStyle.Render("(Press ↑/↓ to select)") + "\n")
			} else {
				b.WriteString(SubtitleStyle.Render("Model: ") + StatusInfoStyle.Render("(Press Tab to select)") + "\n")
			}

			// First option: "No model (use provider default)"
			cursor := "  "
			if m.providerModelCursor == 0 {
				cursor = "> "
			}
			line := cursor + "No model (use provider default)"
			if m.providerModelCursor == 0 {
				b.WriteString(SelectedItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(UnselectedItemStyle.Render(line) + "\n")
			}

			// Then show all available models
			for i, model := range provider.Models {
				cursor = "  "
				if i+1 == m.providerModelCursor {
					cursor = "> "
				}

				line = cursor + model
				if i+1 == m.providerModelCursor {
					b.WriteString(SelectedItemStyle.Render(line) + "\n")
				} else {
					b.WriteString(UnselectedItemStyle.Render(line) + "\n")
				}
			}
			b.WriteString("\n")
		}

		// Instructions
		b.WriteString(SubtitleStyle.Render("This will set:") + "\n")
		b.WriteString("  • ANTHROPIC_AUTH_TOKEN\n")
		b.WriteString("  • ANTHROPIC_BASE_URL\n")
		// Show ANTHROPIC_MODEL if:
		// - Custom provider with model input filled
		// - Other providers with model selection (cursor > 0)
		if isCustomProvider {
			if strings.TrimSpace(m.providerModelInput.Value()) != "" {
				b.WriteString("  • ANTHROPIC_MODEL\n")
			}
		} else if len(provider.Models) > 0 && m.providerModelCursor > 0 {
			b.WriteString("  • ANTHROPIC_MODEL\n")
		}
		b.WriteString("\n")
	}

	// Show error if present (compact in small mode)
	if m.providerError != nil {
		if isCompactMode {
			b.WriteString(StatusErrorStyle.Render("⚠ "+m.providerError.Error()) + "\n\n")
		} else {
			b.WriteString(StatusErrorStyle.Render("Error: "+m.providerError.Error()) + "\n\n")
		}
	}

	// Help text
	if isCompactMode {
		// Very compact help
		b.WriteString(HelpStyle.Render("Tab: Next • ↑↓: Nav • Enter: Save • Esc: Back"))
	} else {
		// Full help
		if isCustomProvider {
			b.WriteString(HelpStyle.Render("Tab: Cycle fields • Enter: Save • Esc: Cancel"))
		} else if len(provider.Models) > 0 {
			b.WriteString(HelpStyle.Render("Tab: Toggle input/model • ↑/↓: Select model • Enter: Save • Esc: Cancel"))
		} else {
			b.WriteString(HelpStyle.Render("Enter: Save • Esc: Cancel"))
		}
	}

	return BoxStyle.Render(b.String())
}

// Helper function to get the field count for a provider
func (m Model) getProviderFieldCount(provider *providers.Provider) string {
	if provider.ID == "custom" {
		return "3" // API Key, Base URL, Model Name
	}
	return "1" // Just API Key for others (base URL is shown as info)
}

// View: Provider saving screen
func (m Model) viewProviderSavingScreen() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Saving Configuration") + "\n\n")
	b.WriteString(m.spinner.View() + " Saving provider configuration...\n")

	return BoxStyle.Render(b.String())
}

// View: Provider complete screen
func (m Model) viewProviderCompleteScreen() string {
	var b strings.Builder

	provider := providers.GetProviderByID(m.selectedProviderID)
	if provider != nil {
		b.WriteString(StatusSuccessStyle.Render(fmt.Sprintf("✓ %s Configured!", provider.Name)) + "\n\n")
	} else {
		b.WriteString(StatusSuccessStyle.Render("✓ Provider Configured!") + "\n\n")
	}

	// Show success message
	if m.providerSuccessMsg != "" {
		b.WriteString(m.providerSuccessMsg + "\n\n")
	}

	// Special message for Claude (default)
	if provider != nil && provider.ID == "claude" {
		b.WriteString(TitleStyle.Render("Configuration:") + "\n\n")
		b.WriteString("Using default Claude API settings.\n")
		b.WriteString("No custom environment variables needed.\n\n")

		scriptPath := providers.GetEnvScriptPath()
		b.WriteString(SubtitleStyle.Render("To ensure default settings:") + "\n")
		b.WriteString(StatusInfoStyle.Render(fmt.Sprintf("   source %s", scriptPath)) + "\n\n")
		b.WriteString(SubtitleStyle.Render("This will unset any custom provider variables.") + "\n\n")
	} else {
		// Instructions for other providers
		scriptPath := providers.GetEnvScriptPath()
		b.WriteString(TitleStyle.Render("Configuration Details:") + "\n\n")

		// Show configured model if available
		if m.dbRepo != nil {
			if config, err := providers.GetProviderConfig(m.dbRepo, provider.ID); err == nil && config != nil {
				if config.ModelName != "" {
					b.WriteString(SubtitleStyle.Render("Model: ") + StatusSuccessStyle.Render(config.ModelName) + "\n\n")
				} else {
					b.WriteString(SubtitleStyle.Render("Model: ") + StatusInfoStyle.Render("Not set (using provider default)") + "\n\n")
				}
			}
		}

		b.WriteString(TitleStyle.Render("Next Steps:") + "\n\n")
		b.WriteString("1. Load environment variables:\n")
		b.WriteString(StatusInfoStyle.Render(fmt.Sprintf("   source %s", scriptPath)) + "\n\n")
		b.WriteString("2. Optionally, add to your shell profile:\n")
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("   echo 'source %s' >> ~/.bashrc", scriptPath)) + "\n")
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("   echo 'source %s' >> ~/.zshrc", scriptPath)) + "\n\n")
	}

	b.WriteString(HelpStyle.Render("Enter/Esc: Back to Main Menu"))

	return BoxStyle.Render(b.String())
}
