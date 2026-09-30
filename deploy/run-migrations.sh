#!/usr/bin/env bash

set -euo pipefail

APP_ROOT="${APP_ROOT:?APP_ROOT is required}"
CONFIG_FILE="${CONFIG_FILE:-$APP_ROOT/config/config.server.yaml}"
MIGRATE_BINARY="${MIGRATE_BINARY:-$APP_ROOT/scripts/migrate}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$APP_ROOT/database/migrations}"

test -f "$MIGRATE_BINARY" || {
  echo "Migration binary not found: $MIGRATE_BINARY" >&2
  exit 1
}
test -r "$CONFIG_FILE" || {
  echo "Migration config is not readable: $CONFIG_FILE" >&2
  exit 1
}
test -d "$MIGRATIONS_DIR" || {
  echo "Migration directory not found: $MIGRATIONS_DIR" >&2
  exit 1
}

chmod 0755 "$MIGRATE_BINARY"

echo "Running database migrations from: $MIGRATIONS_DIR"
"$MIGRATE_BINARY" \
  -config "$CONFIG_FILE" \
  -migrations "$MIGRATIONS_DIR"
