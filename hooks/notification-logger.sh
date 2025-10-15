#!/bin/bash
# Notification Logger Hook for Claude Code
#
# This hook captures notification events (permission requests and idle alerts)
# and stores them in the CCT analytics database for engagement tracking.
#
# Hook Type: Notification
# Input: JSON on stdin with session_id, message
# Output: Silent (no stdout/stderr unless error)

set -euo pipefail

# Read JSON from stdin
INPUT=$(cat)

# Parse JSON fields using jq if available, otherwise use grep/sed
if command -v jq &> /dev/null; then
    SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
    MESSAGE=$(echo "$INPUT" | jq -r '.message // empty')
    CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
else
    # Fallback to basic parsing (less robust)
    SESSION_ID=$(echo "$INPUT" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4 || echo "")
    MESSAGE=$(echo "$INPUT" | grep -o '"message":"[^"]*"' | cut -d'"' -f4 || echo "")
    CWD=""
fi

# Validate required fields
if [[ -z "$SESSION_ID" ]] || [[ -z "$MESSAGE" ]]; then
    # Silent failure - don't block Claude Code
    exit 0
fi

# Get CWD from environment if not in JSON
if [[ -z "$CWD" ]]; then
    CWD=$(pwd)
fi

# Get git branch from working directory
GIT_BRANCH=""
if [[ -d "$CWD/.git" ]]; then
    GIT_BRANCH=$(cd "$CWD" && git branch --show-current 2>/dev/null || echo "")
fi

# Determine notification type and extract tool name if applicable
NOTIFICATION_TYPE="other"
TOOL_NAME=""

if echo "$MESSAGE" | grep -qi "permission"; then
    NOTIFICATION_TYPE="permission_request"
    # Extract tool name from message (e.g., "use Bash" -> "Bash")
    TOOL_NAME=$(echo "$MESSAGE" | grep -oP '(?<=use )\w+' || echo "")
elif echo "$MESSAGE" | grep -qi "waiting.*input"; then
    NOTIFICATION_TYPE="idle_alert"
fi

# Generate friendly session name from session_id
# Use a list of 25 South Park character names and hash the session_id to pick one
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
    "Sheila"
    "Sharon"
    "Chef"
    "Mr-Garrison"
    "Mr-Mackey"
    "Jimmy"
    "Timmy"
    "Bebe"
    "Clyde"
    "Ike"
    "PC-Principal"
    "Towelie"
    "Mr-Hankey"
    "Big-Gay-Al"
    "Satan"
)

# Generate a numeric hash from session_id to pick a name (modulo 25)
if command -v cksum &> /dev/null; then
    HASH=$(echo -n "$SESSION_ID" | cksum | cut -d' ' -f1)
    INDEX=$((HASH % 25))
else
    # Fallback: use character values
    HASH=0
    for ((i=0; i<${#SESSION_ID} && i<8; i++)); do
        CHAR="${SESSION_ID:$i:1}"
        ASCII=$(printf '%d' "'$CHAR")
        HASH=$((HASH + ASCII))
    done
    INDEX=$((HASH % 25))
fi

SESSION_NAME="${SESSION_NAMES[$INDEX]}"

# Analytics server endpoint
NOTIFICATION_ENDPOINT="http://localhost:3333/api/notifications"

# Build JSON payload
if command -v jq &> /dev/null; then
    PAYLOAD=$(jq -n \
        --arg session "$SESSION_ID" \
        --arg sessionName "$SESSION_NAME" \
        --arg notificationType "$NOTIFICATION_TYPE" \
        --arg message "$MESSAGE" \
        --arg toolName "$TOOL_NAME" \
        --arg cwd "$CWD" \
        --arg branch "$GIT_BRANCH" \
        '{
            session_id: $session,
            session_name: $sessionName,
            notification_type: $notificationType,
            message: $message,
            tool_name: $toolName,
            cwd: $cwd,
            branch: $branch
        }')
else
    # Fallback: basic JSON (escape issues possible)
    PAYLOAD=$(cat <<EOF
{
  "session_id": "$SESSION_ID",
  "session_name": "$SESSION_NAME",
  "notification_type": "$NOTIFICATION_TYPE",
  "message": "$MESSAGE",
  "tool_name": "$TOOL_NAME",
  "cwd": "$CWD",
  "branch": "$GIT_BRANCH"
}
EOF
)
fi

# POST to notifications endpoint
if command -v curl &> /dev/null; then
    curl -X POST "$NOTIFICATION_ENDPOINT" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" \
        &> /dev/null &
elif command -v wget &> /dev/null; then
    wget --quiet --post-data="$PAYLOAD" \
        --header="Content-Type: application/json" \
        -O /dev/null \
        "$NOTIFICATION_ENDPOINT" \
        &> /dev/null &
fi

# Exit successfully (don't block Claude Code)
exit 0
