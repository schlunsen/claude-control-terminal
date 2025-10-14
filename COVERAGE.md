# Test Coverage Report

This document tracks test coverage for the claude-control-terminal project, identifies files excluded from coverage metrics, and provides recommendations for improving coverage.

## Current Coverage Status

| Package | Coverage | Status |
|---------|----------|--------|
| **Overall** | **30.1%** | 🟡 Needs Improvement |
| **Filtered** | **32.3%** | 🟡 Needs Improvement |
| cmd/cct | 0.0% | ⚪ Excluded (entry point) |
| internal/analytics | 42.3% | 🟡 Fair |
| internal/cmd | 25.9% | 🔴 Low |
| internal/components | 38.4% | 🟡 Fair |
| internal/database | 43.3% | 🟡 Fair |
| internal/docker | 48.0% | 🟢 Good |
| internal/fileops | 44.5% | 🟡 Fair |
| internal/installer | 38.3% | 🟡 Fair |
| internal/providers | 54.5% | 🟢 Good |
| internal/server | 3.1% | 🔴 Critical |
| internal/tui | 16.3% | ⚪ Excluded (interactive) |
| internal/websocket | 50.0% | 🟢 Good |

**Target**: 60% filtered coverage (excluding untestable files)

## Running Coverage

### Quick Coverage Check
```bash
make test-coverage
```

This will:
1. Run all tests with coverage profiling
2. Filter out untestable files (main.go, static files, interactive TUI)
3. Generate HTML coverage report at `coverage.html`
4. Display both total and filtered coverage percentages

### View Coverage Report
```bash
make test-coverage-html
```

Opens the coverage report in your browser.

### Manual Coverage Analysis
```bash
# Run tests with coverage
go test -coverprofile=coverage.out -covermode=atomic ./...

# View overall summary
go tool cover -func=coverage.out | tail -1

# View filtered summary
./scripts/filter-coverage.sh coverage.out > coverage-filtered.out
go tool cover -func=coverage-filtered.out | tail -1

# View HTML report
go tool cover -html=coverage.out
```

## Files Excluded from Coverage

The following files are excluded from filtered coverage metrics because they are untestable or not worth testing:

### Entry Points
- **cmd/cct/main.go** - Application entry point; just bootstraps the application

### Static/Embedded Files
- **internal/server/static.go** - Serves embedded static files; no business logic

### Interactive Components
- **internal/tui/tui.go** - Interactive TUI launcher; requires user interaction
- **internal/tui/claude_launcher.go** - Interactive launcher; requires user interaction
- **internal/tui/model.go** - Bubbletea model; interactive state management

**Rationale**: These files contain interactive UI code that cannot be effectively unit tested. Integration/E2E tests would be required.

### How Exclusion Works

The `scripts/filter-coverage.sh` script removes excluded files from the coverage report:

```bash
./scripts/filter-coverage.sh coverage.out > coverage-filtered.out
```

This creates a filtered coverage report that more accurately reflects testable code coverage.

## Critical Files Needing Tests

### 🔴 Priority 1: Critical (0-10% coverage)

#### internal/websocket/websocket.go (0.0%)
**Impact**: HIGH - Core analytics feature
**Functions needing tests**:
- `NewHub()` - Hub creation
- `Run()` - Hub event loop
- `Broadcast()` - Message broadcasting
- `HandleWebSocket()` - WebSocket handler

**Recommendation**: Add tests for:
- Hub creation and initialization
- Client registration/unregistration
- Message broadcasting
- Graceful shutdown

#### internal/server/server.go (3.1%)
**Impact**: HIGH - Core analytics/API server
**Functions needing tests**:
- `Setup()` - Server initialization
- `setupRoutes()` - Route registration
- All API handlers (handleGetData, handleGetStats, etc.)
- Reset handlers (handleResetArchive, handleResetClear, etc.)

**Recommendation**: Add tests for:
- API endpoint responses
- Error handling
- Reset operations
- Server lifecycle (start/shutdown)

### 🟡 Priority 2: Important (10-30% coverage)


#### internal/cmd/root.go (25.9%)
**Impact**: MEDIUM - CLI command handling
**Recommendation**: Add tests for:
- Flag parsing
- Command routing
- Error handling

### 🟢 Priority 3: Moderate (30-50% coverage)

#### internal/analytics/conversation_analyzer.go (35.9%)
**Functions at 0%**:
- `LoadConversations()`
- `parseConversationFile()`
- `parseMessages()`
- `ArchiveConversations()`
- `ClearConversations()`

**Recommendation**: Add integration tests for:
- Conversation file parsing
- Archive operations
- Clear operations

#### internal/analytics/file_watcher.go (0% - part of 35.9% package)
**Recommendation**: Add tests for:
- File watcher setup
- Change detection
- Periodic refresh

#### internal/components/mcp.go (38.4% package)
**Recommendation**: Add tests for:
- MCP installation
- Configuration handling
- Error scenarios

## Testing Strategy

### Unit Tests (Current Focus)
- Test individual functions and methods in isolation
- Mock external dependencies (filesystem, network, processes)
- Focus on business logic and data transformations

### Integration Tests (Future)
- Test component interactions
- Test file operations with temporary directories
- Test API endpoints with test server

### E2E Tests (Future)
- Test complete workflows
- Test CLI commands end-to-end
- Test analytics dashboard functionality

## Improving Coverage

### Quick Wins (Easy to test, high impact)

1. **Add websocket tests** - Test hub, broadcast, client management
   ```bash
   touch internal/websocket/websocket_test.go
   ```

2. **Add server handler tests** - Test API endpoints
   ```bash
   # Already exists: internal/server/server_test.go
   # Add more handler tests
   ```


### Medium Effort

4. **Add conversation analyzer tests** - Test file parsing
5. **Add file watcher tests** - Test change detection
6. **Add root command tests** - Test flag handling

### Lower Priority

7. **TUI component tests** - Complex due to interactive nature
8. **Static file tests** - Low value

## Coverage Goals

| Milestone | Filtered Coverage | Target Date |
|-----------|-------------------|-------------|
| Current | 30.7% | - |
| Phase 1 | 45% | Add websocket + server tests |
| Phase 2 | 55% | Add wrapper + analytics tests |
| Phase 3 | 60% | Add remaining core tests |
| Stretch | 70%+ | Comprehensive coverage |

## Notes

- **Filtered coverage** is the primary metric, as it excludes untestable files
- **Total coverage** includes all files, including entry points and interactive TUI
- Focus on testing business logic, not UI interactions
- Use table-driven tests for multiple scenarios
- Mock external dependencies (filesystem, network, system calls)

## CI/CD Integration

Consider adding coverage checks to CI:

```yaml
# .github/workflows/test.yml
- name: Run tests with coverage
  run: |
    go test -coverprofile=coverage.out -covermode=atomic ./...
    ./scripts/filter-coverage.sh coverage.out > coverage-filtered.out

- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage-filtered.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Coverage: $COVERAGE%"
    if (( $(echo "$COVERAGE < 45" | bc -l) )); then
      echo "Coverage below 45% threshold"
      exit 1
    fi
```

## Resources

- [Go Testing](https://golang.org/pkg/testing/)
- [Go Coverage](https://go.dev/blog/cover)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Test Fixtures](https://github.com/go-testfixtures/testfixtures)

---

**Last Updated**: 2025-10-14
**Next Review**: When coverage reaches 45%
