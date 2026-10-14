# CD 迁移实施计划

本文为临时迁移文档，完成迁移和验证后，将最终可用的配置说明、标签规则和操作步骤整理回 [部署 README](README.md)，再删除本文件。当前运行说明仍以 README 和现有脚本为准。

本文记录后续改造的约定，不是当前可执行的发布步骤。当前工作流仍只匹配 `v*`；以下三个入口尚未实现，推送对应标签不会触发新的发布流程。

## 第一版范围：仅测试环境

第一版仅接入 `admin-test-vN`、`app-test-vX.Y.Z`、`back-test-vN`，生产标签只保留为后续设计，不在第一版启用或默认回退到生产。现有工作流和发布脚本将在后续实现阶段统一重写。

测试站点已由用户完成访问验证，后续发布复用现有站点，替换对应版本产物并切换链接；后端程序替换后仍需重启测试宝塔项目并检查健康接口。服务器上的 `config/config.test.yaml` 和 `.env` 原地保留，不由构建上传覆盖。

`deploy/release.yaml`、环境模板和代码一起提交到 Git。第一版 admin 构建使用现有 `pnpm build:test`，产物为 `admin/dist-test/`。构建命令、产物目录及环境模板路径固定在后续构建脚本中，不写入发布 YAML。

### admin 测试构建配置

可提交的完整模板为 `admin/.env.test.example`，本地 `.env`、`.env.test` 继续忽略，不删除或放宽 `.gitignore`。模板包含构建需要的页面、图标、接口、响应码、权限及存储前缀配置，不依赖开发者本地 `.env` 补充变量。

后续 CI 在安装依赖及构建前，将该模板复制为 `admin/.env.test`，然后从 `admin/` 执行 `pnpm build:test`。现有项目的 `admin/.env.example` 仍是通用模板；CI 不将它复制为 `.env`，也不读取本机私有环境文件。

测试模板使用 `VITE_SERVICE_BASE_URL=/api/v1`、关闭 Vite 开发代理，并设置独立存储前缀 `leave-admin-test`。静态站点的同域 API 代理由现有测试 Nginx 配置负责。`VITE_*` 会进入浏览器产物，模板不包含服务器密码、数据库密码或 JWT 密钥；这些值继续留在 GitHub Secrets 或服务器配置中。

模板与统一发布配置已准备，但自动复制和读取配置的 CI/CD 代码尚未实现。生产构建参数在后续生产流程接入时另行确定，不复用第一版测试构建命令来发布生产。

## 标签与发布范围

| 生产标签示例 | 测试标签示例 | 发布范围 | 执行顺序 |
| --- | --- | --- | --- |
| `admin-v1` | `admin-test-v1` | 后端 + admin | 后端检查、构建、迁移及部署成功后，发布管理端 |
| `app-v1.0.0` | `app-test-v1.0.0` | 后端 + App | 后端检查、构建、迁移及部署成功后，发布 APK 并更新客户端版本记录 |
| `back-v1` | `back-test-v1` | 仅后端 | 后端检查、构建、迁移及部署 |

标签格式为生产 `<模块>-<版本>`、测试 `<模块>-test-<版本>`。`test` 明确选择测试环境；未带 `test` 的上述标签选择生产环境。三个发布范围各自提供测试、生产两种标签形式，仍复用同一套模块部署逻辑。

工作流应分别严格匹配 `admin-v*`、`app-v*`、`back-v*` 和 `admin-test-v*`、`app-test-v*`、`back-test-v*`，再校验对应版本格式。不使用宽泛的 `admin-*` 等模式直接触发生产部署，避免测试标签误入生产流程；不符合约定的标签应拒绝发布。

三个入口复用同一套后端部署逻辑。同一环境的生产发布必须串行执行，不能按不同标签各自并发切换同一个后端。三者联合更新时，分别触发 admin 和 App 入口，接受后端重复部署。

保留这三个固定组合入口，不增加 `all-*` 标签，也不通过每次修改发布范围配置来决定发布模块。即使只修改 admin 页面或 App，选择对应入口时也包含后端部署。

后端和管理端使用整数版本，测试从 `v1` 开始；App 使用三段版本，测试从 `1.0.0` 开始。标签前缀用于选择发布入口，不写入服务器版本目录名。例如 `admin-v1` 的管理端产物放在 `admin/releases/v1/`，后端产物放在本次指定的 `backend/releases/vN/`。

`admin-v1`、`app-v1.0.0` 和 `back-v1` 的版本号不要求彼此相同，也不能仅凭标签后缀推断组合发布中的后端版本。组合发布统一读取下文规划的 `deploy/release.yaml` 中的 `backend.version`；同一后端版本只能对应同一程序，不能因另一个入口发布而覆盖已有版本。

测试和生产的版本目录分别位于各自环境中，目录名均不包含 `test` 或模块前缀。例如 `admin-test-v1` 的管理端产物放在 `/www/wwwroot/Leave_test/admin/releases/v1/`；`admin-v1` 的管理端产物放在 `/www/wwwroot/Leave/admin/releases/v1/`。版本禁止覆盖的规则在各环境内分别生效。

测试通过后，如需将同一批代码发布到生产，应将对应生产标签指向通过测试的提交；例如 `admin-test-v1` 与 `admin-v1`。测试、生产的接口地址等构建配置仍需分别提供，不能直接把包含测试地址的管理端产物或 APK 改名作为生产包。两种环境的 App 标签都需要与标签对应提交中的 manifest 版本及各自安装包的真实版本一致。

## 统一发布配置（待迁移）

计划使用 `deploy/release.yaml` 集中维护后端和管理端版本、APK 目录和发布说明；App 版本继续由 `frontend/manifest.json` 唯一维护，完整承接 `deploy/releases/android.yaml` 的职责。不再分别新增 `backend/VERSION`、`admin/VERSION`，也不将管理端模板的 `package.json` 版本作为业务发布版本。

统一配置已创建于 [release.yaml](release.yaml)，目前尚未接入脚本。后端和 admin 初始版本均为 `1`。App 版本不写入统一配置，仍读取 manifest 的 `versionName` 和 `versionCode`；配置只填写 APK 所在目录。测试目标 `1.0.0` 的安装包准备完成后，更新 manifest 并替换目录中的旧 APK。发布说明已从旧文件完整复制。

配置与标签均读取标签指向的提交，APK 目录以仓库根目录为基准。Linux/amd64、CGO 关闭、Go 程序入口、`pnpm build:test`、环境模板路径和 `admin/dist-test/` 等固定项目约定由构建脚本维护，不在发布 YAML 中重复配置。后端输出 Linux 程序及迁移工具，admin 输出静态文件，App 沿用 HBuilderX 本地打包，CD 只接收、校验并发布已准备的 APK。统一配置不会让 App 自动参与 CI 构建，也不修改 App 的打包方式。

| 发布信息 | 迁移后的唯一来源 | 校验及用途 |
| --- | --- | --- |
| 后端版本 | `backend.version` | 正整数，生成 `backend/releases/vN/`；三个入口均读取 |
| 管理端版本 | `admin.version` | 正整数，与 admin 标签后缀一致，生成 `admin/releases/vN/` |
| Android 版本 | `frontend/manifest.json` 的 `versionName`、`versionCode` | 标签与 `versionName` 一致，两个字段均与 APK 内真实版本一致；统一配置不重复填写 |
| Android 安装包目录 | `app.android.apk_directory` | 仅扫描目录直接包含的 `.apk` 文件，必须恰好一个；零个或多个均报错，不按修改时间选择 |
| Android 发布说明 | `app.android.release_notes` | 从旧文件完整迁入，保持字符串列表，供客户端版本 SQL 使用 |

例如 `admin-test-v2` 要求 `admin.version: 2`，但可以搭配 `backend.version: 3`，分别发布测试管理端 `v2` 和后端 `v3`；`app-test-v1.0.0` 则读取统一配置中的后端版本、APK 目录和发布说明，并从同一提交的 manifest 读取 App 版本；`back-test-v3` 要求 `backend.version: 3`。对应生产标签使用同样的版本校验规则。

tag 继续负责选择发布范围及环境，配置负责提供所选模块的构建和发布信息，不增加发布模块勾选字段。仅后端或 admin 发布不要求 APK 存在，也不校验无关 App 版本；App 发布才要求安装包、版本和发布说明齐全。

迁移完成后，App 发布版本仍从标签对应提交的 `frontend/manifest.json` 读取，不在 `release.yaml` 重复维护。发布时需核对标签版本、manifest 版本与 APK 内真实版本；不一致则报错停止。本阶段不修改 manifest 或要求用户改变现有打包流程。

服务器密码、SSH 密钥、MySQL 密码、JWT 密钥仍放在 GitHub Secrets 或服务器 `.env`，不写入统一发布配置。测试、生产的 API 地址等环境参数按标签选择对应构建配置；实际字段在实现阶段补齐，不使用隐式生产默认值。

迁移分两步完成：先确定本文档中的配置结构和职责；统一配置已准备，后续适配构建脚本、版本 SQL 生成脚本及工作流，并在同次迁移中删除 `deploy/releases/android.yaml`。新实现不长期双读两份发布配置。旧文件目前仅有 `release_notes` 业务字段，迁移时原样转入 `app.android.release_notes`。

本阶段只整理文档并准备统一配置：保留仍在使用的旧配置，不修改配置读取代码。当前脚本仍从 manifest 读取 App 版本、从 `android.yaml` 读取发布说明，不能将上述规划视为已经生效。

## 构建产物的存放约定

`deploy/releases/` 是手动打包的 Android APK 输入目录，日常只保留本次待发布的一个 APK，文件名可自定义。App 发布前先替换旧安装包，再将安装包、manifest 和发布说明随对应提交上传 Git；服务器上的历史 APK 仍按版本保存，不因仓库输入目录更换而删除。实际上传文件名由 manifest 中的版本信息生成。

后端和 admin 均从标签对应的源码在 CI 中构建，不要求本地先构建，也不把生成物提交到 `deploy/releases/`。后端程序使用构建工作区的 `bin/`，测试管理端使用 `admin/dist-test/`，构建后直接上传各自的服务器版本目录；跨 job 传递时使用工作流产物。现有 `.gitignore` 继续忽略这些生成目录。

因此管理端本地执行 `pnpm build:test` 后，无需将 `dist-test/` 再搬入 APK 目录。仓库中的手动发布输入和 CI 自动生成物分开存放；服务器上则分别位于 `backend/releases/vN/`、`admin/releases/vN/` 和 `packages/android/`。

## 构建和发布顺序

一次发布固定使用标签指向的提交，相关检查及构建通过后才上传产物。后端部署包括必要的数据库迁移、版本切换、宝塔项目重启及健康检查；后端成功后再发布所选的管理端或 App。

```text
推送发布标签
    ↓
识别发布范围及测试/生产环境，读取同一提交的统一发布配置
    ↓
校验所选模块的版本，检查及准备所需产物
    ↓
进入目标环境的串行发布流程
    ↓
后端迁移、部署和健康检查
    ├── admin-*：发布管理端静态文件，切换 admin/current 并检查访问
    ├── app-*：发布 APK，更新客户端版本记录
    └── back-*：发布结束
```

迁移后只有 App 入口使用统一发布配置中的 `app.android` 选择 APK，并生成 `app-version.sql`。其他入口不要求 APK，也不更新客户端版本记录。后端目录命名不再依赖 App 版本。

目前复制的 APK 仍为真实版本 `0.0.10`，不通过重命名将其标记为 `1.0.0`。测试 App 发布 `1.0.0` 时，需要更新 manifest 并重新打包，本阶段不修改 App 版本或安装包。

## 环境、失败与兼容性

发布根据上述标签规则选择环境，使用该环境自己的根目录、配置文件、数据库、宝塔项目及连接凭据：

| 配置项 | 测试环境（带 `-test-`） | 生产环境（不带 `-test-`） |
| --- | --- | --- |
| 根目录 | `/www/wwwroot/Leave_test` | `/www/wwwroot/Leave` |
| 后端 YAML | `config/config.test.yaml` | `config/config.server.yaml` |
| 数据库 | `leave_list_test` | `leave_list` |
| 宝塔 Go 项目 | `Leave_test` | `Leave` |
| 管理端站点根目录 | `Leave_test/admin/current` | `Leave/admin/current` |

测试 APK、客户端版本 SQL、migration、上传文件和备份均只处理测试环境的目录及数据库。测试 App 应连接测试后端，版本记录写入测试库的 `app_versions`；测试发布不能更新生产客户端版本记录或重启生产项目。

发布并发按环境隔离：所有测试入口共用测试环境的发布锁，所有生产入口共用生产环境的发布锁。同一环境内串行，两个环境可以独立发布。凭据和脚本参数也需按环境绑定，禁止缺少测试配置时回退到生产环境。

后端部署或健康检查失败时，不继续发布管理端或 App。后续管理端或 App 发布失败时，已成功部署的后端不会自动回滚，需根据失败阶段检查并恢复。三个入口也不构成跨程序、数据库和客户端版本记录的原子事务。

后端变更需兼容仍在使用的旧 App，以及发布切换前的旧管理端。数据库 migration、程序回滚和客户端版本记录分别处理，不因回滚程序就自动撤销数据库变更。

## 后续实施清单

- [ ] 第一版实现三个测试标签入口及版本校验，共用后端部署逻辑；不启用新生产入口。
- [x] 准备完整的 `admin/.env.test.example`，保留实际环境文件的忽略规则。
- [ ] CI 构建前从测试模板生成 `.env.test`，使用现有 `pnpm build:test` 并上传 `dist-test/`。
- [x] 创建统一发布配置 `deploy/release.yaml`，填写现有模块版本、APK 目录和发布说明。
- [ ] 接入统一配置读取并校验所选模块字段。
- [ ] 将 `android.yaml` 的发布说明迁入统一配置，同步改造读取脚本并删除旧文件；保留从 manifest 读取 App 版本。
- [ ] 组合发布读取 `backend.version`，按标签绑定测试、生产配置和凭据。
- [ ] 适配 `backend/releases/vN`、`backend/current` 和管理端目录，落实版本复用及禁止覆盖规则。
- [ ] 按环境串行执行发布，不使用每个标签各自独立的并发组保护同一个环境。
- [ ] 完成管理端 CI，验证现有测试站点替换产物及 API 代理的回归行为。
- [ ] 后续单独接入生产标签、生产构建配置和环境凭据。
- [ ] 仅在 App 入口处理 APK 和客户端版本记录，校验 manifest、标签与 APK 真实版本一致，从指定目录唯一确定安装包。
- [ ] 使用 `back-test-v1`、`admin-test-v1`、`app-test-v1.0.0` 验证三个入口、重复后端发布、失败中止和回滚，并验证测试标签不会触发生产发布，再配置生产入口。

未勾选项为后续工作。本次仅整理文档和准备配置，不修改构建脚本、GitHub 工作流或执行服务器发布。
