# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.13] - 2025-10-13

### Added
- Multi-source permissions management with tabbed UI for better control over Claude Code permissions
- GitHub CLI (gh) commands now included in Git Commands permission category

### Improved
- Enhanced permissions management interface with tabbed navigation between different permission sources

## [0.2.12] - 2025-10-13

### Fixed
- Permissions management now uses local `.claude/settings.local.json` instead of global settings file for better project isolation
- Empty permissions object is now properly removed from settings when all permissions are disabled

## [0.2.11] - 2025-10-13

### Fixed
- Permissions management improvements

## [0.2.10] - 2025-10-13

### Added
- 'Launch last Claude session' menu option in TUI for quick access to recent conversations
- Comprehensive godoc package comments across all internal packages for better code documentation

### Fixed
- Search bar state persistence issue when navigating between screens in TUI

## [0.2.9] - 2025-10-13

### Added
- Memory limits to conversation parser to prevent excessive memory usage
- HTTP timeout protection to GitHub downloads for improved reliability
- Graceful shutdown support to analytics server

### Fixed
- WebSocket hub deadlock issue with proper graceful shutdown
- File watcher resource leaks and race conditions
- Command injection vulnerability in process detector
- Performance issue by replacing bubble sort with stdlib sort.Slice

### Security
- Enhanced process detector security to prevent command injection attacks

## [0.2.8] - 2025-10-12

### Fixed
- Enabled CGO in GitHub Actions workflow to fix SQLite database support in Homebrew installations
- Analytics now works correctly when installed via Homebrew

## [0.2.7] - 2025-10-12

### Fixed
- Analytics server now loads conversation data synchronously on startup to ensure data is available before server starts
- Improved initial page load experience with pre-loaded conversation data

## [0.2.6] - 2025-10-12

### Added
- SQLite database for command history tracking
- Automatic command history recording via conversation parsing
- Command history UI with search and filtering capabilities
- User message interception with wrapper script
- User message recording in database

### Security
- Added strict file permissions for database files

### Changed
- Simplified wrapper script implementation
- Enhanced command history tracking with persistent storage

## [0.2.5] - 2025-10-12

### Added
- Analytics server now enabled by default in TUI mode
- Toggle shortcut (Ctrl+A) to start/stop analytics dashboard from TUI
- Quiet mode for analytics server to reduce console output
- Analytics server management integrated into TUI model lifecycle
- Dashboard screenshot in documentation

### Changed
- Analytics server runs automatically when TUI is launched (can be toggled off)
- Improved analytics server lifecycle management with graceful shutdown
- Enhanced TUI experience with integrated analytics control

### Fixed
- Removed debug print statements from WebSocket handler
- Improved analytics server startup/shutdown reliability

## [0.2.4] - 2025-10-12

### Added
- Background shell detection in analytics dashboard
- Process monitoring enhancements for tracking Claude CLI background operations

### Changed
- Improved analytics dashboard to identify and display background shell processes

## [0.2.3] - 2025-10-12

### Added
- Analytics dashboard enhancements for better background process detection

## [0.2.2] - 2025-10-12

### Fixed
- Docker build process now automatically builds cct binary before image build
- Resolved issue where Docker image build would fail if cct binary was missing

## [0.2.1] - 2025-10-12

### Added
- Component removal functionality - ability to uninstall agents, commands, and MCPs
- Interactive component removal in TUI with confirmation prompts

### Changed
- TUI rebranded to "Claude Control Terminal" with updated branding throughout interface
- Improved component management workflow with removal capabilities

## [0.2.0] - 2025-10-12

### Changed - BREAKING
- **Rebrand to Claude Control Terminal (CCT)**: Project renamed from `go-claude-templates` to better reflect its role as a comprehensive control center for Claude Code
- **Module path changed**: `github.com/davila7/go-claude-templates` → `github.com/schlunsen/claude-control-terminal`
- **Repository moved**: `github.com/schlunsen/claude-templates-go` → `github.com/schlunsen/claude-control-terminal`
- All import paths updated across 25 Go files
- CLI descriptions updated to position CCT as "control center and wrapper" for Claude Code

### Added - Docker Support
- **Complete Docker integration** for containerizing Claude Code environments
- New `internal/docker/` package with 3 core modules (~700 lines):
  - `docker.go`: Docker operations (build, run, stop, logs, exec)
  - `dockerfile_generator.go`: Generate 4 types of Dockerfiles
  - `compose_generator.go`: Generate docker-compose.yml templates
- **9 new CLI commands**:
  - `--docker-init`: Generate Dockerfile + .dockerignore
  - `--docker-build`: Build Docker image
  - `--docker-run`: Run containerized Claude environment
  - `--docker-stop`: Stop Docker container
  - `--docker-logs`: View container logs
  - `--docker-compose`: Generate docker-compose.yml
  - `--docker-type`: Select type (base/claude/analytics/full)
  - `--docker-mcps`: Include MCPs in container (comma-separated)
  - `--docker-command`: Custom command to run in container
- **4 Dockerfile templates**:
  - `base`: Minimal CCT-only image
  - `claude`: Full environment (Node.js + Claude CLI + CCT + MCPs)
  - `analytics`: Optimized for analytics dashboard
  - `full`: Complete dev environment with all tools
- **4 docker-compose templates**:
  - `simple`: Claude + CCT
  - `analytics`: Claude + Analytics dashboard
  - `database`: Claude + PostgreSQL
  - `full`: All services (Claude + Analytics + PostgreSQL + Redis)
- MCP integration in Docker containers
- Automatic .dockerignore and .env.example generation

### Improved
- Enhanced TUI with installation status indicators ([G]=Global, [P]=Project)
- Simplified component selection to single-select on Enter
- Improved navigation flow (Enter/Esc returns to list from completion)
- Enhanced "Launch Claude" menu item visibility with special styling
- Better UX showing installation status before installing

### Documentation
- Complete README overhaul with Docker section and migration guide
- Updated CLAUDE.md with new project overview and Docker architecture
- All documentation files updated with new repository URLs
- GitHub workflows updated with new repository references

## [0.1.0] - 2025-10-12

### Changed
- Minor version bump to 0.1.0 marking stable TUI and core functionality

## [0.0.9] - 2025-10-12

### Added
- Claude CLI launcher integration in TUI for direct conversation launching
- Launch Claude Code conversations from selected agents or components
- Interactive component selection with conversation context

## [0.0.8] - 2025-10-12

### Fixed
- TUI MCP installer now properly registers MCP servers in .mcp.json configuration file
- MCPs installed via TUI now work correctly in Claude Code

### Changed
- Added .mcp.json to .gitignore to prevent committing local MCP configurations

## [0.0.7] - 2025-10-12

### Added
- Preview functionality for agents, commands, and MCPs via --preview/-p flag
- Interactive preview screen in TUI with scrollable content viewing
- Preview methods for all component installers
- Ability to view component content before installation in both CLI and TUI modes
- Keyboard navigation in TUI preview: arrow keys, PgUp/PgDn, g/G for top/bottom
- Direct install from preview screen with I key in TUI
- P key to preview selected component from list in TUI

### Fixed
- MCP registration in TUI now uses proper project scope

## [0.0.6] - 2025-10-12

### Added
- MCP installation now properly registers servers in Claude Code configuration files
- Support for project-local vs user-global MCP installation via --scope flag
- Configuration utilities for reading/writing MCP config files

### Changed
- Automated release process with Claude agent integration in justfile

### Fixed
- MCPs not showing up in Claude Code's /mcp command after installation
- MCP servers not being properly registered in .mcp.json or ~/.claude/config.json

## [0.0.5] - 2025-10-12

### Added
- Active filter display with contextual hints when search is not focused
- Two-step Esc behavior: first clears filter, second returns to main screen

### Changed
- Dynamic viewport calculation based on terminal height
- Component list now adapts to any terminal size (min 5, max 20 items)
- Compact help text for terminals with height < 20 lines
- Centered cursor positioning in viewport for better navigation

### Fixed
- TUI elements being cut off in small terminal windows
- Search filter state unclear after exiting search mode
- Help text and component lists truncated in limited height terminals

## [0.0.4] - 2025-10-12

### Added
- Page up/down navigation support in component lists

### Changed
- Organized documentation files into docs/ directory
- Streamlined changelog to follow Keep a Changelog format

### Removed
- Old test scripts from project root

## [0.0.3] - 2025-10-12

### Added
- Modern interactive TUI with theme support
- Bubbles/Bubbletea-based component selection interface
- Visual theme with gradients and modern styling

### Fixed
- Homebrew formula generation in release workflow
- Installation documentation accuracy

## [0.0.2] - 2025-10-12

### Added
- Homebrew formula generation to release workflow
- Automated release commands to justfile
- Professional README improvements
- LICENSE file with MIT License
- CONTRIBUTING.md with development guidelines
- GitHub issue templates (bug report, feature request, question)
- Pull request template
- Badges to README (Go version, license, build status, release)
- Table of contents to README

### Changed
- Streamlined README for professional appearance
- Updated repository URLs from placeholders to actual repository
- Enhanced code blocks with language labels

## [0.0.1] - 2025-10-12

### Added
- Initial Go port of claude-code-templates from Node.js
- CLI with Cobra framework and Pterm terminal UI
- Fiber web server with WebSocket support for real-time updates
- Analytics dashboard with embedded frontend
- Smart category search for agents, commands, and MCPs across 50+ categories
- Cross-platform builds (Linux, macOS, Windows on amd64/arm64)
- File system watching with fsnotify for real-time conversation monitoring
- RESTful API with 6 endpoints (health, data, conversations, processes, stats, refresh)
- Comprehensive test suite with automated category search validation
- Makefile and justfile for build automation
- GitHub Actions workflow for multi-platform releases
- Component management system (agents, commands, MCPs)
- File operations module for template management
- ConversationAnalyzer and FileWatcher modules
- Analytics core modules (StateCalculator, ProcessDetector)

### Fixed
- Component installation 404 errors with comprehensive category search
- Path handling from "cli-tool/templates" to "cli-tool"
- WebSocket unused variable warnings

## Version Comparison Links

[Unreleased]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.13...HEAD
[0.2.13]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.12...v0.2.13
[0.2.12]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.11...v0.2.12
[0.2.11]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.10...v0.2.11
[0.2.10]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.9...v0.2.10
[0.2.9]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.8...v0.2.9
[0.2.8]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.7...v0.2.8
[0.2.7]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/schlunsen/claude-control-terminal/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.9...v0.1.0
[0.0.9]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/schlunsen/claude-control-terminal/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/schlunsen/claude-control-terminal/releases/tag/v0.0.1
