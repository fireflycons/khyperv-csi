#!/usr/bin/env bash

echo "Checking conditions to skip workflow..."

# Get latest commit message
COMMIT_MSG=$(git log -1 --pretty=%B)
echo "Commit message: $COMMIT_MSG"

# Check for NO-CI
if echo "$COMMIT_MSG" | grep -q "NO-CI"; then
    echo "NO-CI found in commit message. Skipping."
    echo "skip=true" >> $GITHUB_OUTPUT
    exit 0
fi

# Get changed files
CHANGED=$(git diff --name-only HEAD~1)
echo "Changed files:"
echo "$CHANGED"

# If all changed files are markdown
if [ -n "$CHANGED" ] && ! echo "$CHANGED" | grep -vqE '\.md$'; then
    echo "Only markdown files changed. Skipping."
    echo "skip=true" >> $GITHUB_OUTPUT
    exit 0
fi

echo "No skip conditions met."
echo "skip=false" >> $GITHUB_OUTPUT
