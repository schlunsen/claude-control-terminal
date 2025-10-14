#!/bin/bash
# filter-coverage.sh - Filter coverage report to exclude untestable files
#
# This script removes files from coverage that are:
# - Entry points (main.go)
# - Embedded static files
# - Interactive TUI components (hard to test)
#
# Usage: ./scripts/filter-coverage.sh coverage.out > filtered-coverage.out

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <coverage.out>"
    exit 1
fi

COVERAGE_FILE="$1"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "Error: Coverage file '$COVERAGE_FILE' not found"
    exit 1
fi

# Files to exclude from coverage (patterns)
EXCLUDE_PATTERNS=(
    "cmd/cct/main.go"                    # Entry point - just calls other code
    "internal/server/static.go"          # Embedded static files - no logic
    "internal/tui/tui.go"                # Interactive TUI launcher
    "internal/tui/claude_launcher.go"    # Interactive launcher
    "internal/tui/model.go"              # Bubbletea model - interactive
)

# Print coverage header (mode line)
head -n 1 "$COVERAGE_FILE"

# Filter out excluded files using grep (much faster than while loop)
tail -n +2 "$COVERAGE_FILE" | \
    grep -v "cmd/cct/main.go" | \
    grep -v "internal/server/static.go" | \
    grep -v "internal/tui/tui.go" | \
    grep -v "internal/tui/claude_launcher.go" | \
    grep -v "internal/tui/model.go"
