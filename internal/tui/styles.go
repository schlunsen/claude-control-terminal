package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// ColorScheme represents a color theme
type ColorScheme struct {
	Primary          lipgloss.Color
	Secondary        lipgloss.Color
	Success          lipgloss.Color
	Warning          lipgloss.Color
	Error            lipgloss.Color
	Info             lipgloss.Color
	BgPrimary        lipgloss.Color
	BgSecondary      lipgloss.Color
	BgTertiary       lipgloss.Color
	BgSelected       lipgloss.Color
	TextPrimary      lipgloss.Color
	TextSecondary    lipgloss.Color
	TextDim          lipgloss.Color
	Border           lipgloss.Color
	BorderSecondary  lipgloss.Color
	BorderDim        lipgloss.Color
}

// Available color schemes
var (
	// Orange scheme (default)
	OrangeScheme = ColorScheme{
		Primary:         lipgloss.Color("#FF6B35"),
		Secondary:       lipgloss.Color("#F7931E"),
		Success:         lipgloss.Color("#4ECDC4"),
		Warning:         lipgloss.Color("#FFE66D"),
		Error:           lipgloss.Color("#FF6B6B"),
		Info:            lipgloss.Color("#95E1D3"),
		BgPrimary:       lipgloss.Color("#1A1A2E"),
		BgSecondary:     lipgloss.Color("#16213E"),
		BgTertiary:      lipgloss.Color("#0F3460"),
		BgSelected:      lipgloss.Color("#E94560"),
		TextPrimary:     lipgloss.Color("#EAEAEA"),
		TextSecondary:   lipgloss.Color("#A0A0A0"),
		TextDim:         lipgloss.Color("#606060"),
		Border:          lipgloss.Color("#FF6B35"),
		BorderSecondary: lipgloss.Color("#4ECDC4"),
		BorderDim:       lipgloss.Color("#404040"),
	}

	// Neon Green scheme
	NeonGreenScheme = ColorScheme{
		Primary:         lipgloss.Color("#39FF14"), // Neon green
		Secondary:       lipgloss.Color("#00FF41"), // Matrix green
		Success:         lipgloss.Color("#7FFF00"), // Chartreuse
		Warning:         lipgloss.Color("#FFFF00"), // Yellow
		Error:           lipgloss.Color("#FF1744"), // Bright red
		Info:            lipgloss.Color("#00FFFF"), // Cyan
		BgPrimary:       lipgloss.Color("#0D0D0D"), // Almost black
		BgSecondary:     lipgloss.Color("#1A1A1A"), // Dark gray
		BgTertiary:      lipgloss.Color("#0A2F0A"), // Dark green tint
		BgSelected:      lipgloss.Color("#1A4D1A"), // Dark green for selection
		TextPrimary:     lipgloss.Color("#E0FFE0"), // Light green tint
		TextSecondary:   lipgloss.Color("#80FF80"), // Medium green
		TextDim:         lipgloss.Color("#4D7F4D"), // Dim green
		Border:          lipgloss.Color("#39FF14"), // Neon green
		BorderSecondary: lipgloss.Color("#00FF41"), // Matrix green
		BorderDim:       lipgloss.Color("#2D5F2D"), // Dark green
	}

	// Cyan/Blue scheme
	CyanScheme = ColorScheme{
		Primary:         lipgloss.Color("#00D9FF"), // Bright cyan
		Secondary:       lipgloss.Color("#00B4D8"), // Blue cyan
		Success:         lipgloss.Color("#4ECDC4"), // Teal
		Warning:         lipgloss.Color("#FFE66D"), // Yellow
		Error:           lipgloss.Color("#FF6B6B"), // Red
		Info:            lipgloss.Color("#90E0EF"), // Light blue
		BgPrimary:       lipgloss.Color("#03045E"), // Navy
		BgSecondary:     lipgloss.Color("#023E8A"), // Medium blue
		BgTertiary:      lipgloss.Color("#0077B6"), // Blue
		BgSelected:      lipgloss.Color("#0096C7"), // Light blue
		TextPrimary:     lipgloss.Color("#CAF0F8"), // Very light blue
		TextSecondary:   lipgloss.Color("#90E0EF"), // Light cyan
		TextDim:         lipgloss.Color("#48CAE4"), // Cyan
		Border:          lipgloss.Color("#00D9FF"), // Bright cyan
		BorderSecondary: lipgloss.Color("#00B4D8"), // Blue cyan
		BorderDim:       lipgloss.Color("#0077B6"), // Blue
	}

	// Purple scheme
	PurpleScheme = ColorScheme{
		Primary:         lipgloss.Color("#BD00FF"), // Bright purple
		Secondary:       lipgloss.Color("#9D4EDD"), // Medium purple
		Success:         lipgloss.Color("#06FFA5"), // Mint green
		Warning:         lipgloss.Color("#FFD60A"), // Gold
		Error:           lipgloss.Color("#FF006E"), // Hot pink
		Info:            lipgloss.Color("#C77DFF"), // Light purple
		BgPrimary:       lipgloss.Color("#10002B"), // Very dark purple
		BgSecondary:     lipgloss.Color("#240046"), // Dark purple
		BgTertiary:      lipgloss.Color("#3C096C"), // Medium dark purple
		BgSelected:      lipgloss.Color("#5A189A"), // Purple
		TextPrimary:     lipgloss.Color("#E0AAFF"), // Light purple
		TextSecondary:   lipgloss.Color("#C77DFF"), // Medium light purple
		TextDim:         lipgloss.Color("#9D4EDD"), // Medium purple
		Border:          lipgloss.Color("#BD00FF"), // Bright purple
		BorderSecondary: lipgloss.Color("#9D4EDD"), // Medium purple
		BorderDim:       lipgloss.Color("#5A189A"), // Dark purple
	}
)

// Current active color scheme
var activeScheme = getSchemeFromEnv()

// getSchemeFromEnv returns color scheme based on CCT_THEME env variable
func getSchemeFromEnv() ColorScheme {
	theme := os.Getenv("CCT_THEME")
	switch theme {
	case "neon", "green":
		return NeonGreenScheme
	case "cyan", "blue":
		return CyanScheme
	case "purple", "magenta":
		return PurpleScheme
	case "orange":
		return OrangeScheme
	default:
		return OrangeScheme
	}
}

// SetColorScheme changes the active color scheme and updates all styles
func SetColorScheme(scheme ColorScheme) {
	activeScheme = scheme
	initializeStyles()
}

// Color palette - Dynamic based on active scheme
var (
	ColorPrimary         lipgloss.Color
	ColorSecondary       lipgloss.Color
	ColorSuccess         lipgloss.Color
	ColorWarning         lipgloss.Color
	ColorError           lipgloss.Color
	ColorInfo            lipgloss.Color
	ColorBgPrimary       lipgloss.Color
	ColorBgSecondary     lipgloss.Color
	ColorBgTertiary      lipgloss.Color
	ColorBgSelected      lipgloss.Color
	ColorTextPrimary     lipgloss.Color
	ColorTextSecondary   lipgloss.Color
	ColorTextDim         lipgloss.Color
	ColorBorder          lipgloss.Color
	ColorBorderSecondary lipgloss.Color
	ColorBorderDim       lipgloss.Color
)

func init() {
	initializeStyles()
}

func initializeStyles() {
	// Update colors from active scheme
	ColorPrimary = activeScheme.Primary
	ColorSecondary = activeScheme.Secondary
	ColorSuccess = activeScheme.Success
	ColorWarning = activeScheme.Warning
	ColorError = activeScheme.Error
	ColorInfo = activeScheme.Info
	ColorBgPrimary = activeScheme.BgPrimary
	ColorBgSecondary = activeScheme.BgSecondary
	ColorBgTertiary = activeScheme.BgTertiary
	ColorBgSelected = activeScheme.BgSelected
	ColorTextPrimary = activeScheme.TextPrimary
	ColorTextSecondary = activeScheme.TextSecondary
	ColorTextDim = activeScheme.TextDim
	ColorBorder = activeScheme.Border
	ColorBorderSecondary = activeScheme.BorderSecondary
	ColorBorderDim = activeScheme.BorderDim

	// Reinitialize all styles with new colors
	updateAllStyles()
}

func updateAllStyles() {
	// Update all styles with new colors
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1).
			MarginTop(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true)

	// Box and container styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			MarginBottom(1)

	ActiveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ColorBorderSecondary).
			Padding(1, 2).
			MarginBottom(1).
			Bold(true)

	// List item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Background(ColorBgSelected).
				Bold(true).
				Padding(0, 1)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorTextSecondary).
				Padding(0, 1)

	CheckedItemStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true).
				Padding(0, 1)

	// Status styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Background(ColorBgSecondary).
			Padding(0, 1)

	StatusSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StatusErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	StatusInfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorBgTertiary).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)

	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Background(ColorBgTertiary).
				Padding(0, 1).
				Border(lipgloss.ThickBorder()).
				BorderForeground(ColorBorderSecondary)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	// Category styles
	CategoryStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			Italic(true)

	// Counter/badge styles
	BadgeStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorPrimary).
			Padding(0, 1).
			Bold(true).
			MarginLeft(1)

	// Spinner/loading styles
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextPrimary).
			Background(ColorBgTertiary).
			Padding(0, 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorderSecondary).
			BorderBottom(false)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorTextSecondary).
				Background(ColorBgSecondary).
				Padding(0, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorBorderDim).
				BorderBottom(false)

	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorBorderDim)
}

// Style variables
var (
	TitleStyle          lipgloss.Style
	SubtitleStyle       lipgloss.Style
	BoxStyle            lipgloss.Style
	ActiveBoxStyle      lipgloss.Style
	SelectedItemStyle   lipgloss.Style
	UnselectedItemStyle lipgloss.Style
	CheckedItemStyle    lipgloss.Style
	StatusBarStyle      lipgloss.Style
	StatusSuccessStyle  lipgloss.Style
	StatusErrorStyle    lipgloss.Style
	StatusInfoStyle     lipgloss.Style
	InputStyle          lipgloss.Style
	InputFocusedStyle   lipgloss.Style
	HelpStyle           lipgloss.Style
	HelpKeyStyle        lipgloss.Style
	CategoryStyle       lipgloss.Style
	BadgeStyle          lipgloss.Style
	SpinnerStyle        lipgloss.Style
	ActiveTabStyle      lipgloss.Style
	InactiveTabStyle    lipgloss.Style
	ProgressBarStyle    lipgloss.Style
	ProgressEmptyStyle  lipgloss.Style
)

// ASCII Art for banner
const BannerArt = `
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║   ░█████╗░░█████╗░████████╗                                  ║
║   ██╔══██╗██╔══██╗╚══██╔══╝                                  ║
║   ██║░░╚═╝██║░░╚═╝░░░██║░░░                                  ║
║   ██║░░██╗██║░░██╗░░░██║░░░                                  ║
║   ╚█████╔╝╚█████╔╝░░░██║░░░                                  ║
║   ░╚════╝░░╚════╝░░░░╚═╝░░░                                  ║
║                                                               ║
║        Claude Code Templates - Interactive Installer         ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
`

// GetBannerStyled returns styled banner
func GetBannerStyled() string {
	banner := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Render(BannerArt)

	subtitle := lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Italic(true).
		Align(lipgloss.Center).
		Width(65).
		Render("Browse, search, and install components with ease")

	return banner + "\n" + subtitle + "\n"
}

// GetCurrentThemeIndex returns the index of the current theme
func GetCurrentThemeIndex() int {
	theme := os.Getenv("CCT_THEME")
	switch theme {
	case "neon", "green":
		return 1
	case "cyan", "blue":
		return 2
	case "purple", "magenta":
		return 3
	case "orange":
		return 0
	default:
		return 0
	}
}

// ApplyThemeByIndex applies a theme by its index
func ApplyThemeByIndex(index int) {
	var scheme ColorScheme
	switch index {
	case 0:
		scheme = OrangeScheme
	case 1:
		scheme = NeonGreenScheme
	case 2:
		scheme = CyanScheme
	case 3:
		scheme = PurpleScheme
	default:
		scheme = OrangeScheme
	}
	SetColorScheme(scheme)
}

// GetThemeName returns the name of a theme by its index
func GetThemeName(index int) string {
	names := []string{"Orange", "Neon Green", "Cyan", "Purple"}
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return "Orange"
}
