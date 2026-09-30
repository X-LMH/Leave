#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
APP_ROOT="${APP_ROOT:-$PROJECT_DIR}"
CONFIG_FILE="${CONFIG_FILE:-$APP_ROOT/config/config.server.yaml}"
RELEASE_SQL="${RELEASE_SQL:?RELEASE_SQL is required}"

command -v mysql >/dev/null 2>&1 || { echo "mysql client is required" >&2; exit 1; }
test -r "$CONFIG_FILE" || { echo "Config file not found: $CONFIG_FILE" >&2; exit 1; }
test -r "$RELEASE_SQL" || { echo "Release SQL not found: $RELEASE_SQL" >&2; exit 1; }

mapfile -t database_config < <(ruby -ryaml -rjson -e '
  config = YAML.safe_load(File.read(ARGV[0]), permitted_classes: [], aliases: false)
  mysql = config.fetch("mysql")
  %w[host port database user password].each { |key| puts mysql.fetch(key, "") }
' "$CONFIG_FILE")

test "${#database_config[@]}" -eq 5 || { echo "Invalid mysql config" >&2; exit 1; }
mysql_host="${database_config[0]}"
mysql_port="${database_config[1]}"
mysql_database="${database_config[2]}"
mysql_user="${database_config[3]}"
mysql_password="${MYSQL_PASSWORD:-${database_config[4]}}"

test -n "$mysql_host" && test -n "$mysql_port" && test -n "$mysql_database" && test -n "$mysql_user" || {
  echo "Incomplete mysql config" >&2
  exit 1
}

echo "Publishing app version from: $RELEASE_SQL"
MYSQL_PWD="$mysql_password" mysql \
  --protocol=tcp \
  --host="$mysql_host" \
  --port="$mysql_port" \
  --user="$mysql_user" \
  "$mysql_database" < "$RELEASE_SQL"
