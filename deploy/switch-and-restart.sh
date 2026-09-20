#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="${APP_ROOT:?APP_ROOT is required}"
VERSION_NAME="${VERSION_NAME:?VERSION_NAME is required}"
release_dir="$APP_ROOT/releases/v$VERSION_NAME"
current_dir="$APP_ROOT/current"
old_current_dir="$APP_ROOT/.current-directory-$VERSION_NAME"

test -f "$release_dir/Leave"
test -r "$APP_ROOT/.env" || { echo "Environment file not readable: $APP_ROOT/.env" >&2; exit 1; }
test -r "$APP_ROOT/config/config.yaml" || { echo "Shared config is not readable: $APP_ROOT/config/config.yaml" >&2; exit 1; }
app_port=$(awk -F': *' '$1 == "  port" {print $2; exit}' "$APP_ROOT/config/config.yaml")
app_port=${app_port//\"/}
app_port="${APP_PORT:-${app_port:-10000}}"
case "$app_port" in
  ''|*[!0-9]*) echo "Invalid app port: $app_port" >&2; exit 1 ;;
esac

set -a
. "$APP_ROOT/.env"
set +a
test -n "${JWT_SECRET:-}" || { echo "JWT_SECRET is empty in $APP_ROOT/.env" >&2; exit 1; }
test -n "${MYSQL_PASSWORD:-}" || { echo "MYSQL_PASSWORD is empty in $APP_ROOT/.env" >&2; exit 1; }

command -v fuser >/dev/null 2>&1 || { echo "fuser is required" >&2; exit 1; }
fuser -k "$app_port/tcp" 2>/dev/null || true
sleep 2

if [ -L "$current_dir" ]; then
  next_current="$APP_ROOT/.current-$VERSION_NAME"
  ln -sfn "releases/v$VERSION_NAME" "$next_current"
  mv -Tf "$next_current" "$current_dir"
elif [ -d "$current_dir" ]; then
  test ! -e "$old_current_dir"
  mv "$current_dir" "$old_current_dir"
  ln -s "releases/v$VERSION_NAME" "$current_dir"
else
  ln -s "releases/v$VERSION_NAME" "$current_dir"
fi

test -L "$current_dir"
test "$(readlink -f "$current_dir")" = "$(readlink -f "$release_dir")"
chmod 0755 "$current_dir/Leave"
cd "$APP_ROOT"
nohup "$APP_ROOT/current/Leave" >/dev/null 2>&1 &

for _ in 1 2 3 4 5 6 7 8 9 10; do
  if curl --fail --silent --show-error "http://127.0.0.1:$app_port/api/v1/health" >/dev/null; then
    exit 0
  fi
  sleep 1
done

echo "Leave did not become healthy on port $app_port" >&2
exit 1
