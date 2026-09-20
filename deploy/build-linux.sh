#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$PROJECT_DIR/bin"
GOARCH="${1:-amd64}"

mkdir -p "$OUTPUT_DIR"

echo "Building Linux binary: linux/$GOARCH"

cd "$PROJECT_DIR/backend"
CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" go build -o "$OUTPUT_DIR/Leave" ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" go build -o "$OUTPUT_DIR/migrate" ./cmd/migrate

chmod +x "$OUTPUT_DIR/Leave" "$OUTPUT_DIR/migrate"
echo "Build complete: $OUTPUT_DIR/Leave"
echo "Build complete: $OUTPUT_DIR/migrate"
