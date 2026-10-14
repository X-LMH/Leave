#!/usr/bin/env bash
set -euo pipefail
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -eq 0 ]; then
  for scenario in back admin app; do
    bash "$0" "$scenario"
  done
  for scenario in pid migration health admin-health; do
    status=0
    bash "$0" "$scenario" || status=$?
    test "$status" -eq 1 || { echo "Failure isolation failed: $scenario" >&2; exit 1; }
  done
  for scenario in backend-conflict admin-conflict apk-conflict backend-incomplete; do
    status=0
    bash "$0" "$scenario" || status=$?
    test "$status" -eq 1 || { echo "Conflict isolation failed: $scenario" >&2; exit 1; }
  done
  bash "$0" migration-conflict
  bash "$0" real-pid-mismatch
  bash "$0" restart-lock
  echo 'Deployment coordination tests passed (only fixtures and mocks).'
  exit 0
fi

scenario="$1"
source "$TEST_DIR/../deploy-release.sh"
APP_ROOT=$(mktemp -d /tmp/leave-deploy-test-XXXXXX)
trap 'rm -rf -- "$APP_ROOT"' EXIT
RUN_ID=123-1
STAGE="$APP_ROOT/.deploy/$RUN_ID"
BUNDLE="$STAGE/bundle"
BACKUP="$APP_ROOT/backups/deploy/first"
CONFIG_FILE="$APP_ROOT/config/config.test.yaml"
BACKEND_VERSION=1
ADMIN_VERSION=1
APK_NAME=leave-android-11-1.0.0.apk
ADMIN_HOST=fixture.invalid
APP_PORT=19999
MODE=admin
if [ "$scenario" = back ]; then MODE=back; fi
if [ "$scenario" = app ] || [ "$scenario" = apk-conflict ]; then MODE=app; fi
mkdir -p "$BUNDLE"/{backend,tools,migrations,admin/assets,app} "$APP_ROOT"/{config,backend/releases/v0,admin/releases/v0,packages/android,fixtures}
printf 'old server' > "$APP_ROOT/backend/releases/v0/Leave"
printf 'new server' > "$BUNDLE/backend/Leave"
printf 'old admin' > "$APP_ROOT/admin/releases/v0/index.html"
printf '<script type="module" src="/assets/entry.js"></script>' > "$BUNDLE/admin/index.html"
printf 'entry' > "$BUNDLE/admin/assets/entry.js"
printf 'SELECT 1;' > "$BUNDLE/migrations/000001_fixture.up.sql"
printf '{"tag":"back-test-v1","mode":"back","backend_version":1}' > "$BUNDLE/release.json"
printf 'new apk' > "$BUNDLE/app/$APK_NAME"
printf 'fixture SQL' > "$BUNDLE/app/app-version.sql"
if [ "$MODE" = app ]; then
  printf '{"tag":"app-test-v1.0.0","mode":"app","backend_version":1,"app":{"version_name":"1.0.0","version_code":11,"apk_name":"leave-android-11-1.0.0.apk","release_notes":[]}}' > "$BUNDLE/release.json"
fi
ln -s releases/v0 "$APP_ROOT/backend/current"
ln -s releases/v0 "$APP_ROOT/admin/current"
cat > "$CONFIG_FILE" <<'YAML'
mysql:
  host: 127.0.0.1
  port: 3306
  database: leave_list_test
  user: fixture
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
  test "$FAIL_PHASE" != pid
}
check_backend_process() {
  test "$FAIL_PHASE" != pid
}
start_backend() {
  echo start >> "$LOG_FILE"
  test "$FAIL_PHASE" != health
}
curl() {
  test "$FAIL_PHASE" != admin-health || return 1
  if [ "${!#}" = http://127.0.0.1/ ]; then cat "$BUNDLE/admin/index.html"; fi
}

if [ "$scenario" = migration-conflict ]; then
  mkdir -p "$APP_ROOT/database/migrations"
  printf 'different SQL' > "$APP_ROOT/database/migrations/000001_fixture.up.sql"
  if sync_migrations; then exit 1; fi
  test "$(cat "$APP_ROOT/backend/current/Leave")" = 'old server'
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
  read_test_pid() { echo 1; }
  verify_test_process() { return 0; }
  start_backend
  test -e "/proc/$$/fd/9"
  exit 0
fi

if [[ "$scenario" =~ ^(backend-conflict|backend-incomplete|admin-conflict|apk-conflict)$ ]]; then
  case "$scenario" in
    backend-conflict|backend-incomplete)
      mkdir -p "$APP_ROOT/backend/releases/v1"
      cp "$BUNDLE/release.json" "$APP_ROOT/backend/releases/v1/release.json"
      cp "$BUNDLE/tools/migrate" "$APP_ROOT/backend/releases/v1/migrate"
      if [ "$scenario" = backend-conflict ]; then printf 'different' > "$APP_ROOT/backend/releases/v1/Leave"; fi
      ;;
    admin-conflict)
      mkdir -p "$APP_ROOT/admin/releases/v1"
      printf 'different' > "$APP_ROOT/admin/releases/v1/index.html"
      ;;
    apk-conflict)
      printf 'different' > "$APP_ROOT/packages/android/$APK_NAME"
      ;;
  esac
  trap 'code=$?; if [ -e "$LOG_FILE" ] || [ "$(cat "$APP_ROOT/backend/current/Leave")" != "old server" ]; then exit 99; fi; exit "$code"' ERR
fi

if [[ "$scenario" =~ ^(pid|migration|health|admin-health)$ ]]; then
  trap 'code=$?; test "$code" -ne 0; test ! -f "$APP_ROOT/scripts/app-published"; if [ "$scenario" = pid ]; then test ! -e "$LOG_FILE" || exit 99; fi; if [ "$scenario" = migration ]; then ! grep -q stop "$LOG_FILE" || exit 99; test "$(cat "$APP_ROOT/backend/current/Leave")" = "old server" || exit 99; fi; if [ "$scenario" = health ]; then test "$(cat "$APP_ROOT/admin/current/index.html")" = "old admin"; fi; exit "$code"' ERR
fi
readonly APP_ROOT CONFIG_FILE
run_release
test "$(cat "$APP_ROOT/backend/current/Leave")" = 'new server'
test -f "$APP_ROOT/backend/releases/v1/migrate"
if [ "$MODE" = admin ]; then cmp "$BUNDLE/admin/index.html" "$APP_ROOT/admin/current/index.html"; fi
if [ "$MODE" = app ]; then
  cmp "$BUNDLE/app/$APK_NAME" "$APP_ROOT/packages/android/$APK_NAME"
  grep -q mysql "$LOG_FILE"
fi
# Re-running identical artifacts reuses version files but still migrates and restarts.
BACKUP="$APP_ROOT/backups/deploy/second"
rm -rf -- "$STAGE/ready-admin"
# Metadata from another combination tag must not invalidate identical backend bytes.
printf '{"tag":"admin-test-v2","mode":"admin","backend_version":1,"admin_version":2}' > "$APP_ROOT/backend/releases/v1/release.json"
before=$(stat -c '%i:%Y' "$APP_ROOT/backend/current/Leave")
run_release
test "$(stat -c '%i:%Y' "$APP_ROOT/backend/current/Leave")" = "$before"
test "$(grep -c '^start$' "$LOG_FILE")" -eq 2
test "$(cat "$APP_ROOT/backend/current/Leave")" = 'new server'
echo "Same-version deployment/retry passed: $scenario"
