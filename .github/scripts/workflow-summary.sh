#!/bin/bash
# Generate workflow summary
# Environment variables required:
#   - ISSUE_NUMBER
#   - BRANCH_NAME
#   - HAS_CHANGES
#   - PR_EXISTS
#   - PR_NUMBER (optional)
#   - PR_URL (optional)

set -euo pipefail

ISSUE_NUMBER="${ISSUE_NUMBER:-}"
BRANCH_NAME="${BRANCH_NAME:-}"
HAS_CHANGES="${HAS_CHANGES:-}"
PR_EXISTS="${PR_EXISTS:-}"

{
  echo "## 🤖 AI Coder Workflow Summary"
  echo ""
  echo "**Issue:** #${ISSUE_NUMBER}"
  echo "**Branch:** \`${BRANCH_NAME}\`"
  echo ""
  
  if [ "$HAS_CHANGES" == "true" ]; then
    echo "✅ **Changes:** Detected"
  else
    echo "ℹ️  **Changes:** None detected"
  fi
  
  if [ "$PR_EXISTS" == "true" ]; then
    PR_NUMBER="${PR_NUMBER:-}"
    PR_URL="${PR_URL:-}"
    echo "✅ **PR:** Already exists - [#${PR_NUMBER}](${PR_URL})"
  elif [ "$PR_EXISTS" == "false" ]; then
    PR_NUMBER="${PR_NUMBER:-}"
    PR_URL="${PR_URL:-}"
    echo "✅ **PR:** Created - [#${PR_NUMBER}](${PR_URL})"
  elif [ "$PR_EXISTS" == "error" ]; then
    echo "❌ **PR:** Failed to create"
  fi
  
  echo ""
  echo "---"
  echo "*Workflow run: ${GITHUB_RUN_ID:-unknown}*"
} >> "$GITHUB_STEP_SUMMARY"
