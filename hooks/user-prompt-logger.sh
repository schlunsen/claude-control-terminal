#!/bin/bash
# User Prompt Logger Hook for Claude Code
#
# This hook captures user prompts submitted to Claude Code and stores them
# in the CCT analytics database for tracking and analysis.
#
# Hook Type: UserPromptSubmit
# Input: JSON on stdin with session_id, cwd, prompt
# Output: Silent (no stdout/stderr unless error)

set -euo pipefail

# Read JSON from stdin
INPUT=$(cat)

# Parse JSON fields using jq if available, otherwise use grep/sed
if command -v jq &> /dev/null; then
    SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
    PROMPT=$(echo "$INPUT" | jq -r '.prompt // empty')
    CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
else
    # Fallback to basic parsing (less robust)
    SESSION_ID=$(echo "$INPUT" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
    PROMPT=$(echo "$INPUT" | grep -o '"prompt":"[^"]*"' | cut -d'"' -f4)
    CWD=$(echo "$INPUT" | grep -o '"cwd":"[^"]*"' | cut -d'"' -f4)
fi

# Validate required fields
if [[ -z "$SESSION_ID" ]] || [[ -z "$PROMPT" ]] || [[ -z "$CWD" ]]; then
    # Silent failure - don't block Claude Code
    exit 0
fi

# Get git branch from working directory
GIT_BRANCH=""
if [[ -d "$CWD/.git" ]]; then
    GIT_BRANCH=$(cd "$CWD" && git branch --show-current 2>/dev/null || echo "")
fi

# Find cct binary
CCT_BIN=""
if command -v cct &> /dev/null; then
    CCT_BIN="cct"
elif [[ -x "/usr/local/bin/cct" ]]; then
    CCT_BIN="/usr/local/bin/cct"
elif [[ -x "$HOME/go/bin/cct" ]]; then
    CCT_BIN="$HOME/go/bin/cct"
elif [[ -x "./cct" ]]; then
    CCT_BIN="./cct"
fi

# If cct not found, try to find it in common locations
if [[ -z "$CCT_BIN" ]]; then
    for path in /usr/local/bin ~/go/bin ~/.local/bin /opt/homebrew/bin; do
        if [[ -x "$path/cct" ]]; then
            CCT_BIN="$path/cct"
            break
        fi
    done
fi

# If still not found, silently exit (don't block Claude Code)
if [[ -z "$CCT_BIN" ]]; then
    exit 0
fi

# Call cct to record the prompt (run in background to not slow down Claude Code)
"$CCT_BIN" record-prompt \
    --session "$SESSION_ID" \
    --prompt "$PROMPT" \
    --cwd "$CWD" \
    --branch "$GIT_BRANCH" \
    &> /dev/null &

# Exit successfully (don't block Claude Code)
exit 0
