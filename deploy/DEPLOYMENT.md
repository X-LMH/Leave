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

## 最简单的发布流程

    编译 Leave 和 migrate
        ↓
    创建新的 releases/版本目录
        ↓
    复制程序、配置和 migrations
        ↓
    备份数据库并执行 migrate
        ↓
    上传 APK，更新 app_versions
        ↓
    将 current 指向新版本
        ↓
    重启 leave 服务并检查健康接口

回滚时，将 current 重新指向上一个 releases 目录，然后重启服务。
