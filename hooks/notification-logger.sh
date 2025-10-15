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
    TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcript_path // empty')
else
    # Fallback to basic parsing (less robust)
    SESSION_ID=$(echo "$INPUT" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4 || echo "")
    MESSAGE=$(echo "$INPUT" | grep -o '"message":"[^"]*"' | cut -d'"' -f4 || echo "")
    TRANSCRIPT_PATH=$(echo "$INPUT" | grep -o '"transcript_path":"[^"]*"' | cut -d'"' -f4 || echo "")
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
COMMAND_DETAILS=""

if echo "$MESSAGE" | grep -qi "permission"; then
    NOTIFICATION_TYPE="permission_request"
    # Extract tool name from message (e.g., "use Bash" -> "Bash")
    TOOL_NAME=$(echo "$MESSAGE" | sed -n 's/.*use \([A-Za-z_0-9-]*\).*/\1/p')

    # Extract actual command from transcript if available
    if [[ -n "$TRANSCRIPT_PATH" ]] && [[ -f "$TRANSCRIPT_PATH" ]] && command -v jq &> /dev/null; then
        # Get the last assistant message and extract tool_use from it
        # The structure is: .message.content[] where type == "tool_use"
        LAST_ASSISTANT=$(tail -20 "$TRANSCRIPT_PATH" | grep '"type":"assistant"' | tail -1)
        if [[ -n "$LAST_ASSISTANT" ]]; then
            # Extract tool_use from the assistant message content array
            if [[ "$TOOL_NAME" == "Bash" ]]; then
                COMMAND_DETAILS=$(echo "$LAST_ASSISTANT" | jq -r '.message.content[] | select(.type == "tool_use") | .input.command // empty' 2>/dev/null | head -1)
            else
                # For other tools, show the input as JSON
                COMMAND_DETAILS=$(echo "$LAST_ASSISTANT" | jq -c '.message.content[] | select(.type == "tool_use") | .input // {}' 2>/dev/null | head -1)
            fi
        fi
    fi
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

# Extract model information from environment variables
# Read ANTHROPIC_MODEL and ANTHROPIC_BASE_URL
MODEL_ID="${ANTHROPIC_MODEL:-}"
BASE_URL="${ANTHROPIC_BASE_URL:-}"

# Determine provider - send ANTHROPIC_BASE_URL as MODEL_PROVIDER
MODEL_PROVIDER="https://api.anthropic.com"
if [[ -n "$BASE_URL" ]]; then
    MODEL_PROVIDER="$BASE_URL"
fi

# Parse model name to human-readable format
MODEL_NAME="Unknown"
if [[ -n "$MODEL_ID" ]]; then
    # Remove claude- prefix if present
    MODEL_CLEAN="${MODEL_ID#claude-}"

    # Try pattern 1: model-major-minor (e.g., sonnet-4-5-20250929)
    if [[ "$MODEL_CLEAN" =~ ^(sonnet|opus|haiku)-([0-9]+)-([0-9]+)- ]]; then
        FAMILY="${BASH_REMATCH[1]}"
        MAJOR="${BASH_REMATCH[2]}"
        MINOR="${BASH_REMATCH[3]}"
        # Capitalize first letter
        FAMILY_CAP="$(echo "${FAMILY:0:1}" | tr '[:lower:]' '[:upper:]')${FAMILY:1}"
        MODEL_NAME="$FAMILY_CAP $MAJOR.$MINOR"
    # Try pattern 2: model-major (e.g., opus-4-20250514)
    elif [[ "$MODEL_CLEAN" =~ ^(sonnet|opus|haiku)-([0-9]+)- ]]; then
        FAMILY="${BASH_REMATCH[1]}"
        MAJOR="${BASH_REMATCH[2]}"
        FAMILY_CAP="$(echo "${FAMILY:0:1}" | tr '[:lower:]' '[:upper:]')${FAMILY:1}"
        MODEL_NAME="$FAMILY_CAP $MAJOR"
    # Try pattern 3: major-minor-model (e.g., 3-5-sonnet-20241022)
    elif [[ "$MODEL_CLEAN" =~ ^([0-9]+)-([0-9]+)-(sonnet|opus|haiku)- ]]; then
        MAJOR="${BASH_REMATCH[1]}"
        MINOR="${BASH_REMATCH[2]}"
        FAMILY="${BASH_REMATCH[3]}"
        FAMILY_CAP="$(echo "${FAMILY:0:1}" | tr '[:lower:]' '[:upper:]')${FAMILY:1}"
        MODEL_NAME="$FAMILY_CAP $MAJOR.$MINOR"
    # Try pattern 4: major-model (e.g., 3-sonnet-20241022)
    elif [[ "$MODEL_CLEAN" =~ ^([0-9]+)-(sonnet|opus|haiku)- ]]; then
        MAJOR="${BASH_REMATCH[1]}"
        FAMILY="${BASH_REMATCH[2]}"
        FAMILY_CAP="$(echo "${FAMILY:0:1}" | tr '[:lower:]' '[:upper:]')${FAMILY:1}"
        MODEL_NAME="$FAMILY_CAP $MAJOR"
    else
        # Fallback: just clean up the model ID
        # Remove date suffix (e.g., -20250929)
        MODEL_NAME=$(echo "$MODEL_CLEAN" | sed 's/-[0-9]\{8\}$//' | tr '-' ' ' | sed 's/\b\(.\)/\u\1/g')
    fi
fi

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
        --arg commandDetails "$COMMAND_DETAILS" \
        --arg cwd "$CWD" \
        --arg branch "$GIT_BRANCH" \
        --arg modelProvider "$MODEL_PROVIDER" \
        --arg modelName "$MODEL_NAME" \
        '{
            session_id: $session,
            session_name: $sessionName,
            notification_type: $notificationType,
            message: $message,
            tool_name: $toolName,
            command_details: $commandDetails,
            cwd: $cwd,
            branch: $branch,
            model_provider: $modelProvider,
            model_name: $modelName
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
  "command_details": "$COMMAND_DETAILS",
  "cwd": "$CWD",
  "branch": "$GIT_BRANCH",
  "model_provider": "$MODEL_PROVIDER",
  "model_name": "$MODEL_NAME"
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
