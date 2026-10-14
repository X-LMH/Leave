#!/usr/bin/env bash
set -euo pipefail
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -eq 0 ]; then
  for scenario in back admin app; do
    bash "$0" "$scenario"
  done
  for scenario in pid migration health admin-health; do
    if bash "$0" "$scenario"; then
      echo "Expected failure: $scenario" >&2
      exit 1
    fi
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
if [ "$scenario" = app ]; then MODE=app; fi
mkdir -p "$BUNDLE"/{backend,tools,migrations,admin/assets,app} "$APP_ROOT"/{config,backend/releases/v1,admin/releases/v1,packages/android,fixtures}
printf 'old server' > "$APP_ROOT/backend/releases/v1/Leave"
printf 'new server' > "$BUNDLE/backend/Leave"
printf 'old admin' > "$APP_ROOT/admin/releases/v1/index.html"
printf '<script type="module" src="/assets/entry.js"></script>' > "$BUNDLE/admin/index.html"
printf 'entry' > "$BUNDLE/admin/assets/entry.js"
printf 'SELECT 1;' > "$BUNDLE/migrations/000001_fixture.up.sql"
printf '{}' > "$BUNDLE/release.json"
printf 'new apk' > "$BUNDLE/app/$APK_NAME"
printf 'old apk' > "$APP_ROOT/packages/android/$APK_NAME"
printf 'fixture SQL' > "$BUNDLE/app/app-version.sql"
ln -s releases/v1 "$APP_ROOT/backend/current"
ln -s releases/v1 "$APP_ROOT/admin/current"
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

if [[ "$scenario" =~ ^(pid|migration|health|admin-health)$ ]]; then
  trap 'code=$?; test "$code" -ne 0; test ! -f "$APP_ROOT/scripts/app-published"; if [ "$scenario" = health ]; then test "$(cat "$APP_ROOT/admin/current/index.html")" = "old admin"; fi; exit "$code"' ERR
fi
readonly APP_ROOT CONFIG_FILE
run_release
test "$(cat "$APP_ROOT/backend/current/Leave")" = 'new server'
test "$(cat "$BACKUP/backend-v1/Leave")" = 'old server'
if [ "$MODE" = admin ]; then cmp "$BUNDLE/admin/index.html" "$APP_ROOT/admin/current/index.html"; fi
if [ "$MODE" = app ]; then
  cmp "$BUNDLE/app/$APK_NAME" "$APP_ROOT/packages/android/$APK_NAME"
  grep -q mysql "$LOG_FILE"
fi
# Re-running the same version must still migrate, replace and restart.
BACKUP="$APP_ROOT/backups/deploy/second"
run_release
test "$(grep -c '^start$' "$LOG_FILE")" -eq 2
test "$(cat "$BACKUP/backend-v1/Leave")" = 'new server'
echo "Same-version deployment/retry passed: $scenario"
