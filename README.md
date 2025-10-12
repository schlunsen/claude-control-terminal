# Claude Control Terminal (CCT)

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://github.com/schlunsen/claude-control-terminal/workflows/Build%20and%20Release/badge.svg)](https://github.com/schlunsen/claude-control-terminal/actions)
[![Release](https://img.shields.io/github/v/release/schlunsen/claude-control-terminal)](https://github.com/schlunsen/claude-control-terminal/releases)

**A powerful wrapper and control center for Claude Code** - Manage components, launch Claude, run analytics, and deploy with Docker.

Rebranded from `go-claude-templates` to better reflect its role as a comprehensive control terminal for Claude Code environments.

**Performance**: 50x faster startup, 5x lower memory usage, single 15MB binary with no dependencies.

<p align="center">
  <img src="docs/images/cct-tui-main.png" alt="CCT TUI Main Screen" width="600">
</p>

<p align="center">
  <img src="docs/images/cct-tui-agents.png" alt="Browse Agents" width="600">
  <img src="docs/images/cct-tui-mcps.png" alt="Browse MCPs" width="600">
</p>

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Component Installation](#component-installation)
- [Docker Support](#docker-support)
- [Analytics Dashboard](#analytics-dashboard)
- [Documentation](#documentation)
- [Development](#development)
- [License](#license)

## Features

- 🎮 **Control Center**: Comprehensive wrapper for managing Claude Code environments
- 🚀 **Interactive TUI**: Modern terminal interface for browsing and installing components
- 📦 **Component Management**: Install agents, commands, and MCPs from 600+ templates
- 🐳 **Docker Support**: Containerize Claude environments with one command
- 📊 **Analytics Dashboard**: Real-time WebSocket-based monitoring with process correlation
- ⚡ **High Performance**: 50-100x faster than Node.js version, 5x lower memory
- 🔧 **Zero Dependencies**: Single 15MB self-contained binary
- 🌐 **Cross-Platform**: Linux, macOS, Windows (amd64/arm64)

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
curl -L https://github.com/schlunsen/claude-control-terminal/releases/latest/download/cct-<platform>-<arch> -o cct
chmod +x cct
sudo mv cct /usr/local/bin/
```

### Option 3: Install with Go

```bash
go install github.com/schlunsen/claude-control-terminal/cmd/cct@latest
```

### Option 4: Build from Source

```bash
git clone https://github.com/schlunsen/claude-control-terminal
cd claude-control-terminal
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
- Installation status indicators ([G]=Global, [P]=Project)
- Modern, hip terminal aesthetic
- Keyboard-driven interface
- Launch Claude CLI directly from TUI

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

## Docker Support

CCT provides comprehensive Docker support for containerizing Claude Code environments.

### Quick Start with Docker

```bash
# Generate Dockerfile and .dockerignore
cct --docker-init --docker-type claude

# Build Docker image
cct --docker-build

# Run containerized Claude environment
cct --docker-run

# View logs
cct --docker-logs

# Stop container
cct --docker-stop
```

### Docker Compose

Generate multi-service setups with docker-compose:

```bash
# Generate docker-compose.yml for Claude + Analytics
cct --docker-compose --docker-type analytics

# Generate full stack (Claude + Analytics + PostgreSQL + Redis)
cct --docker-compose --docker-type full

# Start services
docker-compose up -d
```

### Dockerfile Types

CCT can generate 4 types of Dockerfiles:

| Type | Description | Use Case |
|------|-------------|----------|
| `base` | Minimal CCT-only image | Lightweight component management |
| `claude` | Full Claude environment (Node.js + Claude CLI + CCT) | Complete development setup |
| `analytics` | Optimized for analytics dashboard | Monitoring and metrics |
| `full` | Complete dev environment with all tools | Production-ready setup |

### Docker Compose Templates

| Template | Services | Use Case |
|----------|----------|----------|
| `simple` | Claude + CCT | Basic containerized development |
| `analytics` | Claude + Analytics dashboard | Real-time monitoring |
| `database` | Claude + PostgreSQL | Database-backed projects |
| `full` | Claude + Analytics + PostgreSQL + Redis | Complete production stack |

### Include MCPs in Containers

```bash
# Generate Dockerfile with specific MCPs
cct --docker-init --docker-type claude --docker-mcps "postgresql,github,supabase"

# Generate docker-compose with MCPs
cct --docker-compose --docker-type full --docker-mcps "postgresql,github"
```

### Docker Examples

**Example 1: Simple Development Container**
```bash
# Initialize Docker files
cct --docker-init --docker-type claude

# Build and run
cct --docker-build
cct --docker-run --docker-command "claude"
```

**Example 2: Analytics Dashboard in Docker**
```bash
# Generate analytics-optimized setup
cct --docker-init --docker-type analytics
cct --docker-build
cct --docker-run --docker-command "cct --analytics"

# Access dashboard at http://localhost:3333
```

**Example 3: Full Stack with Docker Compose**
```bash
# Generate complete stack
cct --docker-compose --docker-type full --docker-mcps "postgresql,redis"

# Configure environment
cp .env.example .env
# Edit .env with your API keys

# Start all services
docker-compose up -d

# Access Claude: docker-compose exec claude claude
# Access Analytics: http://localhost:3333
```

## Analytics Dashboard

Real-time monitoring of Claude Code conversations with WebSocket live updates.

```bash
cct --analytics
# Dashboard available at http://localhost:3333
```

<p align="center">
  <img src="docs/images/cct-analytics.png" alt="CCT Analytics Dashboard" width="800">
</p>

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
- `POST /api/reset/soft` - Soft reset with delta tracking (recommended)
- `POST /api/reset/archive` - Archive all conversations (preserves data)
- `POST /api/reset/clear` - Permanently delete all conversations (use with caution!)
- `DELETE /api/reset` - Clear soft reset and restore original counts
- `GET /api/reset/status` - Get current reset status
- `GET /ws` - WebSocket connection for real-time updates

### Resetting Analytics Counts

You can reset the analytics counts in three ways:

**Option 1: Soft Reset (Recommended) - Delta-Based** 🔄
```bash
# Resets counts to zero without deleting any data (reversible)
curl -X POST http://localhost:3333/api/reset/soft
```
This applies a delta to make counts appear as if you're starting from zero, while preserving all conversation data. Perfect for tracking usage from a specific date. You can undo this anytime:
```bash
# Restore original counts
curl -X DELETE http://localhost:3333/api/reset
```

**Option 2: Archive** 📦
```bash
# Archives all conversations to timestamped folder, preserving data
curl -X POST http://localhost:3333/api/reset/archive
```
This moves all `.jsonl` files to `~/.claude/archive/YYYY-MM-DD_HH-MM-SS/`, allowing you to recover them later.

**Option 3: Clear (Permanent)** ⚠️
```bash
# Permanently deletes all conversation files
curl -X POST http://localhost:3333/api/reset/clear
```
This permanently removes all `.jsonl` files (cannot be undone!).

**Using the UI**
The analytics dashboard includes intuitive reset buttons for all three options. Just click the "🔄 Soft Reset" button (recommended) or choose another option. When a soft reset is active, you'll see a yellow banner with reset details and a "Restore Original Counts" button.

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
- **Docker**: Native Docker SDK integration
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

## Migration from go-claude-templates

If you're upgrading from the previous `go-claude-templates`:

**Module Path Changed:**
- Old: `github.com/davila7/go-claude-templates`
- New: `github.com/schlunsen/claude-control-terminal`

**What's New in v0.2.0:**
- Rebranded to Claude Control Terminal (CCT)
- Full Docker support with 4 Dockerfile types
- Docker Compose templates for multi-service setups
- MCP integration in containers
- Enhanced TUI with installation status indicators
- Improved positioning as control center for Claude Code

**Breaking Changes:**
- Import paths must be updated if using CCT as a library
- Binary name remains `cct` (no change needed for CLI users)

## License

MIT License - See [LICENSE](LICENSE) file for details.

Based on [claude-code-templates](https://github.com/davila7/claude-code-templates) by davila7.
