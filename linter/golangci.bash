#!/bin/bash

REMOTE_CONFIG_URL=https://raw.githubusercontent.com/tarmalonchik/golibs/main/linter/.golangci.yml

BASE_CONFIG=".golangci.base.yml"
LOCAL_TEMPLATE=".golangci-template.yml"
FINAL_CONFIG=".golangci.yml"

if ! command -v yq &> /dev/null; then
    echo "❌ spruce is not installed. Please install spruce: ( ~ brew tap starkandwayne/cf; brew install spruce )"
    exit 1
fi

echo "Downloading remote linter file ..."
curl -sSL "$REMOTE_CONFIG_URL" -o "$BASE_CONFIG"

if [ $? -ne 0 ]; then
    echo "❌ error while downloading remote linter config"
    exit 1
fi

if [ -f "$LOCAL_TEMPLATE" ]; then
    echo "🔗 merging local linter file with remote (local have priority)..."
    spruce merge "$BASE_CONFIG" "$LOCAL_TEMPLATE" > "$FINAL_CONFIG"
else
    echo "⚠️ Local template $LOCAL_TEMPLATE is not found. Using remote linter config!"
    cp "$BASE_CONFIG" "$FINAL_CONFIG"
fi

rm "$BASE_CONFIG"

echo "✅ Target file $FINAL_CONFIG is successfully generated!"