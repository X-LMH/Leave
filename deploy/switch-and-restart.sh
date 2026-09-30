#!/usr/bin/env bash
set -euo pipefail

APP_ROOT="${APP_ROOT:?APP_ROOT is required}"
VERSION_NAME="${VERSION_NAME:?VERSION_NAME is required}"
release_dir="$APP_ROOT/releases/v$VERSION_NAME"
current_dir="$APP_ROOT/current"
old_current_dir="$APP_ROOT/.current-directory-$VERSION_NAME"

test -f "$release_dir/Leave"
test -r "$APP_ROOT/.env" || { echo "Environment file not readable: $APP_ROOT/.env" >&2; exit 1; }
test -r "$APP_ROOT/config/config.server.yaml" || { echo "Server config is not readable: $APP_ROOT/config/config.server.yaml" >&2; exit 1; }
app_port=$(awk -F': *' '$1 == "  port" {print $2; exit}' "$APP_ROOT/config/config.server.yaml")
app_port=${app_port//\"/}
app_port="${APP_PORT:-${app_port:-10000}}"
case "$app_port" in
  ''|*[!0-9]*) echo "Invalid app port: $app_port" >&2; exit 1 ;;
esac

set -a
. "$APP_ROOT/.env"
set +a
baota_binary="$current_dir/Leave"
baota_pid_file="${BAOTA_GO_PID_FILE:-/var/tmp/gopids/Leave.pid}"
baota_start_script="${BAOTA_GO_START_SCRIPT:-/www/server/go_project/vhost/scripts/Leave.sh}"
test -n "${JWT_SECRET:-}" || { echo "JWT_SECRET is empty in $APP_ROOT/.env" >&2; exit 1; }
test -n "${MYSQL_PASSWORD:-}" || { echo "MYSQL_PASSWORD is empty in $APP_ROOT/.env" >&2; exit 1; }

test -f "$baota_start_script" || { echo "Baota Go project start script not found: $baota_start_script" >&2; exit 1; }
test -d "$(dirname "$baota_pid_file")" || { echo "Baota Go project PID directory not found: $(dirname "$baota_pid_file")" >&2; exit 1; }

# Baota owns the process lifecycle. Stop only the PID recorded for its Leave
# project, then switch the binary path used by that project's start script.
if [ -s "$baota_pid_file" ]; then
  old_pid=$(cat "$baota_pid_file")
  case "$old_pid" in
    ''|*[!0-9]*) echo "Invalid Baota Go project PID in $baota_pid_file" >&2; exit 1 ;;
  esac
  if [ -r "/proc/$old_pid/cmdline" ]; then
    old_command=$(tr '\000' ' ' < "/proc/$old_pid/cmdline")
    case "$old_command" in
      *"$baota_binary"*) kill "$old_pid" ;;
      *) echo "PID $old_pid is not running $baota_binary; refusing to stop it" >&2; exit 1 ;;
    esac
    for _ in 1 2 3 4 5 6 7 8 9 10; do
      kill -0 "$old_pid" 2>/dev/null || break
      sleep 1
    done
    if kill -0 "$old_pid" 2>/dev/null; then
      echo "Baota Go project process $old_pid did not stop" >&2
      exit 1
    fi
  fi
fi

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
export VERSION="$VERSION_NAME"
bash "$baota_start_script"

for _ in 1 2 3 4 5 6 7 8 9 10; do
  if [ -s "$baota_pid_file" ]; then
    running_pid=$(cat "$baota_pid_file")
    case "$running_pid" in
      ''|*[!0-9]*) echo "Invalid Baota Go project PID in $baota_pid_file" >&2; exit 1 ;;
    esac
    if [ -r "/proc/$running_pid/cmdline" ]; then
      running_command=$(tr '\000' ' ' < "/proc/$running_pid/cmdline")
      running_binary=$(readlink -f "/proc/$running_pid/exe" 2>/dev/null || true)
      expected_binary=$(readlink -f "$release_dir/Leave")
      case "$running_command" in
        *"$baota_binary"*)
          if [ "$running_binary" = "$expected_binary" ] && \
            curl --fail --silent --show-error "http://127.0.0.1:$app_port/api/v1/health" >/dev/null; then
            exit 0
          fi
          ;;
      esac
    fi
  fi
  sleep 1
done

echo "Baota Leave project did not start release v$VERSION_NAME healthy on port $app_port" >&2
exit 1
