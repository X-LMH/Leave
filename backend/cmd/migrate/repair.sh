#!/bin/sh

set -eu

# 脚本目录是 backend/cmd/migrate，回到 backend 目录后执行迁移程序。
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BACKEND_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

cd "$BACKEND_DIR"

# 先修复 dirty 状态，再重新执行迁移。
go run ./cmd/migrate -repair
go run ./cmd/migrate
