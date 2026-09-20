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

Leave_test 通过 current 软链接启动 Go 服务：

    WorkingDirectory=/www/wwwroot/Leave_test/current
    ExecStart=/www/wwwroot/Leave_test/current/Leave

WorkingDirectory 必须设置为 current，因为程序会读取：

    ./config/config.yaml
    ./database/migrations

Go 服务监听本机 `127.0.0.1:10000`，Nginx 负责 HTTPS 和反向代理。

## CD 发布流程

GitHub Actions 的 `Leave CD` 工作流在 `deploy` 分支更新后执行以下步骤：

    从 frontend/manifest.json 读取 versionCode/versionName
        ↓
    从 deploy/releases/*.apk 选择最后修改时间最新的 APK，并校验固定文件名
        ↓
    上传到 packages/android/
        ↓
    使用 CGO_ENABLED=0 编译 backend/cmd/server，并上传为 releases/v{versionName}/Leave
        ↓
    从旧 current 复制 config/config.yaml 到新版本目录
        ↓
    将 current 统一切换为相对路径 releases/v{versionName} 的软链接
        ↓
    从 current/Leave 启动 Leave_test
        ↓
    检查 10000 端口健康接口

本流程不上传或处理 `scripts/`、`backups/`，也不会编译或运行 `migrate`，不会自动执行数据库结构迁移。APK 必须放在 `deploy/releases/`，文件名遵循 `leave-{platform}-{versionCode}-{versionName}.apk`；如果目录中存在多个 APK，使用最后修改时间最新的一个。

## GitHub Secrets

需要配置以下 Secrets：

- `SERVER_HOST`、`SERVER_PORT`、`SERVER_USERNAME`、`SERVER_SSH_KEY`
项目目录固定为 `/www/wwwroot/Leave_test`，运行文件固定为 `/www/wwwroot/Leave_test/current/Leave`。工作流只停止占用 `10000` 端口的旧进程，不生成 `Leave.log` 或 `Leave.pid`，将 `current` 统一切换为项目内相对软链接 `releases/v{version}`；如果旧 `current` 是实体目录，会先移到 `.current-directory-{version}` 保留，不会直接删除。随后通过 `current/Leave` 启动程序，并检查 `127.0.0.1:10000/api/v1/health`。不会使用 `pkill` 或 `killall`，也不会停止其他 Go 进程。

回滚时，将 current 重新指向上一个 releases 目录，然后重启服务。
