# TUI Quick Reference Card

## Launch TUI
```bash
./cct                    # Launch TUI
./cct -d ~/project       # Launch with custom directory
```

## Keyboard Shortcuts

### Universal (All Screens)
| Key | Action |
|-----|--------|
| `Q` | Quit application |
| `Esc` | Go back / Cancel |
| `Ctrl+C` | Force quit |

### Main Screen (Component Type Selection)
| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate up/down |
| `k` / `j` | Navigate up/down (vim-style) |
| `Enter` | Select component type |

### Component List Screen
| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate up/down |
| `k` / `j` | Navigate up/down (vim-style) |
| `Space` | Toggle selection (checkbox) |
| `/` | Activate search |
| `a` | Select all visible items |
| `A` | Deselect all items |
| `Enter` | Proceed to installation |

### Search Mode (Active)
| Key | Action |
|-----|--------|
| Type | Filter components |
| `Esc` | Exit search, clear filter |
| `Enter` | Accept search, continue browsing |

### Confirm Screen
| Key | Action |
|-----|--------|
| `Y` | Confirm installation |
| `Enter` | Confirm installation |
| `N` | Cancel, go back |
| `Esc` | Cancel, go back |

### Complete Screen (Results)
| Key | Action |
|-----|--------|
| `R` | Return to main menu |
| `Q` | Quit application |
| `Enter` | Quit application |

## Visual Indicators

### Selection States
- `☐` - Unselected item
- `☑` - Selected item
- `>` - Current cursor position
- Orange highlight - Current item
- Green text - Selected items

### Status Icons
- `🤖` - Agents
- `⚡` - Commands
- `🔌` - MCPs
- `✓` - Successfully installed
- `✗` - Failed to install
- `⠋` - Loading spinner

## Screen Flow

```
┌───────────┐
│   Main    │ ← Start here
└─────┬─────┘
      ↓ Select type
┌───────────┐
│ Component │ ← Browse, search, select
│   List    │
└─────┬─────┘
      ↓ Enter
┌───────────┐
│  Confirm  │ ← Review selections
└─────┬─────┘
      ↓ Y
┌───────────┐
│Installing │ ← Progress
└─────┬─────┘
      ↓ Complete
┌───────────┐
│ Complete  │ ← Results
└─────┬─────┘
      ↓ R (return) or Q (quit)

```

## Common Workflows

### Install Single Component
1. Launch: `./cct`
2. Select type (Agents/Commands/MCPs)
3. Navigate to desired component
4. Press `Space` to select
5. Press `Enter` to install
6. Press `Y` to confirm

### Install Multiple Components
1. Launch: `./cct`
2. Select type
3. Navigate and press `Space` for each item
4. Press `Enter` when done selecting
5. Press `Y` to confirm

### Search and Install
1. Launch: `./cct`
2. Select type
3. Press `/` to search
4. Type search term (e.g., "security")
5. Press `Esc` to exit search
6. Select items with `Space`
7. Press `Enter` then `Y`

### Select All with Filter
1. Launch: `./cct`
2. Select type
3. Press `/` to search
4. Type filter (e.g., "database")
5. Press `Esc` to exit search
6. Press `a` to select all filtered items
7. Press `Enter` then `Y`

### Browse Without Installing
1. Launch: `./cct`
2. Navigate through different types
3. Press `Esc` to go back
4. Press `Q` to quit

## Tips & Tricks

1. **Fast Navigation**: Use `k`/`j` for quick vim-style movement
2. **Quick Select All**: Use `/` to filter, then `a` to select all matches
3. **Multi-Category Install**: Return to main (press `R` on complete screen) to install from another category
4. **Check Installation**: Review the complete screen carefully before quitting
5. **Cancel Anytime**: Press `Esc` to go back at any point before confirmation

## Component Counts

- **Agents**: 600+ across 25 categories
- **Commands**: 200+ across 18 categories
- **MCPs**: Multiple across 9 categories

## Categories

### Agents (25)
ai-specialists, api-graphql, blockchain-web3, business-marketing, data-ai, database, deep-research-team, development-team, development-tools, devops-infrastructure, documentation, expert-advisors, ffmpeg-clip-team, game-development, git, mcp-dev-team, modernization, obsidian-ops-team, ocr-extraction-team, performance-testing, podcast-creator-team, programming-languages, realtime, security, web-tools

### Commands (18)
automation, database, deployment, documentation, game-development, git, git-workflow, nextjs-vercel, orchestration, performance, project-management, security, setup, simulation, svelte, sync, team, testing, utilities

### MCPs (9)
browser_automation, database, deepgraph, devtools, filesystem, integration, marketing, productivity, web

## Installation Locations

Components are installed to:
```
.claude/
├── agents/[name].md
├── commands/[name].md
└── mcp/[name].json
```

Default location: Current directory
Custom location: Use `-d` flag

## Troubleshooting

### TUI Won't Launch
- Check Go version: `go version` (need 1.23+)
- Update dependencies: `make deps`
- Rebuild: `make build`

### Components Don't Load
- Check internet connection
- Wait if GitHub rate limited
- Try again in a few minutes

### Installation Fails
- Check write permissions
- Verify disk space
- Review error on complete screen

### Terminal Display Issues
- Use modern terminal (iTerm2, Alacritty)
- Ensure ANSI color support
- Check terminal size (min 80x24)

## CLI Fallback

If TUI doesn't work, use CLI flags:

```bash
# Install specific components
./cct --agent security-auditor
./cct --command deploy
./cct --mcp postgresql

# Multiple at once
./cct --agent "security-auditor,code-reviewer"
```

## Need Help?

- Full Guide: [docs/TUI_GUIDE.md](docs/TUI_GUIDE.md)
- Visual Docs: [docs/TUI_SCREENS.md](docs/TUI_SCREENS.md)
- Developer Guide: [docs/TUI_DEVELOPER_GUIDE.md](docs/TUI_DEVELOPER_GUIDE.md)
- Main README: [README.md](README.md)

---

**Quick Reference Version**: 1.0
**Last Updated**: 2025-10-12
