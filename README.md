# Leave · 假条模拟程序

Leave 是一个模仿假条界面与请假流程的程序，围绕请假申请、假条展示和请假记录等场景开发，包含 Android 客户端、Web 管理端和 Go 后端。

项目用于学习与演示移动端页面交互、前后端接口开发和管理后台搭建。

## 功能概览

- **账号与资料**：登录注册、个人资料维护、密码修改。
- **请假与假条**：填写请假申请、展示假条详情、查看请假记录。
- **客户端设置**：主题与外观设置、意见反馈、版本检查与更新。
- **管理后台**：提供用户、请假等业务数据的管理入口。

项目持续开发中，各模块的完成度和接口联调进度不同，部分功能使用 Mock 或占位数据，实际行为以当前代码为准。

## 技术栈

| 模块 | 技术 |
| --- | --- |
| Android 客户端 | uni-app x、UTS、uvue、HBuilderX |
| Web 管理端 | Vue 3、TypeScript、Vite、Naive UI、UnoCSS、Pinia |
| 后端服务 | Go、Gin、GORM、JWT |
| 数据库 | MySQL |

管理端基于 Soybean Admin 开发，客户端目前以 Android 为主要运行平台。

## 目录结构

```text
Leave/
├── frontend/               # uni-app x 客户端
├── admin/                  # Web 管理端
├── backend/
│   ├── cmd/                # 服务及工具命令入口
│   ├── config/             # 配置模板
│   ├── internal/           # 接口、业务逻辑和数据访问
│   └── database/           # 表结构、初始化数据和增量迁移
├── deploy/                 # 构建、部署及客户端发布脚本
├── docs/                   # 设计及相关文档
├── .env.example            # 后端环境变量模板
└── AGENTS.md               # 项目开发约定
```

## 本地开发

### 环境准备

- Go：版本要求见 `backend/go.mod`。
- MySQL：创建本地开发数据库。
- Node.js 与 pnpm：版本要求见 `admin/package.json` 的 `engines` 字段。
- HBuilderX：用于 uni-app x 客户端开发和 Android 调试。
- Android 真机或已配置的 Android 运行环境。

以下命令以项目根目录为起点，分别在各模块的终端中执行。

### 1. 启动后端

复制配置模板：

```powershell
Copy-Item .env.example .env
Copy-Item backend/config/config.example.yaml backend/config/config.local.yaml
```

编辑配置：

- 在根目录 `.env` 中填写 `MYSQL_PASSWORD` 和 `JWT_SECRET`。
- 在 `backend/config/config.local.yaml` 中设置本地 MySQL 连接信息。
- 将 `app.package_dir`、`storage.root_dir` 和 `storage.base_url` 调整为本地实际使用的目录及文件访问地址。

按 [数据库说明](backend/database/README.md) 初始化数据库：新数据库使用 `tables/` 中的完整表结构和 `seed/` 中所需的基础数据；已有数据库升级使用增量迁移。

```powershell
cd backend
go mod download
go run ./cmd/server
```

配置模板的服务端口为 `8080`，API 前缀为 `/api/v1`。端口以实际配置为准。

### 2. 启动管理端

```powershell
cd admin
Copy-Item .env.example .env
pnpm install --frozen-lockfile
pnpm dev
```

通过 `.env` 中的 `VITE_SERVICE_BASE_URL` 配置后端 API 地址，默认模板为 `http://localhost:8080/api/v1`。启动后访问终端显示的开发地址。

更多命令与环境配置见 [管理端说明](admin/README.md)。

### 3. 运行客户端

1. 使用 HBuilderX 打开 `frontend/` 目录。
2. 检查 `frontend/utils/api.uts` 中的接口地址，使其指向当前开发后端。
3. 连接 Android 设备，通过 HBuilderX 的“运行”菜单启动客户端。

客户端开发接口默认使用 `http://127.0.0.1:8080/api/v1`。Android 真机通过 USB 访问电脑上的后端时，需要配置 ADB 反向转发：

```powershell
adb reverse tcp:8080 tcp:8080
```

该命令需在设备已连接、USB 调试已授权且 ADB 可用时执行。也可以使用手机能访问的电脑局域网 IP，并同步调整接口地址。

客户端使用 HBuilderX 的 uni-app x 工具链编译和运行。

## 常用检查

后端，在 `backend/` 目录执行：

```powershell
go test ./...
```

管理端，在 `admin/` 目录执行：

```powershell
pnpm typecheck
pnpm build
```

客户端通过 HBuilderX 编译，并在 Android 环境中检查实际页面和交互。

## 相关文档

- [管理端开发说明](admin/README.md)
- [数据库初始化与迁移](backend/database/README.md)
- [部署与客户端发布](deploy/README.md)
- [项目开发约定](AGENTS.md)
