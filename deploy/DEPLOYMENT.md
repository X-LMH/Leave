# Leave 部署说明

适用于一台 Linux 服务器、不使用 Docker 的部署方式。

## 目录结构

    /www/wwwroot/Leave
    ├── config/
    │   └── config.yaml
    ├── .env
    ├── current -> releases/v0.0.3/
    ├── releases/
    │   ├── v0.0.3/
    │   │   └── Leave
    │   └── v0.0.2/
    ├── packages/android/
    ├── backups/mysql/
    ├── database/migrations/
    └── scripts/

## 目录作用

- current/：软链接，指向当前正在运行的后端版本。
- releases/：保存每次发布的后端程序，保留旧版本用于回滚。
- packages/android/：保存 APK 安装包。
- backups/mysql/：保存 MySQL 数据库备份。
- database/migrations/：保存完整、累积的数据库 migration 历史。
- scripts/：保存发布脚本、迁移调用脚本和长期使用的 migrate 程序。

客户端版本发布记录由 `deploy/releases/android.yaml` 维护。GitHub Actions 使用
`deploy/build-version-sql.rb` 生成事务 SQL，上传到本次 release 后，由服务器
`scripts/publish-version.sh` 通过 MySQL 客户端将旧版本标记为 `archived`，再插入新的 `published` 版本。数据库密码优先从 `MYSQL_PASSWORD` 环境变量读取。

## 运行方式

Leave 使用 current 软链接保存当前发布版本。宝塔面板中的 Go 项目名为 `Leave`，由宝塔项目启动脚本管理。部署会按 `Leave` 项目的 PID 文件停止旧进程，切换 `current` 后执行该项目的启动脚本。启动文件默认是 `/www/wwwroot/Leave/current/Leave`；工作目录设置为项目根目录，因为程序会读取：

    ./config/config.yaml
    ./database/migrations

所有 release 共用 `/www/wwwroot/Leave/config/config.yaml`、`/www/wwwroot/Leave/.env` 和根目录下的 `database/migrations/`，新版本目录不再复制这些文件。

Go 服务监听 `config/config.yaml` 中 `app.port` 配置的本机端口，Nginx 负责 HTTPS 和反向代理。

## CD 发布流程

GitHub Actions 的 `Leave CD` 工作流在推送 `v*` 版本标签后执行以下步骤：

    从 frontend/manifest.json 读取 versionCode/versionName
        ↓
    从 deploy/releases/*.apk 选择最后修改时间最新的 APK，并按版本信息生成标准文件名
        ↓
    生成 app-version.sql，并检查 releases/v{versionName} 是否已存在
        ↓
    上传 APK 到 packages/android/
        ↓
    上传 Leave 到 releases/v{versionName}/
        ↓
    上传 migrate 到服务器 scripts/migrate
        ↓
    同步 backend/database/migrations/* 到根目录 database/migrations/
        ↓
    执行 scripts/migrate，连接根目录 config.yaml 和 .env 中配置的 MySQL
        ↓
    migrate 成功后执行 app-version.sql，更新 app_versions
        ↓
    按 Leave 项目的 PID 停止旧进程
        ↓
    将 current 统一切换为相对路径 releases/v{versionName} 的软链接
        ↓
    切换 current 并执行 Leave 项目启动脚本
        ↓
    检查配置端口的健康接口

本流程不处理 `backups/`。`manifest.json` 和 `android.yaml` 只在 GitHub Actions 中读取，不上传服务器。部署脚本统一放在服务器的 `scripts/`，每个 release 保存程序和本次发布 SQL；数据库 migration 统一保存在根目录，并且只新增、不修改历史文件。APK 必须放在 `deploy/releases/`，原始文件名可以是任意名称；如果目录中存在多个 APK，使用最后修改时间最新的一个，并在上传前统一重命名为 `leave-{platform}-{versionCode}-{versionName}.apk`。配置统一使用项目根目录的 `config/config.yaml` 和 `.env`。

每次 CD 会覆盖服务器 `scripts/` 目录中的同名文件：

    deploy-release.sh
    run-migrations.sh
    publish-version.sh
    switch-and-restart.sh
    migrate

服务器 `scripts/` 中其他未被上传的脚本不会被删除。根目录 `database/migrations/` 会接收本次提交中的 migration 文件；已存在的同名文件可能被覆盖，但服务器中其他历史 migration 不会被删除。

## GitHub Secrets

需要配置以下 Secrets：

- `SERVER_HOST`、`SERVER_PORT`、`SERVER_USERNAME`、`SERVER_SSH_KEY`
项目目录固定为 `/www/wwwroot/Leave`。宝塔 Go 项目名为 `Leave`，默认启动文件 `/www/wwwroot/Leave/current/Leave`、PID 文件 `/var/tmp/gopids/Leave.pid`、启动脚本 `/www/server/go_project/vhost/scripts/Leave.sh`。如服务器中的宝塔脚本或 PID 路径不同，可在服务器 `/www/wwwroot/Leave/.env` 中设置 `BAOTA_GO_PID_FILE` 和 `BAOTA_GO_START_SCRIPT` 覆盖默认值。部署会先验证 PID 对应进程确实运行该启动文件，再停止它；之后切换 `current`、执行 Leave 项目的宝塔启动脚本，并同时核验宝塔 PID、运行中的二进制版本和配置端口的 `/api/v1/health`。不会通过端口批量杀进程，也不会绕过宝塔另行 `nohup` 启动服务。

回滚时，将 current 重新指向上一个 releases 目录，然后重启服务。
