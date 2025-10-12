# Go Claude Templates (CCT)

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/schlunsen/claude-templates-go/workflows/Build%20and%20Release/badge.svg)](https://github.com/schlunsen/claude-templates-go/actions)
[![Release](https://img.shields.io/github/v/release/schlunsen/claude-templates-go)](https://github.com/schlunsen/claude-templates-go/releases)

A high-performance Go port of [claude-code-templates](https://github.com/davila7/claude-code-templates) providing component templates, analytics dashboards, and real-time monitoring for Claude Code projects.

**Performance**: 50x faster startup, 5x lower memory usage, single 15MB binary with no dependencies.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Component Installation](#component-installation)
- [Analytics Dashboard](#analytics-dashboard)
- [Documentation](#documentation)
- [Development](#development)
- [License](#license)

## Features

- **Interactive TUI**: Modern terminal interface for browsing and installing components with search, multi-select, and real-time loading
- **Component Management**: Install agents, commands, and MCPs from 600+ templates with automatic category search across 50+ categories
- **Analytics Dashboard**: Real-time WebSocket-based monitoring of Claude Code conversations with state detection and process correlation
- **Cross-Platform**: Single binary for Linux, macOS, and Windows (amd64/arm64)
- **High Performance**: 50-100x faster build time, 50x faster startup, 5x lower memory vs Node.js version
- **Zero Dependencies**: 15MB self-contained binary vs 50MB+ node_modules

## Installation

### Option 1: Homebrew (macOS/Linux)

```bash
# Tap and install
brew tap schlunsen/cct
brew install cct

# Or one-line installation
brew install schlunsen/cct/cct
```

### Option 2: Download Binary

```bash
# Download for your platform from releases
curl -L https://github.com/schlunsen/claude-templates-go/releases/latest/download/cct-<platform>-<arch> -o cct
chmod +x cct
sudo mv cct /usr/local/bin/
```

### Option 3: Install with Go

```bash
go install github.com/schlunsen/claude-templates-go/cmd/cct@latest
```

### Option 4: Build from Source

```bash
git clone https://github.com/schlunsen/claude-templates-go
cd go-claude-templates
make build  # or: just build
```

## Quick Start

### Interactive TUI (Recommended)

Launch the interactive Terminal User Interface to browse and install components:

```bash
# Run without arguments to launch TUI
cct

# Or with custom directory
cct -d ~/my-project
```

**Features**:
- Browse 600+ agents, 200+ commands, and MCPs
- Real-time search and filtering
- Multi-select with checkboxes
- Modern, hip terminal aesthetic
- Keyboard-driven interface

[View TUI Guide](docs/TUI_GUIDE.md) | [View Screenshots](docs/TUI_SCREENS.md)

### CLI Flags (For Automation)

```bash
# Install components with smart category search
cct --agent api-documenter
cct --agent prompt-engineer --command security-audit --mcp postgresql

# Multiple components at once
cct --agent "api-documenter,database-architect" \
    --command "security-audit,setup-linting" \
    --mcp "postgresql,supabase"

# Launch analytics dashboard
cct --analytics
# Open browser to http://localhost:3333

# Get help
cct --help
cct --version
```

## Component Installation

The CLI automatically searches through all component categories to find what you need. No need to specify full paths.

### Agents (600+ templates across 25+ categories)

```bash
cct --agent api-documenter        # Found in documentation/
cct --agent prompt-engineer       # Found in ai-specialists/
cct --agent database-architect    # Found in database/
cct --agent git-flow-manager      # Found in git/
```

**Categories**: ai-specialists, api-graphql, blockchain-web3, business-marketing, data-ai, database, development-tools, devops-infrastructure, documentation, git, mcp-dev-team, performance-testing, programming-languages, security, and more.

### Commands (200+ templates across 19+ categories)

```bash
cct --command security-audit      # Found in security/
cct --command setup-linting       # Found in setup/
cct --command deploy-production   # Found in deployment/
```

**Categories**: automation, database, deployment, documentation, git, performance, project-management, security, setup, testing, utilities, and more.

### MCPs (Server integrations across 9+ categories)

```bash
cct --mcp postgresql              # Found in database/
cct --mcp supabase                # Found in database/
cct --mcp github                  # Found in integration/
```

**Categories**: browser_automation, database, devtools, filesystem, integration, marketing, productivity, web.

### Custom Installation Directory

```bash
cct --agent api-documenter --directory ~/my-project
```

## Analytics Dashboard

Real-time monitoring of Claude Code conversations with WebSocket live updates.

```bash
cct --analytics
# Dashboard available at http://localhost:3333
```

### Features

- Real-time conversation monitoring with WebSocket updates
- State detection: "Claude Code working...", "Awaiting user input...", etc.
- Process correlation with running Claude Code instances
- System statistics: active conversations, total messages, state distribution
- Auto-refresh every 30 seconds plus instant WebSocket updates
- Responsive purple-themed gradient UI

### API Endpoints

- `GET /api/health` - Health check
- `GET /api/data` - Complete conversation data
- `GET /api/conversations` - Conversation list with metadata
- `GET /api/processes` - Running Claude Code processes
- `GET /api/stats` - System statistics and metrics
- `POST /api/refresh` - Force data refresh
- `GET /ws` - WebSocket connection for real-time updates

## Documentation

- [TUI User Guide (docs/TUI_GUIDE.md)](docs/TUI_GUIDE.md) - Interactive interface guide and keyboard shortcuts
- [TUI Screen Flow (docs/TUI_SCREENS.md)](docs/TUI_SCREENS.md) - Visual documentation of all TUI screens
- [TUI Developer Guide (docs/TUI_DEVELOPER_GUIDE.md)](docs/TUI_DEVELOPER_GUIDE.md) - Technical implementation details
- [Development Guide (CLAUDE.md)](CLAUDE.md) - Architecture, development workflow, and best practices
- [Testing Guide (TESTING.md)](TESTING.md) - Comprehensive testing instructions and examples
- [Changelog (CHANGELOG.md)](CHANGELOG.md) - Version history and migration details
- [Contributing (CONTRIBUTING.md)](CONTRIBUTING.md) - How to contribute to the project

## Development

### Build Commands

```bash
# Build for current platform
make build      # or: just build

# Build for all platforms (Linux, macOS, Windows)
make build-all  # or: just build-all

# Run directly
make run        # or: just run

# Run analytics dashboard
make run-analytics  # or: just analytics

# Clean build artifacts
make clean      # or: just clean
```

### Tech Stack

- **Language**: Go 1.23+ (using go1.24.8 toolchain)
- **CLI**: Cobra (commands), Pterm (terminal UI)
- **TUI**: Bubble Tea (interactive interface), Lipgloss (styling), Bubbles (components)
- **Web**: Fiber v2 (HTTP server), Gorilla WebSocket
- **System**: fsnotify (file watching), gopsutil (process detection)

### Testing

```bash
# Quick automated tests (7 tests)
./TEST_QUICK.sh

# Category search tests (9 tests)
./TEST_CATEGORIES.sh

# Manual testing
# See TESTING.md for comprehensive testing guide
```

For detailed development setup, architecture documentation, and contribution guidelines, see [CONTRIBUTING.md](CONTRIBUTING.md) and [CLAUDE.md](CLAUDE.md).

## License

MIT License - See [LICENSE](LICENSE) file for details.

Based on [claude-code-templates](https://github.com/davila7/claude-code-templates) by davila7.
