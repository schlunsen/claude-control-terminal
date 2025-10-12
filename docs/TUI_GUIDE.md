# Interactive TUI Guide

The go-claude-templates CLI now includes a modern, interactive Terminal User Interface (TUI) that makes it easy to browse and install components.

## Quick Start

Simply run the CLI without any flags to launch the interactive TUI:

```bash
./cct
```

## Features

### 1. Component Type Selection
When you launch the TUI, you'll see a beautiful menu to select the type of component you want to install:
- **Agents** (🤖) - 600+ AI agents for various tasks
- **Commands** (⚡) - 200+ custom commands
- **MCPs** (🔌) - Model Context Protocol integrations

**Navigation:**
- `↑/↓` or `k/j` - Navigate between options
- `Enter` - Select component type
- `Q` or `Esc` - Quit

### 2. Component Browser
After selecting a component type, browse through available components dynamically loaded from GitHub.

**Features:**
- Real-time loading from GitHub repository
- Component categorization
- Multi-select support
- Search functionality

**Navigation:**
- `↑/↓` or `k/j` - Navigate through components
- `Space` - Toggle selection for current component
- `a` - Select all visible components
- `A` - Deselect all components
- `/` - Activate search mode
- `Enter` - Proceed to installation
- `Esc` - Go back to component type selection

### 3. Search
Press `/` to activate search mode:
- Type to filter components by name or category
- `Esc` - Exit search mode
- `Enter` - Accept search and continue browsing

### 4. Confirmation Screen
Review your selections before installation:
- See all selected components
- View target directory
- Confirm or cancel installation

**Navigation:**
- `Y` or `Enter` - Confirm and install
- `N` or `Esc` - Cancel and go back

### 5. Installation Progress
Watch as components are installed in real-time with a beautiful spinner animation.

### 6. Completion Screen
See installation results:
- Successfully installed components (✓)
- Failed installations (✗)
- Summary statistics

**Navigation:**
- `R` - Return to main menu
- `Q` or `Enter` - Exit TUI

## Visual Design

The TUI features a modern, hip aesthetic with:
- **Orange gradient** accent colors (#FF6B35)
- **Cyan highlights** for success states (#4ECDC4)
- **Dark terminal background** for reduced eye strain
- **Smooth animations** and transitions
- **ASCII art banner** on launch
- **Status indicators** with emojis

## Keyboard Shortcuts Reference

| Key | Action |
|-----|--------|
| `↑/↓` or `k/j` | Navigate up/down |
| `Space` | Toggle selection |
| `Enter` | Confirm/Continue |
| `Esc` | Go back/Cancel |
| `Q` | Quit |
| `/` | Search |
| `a` | Select all |
| `A` | Deselect all |
| `Y/N` | Yes/No (confirmation) |
| `R` | Return to main menu |

## Target Directory

By default, components are installed to the current directory (`.`). You can specify a different target directory:

```bash
./cct -d /path/to/target
```

Or:

```bash
./cct --directory /path/to/target
```

## Examples

### Install a single agent
1. Run `./cct`
2. Select "Agents"
3. Navigate to desired agent
4. Press `Space` to select
5. Press `Enter` to install

### Install multiple components with search
1. Run `./cct`
2. Select "Commands"
3. Press `/` to search
4. Type "security"
5. Use `Space` to select multiple security-related commands
6. Press `Enter` twice to confirm and install

### Browse and explore
1. Run `./cct`
2. Navigate through different component types
3. Use `Esc` to go back without installing
4. Press `Q` to quit

## Troubleshooting

### TUI doesn't launch
- Ensure your terminal supports ANSI colors
- Try a different terminal emulator (iTerm2, Alacritty recommended)
- Check that terminal size is adequate (minimum 80x24)

### Components don't load
- Check internet connection
- Verify GitHub API access (rate limits may apply)
- Try again in a few minutes if rate limited

### Installation fails
- Verify write permissions in target directory
- Check disk space
- Ensure `.claude` directory is writable

## Technical Details

### Built With
- **Bubble Tea** - TUI framework by Charm
- **Lipgloss** - Terminal styling
- **Bubbles** - TUI components (textinput, spinner)

### Component Sources
Components are loaded dynamically from:
- Repository: `davila7/claude-code-templates`
- Branch: `main`
- Path: `cli-tool/components/`

### Installation Structure
```
.claude/
├── agents/
│   └── [agent-name].md
├── commands/
│   └── [command-name].md
└── mcp/
    └── [mcp-name].json
```

## Tips & Best Practices

1. **Use search** - With 600+ agents and 200+ commands, search is your friend
2. **Select multiple** - Use `a` to select all filtered results when searching
3. **Explore categories** - Components are organized by category for easy browsing
4. **Check completion screen** - Review results before exiting to ensure success
5. **Return to main menu** - Press `R` on completion screen to install more components

## Fallback to CLI Flags

The original CLI flag interface is still available:

```bash
# Install agent
./cct --agent security-auditor

# Install multiple components
./cct --agent api-tester --command deploy --mcp postgresql

# Install with comma-separated lists
./cct --agent "security-auditor,code-reviewer"
```

The TUI is designed for interactive browsing, while flags are ideal for automation and scripts.
