#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/switch-and-restart.sh"

set_phase() {
  PHASE="$1"
  printf '%s\n' "$PHASE" > "$STAGE/status"
  echo "$DEPLOY_ENV deployment: $PHASE"
}

# Refuse symlinked persistent directories that could redirect deployment writes.
check_directory() {
  local path="$1"
  test ! -L "$path" || { echo "Symlinked deployment directory: $path" >&2; return 1; }
  if [ -e "$path" ]; then
    test -d "$path" && [[ "$(realpath "$path")" == "$APP_ROOT/"* ]] || return 1
  fi
}

sync_migrations() {
  local source target
  mkdir -p "$APP_ROOT/database/migrations"
  # Check the entire set before copying any file.
  for source in "$BUNDLE"/migrations/*.sql; do
    target="$APP_ROOT/database/migrations/$(basename "$source")"
    test ! -L "$target" || return 1
    if [ -e "$target" ] && ! cmp -s "$source" "$target"; then
      echo "Historical migration differs: $(basename "$source")" >&2
      return 1
    fi
  done
  for source in "$BUNDLE"/migrations/*.sql; do
    target="$APP_ROOT/database/migrations/$(basename "$source")"
    if [ ! -e "$target" ]; then install -m 0644 "$source" "$target"; fi
  done
}

prepare_backend() {
  mkdir -p "$STAGE/ready-backend"
  install -m 0755 "$BUNDLE/backend/Leave" "$STAGE/ready-backend/Leave"
  cp "$BUNDLE/release.json" "$STAGE/ready-backend/release.json"
  chown -R "$BACKEND_USER:$BACKEND_USER" "$STAGE/ready-backend"
}

check_versions() {
  ruby "$SCRIPT_DIR/release.rb" check-version backend "$BUNDLE" "$BACKEND_DIR/releases/v$BACKEND_VERSION"
  if [ "$MODE" = admin ]; then
    ruby "$SCRIPT_DIR/release.rb" check-version admin "$BUNDLE" "$APP_ROOT/admin/releases/v$ADMIN_VERSION"
  elif [ "$MODE" = app ]; then
    ruby "$SCRIPT_DIR/release.rb" check-version app "$BUNDLE" "$APP_ROOT/packages/android/$APK_NAME"
  fi
}

install_version() {
  local ready="$1" target="$2"
  check_directory "$target"
  if [ -d "$target" ]; then
    echo "Reusing verified version: $target"
    return 0
  fi
  mkdir -p "$(dirname "$target")"
  mv -T "$ready" "$target"
}

switch_current() {
  local module="$1" version="$2" directory
  directory="$APP_ROOT/$module"
  if [ "$module" = backend ]; then directory="$BACKEND_DIR"; fi
  local link="$directory/current" next="$directory/.current-$RUN_ID"
  test ! -e "$next" && test ! -L "$next" || return 1
  if [ -e "$link" ] && [ ! -L "$link" ]; then
    echo 'current must be a symlink, not a directory' >&2
    return 1
  fi
  ln -s "releases/v$version" "$next"
  mv -Tf "$next" "$link"
}

publish_admin() {
  if [ ! -e "$APP_ROOT/admin/releases/v$ADMIN_VERSION" ]; then
    cp -a "$BUNDLE/admin" "$STAGE/ready-admin"
    chown -R www:www "$STAGE/ready-admin"
    find "$STAGE/ready-admin" -type d -exec chmod 0755 {} +
    find "$STAGE/ready-admin" -type f -exec chmod 0644 {} +
    install_version "$STAGE/ready-admin" "$APP_ROOT/admin/releases/v$ADMIN_VERSION"
  fi
  switch_current admin "$ADMIN_VERSION"
  # Check the existing Nginx vhost without depending on external DNS.
  curl --fail --silent --show-error --max-time 10 -H "Host: $ADMIN_HOST" "http://127.0.0.1:$ADMIN_PORT/" > "$STAGE/admin-served.html"
  cmp "$BUNDLE/admin/index.html" "$STAGE/admin-served.html"
  local asset assets
  assets=$(ruby -e 'html=File.read(ARGV[0]); paths=html.scan(/(?:src|href)=["\x27](\/[^"\x27]+\.(?:js|css)(?:\?[^"\x27]*)?)["\x27]/).flatten; abort "No entry assets in admin HTML" if paths.empty?; puts paths.uniq' "$BUNDLE/admin/index.html")
  while IFS= read -r asset; do
    curl --fail --silent --show-error --max-time 10 -H "Host: $ADMIN_HOST" "http://127.0.0.1:$ADMIN_PORT$asset" >/dev/null
  done <<< "$assets"
  curl --fail --silent --show-error --max-time 10 -H "Host: $ADMIN_HOST" "http://127.0.0.1:$ADMIN_PORT/api/v1/health" >/dev/null
}

publish_app() {
  local target="$APP_ROOT/packages/android/$APK_NAME"
  test ! -L "$target" || return 1
  mkdir -p "$APP_ROOT/packages/android"
  if [ ! -e "$target" ]; then
    install -m 0644 "$BUNDLE/app/$APK_NAME" "$APP_ROOT/packages/android/.$APK_NAME-$RUN_ID"
    mv -T "$APP_ROOT/packages/android/.$APK_NAME-$RUN_ID" "$target"
  fi
  env APP_ROOT="$APP_ROOT" CONFIG_FILE="$CONFIG_FILE" RELEASE_SQL="$BUNDLE/app/app-version.sql" \
    DEPLOY_ENV="$DEPLOY_ENV" \
    bash "$SCRIPT_DIR/publish-version.sh"
}

run_release() {
  set_phase checking-versions
  check_versions
  set_phase preparing-backend
  check_backend_process
  prepare_backend
  # Install immutable files before stopping the service; current still points to the old version.
  install_version "$STAGE/ready-backend" "$BACKEND_DIR/releases/v$BACKEND_VERSION"
  set_phase checking-redis
  check_redis
  set_phase migrations
  sync_migrations
  chmod 0755 "$BUNDLE/scripts/migrate"
  env APP_ROOT="$APP_ROOT" CONFIG_FILE="$CONFIG_FILE" MIGRATE_BINARY="$BUNDLE/scripts/migrate" \
    bash "$SCRIPT_DIR/run-migrations.sh"
  set_phase stopping-backend
  stop_backend
  set_phase replacing-backend
  switch_current backend "$BACKEND_VERSION"
  mkdir -p "$APP_ROOT/scripts"
  install -m 0755 "$BUNDLE/scripts/migrate" "$APP_ROOT/scripts/.migrate-$RUN_ID"
  mv -Tf "$APP_ROOT/scripts/.migrate-$RUN_ID" "$APP_ROOT/scripts/migrate"
  set_phase restarting-backend
  start_backend
  if [ "$MODE" = admin ]; then
    set_phase checking-admin-site
    # 同域健康检查依赖后端；先完成恢复，再检查代理，失败时不切换 admin。
    curl --fail --silent --show-error --max-time 10 -H "Host: $ADMIN_HOST" "http://127.0.0.1:$ADMIN_PORT/api/v1/health" >/dev/null
    set_phase publishing-admin
    publish_admin
  elif [ "$MODE" = app ]; then
    set_phase publishing-app
    publish_app
  fi
  set_phase complete
}

main() {
  test "$#" -ge 1 && test "$#" -le 2 && [[ "$1" =~ ^[0-9]+-[0-9]+$ ]] || { echo 'Expected GitHub RUN_ID-RUN_ATTEMPT [test|prod]' >&2; return 1; }
  # Keep the old one-argument entry confined to test. Production is explicit.
  DEPLOY_ENV="${2:-test}"
  local profile
  profile=$(ruby "$SCRIPT_DIR/release.rb" environment "$DEPLOY_ENV")
  mapfile -t deployment <<< "$profile"
  APP_ROOT="${deployment[0]}"
  RUN_ID="$1"
  STAGE="$APP_ROOT/.deploy/$RUN_ID"
  BUNDLE="$STAGE/bundle"
  CONFIG_FILE="$APP_ROOT/config/${deployment[1]}"
  BAOTA_PID_FILE="/var/tmp/gopids/${deployment[2]}.pid"
  BAOTA_START_SCRIPT="/www/server/go_project/vhost/scripts/${deployment[2]}.sh"
  ADMIN_HOST="${deployment[3]}"
  BACKEND_DIR="${deployment[4]}"
  BACKEND_USER="${deployment[5]}"
  ADMIN_PORT="${deployment[6]}"
  test "$(id -u)" -eq 0 || { echo 'Deployment requires the existing privileged SSH account' >&2; return 1; }
  test "$(realpath "$APP_ROOT")" = "$APP_ROOT" && test ! -L "$APP_ROOT"
  test "$(realpath "$STAGE")" = "$STAGE" && test "$(realpath "$BUNDLE")" = "$BUNDLE"
  for command in ruby flock mysql runuser curl ss sha256sum; do command -v "$command" >/dev/null; done
  exec 9>"$APP_ROOT/.deploy/deploy.lock"
  flock -x 9
  local metadata
  metadata=$(ruby "$SCRIPT_DIR/release.rb" verify "$BUNDLE")
  mapfile -t fields <<< "$metadata"
  test "${fields[4]}" = "$DEPLOY_ENV" || { echo 'Bundle environment does not match deployment target' >&2; return 1; }
  MODE="${fields[0]}"
  BACKEND_VERSION="${fields[1]}"
  ADMIN_VERSION="${fields[2]:-}"
  APK_NAME="${fields[3]:-}"
  test -r "$APP_ROOT/.env" && test -r "$BAOTA_START_SCRIPT"
  for directory in config admin admin/releases database database/migrations packages packages/android scripts backups backups/deploy; do
    check_directory "$APP_ROOT/$directory"
  done
  if [ "$BACKEND_DIR" != "$APP_ROOT" ]; then check_directory "$BACKEND_DIR"; fi
  check_directory "$BACKEND_DIR/releases"
  for module in backend admin; do
    directory="$APP_ROOT/$module"
    if [ "$module" = backend ]; then directory="$BACKEND_DIR"; fi
    if [ -e "$directory/current" ] && [ ! -L "$directory/current" ]; then
      echo "$module/current must be a symlink" >&2
      return 1
    fi
    if [ -L "$directory/current" ]; then
      [[ "$(realpath "$directory/current")" == "$directory/releases/"* ]] || return 1
    fi
  done
  APP_PORT=$(ruby "$SCRIPT_DIR/release.rb" server-config "$CONFIG_FILE" "$DEPLOY_ENV")
  # The service's automatic config search must resolve to the same YAML as migrations.
  for config in config.local.yaml config.test.yaml config.server.yaml; do
    if [ "$config" != "${deployment[1]}" ] && [ -e "$APP_ROOT/config/$config" ]; then
      echo "Unexpected environment config in deployment root: $config" >&2
      return 1
    fi
  done
  readonly DEPLOY_ENV APP_ROOT BACKEND_DIR BACKEND_USER RUN_ID STAGE BUNDLE CONFIG_FILE BAOTA_PID_FILE BAOTA_START_SCRIPT ADMIN_HOST ADMIN_PORT APP_PORT MODE BACKEND_VERSION ADMIN_VERSION APK_NAME
  set -a
  source "$APP_ROOT/.env"
  set +a
  # Viper's automatic environment lookup must not redirect migrations to another DB.
  unset MYSQL_HOST MYSQL_PORT MYSQL_DATABASE MYSQL_USER
  test -n "${MYSQL_PASSWORD:-}" && test -n "${JWT_SECRET:-}"
  BACKUP="$APP_ROOT/backups/deploy/$RUN_ID-$(date +%s%N)"
  mkdir -p "$BACKUP"
  PHASE=preflight
  trap 'status=$?; printf "failed: %s (exit %s)\n" "$PHASE" "$status" > "$STAGE/status"; echo "Deployment failed at $PHASE; artifacts retained in $STAGE; backups in $BACKUP" >&2; exit "$status"' ERR
  run_release
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then main "$@"; fi
