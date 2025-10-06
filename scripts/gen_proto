#!/usr/bin/env bash
set -euo pipefail

if ! command -v buf &>/dev/null; then
  echo "❌ Buf not found. Please install it: https://buf.build/docs/installation"
  exit 1
fi

echo "🔧 Generating protobuf code..."
buf generate

if [ -f "buf.gen.postprocess.yaml" ]; then
  echo "🧩 Running postprocess generators..."
  buf generate --template "buf.gen.postprocess.yaml"
fi

echo "✅ Done!"

