#!/usr/bin/env bash
# Sourced by deploy-release.sh; Baota retains ownership of the process lifecycle.

read_test_pid() {
  local pid
  test -s "$BAOTA_PID_FILE" || return 1
  pid=$(cat "$BAOTA_PID_FILE")
  [[ "$pid" =~ ^[1-9][0-9]*$ ]] || { echo 'Invalid test project PID' >&2; return 2; }
  printf '%s' "$pid"
}

verify_test_process() {
  local pid="$1" expected actual owner
  expected=$(readlink -f "$APP_ROOT/backend/current/Leave")
  actual=$(readlink -f "/proc/$pid/exe") || return 1
  owner=$(ps -o user= -p "$pid" | xargs)
  test "$owner" = www && test "$actual" = "$expected"
}

check_backend_process() {
  local pid status=0
  pid=$(read_test_pid) || status=$?
  if [ "$status" -eq 2 ]; then return 1; fi
  if [ "$status" -eq 0 ] && kill -0 "$pid" 2>/dev/null; then
    verify_test_process "$pid" || { echo 'PID is not the Baota test backend; refusing deployment' >&2; return 1; }
  else
    test -z "$(ss -ltnH "sport = :$APP_PORT")" || { echo 'Test port is occupied without a verified backend PID' >&2; return 1; }
  fi
}

stop_backend() {
  local pid status=0
  pid=$(read_test_pid) || status=$?
  if [ "$status" -eq 2 ]; then return 1; fi
  if [ "$status" -eq 0 ] && kill -0 "$pid" 2>/dev/null; then
    verify_test_process "$pid" || { echo 'PID is not the Baota test backend; refusing to stop it' >&2; return 1; }
    kill "$pid"
    for _ in {1..20}; do
      kill -0 "$pid" 2>/dev/null || break
      sleep 1
    done
    if kill -0 "$pid" 2>/dev/null; then
      echo 'Test backend did not stop; files were not replaced' >&2
      return 1
    fi
  fi
  test -z "$(ss -ltnH "sport = :$APP_PORT")" || { echo 'Test port is still occupied' >&2; return 1; }
}

start_backend() {
  # Execute the panel-generated script as the same user configured in Baota.
  # A long-lived backend must not inherit the deployment flock descriptor.
  runuser -u www -- bash "$BAOTA_START_SCRIPT" 9>&-
  local pid
  for _ in {1..30}; do
    if pid=$(read_test_pid) && verify_test_process "$pid" && \
      curl --fail --silent --show-error --max-time 5 "http://127.0.0.1:$APP_PORT/api/v1/health" >/dev/null; then
      return 0
    fi
    sleep 1
  done
  echo 'Baota test project failed PID/binary/health verification' >&2
  return 1
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  echo 'Use deploy-release.sh to coordinate a test release' >&2
  exit 1
fi
