# 请假管理平台管理后台

本目录是 Leave 项目的 Web 管理后台，基于 SoybeanAdmin、Vue 3、Vite、TypeScript、Pinia、Vue Router、Naive UI 和 UnoCSS 构建。

## 项目结构

```text
admin/
├── build/                 # Vite 构建和插件配置
├── packages/              # 项目内部可复用包
├── public/                # 不参与构建处理的静态资源
├── src/
│   ├── components/        # 通用组件
│   ├── layouts/            # 后台布局
│   ├── router/             # 路由和菜单
│   ├── service/            # API 请求和接口类型
│   ├── store/              # Pinia 状态管理
│   ├── views/              # 业务页面
│   ├── locales/            # 国际化资源
│   └── styles/             # 全局样式
├── .env                   # 通用环境配置
├── .env.test              # 开发/测试接口配置
├── .env.prod              # 生产接口配置
├── package.json
├── pnpm-workspace.yaml
└── vite.config.ts
```

## 开发命令

在本目录执行：

```bash
pnpm install       # 安装依赖
pnpm dev           # 启动开发环境
pnpm dev:prod      # 使用生产模式启动
pnpm build         # 构建生产版本
pnpm preview       # 预览构建结果
pnpm typecheck     # TypeScript 类型检查
pnpm lint          # ESLint、oxlint 检查并修复
pnpm fmt           # 使用 oxfmt 格式化
```

## 环境配置

接口地址在 `.env.test` 和 `.env.prod` 中配置，当前默认使用 SoybeanAdmin 的 Mock 服务。接入 Leave 的 Go 后端时，修改 `VITE_SERVICE_BASE_URL`，不要把接口地址硬编码到业务页面中。

环境文件包含本地配置，已由根目录 `.gitignore` 忽略；如需提供配置模板，应新增 `admin/.env.example`。

## 与项目其他部分的关系

```text
frontend/   # uni-app x 用户端
admin/      # Web 管理后台
backend/    # Go 后端服务
```

管理后台与用户端是两个独立的前端应用，共享后端接口，但不要把 uni-app x 的页面或组件直接移植到本目录。
