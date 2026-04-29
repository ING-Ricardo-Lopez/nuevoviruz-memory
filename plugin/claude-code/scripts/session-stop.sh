#!/bin/bash
# Engram — Stop hook for Claude Code (async)
#
# Marks the session as ended via the HTTP API.
# Runs async so it doesn't block Claude's response.

NV_PORT="${NV_PORT:-7437}"
NV_URL="http://127.0.0.1:${NV_PORT}"

INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')

if [ -z "$SESSION_ID" ]; then
  exit 0
fi

curl -sf "${NV_URL}/sessions/${SESSION_ID}/end" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{}' \
  > /dev/null 2>&1

exit 0
