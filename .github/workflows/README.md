# GitHub Actions Workflows

This directory contains the GitHub Actions workflows for the claude-control-terminal project.

## Workflows Overview

### 1. CI Workflow (`ci.yml`)

**Triggers:**
- Push to `main` branch
- Push to any feature branch (`feature/**`, `fix/**`, `refactor/**`, etc.)
- Pull requests to `main`
- Manual trigger via GitHub Actions UI

**Jobs:**
- **Test Matrix**: Runs tests on Ubuntu, macOS, and Windows with Go 1.23 and 1.24
  - Race condition detection
  - Code coverage (20% minimum threshold)
  - Codecov upload (Ubuntu + Go 1.24 only)
- **Linting**: Runs `golangci-lint` with comprehensive checks
- **Build Check**: Verifies the project builds successfully
- **Security Scan**: Uses `gosec` to scan for security vulnerabilities
- **CI Summary**: Aggregates results and posts PR comments

**Status Checks:**
All jobs must pass before a PR can be merged into `main`.

### 2. Scheduled Tests (`scheduled-tests.yml`)

**Triggers:**
- Daily at 8:00 AM UTC
- Manual trigger via GitHub Actions UI

**Jobs:**
- **Test Matrix**: Same as CI workflow (3 OS × 2 Go versions)
- **Linting**: Same as CI workflow
- **Test Summary**: Aggregates results with detailed summary

**Purpose:**
Catch regressions and dependency issues that may occur over time.

### 3. Release Workflow (`release.yml`)

**Triggers:**
- Push tags matching `v*` (e.g., `v1.0.0`)
- Manual trigger via GitHub Actions UI

**Jobs:**
- **Build Linux**: Builds binaries for Linux (amd64, arm64)
- **Build Darwin**: Builds binaries for macOS (amd64, arm64)
- **Release**: Creates GitHub release with binaries and Homebrew formula

**Artifacts:**
- Linux binaries: `cct-linux-amd64`, `cct-linux-arm64`
- macOS binaries: `cct-darwin-amd64`, `cct-darwin-arm64`
- Homebrew formula: `cct.rb`
- SHA256 checksums

### 4. Deploy Website (`deploy-website.yml`)

**Triggers:**
- Push to `main` branch (when website files change)

**Purpose:**
Deploys the project website to GitHub Pages.

## Branch Strategy

### Main Branch Protection

The `main` branch is protected with:
- ✅ All CI checks must pass
- ✅ Pull request reviews required
- ✅ No direct commits (except releases)
- ✅ Branch must be up-to-date

### Feature Branches

All development should be done on feature branches:

```bash
# Feature branch naming conventions
feature/*   # New features
fix/*       # Bug fixes
refactor/*  # Code refactoring
docs/*      # Documentation updates
test/*      # Test additions/updates
chore/*     # Maintenance tasks
```

CI runs automatically on all branch pushes.

## Secrets Configuration

Required secrets (configured by maintainers):

| Secret | Purpose | Required |
|--------|---------|----------|
| `CODECOV_TOKEN` | Upload coverage to Codecov | Optional |
| `GITHUB_TOKEN` | GitHub API access | Auto-provided |

## Local Testing

Before pushing your branch, replicate CI checks locally:

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests with race detector
make test-verbose

# Run tests with coverage
make test-coverage

# Build binary
make build

# Verify binary
./cct --help
```

## CI Failure Debugging

If CI fails on your PR:

1. **Check Logs**: Click "Details" next to the failed check
2. **Identify Issue**: Review the specific job that failed
3. **Reproduce Locally**: Run the same commands locally
4. **Fix and Push**: Push the fix (CI will re-run automatically)

### Common Failures

**Test Failures:**
```bash
# Run tests locally
make test-verbose
```

**Lint Failures:**
```bash
# Run linter with autofix
make lint-fix
```

**Build Failures:**
```bash
# Build locally
make build
```

**Coverage Below Threshold:**
```bash
# Check coverage
make test-coverage
# Aim for >20% overall coverage
```

## Workflow Maintenance

### Adding New Workflows

1. Create `.github/workflows/workflow-name.yml`
2. Follow existing workflow patterns
3. Test with manual trigger first
4. Update this README

### Modifying Existing Workflows

1. Create a feature branch
2. Modify the workflow file
3. Test via manual trigger or push
4. Create PR for review

### Best Practices

- ✅ Use caching for Go modules (`cache: true`)
- ✅ Run jobs in parallel when possible
- ✅ Use matrix builds for multi-platform testing
- ✅ Set appropriate timeouts
- ✅ Provide clear job names
- ✅ Use GitHub Actions marketplace actions when appropriate
- ✅ Document secrets and environment variables

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Configuration](.golangci.yml)
- [Contributing Guide](../../CONTRIBUTING.md)
- [Testing Guide](../../TESTING.md)
