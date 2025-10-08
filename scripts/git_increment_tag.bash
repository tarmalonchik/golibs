#!/bin/bash

git fetch --tags

latestVersion=$(git ls-remote --tags --sort=creatordate | grep -o 'v[0-9]*.[0-9]*.[0-9]*$' | tac | head -1)

incrementType=bug

major=0
minor=0
build=0

# break down the version number into it's components
regex="([0-9]+).([0-9]+).([0-9]+)"
if [[ $latestVersion =~ $regex ]]
then
  major="${BASH_REMATCH[1]}"
  minor="${BASH_REMATCH[2]}"
  build="${BASH_REMATCH[3]}"
fi

if [ "$incrementType" == "feature" ]
then
  minor=$(echo "$minor" + 1 | bc)
elif [ "$incrementType" == "bug" ]
then
  build=$(echo "$build" + 1 | bc)
elif [ "$incrementType" == "major" ]
then
  major=$(echo "$major" + 1 | bc)
fi

newVersion=${major}.${minor}.${build}
newVersionText=v$newVersion
git tag -a "$newVersionText" -m "$newVersionText"
git push origin "$newVersionText"

echo "$newVersionText" 2>&1
