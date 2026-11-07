# 参与贡献

欢迎报告问题、讨论功能或提交修复。Leave 是假条模拟与学习项目，不实现学校审批流程。当前能力见 [功能状态](docs/feature-status.md)，启动步骤见 [README](README_zh_CN.md)。

## 开始前

- 修改前阅读根目录 [AGENTS.md](AGENTS.md) 及所修改模块的同名规范。
- Bug 请使用 Issue 表单，提供最小复现、预期与实际结果、模块和环境版本。功能建议说明场景、当前限制和希望解决的问题。
- 涉及多个模块或业务行为变化时，先在 Issue 中说明方案；小修复可以直接提交 PR。
- 不公开真实姓名、学号、反馈附件、密码、Token、服务器凭据及本地私有配置。漏洞报告见 [安全说明](SECURITY.md)。

## 开发与验证

从 fork 或工作分支提交 PR，目标分支为 `master`。保持一个 PR 聚焦一个完整目标，不混入无关格式化、依赖升级或重构；已有真实请求保持原有链路，不擅自改为 Mock。

| 模块 | 检查方式 |
| --- | --- |
| 后端（在 `backend/`） | `gofmt -l .` 无输出；`go vet ./...`、`go test ./...`、`go build ./...` |
| 管理端（在 `admin/`） | `pnpm typecheck`、`pnpm exec oxlint`、`pnpm exec eslint .`、`pnpm exec oxfmt --check src packages build tests`、`node --test tests/*.test.mjs`、`pnpm build` |
| Android 客户端 | HBuilderX 编译及 Android 运行；Node 测试不能替代原生验证 |
| 纯文档 | 核对内容、命令、链接，运行 `git diff --check` |

管理端先按 README 安装锁定依赖并准备构建环境模板；后端集成测试只使用独立本地测试实例，配置方式见 `backend/AGENTS.md`。未配置而跳过的测试不能写成通过。`pnpm lint` 和 `pnpm fmt` 会修改文件，提交前检查 diff。

## 提交与 PR

- 使用 Conventional Commits，中文描述，例如 `fix(app): 修复记录加载失败提示`。
- 每个提交表达独立完整变化，相关测试和文档随代码提交。
- 仅暂存本次改动，检查 `git diff --cached` 和 `git diff --cached --check`。
- PR 说明问题、改动后行为、验证命令及结果；UI 改动附脱敏后的实际截图。
- 明确列出测试跳过、环境阻塞及未验证的平台行为。

项目自有代码的根许可证尚未确定，勿把 `admin/LICENSE` 中上游 Soybean 的 MIT 声明视为整个仓库的统一授权；现有上游声明必须保留。
