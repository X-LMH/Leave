<p align="center">
  <img src="frontend/static/manifest/app.png" width="112" height="112" alt="Leave app icon" />
</p>

<h1 align="center">Leave</h1>

<p align="center">
  一个假条模拟与学习项目，包含 Android 客户端、Web 管理端和 Go 后端。
</p>

<p align="center">
  <a href="README.md">English</a> · <strong>简体中文</strong>
</p>

<p align="center">
  <a href="https://github.com/avery-my/Leave/releases/latest"><img src="https://img.shields.io/github/v/release/avery-my/Leave?style=flat-square&amp;label=release" alt="Release" /></a>
  <a href="https://github.com/avery-my/Leave/releases"><img src="https://img.shields.io/github/downloads/avery-my/Leave/total?style=flat-square&amp;logo=github" alt="Downloads" /></a>
</p>

项目用于学习假条生成及相关账号、数据管理，不包含学校审批流程，也不提供正式请假授权。

## 功能

- 账号注册、登录、个人资料维护和密码修改。
- 请假申请、假条详情和请假记录。
- 明暗主题、外观设置、意见反馈和客户端更新。
- 管理端提供学生账号、请假记录、班级、宿舍、请假类型、反馈、版本记录及数据概览等功能。
- 基于 Redis 的服务端登录会话和请求限流。

项目持续开发中，部分功能仍使用 Mock 或占位数据，各模块的接口联调进度不同。

## 下载

从 [GitHub Releases](https://github.com/avery-my/Leave/releases/latest) 获取最新 Android 安装包，或查看[历史版本](https://github.com/avery-my/Leave/releases)。

| 平台 | 安装包 | 下载 |
| --- | --- | --- |
| Android | APK | [最新 Android APK](https://github.com/avery-my/Leave/releases/latest) |

Web 管理端和 Go 后端需自行部署，参见下方的[从源码运行](#从源码运行)和[部署说明](deploy/README.md)。

## 技术栈

| 模块 | 技术 |
| --- | --- |
| Android 客户端 | uni-app x、UTS、uvue、HBuilderX |
| Web 管理端 | Vue 3、TypeScript、Vite、Naive UI、UnoCSS、Pinia |
| 后端服务 | Go、Gin、GORM、Zap |
| 数据存储 | MySQL（业务数据）、Redis（登录会话与请求限流） |

管理端基于 Soybean Admin 开发，客户端目前以 Android 为主要运行平台。

## 从源码运行

需要 Go、MySQL、Redis、Node.js、pnpm 和 HBuilderX。工具链版本要求见 [go.mod](backend/go.mod) 和 [package.json](admin/package.json)。客户端运行还需要 Android 真机或已配置的 Android 运行环境。

各模块使用独立终端，以下命令均以项目根目录为起点。

### 后端

使用 PowerShell 复制配置模板：

```powershell
Copy-Item .env.example .env
Copy-Item backend/config/config.example.yaml backend/config/config.local.yaml
```

启动前完成以下配置：

- 在根目录 `.env` 中填写 `MYSQL_PASSWORD`；Redis 启用认证时填写 `REDIS_PASSWORD`，否则留空。
- 在 `backend/config/config.local.yaml` 中设置 MySQL 和 Redis 连接信息，确保两个服务均可连接。
- 设置当前环境独立的 `app.namespace`，会话有效期可沿用模板值。
- 将 `app.package_dir` 设为本地 APK 目录，配置上传文件的 `storage.root_dir` 和 `storage.base_url`。上传文件通过 Nginx 等静态服务访问，目录与 URL 前缀需相互对应。

按 [数据库说明](backend/database/README.md) 初始化独立的开发数据库。新数据库使用当前完整表结构和所需的 seed 数据，已有数据库使用增量迁移。初始化数据不提供默认管理员账号或固定密码。

```powershell
cd backend
go mod download
go run ./cmd/server
```

模板使用 `8080` 端口和 `/api/v1` 前缀。可在另一个终端检查服务：

```powershell
Invoke-RestMethod http://127.0.0.1:8080/api/v1/health
```

App 请求还会校验客户端版本。联调前需准备与 `frontend/manifest.json` 对应的已发布版本数据，版本数据与 APK 的准备方式见 [发布说明](deploy/README.md)。

### Web 管理端

```powershell
cd admin
Copy-Item .env.example .env
pnpm install --frozen-lockfile
pnpm dev
```

通过 `admin/.env` 中的 `VITE_SERVICE_BASE_URL` 设置后端 API 地址，模板默认使用 `http://localhost:8080/api/v1`。启动后访问终端显示的开发地址。

环境配置和其他命令见 [管理端说明](admin/README.md)。

### Android 客户端

1. 使用 HBuilderX 打开 `frontend/`。
2. 检查 `frontend/utils/api.uts` 中的开发接口和头像地址。运行使用 development 配置，发行打包使用 production 配置。
3. 连接并授权 Android 设备，通过 HBuilderX 运行客户端。

USB 调试时，开发接口默认使用 `http://127.0.0.1:8080/api/v1`。ADB 可用且设备已授权后，配置后端端口反向转发：

```powershell
adb reverse tcp:8080 tcp:8080
```

默认开发头像地址使用 `8081` 端口的静态服务。使用时先启动该服务，再执行 `adb reverse tcp:8081 tcp:8081`。后端的 `storage.base_url` 也需能被设备访问。也可以使用设备可访问的电脑局域网地址，并同步调整开发地址。

客户端使用 HBuilderX 的 uni-app x 工具链，没有 npm/pnpm 构建命令。

### 可选开发脚本

在项目根目录执行 `make phone`，可检查设备和本地服务并配置 USB 反向转发。需要 Make、Windows PowerShell 和 ADB；多设备时传入 `DEVICE_SERIAL=<设备序列号>`。

`make air` 会先建立连接远端 MySQL／Redis 的 SSH 隧道，再启动 Air 热重载。需要 Make、Air、OpenSSH、已验证的主机密钥及本机忽略的 `.private/dev-ssh.json` 配置，本地 MySQL／Redis 连接端口分别设为 `13306`／`16379`。确认目标环境和访问范围后再使用；纯本地开发使用 `go run ./cmd/server` 即可。

## 检查

在 `backend/` 中执行：

```powershell
go test ./...
go build ./...
```

在 `admin/` 中执行：

```powershell
pnpm typecheck
pnpm exec oxlint
pnpm exec eslint .
pnpm exec oxfmt --check src packages build tests
node --test tests/*.test.mjs
pnpm build
```

以上 Lint 和格式命令仅检查文件，`pnpm lint`、`pnpm fmt` 会修改文件。`pnpm build` 使用生产模式，测试模式使用 `pnpm build:test`，对应环境模板为 `.env.prod.example` 和 `.env.test.example`。

数据库和 Redis 集成测试需要独立的本机测试实例，详见 [后端开发约定](backend/AGENTS.md)。跳过不代表通过。Node 测试覆盖部分客户端逻辑，客户端编译和设备交互仍需通过 HBuilderX 与 Android 验证。

## 相关文档

- [认证与 Redis 会话](docs/authentication.md)
- [管理端开发](admin/README.md)
- [数据库初始化与迁移](backend/database/README.md)
- [部署与客户端发布](deploy/README.md)
- [项目开发约定](AGENTS.md)
- [功能实现状态](docs/feature-status.md)
- [参与贡献](CONTRIBUTING.md)
- [安全问题报告](SECURITY.md)
- [Android 发布验收清单](docs/release-checklist.md)
- [维护者待办](docs/maintainer-checklist.md)
