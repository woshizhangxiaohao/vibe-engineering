#!/bin/bash
# Run Aider with issue context
# Environment variables required:
#   - OPENROUTER_API_KEY
#   - ISSUE_NUMBER
#   - ISSUE_TITLE
#   - ISSUE_BODY
#   - BRANCH_NAME (from previous step)

set -euo pipefail

# Validate required environment variables
if [ -z "${OPENROUTER_API_KEY:-}" ]; then
  echo "::error::OPENROUTER_API_KEY is not set"
  exit 1
fi

if [ -z "${ISSUE_NUMBER:-}" ] || [ -z "${ISSUE_TITLE:-}" ] || [ -z "${ISSUE_BODY:-}" ]; then
  echo "::error::Issue information is incomplete"
  exit 1
fi

# Write issue context to a temporary file to safely pass to aider
# This prevents any script injection through issue title or body
ISSUE_CONTEXT_FILE=$(mktemp)
trap 'rm -f "$ISSUE_CONTEXT_FILE"' EXIT

cat > "$ISSUE_CONTEXT_FILE" <<EOF
Issue #${ISSUE_NUMBER}: ${ISSUE_TITLE}

${ISSUE_BODY}
EOF

echo "::group::🤖 Running Aider"
echo "Issue: #${ISSUE_NUMBER}"
echo "Title: ${ISSUE_TITLE}"
echo "Model: openrouter/anthropic/claude-sonnet-4.5"
echo "API Provider: OpenRouter"
echo "::endgroup::"

# Configure Aider to use OpenRouter
# Aider uses litellm which supports OpenRouter when:
# 1. OPENROUTER_API_KEY is set (already set)
# 2. Model identifier uses 'openrouter/' prefix
# 3. API base URL may need to be set for some versions

# Export OPENROUTER_API_KEY for litellm to use
export OPENROUTER_API_KEY

# Run Aider in fully automated, non-interactive mode with OpenRouter
# Automation flags for maximum automation:
# --yes: Auto-confirm all prompts (required for headless execution)
# --auto-commits: Let Aider handle git commits automatically (creates commits as it works)
# --message: Pass issue context from file to prevent injection
# Model format: openrouter/anthropic/claude-sonnet-4.5 (OpenRouter format)
#
# Additional automation via environment variables:
# - NO_COLOR: Disable ANSI colors for cleaner CI logs
# - AIDER_NO_INPUT: Ensure non-interactive mode (if supported)
# - CI=true: Common CI environment flag that many tools respect

# Capture aider output to check for errors
AIDER_OUTPUT_FILE=$(mktemp)
trap 'rm -f "$AIDER_OUTPUT_FILE"' EXIT

# Set environment variables for maximum automation and clean output
export NO_COLOR=1           # Disable colors for cleaner logs
export CI=true              # Indicate CI environment (many tools auto-disable prompts)
export AIDER_NO_INPUT=1     # Ensure non-interactive mode (if supported by aider)

# Run aider in fully automated mode and capture both stdout and stderr
set +e  # Temporarily disable exit on error to handle aider's exit code
aider \
  --model openrouter/anthropic/claude-sonnet-4.5 \
  --yes \
  --auto-commits \
  --message "$(cat "$ISSUE_CONTEXT_FILE")" > "$AIDER_OUTPUT_FILE" 2>&1
AIDER_EXIT_CODE=$?
set -e  # Re-enable exit on error

# Display aider output
cat "$AIDER_OUTPUT_FILE"

# Check for real errors (authentication, API errors, etc.)
if [ $AIDER_EXIT_CODE -ne 0 ]; then
  echo "aider_success=false" >> "$GITHUB_OUTPUT"
  
  # Check if it's a real error (authentication, API failure, etc.)
  # Use set +e temporarily to avoid grep failure causing script failure
  set +e
  HAS_ERROR=$(grep -qiE "(error|failed|exception|authentication|invalid|missing|unauthorized|litellm\.AuthenticationError)" "$AIDER_OUTPUT_FILE" 2>/dev/null && echo "yes" || echo "no")
  set -e
  
  if [ "$HAS_ERROR" = "yes" ]; then
    echo "::error::Aider execution failed with critical errors"
    echo "Exit code: $AIDER_EXIT_CODE"
    echo "Error details found in output above"
    exit 1  # Fail the workflow on real errors
  else
    # Aider may exit with non-zero code even when it just made no changes
    # This is acceptable and shouldn't fail the workflow
    echo "::warning::Aider exited with code $AIDER_EXIT_CODE but no critical errors detected"
    echo "This may indicate no changes were needed (acceptable)"
  fi
else
  echo "aider_success=true" >> "$GITHUB_OUTPUT"
  echo "✅ Aider completed successfully"
fi

# Check if there are any changes
# Since Aider uses --auto-commits, changes are already committed
# We need to check for new commits, not uncommitted changes

BRANCH_NAME="${BRANCH_NAME:-}"
DEFAULT_BRANCH="${DEFAULT_BRANCH:-main}"
INITIAL_COMMIT="${INITIAL_COMMIT:-}"
CURRENT_COMMIT=$(git rev-parse HEAD)

echo "::group::🔍 Checking for changes"
echo "Current commit: $CURRENT_COMMIT"
if [ -n "$INITIAL_COMMIT" ]; then
  echo "Initial commit: $INITIAL_COMMIT"
fi
echo "Branch: $BRANCH_NAME"
echo "Default branch: $DEFAULT_BRANCH"
echo "::endgroup::"

# Check for uncommitted changes first (in case aider didn't commit everything)
HAS_UNCOMMITTED=false
if ! git diff --quiet || ! git diff --cached --quiet; then
  HAS_UNCOMMITTED=true
  echo "✅ Uncommitted changes detected"
fi

# Check for new commits
# Method 1: Compare with initial commit if available (most accurate for new branches)
HAS_NEW_COMMITS=false

if [ -n "$INITIAL_COMMIT" ] && [ "$INITIAL_COMMIT" != "$CURRENT_COMMIT" ]; then
  # Compare with initial commit (when branch was created)
  COMMITS_SINCE_INITIAL=$(git rev-list --count "$INITIAL_COMMIT"..HEAD 2>/dev/null || echo "0")
  if [ "$COMMITS_SINCE_INITIAL" -gt 0 ]; then
    HAS_NEW_COMMITS=true
    echo "✅ Found $COMMITS_SINCE_INITIAL new commit(s) since branch creation"
  fi
elif git ls-remote --heads origin "$BRANCH_NAME" 2>/dev/null | grep -q "$BRANCH_NAME"; then
  # Branch exists remotely, check for commits ahead
  git fetch origin "$BRANCH_NAME" 2>/dev/null || true
  COMMITS_AHEAD=$(git rev-list --count origin/"$BRANCH_NAME"..HEAD 2>/dev/null || echo "0")
  if [ "$COMMITS_AHEAD" -gt 0 ]; then
    HAS_NEW_COMMITS=true
    echo "✅ Found $COMMITS_AHEAD new commit(s) ahead of remote branch"
  fi
else
  # New branch, check commits compared to default branch
  # Fetch default branch first
  git fetch origin "$DEFAULT_BRANCH" 2>/dev/null || true
  COMMITS_COUNT=$(git rev-list --count origin/"$DEFAULT_BRANCH"..HEAD 2>/dev/null || echo "0")
  if [ "$COMMITS_COUNT" -gt 0 ]; then
    HAS_NEW_COMMITS=true
    echo "✅ Found $COMMITS_COUNT new commit(s) compared to $DEFAULT_BRANCH"
  fi
fi

# Determine if there are changes
if [ "$HAS_UNCOMMITTED" = "true" ] || [ "$HAS_NEW_COMMITS" = "true" ]; then
  echo "has_changes=true" >> "$GITHUB_OUTPUT"
  echo "✅ Changes detected"
else
  echo "has_changes=false" >> "$GITHUB_OUTPUT"
  echo "ℹ️  No changes detected (no uncommitted changes or new commits)"
fi
