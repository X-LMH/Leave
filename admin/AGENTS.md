# 管理后台开发规范

## 基本原则

- 当前目录是 Leave 项目的管理后台，不是独立 Git 仓库。
- 优先复用 SoybeanAdmin 已有的布局、组件、路由和请求封装。
- 只修改当前需求涉及的页面和模块，不要随意升级依赖或重构模板基础设施。
- 先使用 Mock 数据完成页面和交互，再进行后端接口联调。

## 目录职责

- `src/views/`：业务页面。
- `src/router/`：路由、菜单和权限相关配置。
- `src/service/api/`：接口定义和请求参数类型。
- `src/service/request/`：统一请求、错误处理和 Token 逻辑。
- `src/store/`：跨页面共享状态。
- `src/components/`：跨业务复用的组件。
- `src/layouts/`：后台整体布局，非必要不要修改。

## 页面开发

- 新业务页面放在 `src/views/<业务名>/` 下。
- 页面路由应放在现有路由模块中，并保持页面、路由名称和菜单标题语义一致。
- 优先使用项目已有的 Naive UI 组件和公共表格、表单封装。
- 不要在页面中直接写重复的请求、Token 或权限判断逻辑。
- 页面交互先使用当前模板的 Mock 方案验证。

## 接口开发

- API 请求集中放在 `src/service/api/`。
- 请求地址通过 `.env` 或 `.env.prod` 的 `VITE_SERVICE_BASE_URL` 配置。
- 不要提交 `.env`、`.env.prod` 等本地环境文件。
- 接口返回码和登录失效处理沿用 `src/service/request/` 的现有约定。

## 提交前检查

在 `admin/` 目录执行：

```bash
pnpm typecheck
pnpm lint
pnpm build
```

Git 操作必须在项目根目录 `/Users/mac/my/Leave` 执行，`admin` 不再作为独立仓库管理。
