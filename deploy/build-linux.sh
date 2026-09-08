#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$PROJECT_DIR/bin"
OUTPUT_FILE="$OUTPUT_DIR/Leave"
GOARCH="${1:-amd64}"

mkdir -p "$OUTPUT_DIR"

echo "Building Linux binary: linux/$GOARCH"

cd "$PROJECT_DIR/backend"
CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" go build -o "$OUTPUT_FILE" .

chmod +x "$OUTPUT_FILE"
echo "Build complete: $OUTPUT_FILE"
