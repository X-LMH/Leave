# Leave 部署说明

适用于一台 Linux 服务器、不使用 Docker 的部署方式。

## 目录结构

    /www/wwwroot/Leave_test
    ├── current -> releases/v0.0.3/
    ├── releases/
    │   ├── v0.0.3/
    │   │   ├── Leave
    │   │   ├── migrate
    │   │   ├── config/config.yaml
    │   │   └── database/migrations/
    │   └── v0.0.2/
    ├── packages/android/
    ├── backups/mysql/
    └── scripts/

## 目录作用

- current/：软链接，指向当前正在运行的后端版本。
- releases/：保存每次发布的后端程序，保留旧版本用于回滚。
- packages/android/：保存 APK 安装包。
- backups/mysql/：保存 MySQL 数据库备份。
- scripts/：保存发布、备份、迁移等重复执行的脚本。

## 运行方式

使用 systemd 管理 Go 服务：

    WorkingDirectory=/www/wwwroot/Leave_test/current
    ExecStart=/www/wwwroot/Leave_test/current/Leave

WorkingDirectory 必须设置，因为程序会读取：

    ./config/config.yaml
    ./database/migrations

Go 服务监听本机 127.0.0.1:8080，Nginx 负责 HTTPS 和反向代理。

## CD 发布流程

GitHub Actions 的 `Leave CD` 工作流在 `deploy` 分支更新后执行以下步骤：

    从 frontend/manifest.json 读取 versionCode/versionName
        ↓
    将 artifacts/android/app.apk 命名为 leave-android-{versionCode}-{versionName}.apk
        ↓
    上传到 packages/android/
        ↓
    使用 CGO_ENABLED=0 编译 backend/cmd/server，并上传为 releases/v{versionName}/Leave
        ↓
    从旧 current 复制 config/config.yaml 到新版本目录
        ↓
    将 current 切换到新版本目录
        ↓
    重启指定的 systemd Go 服务并检查 ActiveState/MainPID

本流程不上传或处理 `scripts/`、`backups/`，也不会编译或运行 `migrate`，不会自动执行数据库迁移。

## GitHub Secrets

需要配置以下 Secrets：

- `SERVER_HOST`、`SERVER_PORT`、`SERVER_USERNAME`、`SERVER_SSH_KEY`
- `SERVER_APP_ROOT`：例如 `/www/wwwroot/Leave_test`
- `SERVER_SERVICE_NAME`：宝塔/systemd 中实际管理该 Go 程序的服务名

发布前应确认 `current/config/config.yaml` 已存在，并确认该 systemd 服务的 `WorkingDirectory` 和 `ExecStart` 都通过 `current` 指向当前版本。工作流只重启 `SERVER_SERVICE_NAME` 指定的服务，不会使用 `pkill` 或 `killall`。

回滚时，将 current 重新指向上一个 releases 目录，然后重启服务。
