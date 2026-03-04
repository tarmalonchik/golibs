#!/bin/bash

REMOTE="origin"
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
FORCE=false

if [[ -z "$BRANCH" ]]; then
    echo "Error: Not a git repository."
    exit 1
fi

if [[ "$1" == "-f" ]]; then
    FORCE=true
fi

echo "Sync branch [$BRANCH] with $REMOTE/$BRANCH..."

if [ "$FORCE" = false ]; then
    read -p "Do you really want to sync current branch with origin? (y/n) " -n 1 REPLY
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancel operation."
        exit 1
    fi
fi

echo "Running git fetch..."
git fetch $REMOTE

echo "Rollback to $REMOTE/$BRANCH..."
git reset --hard $REMOTE/$BRANCH

echo "Remove garbage (git clean)..."
git clean -fd

echo "Done!"