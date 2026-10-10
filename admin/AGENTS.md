# 管理后台开发规范

本文件适用于 `admin/`，与根目录 `AGENTS.md` 的共同约定一起遵循。

## 基本原则

- 优先复用 SoybeanAdmin 已有的布局、组件、路由和请求封装。
- 只修改当前需求涉及的页面和模块，不要随意升级依赖或重构模板基础设施。
- 新页面未要求联调时，先使用 Mock 数据完成页面和交互；已联调页面保持真实请求，不改回 Mock。

## 目录职责

- `src/views/`：业务页面。
- `src/router/`：路由、菜单和权限相关配置。
- `src/service/api/`：接口定义和请求参数类型。
- `src/typings/api/`：已有的全局接口类型声明；新增或修改类型时沿用对应业务模块的组织方式，不重复维护同一类型。
- `src/service/request/`：统一请求、错误处理和 Token 逻辑。
- `src/styles/scss/management.scss`：管理页面共享样式。
- `src/store/`：跨页面共享状态。
- `src/components/`：跨业务复用的组件。
- `src/layouts/`：后台整体布局，非必要不要修改。

## 页面开发

- 新业务页面放在 `src/views/<业务名>/` 下。
- 新页面沿用 Elegant Router 生成流程，保持页面、路由名称和菜单标题语义一致；菜单元信息检查 `build/plugins/router.ts`，菜单标题检查 `src/locales/` 对应语言文件。
- 路由调整应修改对应源配置并按现有流程生成，不要仅手改 `src/router/elegant/` 或 `src/typings/elegant-router.d.ts` 等生成文件，避免重新生成后丢失改动。
- 优先使用项目已有的 Naive UI 组件和公共表格、表单封装。
- 管理页面优先复用 `src/styles/scss/management.scss`；抽屉、弹窗等通过 Teleport 挂载到页面外部的内容，样式不能依赖 `.management-page` 祖先选择器，应使用组件自身的 class。
- 不要在页面中直接写重复的请求、Token 或权限判断逻辑。
- 页面交互按当前数据来源验证；Mock 页面使用现有 Mock 方案，已联调页面检查真实请求及错误处理。

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

`pnpm lint` 会自动修复代码，`pnpm fmt` 也会修改文件，执行后检查 diff，避免混入无关改动；涉及 `tests/` 已覆盖的功能时，运行对应的 `node --test tests/<文件名>.test.mjs`。纯文档修改按根目录规范检查内容和 diff 即可。
