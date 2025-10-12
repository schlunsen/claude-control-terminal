# TUI Developer Guide

A technical guide for developers working on or extending the TUI implementation.

## Architecture Overview

### Bubble Tea Pattern (Elm Architecture)

The TUI follows Bubble Tea's Model-View-Update pattern:

```go
type Model struct {
    // State
    screen Screen
    components []ComponentItem
    cursor int
    // ...
}

func (m Model) Init() tea.Cmd {
    // Initialize (return initial commands)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages, update state, return new commands
}

func (m Model) View() string {
    // Render UI based on current state
}
```

### State Machine

The TUI operates as a state machine with these states:

```go
const (
    ScreenMain           // Component type selection
    ScreenComponentList  // Browse/search components
    ScreenConfirm        // Confirm installation
    ScreenInstalling     // Installation progress
    ScreenComplete       // Results display
)
```

## Project Structure

```
internal/tui/
├── component_item.go    # Data models
├── loader.go           # GitHub API integration
├── styles.go           # Visual styling
├── model.go            # Main Bubble Tea model
├── installer.go        # Component installation
└── tui.go              # Entry point
```

## Key Components

### 1. Component Data Model

```go
type ComponentItem struct {
    Name        string
    Category    string
    Description string
    Type        string // "agent", "command", "mcp"
    Selected    bool
}
```

### 2. Component Metadata

```go
type ComponentMetadata struct {
    Type       string
    Icon       string
    Path       string
    Extension  string
    Categories []string
}
```

Located in `component_item.go`, provides configuration for each component type.

### 3. Component Loader

```go
type ComponentLoader struct {
    config *fileops.GitHubConfig
}

func (cl *ComponentLoader) LoadComponents(componentType string) ([]ComponentItem, error)
```

Fetches component lists from GitHub API asynchronously.

### 4. Style System

All styles are defined as `lipgloss.Style` constants in `styles.go`:

```go
var TitleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(ColorPrimary).
    MarginBottom(1)
```

## Message Flow

### Message Types

```go
// Keyboard input
tea.KeyMsg

// Window resize
tea.WindowSizeMsg

// Component loading complete
componentsLoadedMsg struct {
    components []ComponentItem
    err        error
}

// Installation complete
installCompleteMsg struct {
    success []string
    failed  []string
    err     error
}

// Spinner tick
spinner.TickMsg
```

### Async Operations

Commands return functions that produce messages:

```go
func loadComponentsCmd(componentType string) tea.Cmd {
    return func() tea.Msg {
        loader := NewComponentLoader()
        components, err := loader.LoadComponents(componentType)
        return componentsLoadedMsg{
            components: components,
            err:        err,
        }
    }
}
```

## Screen Implementation

### Screen Template

Each screen follows this pattern:

```go
// 1. Handle keyboard input
func (m Model) handleScreenName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        // Action
        return m, someCmd()
    case "esc":
        // Go back
        m.screen = PreviousScreen
        return m, nil
    }
    return m, nil
}

// 2. Render view
func (m Model) viewScreenName() string {
    var b strings.Builder

    // Title
    b.WriteString(TitleStyle.Render("Screen Title") + "\n\n")

    // Content
    // ...

    // Help
    b.WriteString(HelpStyle.Render("Keys: ..."))

    return BoxStyle.Render(b.String())
}
```

## Adding a New Screen

### Step 1: Define Screen Constant

```go
const (
    // ...existing screens...
    ScreenNewScreen Screen = iota + 5
)
```

### Step 2: Add Handler

```go
func (m Model) handleNewScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        m.screen = NextScreen
        return m, nil
    case "esc":
        m.screen = PreviousScreen
        return m, nil
    }
    return m, nil
}
```

### Step 3: Add View

```go
func (m Model) viewNewScreen() string {
    var b strings.Builder
    b.WriteString(TitleStyle.Render("New Screen") + "\n\n")
    // Add content
    return BoxStyle.Render(b.String())
}
```

### Step 4: Wire Up in Main Update/View

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ...
    switch m.screen {
    // ...
    case ScreenNewScreen:
        return m.handleNewScreen(msg)
    }
}

func (m Model) View() string {
    switch m.screen {
    // ...
    case ScreenNewScreen:
        return m.viewNewScreen()
    }
}
```

## Styling Guide

### Creating New Styles

```go
var MyStyle = lipgloss.NewStyle().
    Foreground(ColorPrimary).
    Background(ColorBgSecondary).
    Bold(true).
    Padding(1, 2).
    Border(lipgloss.RoundedBorder()).
    BorderForeground(ColorBorder)
```

### Common Style Patterns

#### Boxes
```go
BoxStyle.Render(content)
ActiveBoxStyle.Render(content)
```

#### Lists
```go
if selected {
    line = SelectedItemStyle.Render(text)
} else {
    line = UnselectedItemStyle.Render(text)
}
```

#### Status
```go
StatusSuccessStyle.Render("Success!")
StatusErrorStyle.Render("Error!")
```

## GitHub API Integration

### Loading Components

```go
loader := NewComponentLoader()
components, err := loader.LoadComponents("agent")
```

### API Endpoints

```go
// List directory contents
apiURL := fmt.Sprintf(
    "https://api.github.com/repos/%s/%s/contents/%s/%s?ref=%s",
    owner, repo, templatesPath, category, branch
)
```

### Rate Limiting

The loader handles rate limiting gracefully:
- Continues if a category fails
- Returns partial results
- No crashes on API errors

## Installation System

### Installer Interface

```go
type AgentInstallerForTUI struct {
    config *fileops.GitHubConfig
}

func (ai *AgentInstallerForTUI) InstallAgent(
    agentName, category, targetDir string
) error
```

### Installation Flow

1. Download component content from GitHub
2. Create target directory (`.claude/agents/`)
3. Write file to disk
4. Return error or nil

### Parallel Installation

```go
for _, comp := range components {
    err := installer.Install(comp.Name, comp.Category, targetDir)
    if err != nil {
        failed = append(failed, comp.Name)
    } else {
        success = append(success, comp.Name)
    }
}
```

## Testing Strategy

### Manual Testing

```bash
# Test TUI launch
go run ./cmd/cct

# Test with custom directory
go run ./cmd/cct -d /tmp/test

# Test component installation
# (use TUI interactively)
```

### Unit Testing (Future)

```go
func TestComponentLoader_LoadComponents(t *testing.T) {
    loader := NewComponentLoader()
    components, err := loader.LoadComponents("agent")

    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if len(components) == 0 {
        t.Error("expected components, got empty list")
    }
}
```

### Integration Testing (Future)

```go
func TestTUI_CompleteFlow(t *testing.T) {
    // Create model
    m := NewModel("/tmp/test")

    // Simulate user input
    // Test state transitions
    // Verify installation
}
```

## Debugging

### Enable Verbose Output

The TUI runs in alt-screen mode, so debugging requires:

1. **Log to File**:
```go
f, _ := os.OpenFile("/tmp/tui-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
defer f.Close()
fmt.Fprintf(f, "Debug: %+v\n", someVar)
```

2. **Use Bubble Tea Debug Mode**:
```go
p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
```

3. **Test Without TUI**:
```bash
# Test component loading separately
go run ./cmd/cct --agent test-agent
```

### Common Issues

#### Screen Doesn't Refresh
- Ensure `Update()` returns a command for continuous updates
- Check spinner is ticking: `m.spinner.Tick`

#### Keyboard Input Not Working
- Verify key handling in `handleKeyPress()`
- Check screen-specific handlers
- Ensure not blocking on input

#### Styles Not Applying
- Check terminal supports ANSI colors
- Verify style definitions in `styles.go`
- Test with different terminals

## Performance Optimization

### Component Loading

```go
// Load categories in parallel (future enhancement)
var wg sync.WaitGroup
results := make(chan []ComponentItem)

for _, category := range categories {
    wg.Add(1)
    go func(cat string) {
        defer wg.Done()
        items := loadCategory(cat)
        results <- items
    }(category)
}
```

### Rendering

```go
// Only render visible items
start := m.cursor - 5
end := start + 15

for i := start; i < end; i++ {
    // Render item
}
```

### Memory Management

- Clear component cache after installation
- Limit component list size if needed
- Use pointers for large structs

## Extending the TUI

### Add Component Preview

```go
// In component_item.go
type ComponentItem struct {
    // ...existing fields...
    Content string // Full component content
}

// In model.go
case "p": // Preview key
    if len(m.components) > 0 {
        m.screen = ScreenPreview
        m.previewContent = m.components[m.cursor].Content
    }
```

### Add Filtering by Category

```go
// In model.go
func (m *Model) filterByCategory(category string) {
    m.filteredIndices = nil
    for i, comp := range m.components {
        if comp.Category == category {
            m.filteredIndices = append(m.filteredIndices, i)
        }
    }
}

// Add key handler
case "c": // Category filter
    // Show category selection
```

### Add Installation History

```go
// Create new file: internal/tui/history.go
type InstallHistory struct {
    Timestamp  time.Time
    Components []string
}

func (h *InstallHistory) Save() error {
    // Save to ~/.cct-history.json
}

func LoadHistory() (*InstallHistory, error) {
    // Load from ~/.cct-history.json
}
```

## Best Practices

### 1. State Management
- Keep state in Model struct
- Avoid global variables
- Use message passing for async operations

### 2. Error Handling
- Always check errors
- Display user-friendly messages
- Allow recovery (don't crash)

### 3. UI/UX
- Provide clear feedback
- Show loading indicators
- Display help text
- Use consistent styling

### 4. Code Organization
- One screen per function
- Separate concerns (model, view, update)
- Keep files focused (< 500 lines)

### 5. Performance
- Lazy load when possible
- Limit list rendering
- Avoid unnecessary re-renders

## Resources

### Bubble Tea Documentation
- [GitHub](https://github.com/charmbracelet/bubbletea)
- [Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)

### Lipgloss Documentation
- [GitHub](https://github.com/charmbracelet/lipgloss)
- [Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)

### Bubbles Components
- [GitHub](https://github.com/charmbracelet/bubbles)
- [Components](https://github.com/charmbracelet/bubbles/tree/master/examples)

## Troubleshooting

### Import Issues
```bash
go mod tidy
go mod download
```

### Build Issues
```bash
go clean -cache
go build ./cmd/cct
```

### Runtime Issues
```bash
# Check Go version
go version  # Should be 1.23+

# Check dependencies
go list -m all

# Verify imports
go mod verify
```

## Contributing

When contributing to the TUI:

1. Follow existing code structure
2. Add tests for new features
3. Update documentation
4. Test on multiple terminals
5. Ensure backward compatibility
6. Follow Go style guidelines

## Contact & Support

For questions or issues:
- Open a GitHub issue
- Check existing documentation
- Review Bubble Tea examples
- Test in different environments
