# AGENTS.md

## 项目概览

本项目是请销假管理系统，包含 uni-app x 客户端、Web 管理端和 Go 后端。主要业务围绕用户账号、个人资料、请假申请及记录展开，并包含意见反馈和客户端版本更新等功能。

客户端 App 主要运行在 Android 平台。涉及平台 API、权限、文件路径、安装包更新和页面交互时，优先考虑 Android 的实际行为与兼容性；其他平台的适配以具体任务要求为准。

### 目录与技术栈

* `frontend/`：uni-app x 客户端，使用 UTS 和 uvue，包含登录注册、请假申请、请假详情、请假记录、个人中心和版本更新等页面。页面注册位于 `frontend/pages.json`，应用配置位于 `frontend/manifest.json`。
* `admin/`：基于 Soybean Admin 的 Web 管理端，使用 Vue 3、TypeScript、Vite、Naive UI 和 UnoCSS。
* `backend/`：Go 后端，使用 Gin 提供 HTTP 接口，使用 GORM 操作 MySQL。命令入口位于 `backend/cmd/`，配置文件位于 `backend/config/`。
* `backend/internal/`：后端主要代码。`controller/` 对应 Handler 层，`service/` 处理业务逻辑，`dao/` 处理数据库操作；`router/` 管理路由，`middleware/` 管理中间件，`models/` 定义数据模型，`dto/` 定义请求和响应结构。
* `backend/database/`：数据库相关文件。`tables/` 保存当前完整表结构，`seed/` 保存初始化数据，`migrations/` 保存增量变更。具体操作参见 `backend/database/README.md`。
* `deploy/`：构建、部署、数据库迁移和客户端版本发布脚本；部署说明位于 `deploy/DEPLOYMENT.md`，客户端发布配置位于 `deploy/releases/`。

### 当前开发边界

* 客户端、管理端和后端是独立的开发模块，修改前先确认需求涉及哪个模块。
* 当前各业务模块的联调进度不同：已有真实请求的功能保持现有数据链路，使用 Mock 或占位的功能按当前阶段开发；不要仅因存在后端接口就接入，也不要将已联调功能改回 Mock。
* 开发时以现有代码和配置为准；项目概览仅用于定位模块，不代表所有页面或业务流程已经完成。

## 规范适用范围

* 根目录维护跨模块边界、Git、通用代码原则及交付要求；修改模块时，同时遵循对应目录的规范：
  * `admin/AGENTS.md`：管理端页面、请求封装及 pnpm 验证。
  * `backend/AGENTS.md`：Go 分层、DTO、分页、数据库变更及测试。
  * `frontend/AGENTS.md`：UTS、uvue、Android 行为、CSS 约束及 HBuilderX 验证。
* 模块文件补充本模块规则，不重复维护根目录共同约定；跨模块任务分别遵循相关模块规范。

## 可用技能

* 前端界面开发、美化或重设计任务可参考 `.agents/skills/frontend-design/SKILL.md`，结合本项目实际技术栈、平台限制和现有 UI 规范使用。

## 基本原则

* 优先保证代码正确、清晰、可维护。
* 不以减少代码行数为目标。
* 优先使用简单、直接、符合当前项目规模的实现。
* 不要为了所谓“架构规范”进行过度封装或过度设计。
* 只修改当前需求涉及的代码，不要顺手大范围重构无关模块。
* 如果现有代码已经足够清晰，不要为了“优化”强行重构。

## 代码修改边界

* 修改前先阅读相关代码和目录内的 `AGENTS.md`，确认需求范围、现有行为及受影响的调用方。
* 每次改动围绕一个明确目标，保持实现完整；必要的调用方、测试和文档调整应一起完成，不以文件数量或代码行数作为拆分标准。
* 功能修改与独立重构尽量分开，不混入无关的重命名、格式化、依赖升级或目录调整。
* 纯重构保持业务行为、接口契约和现有 UI 不变；替换数据层时保留已有表单、交互和确认流程。需求涉及行为变化时，明确说明变化及其影响。
* 修改共享函数、DTO 或配置时，检查所有相关引用，保持调用方各自的业务规则，不只修复当前看到的一处。
* 遵循用户指定的阶段和范围；要求先完成页面、Mock 或单个示例时，只完成该阶段，不提前扩展到接口联调或批量修改。

## Git 协作规范

* 本项目根目录是统一的 Git 仓库，`frontend/`、`admin/`、`backend/` 不作为独立仓库管理；Git 操作以当前检出的项目根目录为准，不依赖某台机器的绝对路径。
* 开始修改和提交前检查 `git status` 与相关 diff，区分已有改动和本次改动；保留用户及其他任务的修改，不擅自覆盖、撤销或纳入提交。
* 用户要求提交时再创建 commit；提交授权不代表推送授权，只有明确要求推送时才执行 `git push`。
* 每个 commit 表达一个完整的逻辑变化；可独立的功能、修复、重构和文档调整按职责拆分，必要的测试与对应代码一起提交，拆分后保持可构建。
* 仅暂存本次提交涉及的文件或代码块；提交前检查 `git diff --cached` 和 `git diff --cached --check`，不要未经检查直接暂存整个工作区。
* 提交说明使用 Conventional Commits 格式，描述用中文，例如 `feat: 新增学生状态管理`、`fix: 修复请假时间校验`、`refactor: 统一分页查询`、`docs: 更新开发规范`；需要时添加 scope。
* 遵循根目录 `.gitignore`：不提交本地 `.env`、后端本地及服务器配置、上传文件、`node_modules/`、`frontend/unpackage/` 等运行或构建产物，不使用强制暂存绕过忽略规则提交真实凭据。
* 保留项目必需的 `admin/pnpm-lock.yaml`、`backend/go.mod`、`backend/go.sum` 和相关 migration；`admin/build/`、`admin/packages/*/src/` 是源码，不当作构建产物清理。
* 覆盖或丢弃已有改动、改写提交历史、强制推送等操作，必须有用户针对该操作的明确授权；不要为了获得干净工作区而执行 `reset --hard` 或 `clean -fd`。
* 提交后检查最新提交和工作区状态，说明提交内容及剩余未提交改动；不要将“仍有其他改动”当作提交失败。

## 验证与交付

* 根据改动影响运行相关格式检查、类型检查、测试或构建；业务逻辑和缺陷修复优先验证正常路径、边界条件及回归场景，纯文档修改检查内容和 diff 即可。
* 使用对应模块规范中的现有工具和验证流程，不为了验证引入无关依赖、升级工具链或修改工具配置。
* UI 改动在可用环境中检查实际渲染和交互；无法运行或查看时，明确说明未完成视觉验证。
* 交付时说明改动内容、验证结果和已知限制；区分检查通过、测试跳过和环境阻塞，不将局部测试通过表述为全量验证通过。

## 代码结构

* 保持主业务流程从上到下清晰可读，优先使用 early return，减少多层嵌套。
* 函数过长或包含明显独立职责时，可以合理拆分。
* 不要为了缩短函数而拆出大量一两行的小函数。
* 只有在能明显提升可读性、复用性、可测试性或可维护性时，才新增抽象。
* 不要随意增加 `Interface`、`Repository`、`Manager`、`Factory`、`UseCase` 等额外层级。

## 代码格式

* 保持正常缩进和换行。
* 不要将多段逻辑、条件或大量属性堆在同一行。
* 代码排版应便于人工查看和编辑；字段较多的结构体、参数或配置项优先纵向逐项排列，不为压缩行数挤在一行。
* 命名应清晰表达真实用途，避免模糊或过度冗长。

## 部署与客户端发布边界

* 部署流程以 `deploy/DEPLOYMENT.md` 和现有脚本为准；源码修改与执行部署、发布或业务数据库操作分开处理，是否执行以当前任务授权为准。
* 客户端版本发布沿用 `deploy/releases/` 和发布脚本，版本号由 `frontend/manifest.json` 读取；发布版本数据不属于表结构 migration，不为每次版本发布新增 migration。

## 规范参考

以下资料用于参考通用实践，具体执行以本项目约定和当前任务要求为准：

* [Google Engineering Practices：Small CLs](https://google.github.io/eng-practices/review/developer/small-cls.html)：保持改动聚焦、完整，合理拆分重构与功能变化。
* [Git 官方书籍：记录每次更新到仓库](https://git-scm.com/book/en/v2/Git-Basics-Recording-Changes-to-the-Repository)：理解工作区和暂存区，检查实际提交内容。
* [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/)：统一提交类型和说明格式，中文描述为本项目约定。
* [GitHub Docs：Helping others review your changes](https://docs.github.com/en/pull-requests/concepts/helping-others-review-your-changes)：提供聚焦的改动及清楚的背景、影响和验证说明。

## 最终原则

在实现或重构代码前，优先判断：

> 是否可以用更简单、清晰、可维护的方式完成当前需求？

不要把简单问题复杂化，也不要为了追求“简洁”牺牲代码可读性。
