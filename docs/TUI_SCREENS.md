# TUI Screen Flow Documentation

This document describes the visual flow and structure of each screen in the TUI.

## Screen Flow Diagram

```
┌─────────────────┐
│                 │
│   Main Screen   │◄──────────────────────────┐
│                 │                           │
└────────┬────────┘                           │
         │ Select Type                        │
         ▼                                    │
┌─────────────────┐                           │
│                 │                           │
│ Component List  │                           │
│                 │                           │
└────────┬────────┘                           │
         │ Select Components                  │
         ▼                                    │
┌─────────────────┐                           │
│                 │                           │
│ Confirm Screen  │                           │
│                 │                           │
└────────┬────────┘                           │
         │ Confirm                            │
         ▼                                    │
┌─────────────────┐                           │
│                 │                           │
│   Installing    │                           │
│                 │                           │
└────────┬────────┘                           │
         │ Complete                           │
         ▼                                    │
┌─────────────────┐                           │
│                 │                           │
│ Complete Screen │                           │
│                 │                           │
└────────┬────────┘                           │
         │ Return (R)                         │
         └────────────────────────────────────┘
```

## Screen 1: Main Screen (Component Type Selection)

### Visual Layout
```
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

   Browse, search, and install components with ease

┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  Select Component Type                                        │
│                                                               │
│  > 🤖 Agents      ← Selected (highlighted)                   │
│    ⚡ Commands                                                │
│    🔌 MCPs                                                    │
│                                                               │
│  ↑/↓: Navigate • Enter: Select • Q/Esc: Quit                │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Features
- ASCII art banner with gradient orange colors
- Three component types with icons
- Cursor indicator (">") for selected item
- Visual highlighting of current selection
- Help text at bottom

### Navigation
- `↑/↓` or `k/j` - Move selection
- `Enter` - Select type and proceed
- `Q` or `Esc` - Quit application

## Screen 2: Component List Screen

### Visual Layout
```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  🤖 Browse Agents                                            │
│                                                               │
│  ┌─────────────────────────────────────────┐                │
│  │ Press / to search                       │                │
│  └─────────────────────────────────────────┘                │
│                                                               │
│  > ☐ api-security-audit (security)       ← Cursor here      │
│    ☑ code-reviewer (development-tools)   ← Selected         │
│    ☐ data-analyst (data-ai)                                 │
│    ☐ debug-assistant (development-tools)                    │
│    ☐ devops-engineer (devops-infrastructure)                │
│    ☐ documentation-writer (documentation)                   │
│    ☐ frontend-specialist (development-team)                 │
│    ☐ git-workflow-manager (git)                             │
│    ☐ performance-optimizer (performance-testing)            │
│    ☐ security-auditor (security)                            │
│    ☐ test-automation-expert (testing)                       │
│    ☐ ui-ux-designer (web-tools)                             │
│                                                               │
│  ┌─────────────────────────────────────────┐                │
│  │ Selected: 1/250                         │                │
│  └─────────────────────────────────────────┘                │
│                                                               │
│  ↑/↓: Navigate • Space: Toggle • A: Select All • a: Deselect│
│  /: Search • Enter: Install • Esc: Back                     │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### With Search Active
```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  🤖 Browse Agents                                            │
│                                                               │
│  ┌─────────────────────────────────────────┐                │
│  │ Search: security_                       │ ← Active input │
│  └─────────────────────────────────────────┘                │
│                                                               │
│  > ☐ api-security-audit (security)                          │
│    ☐ security-auditor (security)                            │
│                                                               │
│  ┌─────────────────────────────────────────┐                │
│  │ Selected: 0/2 (filtered)                │                │
│  └─────────────────────────────────────────┘                │
│                                                               │
│  ↑/↓: Navigate • Esc: Exit Search • Enter: Accept          │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Features
- Dynamic loading indicator (spinner) on first load
- Search input field (activated with `/`)
- Scrollable list with visible window
- Checkboxes for selection state
- Category labels in parentheses
- Selection counter
- Multi-line help text

### Navigation
- `↑/↓` or `k/j` - Move cursor
- `Space` - Toggle selection
- `/` - Activate search
- `a` - Select all visible
- `A` - Deselect all
- `Enter` - Proceed to confirmation
- `Esc` - Go back to main screen

## Screen 3: Confirm Screen

### Visual Layout
```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  Confirm Installation                                         │
│                                                               │
│  You are about to install 3 component(s):                   │
│                                                               │
│    🤖 code-reviewer (development-tools)                      │
│    🤖 security-auditor (security)                            │
│    🤖 api-security-audit (security)                          │
│                                                               │
│  Target directory: /Users/username/project                   │
│                                                               │
│  Y/Enter: Install • N/Esc: Cancel                           │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Features
- Component count summary
- Full list of selected components with icons and categories
- Target directory display
- Simple yes/no prompt

### Navigation
- `Y` or `Enter` - Confirm and start installation
- `N` or `Esc` - Cancel and return to component list

## Screen 4: Installing Screen

### Visual Layout
```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  Installing Components                                        │
│                                                               │
│  ⠋ Installing components, please wait...                     │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Features
- Animated spinner (rotating dots)
- Simple loading message
- Non-interactive (no keyboard input)

### Behavior
- Automatically transitions to Complete screen when done
- Runs installations in parallel

## Screen 5: Complete Screen

### Success Case
```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  Installation Complete!                                       │
│                                                               │
│  Successfully installed 3 component(s):                      │
│    ✓ code-reviewer                                           │
│    ✓ security-auditor                                        │
│    ✓ api-security-audit                                      │
│                                                               │
│  R: Return to Main • Q/Enter: Quit                          │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Partial Success Case
```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  Partial Installation Complete                                │
│                                                               │
│  Successfully installed 2 component(s):                      │
│    ✓ code-reviewer                                           │
│    ✓ security-auditor                                        │
│                                                               │
│  Failed to install 1 component(s):                           │
│    ✗ api-security-audit                                      │
│                                                               │
│  R: Return to Main • Q/Enter: Quit                          │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Error Case
```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│  Installation Error                                           │
│                                                               │
│  Failed to install 3 component(s):                           │
│    ✗ code-reviewer                                           │
│    ✗ security-auditor                                        │
│    ✗ api-security-audit                                      │
│                                                               │
│  Error: Network connection failed                            │
│                                                               │
│  R: Return to Main • Q/Enter: Quit                          │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Features
- Clear success/failure indication
- Separate lists for successful and failed installations
- Error message if applicable
- Options to retry or exit

### Navigation
- `R` - Return to main screen (start over)
- `Q` or `Enter` - Quit application

## Color Scheme

### Primary Colors
- **Orange (#FF6B35)**: Primary accent, borders, highlights
- **Light Orange (#F7931E)**: Secondary accent, titles
- **Cyan (#4ECDC4)**: Success states, active elements
- **Yellow (#FFE66D)**: Warnings
- **Red (#FF6B6B)**: Errors

### Background Colors
- **Dark Blue-Black (#1A1A2E)**: Primary background
- **Lighter Dark (#16213E)**: Secondary background
- **Accent Blue (#0F3460)**: Input fields, tertiary background
- **Pink-Red (#E94560)**: Selection highlight

### Text Colors
- **Almost White (#EAEAEA)**: Primary text
- **Gray (#A0A0A0)**: Secondary text
- **Dark Gray (#606060)**: Dim text, help text

## Responsive Behavior

### Minimum Terminal Size
- Width: 80 characters
- Height: 24 lines

### Adaptive Features
- Component list adjusts visible window based on terminal height
- Help text wraps on narrow terminals
- Banner scales with terminal width

## Visual Feedback

### Selection States
- **Unselected**: Gray text with empty checkbox (☐)
- **Selected**: Green text with filled checkbox (☑)
- **Cursor**: Orange background highlight with ">" prefix

### Loading States
- **Loading**: Animated spinner with orange color
- **Error**: Red text with error icon
- **Success**: Cyan/green text with checkmark

### Interactive States
- **Normal**: Standard color scheme
- **Focused**: Bright colors, thick borders
- **Hover**: (not applicable, keyboard-only)
- **Active**: Bold text, accent colors

## Accessibility Features

1. **Keyboard-Only**: No mouse required
2. **High Contrast**: Clear color differentiation
3. **Clear Labels**: Descriptive text for all actions
4. **Status Indicators**: Visual feedback for all states
5. **Help Text**: Always visible on every screen

## Animation Details

### Spinner Animation
- Type: Dot spinner (⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)
- Speed: 10 frames per second
- Color: Orange (#FF6B35)

### Transitions
- Screen changes: Instant (no animation)
- Text input: Cursor blink (standard terminal)
- Selection: Immediate color change

## Error Handling

### Network Errors
- Display clear error message
- Allow return to main screen
- Preserve user selections where possible

### API Rate Limiting
- Show user-friendly message
- Suggest retry after delay
- Don't crash the application

### Component Not Found
- List specific components that failed
- Continue with successful installations
- Show partial success screen
