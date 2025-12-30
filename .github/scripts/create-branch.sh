#!/bin/bash
# Create or checkout branch for issue
# Usage: create-branch.sh <issue_number> <default_branch>

set -euo pipefail

ISSUE_NUMBER="${1:-}"
DEFAULT_BRANCH="${2:-main}"

if [ -z "$ISSUE_NUMBER" ]; then
  echo "::error::Issue number is required"
  exit 1
fi

BRANCH_NAME="feat/issue-${ISSUE_NUMBER}"

# Ensure we're on the default branch first
git checkout "$DEFAULT_BRANCH" 2>/dev/null || git checkout -b "$DEFAULT_BRANCH"
git pull origin "$DEFAULT_BRANCH" || true

# Check if branch exists remotely
if git ls-remote --heads origin "$BRANCH_NAME" | grep -q "$BRANCH_NAME"; then
  echo "Branch $BRANCH_NAME already exists remotely. Checking it out."
  git fetch origin "$BRANCH_NAME":"$BRANCH_NAME" || true
  git checkout "$BRANCH_NAME"
  echo "branch_exists=true" >> "$GITHUB_OUTPUT"
elif git show-ref --verify --quiet "refs/heads/$BRANCH_NAME"; then
  echo "Branch $BRANCH_NAME exists locally. Checking it out."
  git checkout "$BRANCH_NAME"
  echo "branch_exists=true" >> "$GITHUB_OUTPUT"
else
  echo "Creating new branch: $BRANCH_NAME"
  git checkout -b "$BRANCH_NAME"
  echo "branch_exists=false" >> "$GITHUB_OUTPUT"
fi

echo "branch_name=$BRANCH_NAME" >> "$GITHUB_OUTPUT"

# Record the initial HEAD commit for change detection later
INITIAL_COMMIT=$(git rev-parse HEAD)
echo "initial_commit=$INITIAL_COMMIT" >> "$GITHUB_OUTPUT"

echo "✅ Branch ready: $BRANCH_NAME"
