#!/usr/bin/env bash
set -euo pipefail

test "$#" -eq 3 || { echo 'Usage: verify-apk.sh APK VERSION_NAME VERSION_CODE' >&2; exit 1; }
analyzer="${APKANALYZER:-${ANDROID_HOME:-}/cmdline-tools/latest/bin/apkanalyzer}"
if [ ! -x "$analyzer" ]; then
  analyzer=$(command -v apkanalyzer) || { echo 'Android SDK apkanalyzer is required' >&2; exit 1; }
fi
actual_name=$("$analyzer" manifest version-name "$1")
actual_code=$("$analyzer" manifest version-code "$1")
test "$actual_name" = "$2" && test "$actual_code" = "$3" || {
  echo 'APK versionName/versionCode does not match frontend/manifest.json' >&2
  exit 1
}
