#!/bin/bash
# Push changes to remote branch with retry logic
# Usage: push-changes.sh <branch_name> <default_branch>

set -euo pipefail

BRANCH_NAME="${1:-}"
DEFAULT_BRANCH="${2:-main}"

if [ -z "$BRANCH_NAME" ]; then
  echo "::error::Branch name is required"
  exit 1
fi

echo "::group::📤 Pushing changes"

# Verify we're on the correct branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "$BRANCH_NAME" ]; then
  echo "::error::Expected branch $BRANCH_NAME but on $CURRENT_BRANCH"
  exit 1
fi

# Check if there are commits to push
if git ls-remote --heads origin "$BRANCH_NAME" | grep -q "$BRANCH_NAME"; then
  # Branch exists remotely, check if we have new commits
  git fetch origin "$BRANCH_NAME" 2>/dev/null || true
  COMMITS_AHEAD=$(git rev-list --count origin/"$BRANCH_NAME"..HEAD 2>/dev/null || echo "0")
  if [ "$COMMITS_AHEAD" -eq 0 ]; then
    echo "ℹ️  No new commits to push"
    echo "push_success=true" >> "$GITHUB_OUTPUT"
    echo "::endgroup::"
    exit 0
  fi
  echo "📝 Found $COMMITS_AHEAD new commit(s) to push"
else
  # New branch, check if we have any commits
  COMMITS_COUNT=$(git rev-list --count HEAD ^origin/"$DEFAULT_BRANCH" 2>/dev/null || echo "0")
  if [ "$COMMITS_COUNT" -eq 0 ]; then
    echo "ℹ️  No commits to push"
    echo "push_success=false" >> "$GITHUB_OUTPUT"
    echo "::endgroup::"
    exit 0
  fi
  echo "📝 Found $COMMITS_COUNT commit(s) to push"
fi

# Push changes with retry logic
MAX_RETRIES=3
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if git push -u origin "$BRANCH_NAME"; then
    echo "✅ Successfully pushed to $BRANCH_NAME"
    echo "push_success=true" >> "$GITHUB_OUTPUT"
    echo "::endgroup::"
    exit 0
  else
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
      echo "⚠️  Push failed, retrying ($RETRY_COUNT/$MAX_RETRIES)..."
      sleep 2
    else
      echo "::error::Failed to push changes after $MAX_RETRIES attempts"
      echo "push_success=false" >> "$GITHUB_OUTPUT"
      echo "::endgroup::"
      exit 1
    fi
  fi
done
