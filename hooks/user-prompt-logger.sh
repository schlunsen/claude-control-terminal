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

# Generate friendly session name from session_id
# Use a list of 10 South Park character names and hash the session_id to pick one
SESSION_NAMES=(
    "Cartman"
    "Stan"
    "Kyle"
    "Kenny"
    "Butters"
    "Randy"
    "Tweek"
    "Craig"
    "Token"
    "Wendy"
)

# Generate a numeric hash from session_id to pick a name (modulo 10)
# Use first 8 characters of session_id to get a stable hash
if command -v cksum &> /dev/null; then
    HASH=$(echo -n "$SESSION_ID" | cksum | cut -d' ' -f1)
    INDEX=$((HASH % 10))
else
    # Fallback: use character values
    HASH=0
    for ((i=0; i<${#SESSION_ID} && i<8; i++)); do
        CHAR="${SESSION_ID:$i:1}"
        ASCII=$(printf '%d' "'$CHAR")
        HASH=$((HASH + ASCII))
    done
    INDEX=$((HASH % 10))
fi

SESSION_NAME="${SESSION_NAMES[$INDEX]}"

# Analytics server endpoint (default port)
ANALYTICS_URL="http://localhost:3333/api/prompts"

# Build JSON payload
if command -v jq &> /dev/null; then
    # Use jq for proper JSON encoding
    PAYLOAD=$(jq -n \
        --arg session "$SESSION_ID" \
        --arg sessionName "$SESSION_NAME" \
        --arg prompt "$PROMPT" \
        --arg cwd "$CWD" \
        --arg branch "$GIT_BRANCH" \
        '{session_id: $session, session_name: $sessionName, prompt: $prompt, cwd: $cwd, branch: $branch}')
else
    # Fallback: basic JSON (less robust, but works for most cases)
    # Note: This doesn't escape special characters properly
    PAYLOAD=$(cat <<EOF
{
  "session_id": "$SESSION_ID",
  "session_name": "$SESSION_NAME",
  "prompt": "$PROMPT",
  "cwd": "$CWD",
  "branch": "$GIT_BRANCH"
}
EOF
)
fi

# POST to analytics server (run in background to not slow down Claude Code)
# Use curl if available, otherwise try wget
if command -v curl &> /dev/null; then
    curl -X POST "$ANALYTICS_URL" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" \
        &> /dev/null &
elif command -v wget &> /dev/null; then
    wget --quiet --post-data="$PAYLOAD" \
        --header="Content-Type: application/json" \
        -O /dev/null \
        "$ANALYTICS_URL" \
        &> /dev/null &
fi

# Exit successfully (don't block Claude Code)
exit 0
