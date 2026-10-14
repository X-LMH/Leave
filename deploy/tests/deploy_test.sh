#!/usr/bin/env bash
set -euo pipefail
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -eq 0 ]; then
  bash "$0" all test
  bash "$0" all prod
  echo 'Deployment coordination tests passed (only fixtures and mocks).'
  exit 0
fi
if [ "$1" = all ]; then
  deploy_env="$2"
  for scenario in back admin app admin-backend-down; do
    bash "$0" "$scenario" "$deploy_env"
  done
  for scenario in pid migration health admin-health; do
    status=0
    bash "$0" "$scenario" "$deploy_env" || status=$?
    test "$status" -eq 1 || { echo "Failure isolation failed: $scenario" >&2; exit 1; }
  done
  for scenario in backend-conflict admin-conflict apk-conflict backend-incomplete; do
    status=0
    bash "$0" "$scenario" "$deploy_env" || status=$?
    test "$status" -eq 1 || { echo "Conflict isolation failed: $scenario" >&2; exit 1; }
  done
  bash "$0" migration-conflict "$deploy_env"
  bash "$0" real-pid-mismatch "$deploy_env"
  bash "$0" restart-lock "$deploy_env"
  echo 'Deployment coordination tests passed (only fixtures and mocks).'
  exit 0
fi

scenario="$1"
DEPLOY_ENV="${2:-test}"
source "$TEST_DIR/../deploy-release.sh"
APP_ROOT=$(mktemp -d /tmp/leave-deploy-test-XXXXXX)
trap 'rm -rf -- "$APP_ROOT"' EXIT
BACKEND_DIR="$APP_ROOT/backend"
BACKEND_USER=www
TAG_PREFIX=test-
SYSTEM_ROOT=/www/wwwroot/Leave_test
DATABASE=leave_list_test
CONFIG_NAME=config.test.yaml
if [ "$DEPLOY_ENV" = prod ]; then
  BACKEND_DIR="$APP_ROOT"
  BACKEND_USER=root
  TAG_PREFIX=
  SYSTEM_ROOT=/www/wwwroot/Leave
  DATABASE=leave_list
  CONFIG_NAME=config.server.yaml
fi
RUN_ID=123-1
STAGE="$APP_ROOT/.deploy/$RUN_ID"
BUNDLE="$STAGE/bundle"
BACKUP="$APP_ROOT/backups/deploy/first"
CONFIG_FILE="$APP_ROOT/config/$CONFIG_NAME"
BACKEND_VERSION=1
ADMIN_VERSION=1
APK_NAME=leave-android-11-1.0.0.apk
ADMIN_HOST=fixture.invalid
ADMIN_PORT=80
if [ "$DEPLOY_ENV" = prod ]; then ADMIN_PORT=8001; fi
APP_PORT=19999
MODE=admin
BACKEND_RUNNING=1
if [ "$scenario" = admin-backend-down ]; then BACKEND_RUNNING=0; fi
if [ "$scenario" = back ]; then MODE=back; fi
if [ "$scenario" = app ] || [ "$scenario" = apk-conflict ]; then MODE=app; fi
mkdir -p "$BUNDLE"/{backend,tools,migrations,admin/assets,app} "$APP_ROOT"/{config,admin/releases/v0,packages/android,fixtures} "$BACKEND_DIR/releases/v0"
printf 'old server' > "$BACKEND_DIR/releases/v0/Leave"
printf 'new server' > "$BUNDLE/backend/Leave"
printf 'old admin' > "$APP_ROOT/admin/releases/v0/index.html"
printf '<script type="module" src="/assets/entry.js"></script>' > "$BUNDLE/admin/index.html"
printf 'entry' > "$BUNDLE/admin/assets/entry.js"
printf 'SELECT 1;' > "$BUNDLE/migrations/000001_fixture.up.sql"
printf '{"tag":"back-%sv1","mode":"back","backend_version":1}' "$TAG_PREFIX" > "$BUNDLE/release.json"
printf 'new apk' > "$BUNDLE/app/$APK_NAME"
printf 'fixture SQL' > "$BUNDLE/app/app-version.sql"
if [ "$MODE" = app ]; then
  printf '{"tag":"app-%sv1.0.0","mode":"app","backend_version":1,"app":{"version_name":"1.0.0","version_code":11,"apk_name":"leave-android-11-1.0.0.apk","release_notes":[]}}' "$TAG_PREFIX" > "$BUNDLE/release.json"
fi
ln -s releases/v0 "$BACKEND_DIR/current"
ln -s releases/v0 "$APP_ROOT/admin/current"
cat > "$CONFIG_FILE" <<YAML
mysql:
  host: 127.0.0.1
  port: 3306
  database: $DATABASE
  user: fixture
app:
  port: 19999
  package_dir: $SYSTEM_ROOT/packages/android
storage:
  root_dir: $SYSTEM_ROOT/uploads
YAML
export LOG_FILE="$APP_ROOT/events" FAIL_PHASE="$scenario" MYSQL_PASSWORD=fixture
cat > "$BUNDLE/tools/migrate" <<'SH'
#!/bin/sh
echo migration >> "$LOG_FILE"
test "$FAIL_PHASE" != migration
SH
cat > "$APP_ROOT/fixtures/mysql" <<'SH'
#!/bin/sh
cat >/dev/null
echo mysql >> "$LOG_FILE"
SH
chmod +x "$APP_ROOT/fixtures/mysql"
export PATH="$APP_ROOT/fixtures:$PATH"
chown() { :; }
stop_backend() {
  echo stop >> "$LOG_FILE"
  test "$FAIL_PHASE" != pid || return 1
  BACKEND_RUNNING=0
}
check_backend_process() {
  test "$FAIL_PHASE" != pid
}
start_backend() {
  echo start >> "$LOG_FILE"
  test "$FAIL_PHASE" != health || return 1
  BACKEND_RUNNING=1
}
curl() {
  test "$FAIL_PHASE" != admin-health || return 1
  if [ "${!#}" = "http://127.0.0.1:$ADMIN_PORT/api/v1/health" ]; then
    # 模拟 Nginx：后端停止时返回 502，对应 curl --fail 的退出码 22。
    test "$BACKEND_RUNNING" -eq 1 || return 22
  fi
  if [ "${!#}" = "http://127.0.0.1:$ADMIN_PORT/" ]; then cat "$BUNDLE/admin/index.html"; fi
}

if [ "$scenario" = migration-conflict ]; then
  mkdir -p "$APP_ROOT/database/migrations"
  printf 'different SQL' > "$APP_ROOT/database/migrations/000001_fixture.up.sql"
  if sync_migrations; then exit 1; fi
  test "$(cat "$BACKEND_DIR/current/Leave")" = 'old server'
  exit 0
fi
if [ "$scenario" = real-pid-mismatch ]; then
  # Use this shell PID, whose executable is not the fixture Leave program.
  source "$TEST_DIR/../switch-and-restart.sh"
  BAOTA_PID_FILE="$APP_ROOT/pid"
  echo $$ > "$BAOTA_PID_FILE"
  APP_PORT=19999
  if check_backend_process; then exit 1; fi
  if stop_backend; then exit 1; fi
  kill -0 $$
  exit 0
fi

if [ "$scenario" = restart-lock ]; then
  source "$TEST_DIR/../switch-and-restart.sh"
  BAOTA_START_SCRIPT=fixture
  exec 9>"$APP_ROOT/lock"
  runuser() { bash -c 'test ! -e /proc/self/fd/9'; }
  read_backend_pid() { echo 1; }
  verify_backend_process() { return 0; }
  start_backend
  test -e "/proc/$$/fd/9"
  exit 0
fi

if [[ "$scenario" =~ ^(backend-conflict|backend-incomplete|admin-conflict|apk-conflict)$ ]]; then
  case "$scenario" in
    backend-conflict|backend-incomplete)
      mkdir -p "$BACKEND_DIR/releases/v1"
      cp "$BUNDLE/release.json" "$BACKEND_DIR/releases/v1/release.json"
      cp "$BUNDLE/tools/migrate" "$BACKEND_DIR/releases/v1/migrate"
      if [ "$scenario" = backend-conflict ]; then printf 'different' > "$BACKEND_DIR/releases/v1/Leave"; fi
      ;;
    admin-conflict)
      mkdir -p "$APP_ROOT/admin/releases/v1"
      printf 'different' > "$APP_ROOT/admin/releases/v1/index.html"
      ;;
    apk-conflict)
      printf 'different' > "$APP_ROOT/packages/android/$APK_NAME"
      ;;
  esac
  trap 'code=$?; if [ -e "$LOG_FILE" ] || [ "$(cat "$BACKEND_DIR/current/Leave")" != "old server" ]; then exit 99; fi; exit "$code"' ERR
fi

if [[ "$scenario" =~ ^(pid|migration|health|admin-health)$ ]]; then
  assert_failed_release() {
    local code=$?
    test "$code" -ne 0
    test ! -f "$APP_ROOT/scripts/app-published"
    case "$scenario" in
      pid)
        test ! -e "$LOG_FILE" || exit 99
        ;;
      migration)
        ! grep -q stop "$LOG_FILE" || exit 99
        test "$(cat "$BACKEND_DIR/current/Leave")" = 'old server' || exit 99
        ;;
      health)
        test "$(cat "$APP_ROOT/admin/current/index.html")" = 'old admin' || exit 99
        ;;
      admin-health)
        test "$(cat "$BACKEND_DIR/current/Leave")" = 'new server' || exit 99
        test "$(cat "$APP_ROOT/admin/current/index.html")" = 'old admin' || exit 99
        grep -q '^start$' "$LOG_FILE" || exit 99
        ;;
    esac
    exit "$code"
  }
  trap assert_failed_release ERR
fi
readonly APP_ROOT CONFIG_FILE
run_release
test "$(cat "$BACKEND_DIR/current/Leave")" = 'new server'
test -f "$BACKEND_DIR/releases/v1/migrate"
if [ "$MODE" = admin ]; then cmp "$BUNDLE/admin/index.html" "$APP_ROOT/admin/current/index.html"; fi
if [ "$MODE" = app ]; then
  cmp "$BUNDLE/app/$APK_NAME" "$APP_ROOT/packages/android/$APK_NAME"
  grep -q mysql "$LOG_FILE"
fi
# Re-running identical artifacts reuses version files but still migrates and restarts.
BACKUP="$APP_ROOT/backups/deploy/second"
rm -rf -- "$STAGE/ready-admin"
# Metadata from another combination tag must not invalidate identical backend bytes.
printf '{"tag":"admin-%sv2","mode":"admin","backend_version":1,"admin_version":2}' "$TAG_PREFIX" > "$BACKEND_DIR/releases/v1/release.json"
before=$(stat -c '%i:%Y' "$BACKEND_DIR/current/Leave")
if [ "$scenario" = admin-backend-down ]; then BACKEND_RUNNING=0; fi
run_release
test "$(stat -c '%i:%Y' "$BACKEND_DIR/current/Leave")" = "$before"
test "$(grep -c '^start$' "$LOG_FILE")" -eq 2
test "$(cat "$BACKEND_DIR/current/Leave")" = 'new server'
echo "Same-version deployment/retry passed: $DEPLOY_ENV $scenario"
