# Contributing to Go Claude Templates (CCT)

Thank you for your interest in contributing to Go Claude Templates! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Contributing Code](#contributing-code)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Style Requirements](#code-style-requirements)
- [Testing Requirements](#testing-requirements)
- [CI/CD Pipeline](#cicd-pipeline)
- [Commit Message Conventions](#commit-message-conventions)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow. Please be respectful and constructive in all interactions.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**: Use a clear and descriptive title
- **Steps to reproduce**: Detailed steps to reproduce the issue
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Environment details**:
  - Go version (`go version`)
  - Operating system and version
  - CCT version (`./cct --version`)
- **Additional context**: Screenshots, error messages, logs

Use the bug report template when creating issues.

### Suggesting Features

Feature suggestions are welcome! When suggesting a feature:

- **Check existing issues**: Someone may have already suggested it
- **Clear use case**: Explain why this feature would be useful
- **Implementation ideas**: If you have thoughts on implementation, share them
- **Scope**: Keep features focused and aligned with project goals

Use the feature request template when creating issues.

### Contributing Code

We welcome pull requests for:

- Bug fixes
- New features
- Performance improvements
- Documentation improvements
- Test coverage improvements

## Development Setup

### Prerequisites

- **Go 1.23 or higher**: [Install Go](https://go.dev/doc/install)
- **Git**: For version control
- **Make or Just**: Build automation tools
  - Make: Usually pre-installed on macOS/Linux
  - Just: `brew install just` (macOS) or see [just documentation](https://github.com/casey/just)

### Clone and Build

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/YOUR_USERNAME/claude-control-terminal
cd go-claude-templates

# Install dependencies
go mod download

# Build the project
make build
# or
just build

# Verify the build
./cct --version
```

### Project Structure

```text
go-claude-templates/
├── cmd/cct/                    # CLI entry point
├── internal/                   # Private application code
│   ├── analytics/              # Analytics backend modules
│   ├── cmd/                    # CLI commands & UI
│   ├── components/             # Component installers
│   ├── fileops/                # File operations
│   ├── server/                 # Web server
│   └── websocket/              # WebSocket server
├── Makefile                    # Make build automation
├── justfile                    # Just task runner
└── README.md                   # User documentation
```

## Development Workflow

### Making Changes

1. **Create a branch**: Use descriptive branch names
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/bug-description
   ```

2. **Make your changes**: Follow the code style guidelines below

3. **Run tests**: Ensure all tests pass
   ```bash
   make test
   # or
   just test
   ```

4. **Format code**: Format your code before committing
   ```bash
   make fmt
   # or
   just fmt
   ```

5. **Build and test**: Verify your changes work
   ```bash
   make build
   ./cct --version
   # Test relevant functionality
   ```

### Running the Application

```bash
# Run directly
go run ./cmd/cct

# Run analytics dashboard for testing
./cct --analytics

# Test component installation
./cct --agent test --directory /tmp/test
```

## Code Style Requirements

### Go Style Guide

Follow the official Go style guidelines:

- **Effective Go**: [https://golang.org/doc/effective_go](https://golang.org/doc/effective_go)
- **Go Code Review Comments**: [https://github.com/golang/go/wiki/CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)

### Key Conventions

1. **Error Handling**: Always check and handle errors explicitly
   ```go
   if err != nil {
       return fmt.Errorf("failed to do X: %w", err)
   }
   ```

2. **Naming**:
   - Packages: lowercase, single word (`analytics`, `server`)
   - Structs: PascalCase (`ConversationAnalyzer`, `ProcessDetector`)
   - Functions: camelCase for private, PascalCase for exported
   - Constants: PascalCase

3. **Comments**: Document all exported types, functions, and methods
   ```go
   // ConversationAnalyzer handles conversation data loading and analysis.
   // It provides methods for parsing JSONL files and extracting metrics.
   type ConversationAnalyzer struct { ... }
   ```

4. **File Naming**: Use snake_case for Go files (`state_calculator.go`)

5. **Formatting**: Always run `go fmt` or `make fmt` before committing

### Code Organization

- Keep functions focused and small (single responsibility principle)
- Group related functionality in the same file
- Use meaningful variable and function names
- Avoid deep nesting (prefer early returns)

## Testing Requirements

### Running Tests

```bash
# Run all tests
make test
# or
just test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/analytics/...
```

### Writing Tests

- Place tests in `*_test.go` files next to the code they test
- Use table-driven tests where appropriate
- Test both success and error cases
- Aim for meaningful test coverage (not just high percentage)

Example test structure:

```go
func TestStateCalculator_DetermineState(t *testing.T) {
    tests := []struct {
        name     string
        messages []Message
        want     string
    }{
        {
            name:     "active conversation",
            messages: []Message{...},
            want:     "Claude Code working...",
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            sc := NewStateCalculator()
            got := sc.DetermineConversationState(tt.messages, time.Now(), nil)
            if got != tt.want {
                t.Errorf("DetermineConversationState() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Run the automated test suites:

```bash
# Quick tests (7 automated tests)
./TEST_QUICK.sh

# Category search tests (9 tests)
./TEST_CATEGORIES.sh
```

## CI/CD Pipeline

### Continuous Integration

The project uses GitHub Actions for automated testing and quality checks. CI runs automatically on:

- **All branches**: When you push commits to any branch
- **Pull Requests**: When PRs are created or updated against `main`
- **Main branch**: On direct commits (only during releases)

### CI Workflow Jobs

The CI pipeline (`ci.yml`) includes four main jobs:

#### 1. Test Matrix

Runs comprehensive tests across multiple platforms and Go versions:

- **Operating Systems**: Ubuntu, macOS, Windows
- **Go Versions**: 1.23, 1.24
- **Features**:
  - Race condition detection (`-race` flag)
  - Code coverage tracking
  - Coverage threshold check (≥20%)
  - Codecov upload (Ubuntu + Go 1.24 only)

```bash
# Locally replicate CI tests
make test-verbose  # Runs tests with race detector
make test-coverage # Generates coverage report
```

#### 2. Linting

Ensures code quality and consistency:

- Runs `golangci-lint` with comprehensive checks
- Configuration: `.golangci.yml`
- Timeout: 5 minutes

```bash
# Install golangci-lint
brew install golangci-lint  # macOS
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run locally
make lint           # Lint with fallback to go vet
make lint-fix       # Lint with automatic fixes
```

#### 3. Build Check

Verifies the project builds successfully:

- Compiles the binary
- Runs `--help` to verify functionality
- Ensures no build errors

```bash
# Locally test build
make build
./cct --help
```

#### 4. Security Scan

Scans for security vulnerabilities:

- Uses `gosec` security scanner
- Generates SARIF reports
- Uploads to GitHub Security tab
- Does not fail CI (informational only)

### Scheduled Tests

Daily comprehensive testing runs at 8:00 AM UTC:

- **Workflow**: `scheduled-tests.yml`
- **Purpose**: Catch regressions and dependency issues
- **Matrix**: Same as CI (3 OS × 2 Go versions)
- **Coverage**: Full coverage reports with HTML output
- **Trigger**: Can be manually triggered via GitHub Actions UI

### Pre-Push Checklist

Before pushing your branch, run these commands locally:

```bash
# 1. Format code
make fmt

# 2. Run linter
make lint

# 3. Run tests with race detector
make test-verbose

# 4. Build binary
make build

# 5. Verify binary works
./cct --help
```

### CI Failures

If CI fails on your PR:

1. **Check the failure**: Click "Details" next to the failed check
2. **Review logs**: Identify which job failed and why
3. **Fix locally**: Reproduce and fix the issue locally
4. **Re-run tests**: Ensure `make test` and `make lint` pass
5. **Push fix**: Push the fix to your branch (CI will re-run automatically)

### GitHub Actions Secrets

Required secrets (already configured for maintainers):

- `CODECOV_TOKEN`: For uploading coverage to Codecov (optional)
- `GITHUB_TOKEN`: Automatically provided by GitHub

### Branch Protection

The `main` branch is protected with the following rules:

- ✅ All CI checks must pass before merging
- ✅ Pull request reviews required
- ✅ No direct commits to `main` (except releases)
- ✅ Branch must be up-to-date with `main`

### Local CI Testing

To test your changes like CI does:

```bash
# Full CI simulation
make fmt && \
make lint && \
make test-verbose && \
make test-coverage && \
make build

# Quick pre-push check
make fmt && make lint && make test
```

### Coverage Reports

Coverage reports are generated automatically:

- **Local**: Run `make test-coverage-html` to open coverage in browser
- **CI**: Coverage uploaded to Codecov for all PRs
- **Threshold**: Minimum 20% coverage enforced
- **Filtered**: Excludes `main.go`, static files, and interactive TUI

View coverage badge in README.md (updated on main branch).

## Commit Message Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
<type>: <subject>

<body>

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>
```

### Commit Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **refactor**: Code refactoring (no functional changes)
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (dependencies, build, etc.)
- **perf**: Performance improvements

### Examples

```bash
# Good commit messages
git commit -m "feat: add WebSocket support for real-time updates"
git commit -m "fix: resolve component installation 404 errors"
git commit -m "docs: update README with installation instructions"
git commit -m "refactor: simplify state calculation logic"

# Include body for complex changes
git commit -m "feat: add comprehensive category search

Implements automatic search across 25+ agent categories,
19+ command categories, and 9+ MCP categories to find
components without requiring full paths.

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"
```

## Pull Request Process

### Before Submitting

1. **Update documentation**: If your change affects user-facing behavior
2. **Add tests**: Ensure your changes are tested
3. **Run all tests**: `make test` and automated test scripts
4. **Format code**: Run `make fmt`
5. **Update CHANGELOG**: Add entry under "Unreleased" section
6. **Verify build**: Test on your platform
   ```bash
   make build
   ./cct --version
   ```

### Submitting Your PR

1. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request**: Go to GitHub and create a PR
   - Use the PR template
   - Provide clear description of changes
   - Link related issues (e.g., "Fixes #123")
   - Add screenshots for UI changes

3. **PR Title**: Follow conventional commit format
   ```text
   feat: add new feature
   fix: resolve bug
   docs: improve documentation
   ```

4. **Wait for review**: A maintainer will review your PR
   - Address review comments
   - Make requested changes
   - Push updates (no force push unless requested)

### PR Template Checklist

When submitting a PR, ensure you've completed:

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG.md updated
- [ ] Commit messages follow conventions
- [ ] No merge conflicts with main branch
- [ ] Build succeeds (`make build`)
- [ ] Tested locally

### Cross-Platform Testing

If possible, test on multiple platforms:

```bash
# Build for all platforms
make build-all
# or
just build-all

# Verify outputs in dist/
ls -lh dist/
```

## Getting Help

- **Documentation**: Check [CLAUDE.md](CLAUDE.md) for architecture details
- **Testing Guide**: See [TESTING.md](TESTING.md) for testing instructions
- **Questions**: Open a discussion or issue on GitHub
- **Original Project**: Reference [claude-code-templates](https://github.com/davila7/claude-code-templates)

## Recognition

Contributors will be:
- Listed in release notes
- Credited in commit history
- Acknowledged in project documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Go Claude Templates! Your efforts help make this tool better for everyone.
