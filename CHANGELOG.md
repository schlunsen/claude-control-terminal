# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/schlunsen/claude-templates-go/compare/v0.0.4...HEAD
[0.0.4]: https://github.com/schlunsen/claude-templates-go/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/schlunsen/claude-templates-go/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/schlunsen/claude-templates-go/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/schlunsen/claude-templates-go/releases/tag/v0.0.1
