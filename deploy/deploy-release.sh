#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="${APP_ROOT:?APP_ROOT is required}"
VERSION_NAME="${VERSION_NAME:?VERSION_NAME is required}"
release_dir="$APP_ROOT/releases/v$VERSION_NAME"
release_binary="$release_dir/Leave"
publish_script="$release_dir/publish-version.sh"
switch_script="$release_dir/switch-and-restart.sh"

test -f "$release_binary"
test -r "$APP_ROOT/.env" || { echo "Environment file not readable: $APP_ROOT/.env" >&2; exit 1; }
test -r "$APP_ROOT/config/config.yaml" || { echo "Shared config is not readable: $APP_ROOT/config/config.yaml" >&2; exit 1; }
test -f "$publish_script" || { echo "Publish script not found: $publish_script" >&2; exit 1; }
test -f "$switch_script" || { echo "Switch script not found: $switch_script" >&2; exit 1; }

set -a
. "$APP_ROOT/.env"
set +a
test -n "${JWT_SECRET:-}" || { echo "JWT_SECRET is empty in $APP_ROOT/.env" >&2; exit 1; }
test -n "${MYSQL_PASSWORD:-}" || { echo "MYSQL_PASSWORD is empty in $APP_ROOT/.env" >&2; exit 1; }

chmod 0755 "$publish_script" "$switch_script"
APP_ROOT="$APP_ROOT" \
CONFIG_FILE="$APP_ROOT/config/config.yaml" \
MANIFEST_FILE="$release_dir/manifest.json" \
RELEASE_FILE="$release_dir/android.yaml" \
BUILD_SQL_SCRIPT="$release_dir/build-version-sql.rb" \
  "$publish_script"

APP_ROOT="$APP_ROOT" VERSION_NAME="$VERSION_NAME" "$switch_script"
