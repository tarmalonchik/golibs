#!/bin/bash

git fetch --tags

latestVersion=$(git ls-remote --tags --sort=creatordate | grep -o 'v[0-9]*.[0-9]*.[0-9]*$' | tail -r | head -1)

echo "$latestVersion" 2>&1
