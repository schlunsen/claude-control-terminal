# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.15] - Browser-Based Voice Recording & Transcription - 2025-10-23

### Added
- Browser-based voice recording with Whisper.js transcription for hands-free agent interaction
- Three Whisper model options: Tiny (~35MB, fast), Base (~75MB, balanced), Small (~150MB, accurate)
- Voice recording settings UI with model selection and visual size/speed/accuracy indicators
- Auto-start recording after Whisper model loads for streamlined voice input workflow
- Keyboard shortcuts for voice recording: Ctrl/Cmd+R to start, Space to stop, Escape to cancel
- Real-time progress tracking during Whisper model downloads with per-file progress calculation
- Vertical button layout in voice recording modal optimized for mobile UX
- Visual feedback showing keyboard shortcuts during recording session
- Smart model switching that avoids reloading when same model already loaded
- Complete documentation in VOICE_RECORDING.md with architecture and usage guide

### Changed
- Voice recording UX improved with immediate auto-start after model initialization
- Recording modal buttons stacked vertically for better mobile accessibility
- Progress bar now accurately reflects multi-file Whisper model download status
- Button labels enhanced to display keyboard shortcut hints

### Technical
- Integrated Whisper.js (@xenova/transformers) for client-side speech recognition
- Dynamic model loading based on user preferences stored in settings
- Per-file progress tracking with overall percentage calculation
- Model caching to prevent unnecessary redownloads

## [0.5.14] - Username/Password Authentication System - 2025-10-23

### Added
- Comprehensive username/password authentication system with session-based login
- Admin user management interface accessible via TUI ('U' key)
- Dedicated login page with beautiful UI and form validation
- User menu dropdown in navbar showing username and admin badge
- Configurable authentication modes: disabled, optional, or required
- Session-based authentication with 24-hour token expiration
- Password change functionality for authenticated users
- Admin-only user management endpoints (create, list, delete users)
- Rate limiting on authentication endpoints (5 login attempts, 3 password changes per 15min)
- Interactive setup wizard for creating first admin user via TUI
- useAuth composable for managing authentication state across frontend
- Automatic login prompt when authentication is required
- Comprehensive USER_AUTHENTICATION.md documentation

### Security
- Bcrypt password hashing with cost factor 12 for secure password storage
- 256-bit random session tokens for secure session management
- HttpOnly session cookies to prevent XSS attacks
- Secure flag for HTTPS to prevent man-in-the-middle attacks
- Timing attack protection using crypto/subtle constant-time comparison for API keys
- File permissions: 0600 for sensitive files (users.json, sessions.json)
- Minimum 8 character password requirement
- Session middleware protecting endpoints based on configuration
- .gitignore updated to exclude sensitive files (users.json, sessions.json, .secret, *.key, *.crt)

### Changed
- Authentication middleware now properly enforces authentication based on require_login setting
- Static assets (/, /assets/, /_nuxt/, favicon) allowed without authentication
- API access properly blocked when user authentication is enabled
- Improved middleware logic for strict vs optional authentication modes
- User menu appears in top-right navbar when authenticated with logout option

### API
- POST /api/auth/login - Login with username/password credentials
- POST /api/auth/logout - Logout current session
- GET /api/auth/status - Check authentication status
- POST /api/auth/change-password - Change user password
- POST /api/auth/users - Create user (admin only)
- GET /api/auth/users - List all users (admin only)
- DELETE /api/auth/users/:username - Delete user (admin only)

### Files
- ~/.claude/analytics/users/users.json - User account storage
- ~/.claude/analytics/users/sessions.json - Active session tracking
- USER_AUTHENTICATION.md - Complete authentication documentation

## [0.5.13] - Settings UI & Unified Diff Display - 2025-10-23

### Added
- Comprehensive Settings UI page for user preference management
- Professional unified diff algorithm with contextual display using jsdiff library
- Smart context-aware diff with automatic collapsing for large file edits
- Conditional diff display based on user preferences (inline chat vs overlay)
- Diff statistics showing additions/deletions/context lines
- Side-by-side diff support infrastructure for future enhancements

### Changed
- Edit diff display now defaults to 'In Chat' for better inline visibility
- Removed 'In Panel' option from diff display settings for simplified UX
- Edit tool data now persists on messages for reliable diff display
- Diff display shows 3 lines of context around changes matching git diff standards
- Improved diff rendering with proper line-by-line matching instead of simple before/after

### Fixed
- Resolved variable shadowing bug in useDiff composable causing initialization errors
- Fixed Edit diff persistence issue by storing data directly on message objects
- Corrected default diff display values across database schema and TypeScript types
- Edit diffs now display reliably even after tool dismissal

### Removed
- Verbose debug console logging across frontend codebase for better performance
- Removed debug logs from TodoWrite parser and WebSocket connection handling
- Cleaned up console noise while preserving error and warning visibility

## [0.5.12] - Sidebar Improvements & Session State Persistence - 2025-10-23

### Added
- Enhanced sidebar with improved session state persistence for better user experience
- Session state now reliably persists across page refreshes and navigation
- Improved sidebar UI with better visual feedback and interaction patterns

### Changed
- Sidebar architecture refactored for more reliable state management
- Enhanced session tracking to maintain user context during navigation
- Improved sidebar responsiveness and performance

## [0.5.11] - Image Support & Agent UI Enhancements - 2025-10-20

### Added
- Image support in agent chat with drag-and-drop and paste functionality for visual asset sharing
- Lightbox viewer for full-screen image viewing with zoom and navigation controls
- Image gallery support in agent conversations for better visual context

### Changed
- Improved agent session creation preserving provider/model settings from TUI configuration
- Complete modularization of agents.vue component achieving 83% size reduction for better performance
- Enhanced UI responsiveness and maintainability through component refactoring

### Fixed
- Provider and model selection now correctly persists when creating new agent sessions from TUI
- Session creation workflow now respects TUI-configured settings instead of reverting to defaults

## [0.5.10] - Maintenance Release - 2025-10-20

### Changed
- Version bump for maintenance release
- Updated internal version numbers and dependencies

## [0.5.9] - Character Avatar Enhancement & Session Management - 2025-10-20

### Added
- Deterministic character avatar assignment for session IDs using hash-based mapping
- Character avatars now displayed in session selector on agents page for visual identification
- Session start time tracking and display in dropdown selectors with relative timestamps
- Enhanced session metadata showing both character name and session ID for better clarity

### Changed
- Session selector dropdown now shows character names instead of raw session IDs for better UX
- Character avatar composable enhanced with fallback logic: direct lookup first, then hash-based assignment
- Sessions now sorted by most recent first in activity history and selector dropdowns
- Activity history displays character names as primary label with session ID as metadata

### Improved
- Consistent visual identification across all session displays (agents page, activity history, dropdowns)
- Better user experience with memorable character-based session naming system
- Enhanced session management with character avatars providing visual cues throughout UI

## [0.5.8] - Multi-Provider Support & UI Enhancements - 2025-10-20

### Added
- Multi-provider support with custom model selection for flexible AI provider configuration
- Custom model input allowing users to specify any model name beyond predefined options
- Enhanced statistics page with comprehensive provider metrics and usage insights
- Interactive website features section with show/hide screenshots for better documentation
- New website screenshots showcasing analytics, live agents, shortcuts, statistics, and themes

### Changed
- Improved provider configuration with dynamic model selection across multiple AI providers
- Enhanced session metrics component with comprehensive conversation visibility
- Upgraded website UI with optimized screenshot gallery and improved navigation
- Better state calculator with tool result message type distinction

### Fixed
- Fixed viewport layout to provide desktop app experience with proper scroll behavior
- Tool results now correctly distinguished from user messages in analytics
- Provider model selection now persists custom model choices correctly
- Enhanced chat scroll behavior with deep watch mode for improved reactivity

## [0.5.7] - South Park Theme & Theme Carousel Enhancement - 2025-10-20

### Added
- South Park Dark and Light themes with Comic Sans typography and South Park-inspired color palette
- Theme selector component in header for quick theme switching
- Comprehensive theme carousel navigation page with all 5 theme families (Default, Neon, Nord, Dracula, South Park)
- Theme documentation with THEMES.md describing all theme options and customization
- Enhanced theme composable with support for theme persistence and system-wide theme management

### Changed
- Increased theme carousel scroll amount from 320px to 650px for faster navigation and improved user experience
- Improved theme toggle component with better integration for all theme families
- Enhanced sidebar with theme selector for easy access to theme customization

## [0.5.6] - Historical Message Loading & Session Deletion - 2025-10-19

### Added
- Historical message loading for agent sessions to view complete conversation history
- Session deletion functionality allowing users to remove agent sessions from the dashboard
- Deep watch mode for agent chat auto-scroll ensuring smooth scrolling behavior

### Fixed
- Agent chat auto-scroll now uses deep watch for improved reactivity and reliability
- Session management improved with proper cleanup when sessions are deleted

## [0.5.5] - Permission Streaming & Logging Enhancements - 2025-10-19

### Fixed
- Agent permission callbacks now use streaming mode for improved real-time responsiveness
- Enhanced logging for permission request handling with detailed debugging information
- Improved error handling for missing permission_id in frontend responses
- Better visibility of permission request flow through enhanced logging

### Changed
- Switched permission callbacks from non-streaming to streaming mode for better performance
- Improved logging infrastructure for permission request lifecycle tracking

## [0.5.4] - Permission Callback Improvements - 2025-10-19

### Changed
- Updated claude-agent-sdk-go dependency from v0.2.1 to v0.2.2 for improved stability

### Fixed
- Permission popup now displays tool descriptions correctly instead of showing 'undefined'
- Agent permission callbacks switched to streaming mode for better real-time responsiveness
- Enhanced permission request UI with proper tool information display
- Removed local replace directive for claude-agent-sdk-go to use official package version

## [0.5.3] - SQLite Agent Session Persistence - 2025-10-19

### Added
- SQLite-based agent session persistence for reliable session state management across server restarts
- Database-backed session storage providing durability and crash recovery
- Persistent session tracking for long-running agent conversations
- Enhanced session recovery capabilities after server interruptions

### Changed
- Agent sessions now stored in SQLite database instead of in-memory only
- Improved session reliability with persistent storage backend
- Enhanced session lifecycle management with database integration

## [0.5.2] - Agent Session Timeout Improvement - 2025-10-19

### Changed
- Increased agent session timeout from 60 seconds to 5 minutes for better stability with long-running agent operations
- Enhanced session management to accommodate complex tasks requiring extended execution time

## [0.5.1] - Global Keyboard Shortcuts - 2025-01-19

### Added
- Global keyboard shortcuts system for enhanced accessibility and efficiency
- Shortcuts dialog (press '?' key) displaying all available keyboard commands
- Comprehensive keyboard navigation support across all dashboard screens
- Help modal with organized shortcut reference grouped by functionality

### Changed
- Improved user experience with discoverable keyboard shortcuts
- Enhanced navigation workflow with consistent keyboard controls

## [0.5.0] - Unified Go Server & Enhanced Logging - 2025-01-19

### Added
- Comprehensive logging system with structured logging for enhanced debugging and monitoring
- Enhanced tool tracking with detailed execution information and metrics
- New internal/logging package providing centralized logging infrastructure
- Debug logging throughout agent server for full conversation visibility
- TodoWrite and tool execution event overlays in agents page for real-time visibility
- API key validation warnings on server startup for better troubleshooting

### Changed
- **BREAKING**: Migrated agent server from Python FastAPI to Go with native implementation
- Complete unified server implementation integrating analytics and agent functionality
- Agent server now uses Gorilla WebSocket and claude-agent-sdk-go natively in Go
- Improved session management with Go-native implementation
- Enhanced agent handler with comprehensive logging and error tracking
- Streamlined server architecture with single-process unified server
- Converted all log.Printf statements to internal/logging package for consistent verbose output

### Removed
- Python-based agent server (internal/agents/agents_server/)
- Python dependencies and virtual environment management
- FastAPI and Python FastAPI WebSocket implementation
- Legacy Python agent manager, auth, and session modules (~2500 lines of Python code)

### Fixed
- **CRITICAL**: Downgraded claude-agent-sdk-go from v0.2.0 to v0.1.3 for stability and compatibility
- **CRITICAL**: Restored WithVerbose option that was removed in SDK v0.2.0 for better debugging
- Agent connection issues caused by SDK v0.2.0 incompatibilities
- Server tests updated to match NewServerWithOptions signature with verbose parameter
- Improved error handling and logging throughout agent lifecycle
- Better signal handling and cleanup for agent server process lifecycle
- Enhanced verbose logging provides comprehensive session and tool execution visibility

## [0.4.4] - Real-Time Session Metrics Dashboard - 2025-01-16

### Added
- Real-time session metrics dashboard in agents page with comprehensive conversation visibility
- SessionMetrics.vue component displaying live session statistics:
  - Session status and duration tracking
  - Message count tracking with visual progress bars
  - Tool usage statistics and breakdown by tool type
  - Permission approval/denial rates with visual indicators
  - Working directory and configuration details
- Comprehensive debug logging to agent server for full conversation visibility
- Live tracking of tool executions, permissions, and message counts
- Detailed tool execution information extraction (files, commands, patterns) in execution bars
- Status updates during message streaming and tool execution phases
- Responsive design supporting desktop, tablet, and mobile views

### Changed
- Enhanced agents page with integrated metrics sidebar for better monitoring
- Improved real-time WebSocket updates for session metrics

## [0.4.3] - MCP Server Integration - 2025-10-16

### Added
- MCP (Model Context Protocol) server integration for agents_server to extend Claude's capabilities with custom tools and resources
- Support for stdio-based MCP server processes with configurable commands, arguments, and environment variables
- MCPServerConfig model for managing MCP server configurations
- Comprehensive MCP tool permission handling integrated with existing permission system
- MCP tool naming convention: mcp__<server_name>__<tool_name>
- Configuration options for per-server permission requirements via require_permission flag
- Documentation and examples for calculator and GitHub MCP servers

### Changed
- Enhanced agent_manager.py with _build_mcp_servers() method to register and manage MCP servers
- Improved models.py with MCP server configuration data structures

## [0.4.2] - Live Agent TodoWrite & Tool Tracking - 2025-10-16

### Added
- Real-time TodoWrite event streaming for live agent sessions in analytics dashboard
- Tool execution tracking with live updates showing tool calls and results
- Enhanced TodoWrite parsing to capture structured task updates from agents
- Live visualization of agent task progress with status indicators
- Tool event timeline showing execution history and results

### Changed
- Improved agent manager to emit TodoWrite and tool execution events via WebSocket
- Enhanced agents.vue page with dedicated TodoWrite panel showing live task updates
- Refined todo auto-hide timing with 2-second delay after completion for better visibility

### Fixed
- TodoWrite parsing now handles new agent SDK input format correctly
- Todo cleanup logic improved to prevent premature hiding of active tasks
- Better signal handling and cleanup for agent server process lifecycle

## [0.4.1] - UI Default Value Cleanup - 2025-01-16

### Fixed
- Removed hardcoded user-specific working directory defaults from agent session forms
- Working directory fields now use empty string defaults instead of '/Users/schlunsen/projects'
- Improved UI generalization for all users

## [0.4.0] - Agent Server & Live Agent Integration - 2025-10-16

### Added
- 🤖 **Agent Server**: New Python FastAPI WebSocket server for real-time Claude agent conversations
- Full Claude Agent SDK integration with WebSocket support
- Session management for multiple concurrent agent conversations
- Embedded Python runtime for agent server functionality
- Automatic Python dependency management via virtual environments
- Comprehensive tool support (Read, Write, Edit, Bash, etc.)
- Real-time agent communication with streaming responses

### Changed
- Enhanced release workflow documentation with descriptive release names in CHANGELOG format
- Improved release process guide in workflow documentation for better clarity
- Streamlined project structure to support agent server integration
- Updated CLI commands to support agent server management

### Fixed
- Disabled Windows builds in GitHub Actions release workflow due to compilation issues
- Improved process management for agent server lifecycle

### Security
- API key authentication for agent server
- Secure session management
- Isolated Python virtual environment for dependencies

## [0.3.5] - TLS/HTTPS Security & API Authentication - 2025-10-16

### Added
- TLS/HTTPS encryption enabled by default for analytics server with auto-generated self-signed certificates
- API key authentication system protecting write operations to analytics endpoints
- Automatic API key generation and storage in `~/.claude/analytics/.secret`
- TLS certificate auto-generation with 1-year validity and expiration warnings
- Comprehensive security configuration in `~/.claude/analytics/config.json`
- Enhanced hook scripts with automatic API key authentication and TLS support
- Security documentation covering TLS, API keys, and best practices

### Changed
- Analytics server now runs on HTTPS by default (https://localhost:3333)
- All hooks updated to use API key authentication via Authorization header
- TUI analytics dashboard URLs updated to use HTTPS protocol
- Analytics server configuration now supports enabling/disabling TLS and auth independently
- Hook scripts enhanced with self-signed certificate support (-k flag for curl)

### Security
- All POST/PUT/DELETE/PATCH requests now require API key authentication
- GET requests remain unauthenticated for browser access
- Server binds to localhost (127.0.0.1) by default for security
- Self-signed certificates stored in `~/.claude/analytics/certs/`

### Fixed
- Analytics header UI cleaned up by removing non-functional "Open Dashboard" button

## [0.3.4] - AI Model Provider Tracking - 2025-10-15

### Added
- AI model provider tracking with color-coded badges in analytics dashboard
- Provider badges showing AI service (Anthropic, OpenAI, Google, etc.) with distinct colors
- Enhanced analytics UI with visual provider identification for conversations

### Changed
- Bumped GitHub Pages deployment version for improved website stability

## [0.3.3] - Browser Integration and Character Avatars - 2025-10-15

### Added
- Browser integration: Press 'O' in TUI menu to open analytics dashboard in default browser (when analytics is enabled)
- South Park character avatars for session names in analytics dashboard (26 optimized character images)
- Modern session selector dropdown in analytics with character avatars and session metadata
- Session start time tracking and sorting (most recent sessions first)
- Character avatar composable with 25+ South Park character mappings

### Changed
- Enhanced TUI help text to show 'O: Open Dashboard' when analytics is enabled
- Improved ActivityHistory component with visual session identification using character avatars
- Analytics dashboard now displays session info with avatars, IDs, and relative timestamps

## [0.3.2] - Website Branding Updates - 2025-10-15

### Changed
- Updated website favicons for improved branding consistency
- Enhanced website header with current version display (v0.3.1)

## [0.3.1] - Hook Installation Improvements - 2025-10-15

### Fixed
- Hook scripts now embedded in binary using Go's embed package for portability
- `cct --install-all-hooks` now works from any directory without requiring hook source files
- Created hooks package with embedded .sh files for reliable hook installation

## [0.3.0] - PostToolUse Hooks & Test Coverage - 2025-10-15

### Added
- Nuxt-based documentation website with modern UI at website/
- Lightbox viewer for screenshot galleries in documentation
- GitHub Actions workflow for automated website deployment
- Enhanced website UI with responsive design and improved navigation
- COVERAGE.md file documenting test coverage status

### Changed
- Improved test coverage across analytics, file watcher, and reset tracker modules
- Enhanced mobile responsiveness for documentation website
- Hero section spacing and badge alignment improvements
- Refactored version management to dedicated internal/version package

### Removed
- Non-working wrapper functionality and related scripts (internal/wrapper, scripts/install-wrapper.sh)
- Deprecated wrapper tests that were no longer functional

### Fixed
- Mobile responsive layout issues in documentation website
- Hero section spacing and component alignment

## [0.2.20] - MCP Metadata Tracking - 2025-10-14

### Added
- MCP metadata tracking system (.mcp-metadata.json) that maintains install name to server keys mapping
- Reliable MCP uninstall using metadata for exact server key removal from .mcp.json
- Comprehensive test suite with 15 new tests for MCP metadata and detection functionality
- Backward compatibility with legacy MCP installs through substring matching fallback

### Changed
- Improved TUI installation status detection for MCPs with complex names (e.g., deepgraph-vue → DeepGraph Vue MCP)
- Enhanced MCP detection with bidirectional name matching between install names and server keys

### Fixed
- MCP uninstall now accurately removes correct server entries using metadata tracking
- Installation status indicator in TUI now correctly identifies installed MCPs regardless of name complexity

## [0.2.19] - Batch Component Operations - 2025-10-14

### Added
- Multi-component selection support with Space key for batch operations
- Action chooser screen to select install/uninstall for multiple components
- Auto-refresh component list after install/remove operations to show updated status
- Visual selection indicators (checkmark) and improved help text

### Changed
- Component operations now support batch install/uninstall workflows
- Silent skip for non-installed components instead of errors during uninstall

### Fixed
- MCP removal now properly cleans up all servers from .mcp.json configuration
- Fixed broken string matching for MCP server removal

## [0.2.18] - Claude CLI Detection Improvements - 2025-10-14

### Fixed
- Claude CLI detection now works when installed in ~/.local/bin but not in PATH
- TUI launcher now properly finds Claude binary in common installation locations (/usr/local/bin, /opt/homebrew/bin, ~/.local/bin)
- FindClaudePath() function enhanced to check common locations beyond PATH

### Testing
- Added comprehensive tests for FindClaudePath() function in wrapper package
- Added 160+ lines of test coverage for Claude binary detection
- Tests cover PATH detection, common location fallback, and error cases

## [0.2.17] - Automatic Claude CLI Installer - 2025-10-14

### Added
- Automatic Claude CLI installer with native binary and npm fallback support
- New `--install-claude` flag for automated Claude CLI installation
- Node.js version detection and compatibility checking (v18+ required for npm fallback)
- TUI integration for Claude CLI installation detection and prompts
- Interactive installation prompts with progress feedback
- Detection of existing Claude installations to avoid redundant installs

### Changed
- TUI claude_launcher now detects missing Claude CLI and suggests installation
- Improved user experience with automatic installation option instead of manual setup

### Documentation
- Updated README with comprehensive installation documentation
- Added automatic installer benefits and usage examples

### Testing
- Added 40+ new tests for installer functionality
- Docker-based testing environment for clean installation verification
- Comprehensive test coverage for edge cases and error handling

## [0.2.16] - Test Coverage Infrastructure - 2025-10-14

### Added
- Comprehensive test coverage for core packages (analytics, server, websocket, components)
- Testing infrastructure with coverage thresholds and CI integration
- Scheduled CI runs for continuous test validation

### Changed
- Improved test coverage from minimal to 60%+ across critical packages
- Enhanced CI workflow with test coverage reporting

## [0.2.15] - Persistent Provider Storage - 2025-10-14

### Added
- Persistent provider token storage with SQLite database for secure credential management
- Custom model input support for AI providers allowing users to specify any model name
- Responsive compact UI for provider configuration that adapts to terminal size
- Enhanced provider configuration screen with improved layout and usability

### Changed
- Provider tokens and configurations now persist across sessions in ~/.claude/cct_history.db
- Provider UI now displays in compact mode on smaller terminals for better accessibility
- Improved provider model selection with custom input option

## [0.2.14] - AI Provider Configuration - 2025-10-13

### Added
- AI provider configuration and management system for flexible model selection
- Support for multiple AI providers (Anthropic, OpenAI, Google, Mistral, etc.)
- Provider configuration UI in TUI for easy setup
- Model selection and API key management per provider

## [0.2.13] - Multi-Source Permissions - 2025-10-13

### Added
- Multi-source permissions management with tabbed UI for better control over Claude Code permissions
- GitHub CLI (gh) commands now included in Git Commands permission category

### Improved
- Enhanced permissions management interface with tabbed navigation between different permission sources

## [0.2.12] - Local Permissions Fix - 2025-10-13

### Fixed
- Permissions management now uses local `.claude/settings.local.json` instead of global settings file for better project isolation
- Empty permissions object is now properly removed from settings when all permissions are disabled

## [0.2.11] - Permissions Improvements - 2025-10-13

### Fixed
- Permissions management improvements

## [0.2.10] - Session Launcher & Documentation - 2025-10-13

### Added
- 'Launch last Claude session' menu option in TUI for quick access to recent conversations
- Comprehensive godoc package comments across all internal packages for better code documentation

### Fixed
- Search bar state persistence issue when navigating between screens in TUI

## [0.2.9] - Security & Stability Improvements - 2025-10-13

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

## [0.2.8] - Homebrew CGO Fix - 2025-10-12

### Fixed
- Enabled CGO in GitHub Actions workflow to fix SQLite database support in Homebrew installations
- Analytics now works correctly when installed via Homebrew

## [0.2.7] - Analytics Startup Fix - 2025-10-12

### Fixed
- Analytics server now loads conversation data synchronously on startup to ensure data is available before server starts
- Improved initial page load experience with pre-loaded conversation data

## [0.2.6] - Command History Database - 2025-10-12

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

## [0.2.5] - Integrated Analytics - 2025-10-12

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

## [0.2.4] - Background Shell Detection - 2025-10-12

### Added
- Background shell detection in analytics dashboard
- Process monitoring enhancements for tracking Claude CLI background operations

### Changed
- Improved analytics dashboard to identify and display background shell processes

## [0.2.3] - Analytics Enhancements - 2025-10-12

### Added
- Analytics dashboard enhancements for better background process detection

## [0.2.2] - Docker Build Fix - 2025-10-12

### Fixed
- Docker build process now automatically builds cct binary before image build
- Resolved issue where Docker image build would fail if cct binary was missing

## [0.2.1] - Component Removal - 2025-10-12

### Added
- Component removal functionality - ability to uninstall agents, commands, and MCPs
- Interactive component removal in TUI with confirmation prompts

### Changed
- TUI rebranded to "Claude Control Terminal" with updated branding throughout interface
- Improved component management workflow with removal capabilities

## [0.2.0] - Claude Control Terminal Rebrand & Docker Support - 2025-10-12

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

## [0.1.0] - Stable TUI Release - 2025-10-12

### Changed
- Minor version bump to 0.1.0 marking stable TUI and core functionality

## [0.0.9] - Claude CLI Launcher - 2025-10-12

### Added
- Claude CLI launcher integration in TUI for direct conversation launching
- Launch Claude Code conversations from selected agents or components
- Interactive component selection with conversation context

## [0.0.8] - MCP Registration Fix - 2025-10-12

### Fixed
- TUI MCP installer now properly registers MCP servers in .mcp.json configuration file
- MCPs installed via TUI now work correctly in Claude Code

### Changed
- Added .mcp.json to .gitignore to prevent committing local MCP configurations

## [0.0.7] - Component Preview - 2025-10-12

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

## [0.0.6] - MCP Configuration - 2025-10-12

### Added
- MCP installation now properly registers servers in Claude Code configuration files
- Support for project-local vs user-global MCP installation via --scope flag
- Configuration utilities for reading/writing MCP config files

### Changed
- Automated release process with Claude agent integration in justfile

### Fixed
- MCPs not showing up in Claude Code's /mcp command after installation
- MCP servers not being properly registered in .mcp.json or ~/.claude/config.json

## [0.0.5] - TUI Improvements - 2025-10-12

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

## [0.0.4] - Navigation & Documentation - 2025-10-12

### Added
- Page up/down navigation support in component lists

### Changed
- Organized documentation files into docs/ directory
- Streamlined changelog to follow Keep a Changelog format

### Removed
- Old test scripts from project root

## [0.0.3] - Modern TUI - 2025-10-12

### Added
- Modern interactive TUI with theme support
- Bubbles/Bubbletea-based component selection interface
- Visual theme with gradients and modern styling

### Fixed
- Homebrew formula generation in release workflow
- Installation documentation accuracy

## [0.0.2] - Homebrew & Documentation - 2025-10-12

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

## [0.0.1] - Initial Go Port - 2025-10-12

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

[0.5.14]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.13...v0.5.14
[0.5.13]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.12...v0.5.13
[0.5.12]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.11...v0.5.12
[0.5.11]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.10...v0.5.11
[0.5.10]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.9...v0.5.10
[0.5.9]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.8...v0.5.9
[0.5.8]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.7...v0.5.8
[0.5.7]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.6...v0.5.7
[0.5.6]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.5...v0.5.6
[0.5.5]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/schlunsen/claude-control-terminal/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/schlunsen/claude-control-terminal/compare/v0.4.4...v0.5.0
[0.4.4]: https://github.com/schlunsen/claude-control-terminal/compare/v0.4.3...v0.4.4
[0.4.3]: https://github.com/schlunsen/claude-control-terminal/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/schlunsen/claude-control-terminal/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/schlunsen/claude-control-terminal/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/schlunsen/claude-control-terminal/compare/v0.3.5...v0.4.0
[0.3.5]: https://github.com/schlunsen/claude-control-terminal/compare/v0.3.4...v0.3.5
[0.3.4]: https://github.com/schlunsen/claude-control-terminal/compare/v0.3.3...v0.3.4
[0.3.3]: https://github.com/schlunsen/claude-control-terminal/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/schlunsen/claude-control-terminal/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/schlunsen/claude-control-terminal/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.20...v0.3.0
[0.2.20]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.19...v0.2.20
[0.2.19]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.18...v0.2.19
[0.2.18]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.17...v0.2.18
[0.2.17]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.16...v0.2.17
[0.2.16]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.15...v0.2.16
[0.2.15]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.14...v0.2.15
[0.2.14]: https://github.com/schlunsen/claude-control-terminal/compare/v0.2.13...v0.2.14
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
