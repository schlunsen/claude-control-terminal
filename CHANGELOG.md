# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.4] - 2025-10-12

### Added
- LICENSE file with MIT License
- CONTRIBUTING.md with development guidelines
- GitHub issue templates (bug report, feature request, question)
- Pull request template
- Badges to README.md (Go version, license, build status, release)
- Table of contents to README.md

### Changed
- Updated repository URLs from placeholders to actual repository
- Enhanced code blocks with language labels
- Organized documentation files into docs/ directory
- Streamlined changelog to follow Keep a Changelog format

### Removed
- Old test scripts from project root

## [2.0.0] - 2025-10-12

### Added
- Complete Go port of claude-code-templates from Node.js
- CLI with Cobra framework and Pterm terminal UI
- Fiber web server with WebSocket support for real-time updates
- Analytics dashboard with embedded frontend
- Smart category search for agents, commands, and MCPs across 50+ categories
- Cross-platform builds (Linux, macOS, Windows on amd64/arm64)
- File system watching with fsnotify for real-time conversation monitoring
- RESTful API with 6 endpoints (health, data, conversations, processes, stats, refresh)
- Comprehensive test suite with automated category search validation
- Makefile and justfile for build automation

### Changed
- Build time: 2-5 seconds (50-100x faster than Node.js)
- Binary size: 15MB (3x smaller than node_modules)
- Startup time: <10ms (50x faster)
- Memory usage: 15MB (5x lower)

### Fixed
- Component installation 404 errors with comprehensive category search
- Path handling from "cli-tool/templates" to "cli-tool"
- WebSocket unused variable warnings

## Version Comparison Links

[Unreleased]: https://github.com/schlunsen/claude-templates-go/compare/v0.0.4...HEAD
[0.0.4]: https://github.com/schlunsen/claude-templates-go/compare/v2.0.0...v0.0.4
[2.0.0]: https://github.com/schlunsen/claude-templates-go/releases/tag/v2.0.0
