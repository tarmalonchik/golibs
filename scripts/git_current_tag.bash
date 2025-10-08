#!/bin/bash

git fetch --tags

latestVersion=$(git ls-remote --tags --sort=creatordate | grep -o 'v[0-9]*.[0-9]*.[0-9]*$' | tac | head -1)

echo "$latestVersion" 2>&1
