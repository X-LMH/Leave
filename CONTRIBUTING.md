# 参与贡献

欢迎报告问题、提出建议或提交修复。启动与检查命令见 [README](README_zh_CN.md)，当前能力见 [功能状态](docs/feature-status.md)。

## 提交 Issue

选择“问题报告”或“功能建议”，按表单填写即可。截图和日志请脱敏，不要提交密码、Token 或私人数据；安全漏洞请按 [安全说明](SECURITY.md) 私密报告。

## 提交代码

- 修改前阅读 [AGENTS.md](AGENTS.md) 和对应模块规范。较大的改动建议先通过 Issue 讨论。
- 从 fork 或工作分支向 `master` 提交 PR，一个 PR 聚焦一个目标，避免无关改动。
- 提交说明使用 Conventional Commits 和中文描述，例如 `fix(app): 修复记录加载失败提示`；每个提交包含一个完整变化。
- 按 README 和模块规范运行相关检查，在 PR 中说明改动及验证结果；跳过或未验证的项目请明确写出。Android 改动需通过 HBuilderX 和真机验证。
- 提交前检查 diff，保留相关测试与文档，不提交凭据、私人数据和运行产物。

项目自有代码的根许可证尚未确定，`admin/LICENSE` 是上游 Soybean 的许可声明，不代表整个仓库的统一授权；请保留现有上游声明。
