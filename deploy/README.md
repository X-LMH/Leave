# Leave 测试与生产环境 CI/CD

测试与生产共用构建和发布流程，通过标签选择环境。App 继续由 HBuilderX 本地打包，后端和 admin 在 GitHub Runner 从标签对应的提交构建。日常发布保留服务器维护的配置、Nginx 和宝塔项目；首次生产 admin 建站见下文。

## 发布入口与版本

| 测试标签 | 生产标签 | 发布内容 | 版本来源 |
| --- | --- | --- | --- |
| `back-test-vN` | `back-vN` | 后端 | `deploy/release.yaml` 的 `backend.version` |
| `admin-test-vN` | `admin-vN` | 后端成功后发布 admin | admin 标签匹配 `admin.version`，后端使用 `backend.version` |
| `app-test-vX.Y.Z` | `app-vX.Y.Z` | 后端成功后发布 APK、更新客户端版本记录 | manifest 的 `versionName`、`versionCode`；后端使用 `backend.version` |

后端和 admin 使用正整数版本，目录名为 `v1`、`v2`；App 使用三段版本号及递增的 versionCode。两个环境共用 release.yaml 的模块版本，同一提交可以先发布测试标签，再发布生产标签。生产标签不带 `test-`，不接受 `v*` 或 `*-prod-v*`。组合发布的后端版本可以与 admin 版本不同；仅后端或 admin 发布不要求 manifest 或 APK 存在。

版本校验读取标签指向的提交，各标签只比较自身模块的版本。不一致时在 metadata 阶段中止，列出标签、字段和配置版本。修改配置后需提交并为正确提交创建符合版本的标签；不能只修改工作区后重跑旧标签，也不要覆盖已发布的远端标签。

`deploy/release.yaml` 只维护版本、APK 目录和发布说明，不重复维护固定构建参数或 App 版本。App APK 目录直接包含的 `.apk` 文件必须恰好一个，允许任意输入文件名，不按修改时间挑选。CI 用 Android SDK `apkanalyzer` 核对真实版本，发布名为 `leave-android-<versionCode>-<versionName>.apk`。重命名不能改变安装包版本；脚本不修改 manifest、不自动打包 App。

## 服务器环境与目录

| 项目 | test | prod |
| --- | --- | --- |
| 根目录 | `/www/wwwroot/Leave_test` | `/www/wwwroot/Leave` |
| 后端 current | `backend/current` | `backend/current` |
| 后端版本目录 | `backend/releases/vN` | `backend/releases/vN` |
| admin current | `admin/current` | `admin/current` |
| admin 版本目录 | `admin/releases/vN` | `admin/releases/vN` |
| YAML | `config/config.test.yaml` | `config/config.server.yaml` |
| loopback MySQL | `leave_list_test` | `leave_list` |
| 后端 HTTP 端口 | `9000` | `8000` |
| admin HTTP 端口 | `80` | `80` |
| 宝塔 Go 项目 | `Leave_test` | `Leave` |
| 宝塔配置的运行用户 | `www` | `www` |
| admin 域名 | `admin-test.leave.mhspace.dpdns.org` | `admin.leave.mhspace.dpdns.org` |

生产与测试使用相同的模块目录布局。旧生产目录完整保留为同级 `/www/wwwroot/Leave_old`，新的生产目录为 `/www/wwwroot/Leave`。宝塔生产项目的启动文件设为 `/www/wwwroot/Leave/backend/current/Leave`，使用面板配置的 `www` 用户启动；配置文件和 .env 需对该用户可读。

各环境根目录同时维护：

```text
config/                             # 服务器 YAML，不随发布覆盖
.env                                # MYSQL_PASSWORD、REDIS_PASSWORD、JWT_SECRET，不上传
admin/current -> releases/vN/
admin/releases/vN/                  # index.html、assets 等
packages/android/                   # 保留历史 APK
uploads/                            # 持久上传文件
database/migrations/                # 累积历史，不修改同名文件
scripts/migrate                     # 最新迁移工具
backups/mysql/                      # 数据库备份
backups/deploy/                     # 历史备份及运行记录目录
.deploy/
├── deploy.lock
└── <运行ID-尝试次数>/
    ├── bundle.tar.gz
    ├── bundle/                     # 校验和、元数据、产物、本次脚本
    └── status                      # 当前或失败阶段
```

后端每个新版本目录包含 Leave、release.json。迁移工具仅放在 scripts/migrate，不随 backend/releases/vN 保存副本。旧生产版本保留在 Leave_old 中；新生产目录使用整数版本，不将历史目录自动补齐为新版本。

从旧发布布局升级时，先确认 scripts/migrate 存在且可执行；核对 backend/releases/vN/migrate 与其内容一致后，清理版本目录中的重复副本。新版校验要求后端版本目录仅包含 Leave 和 release.json，保留旧副本会阻止同版本复用发布。

环境、根目录、数据库、宝塔项目及构建目录在 `deploy/release.rb` 中固定映射，不使用任意 `SERVER_APP_ROOT`。服务端再次检查发布包环境与目标环境一致。YAML 的 MySQL 必须是对应数据库和 loopback 地址，package_dir 必须为该环境根目录下的 `packages/android`，storage.root_dir 必须为该环境根目录下的 `uploads`。

参考生产模板 [`config.server.yaml.example`](../backend/config/config.server.yaml.example)，首次启用时将现有生产 YAML 的 app.port 从 8080 改为 8000，并将 API 站点的上游改为 127.0.0.1:8000；保留其余配置和凭据。生产根目录不得混放 config.local.yaml 或 config.test.yaml，测试根目录也不得混放生产 YAML，以保证服务自动搜索与迁移工具选择相同配置；旧 config.yaml 不参与当前搜索。配置及 .env 需让对应运行用户可读，例如测试使用 root:www 0640，真实凭据不提交 Git。

宝塔 PID 文件为 `/var/tmp/gopids/<项目名>.pid`，启动脚本为 `/www/server/go_project/vhost/scripts/<项目名>.sh`。发布不改写面板脚本，不另建启动流程；按 PID 核验程序路径及用户后停止，再调用既有脚本重启。PID 不符、停止超时或端口仍被占用均中止，不按端口批量杀进程。

## admin 构建与打包

完整可提交模板为 [`admin/.env.test.example`](../admin/.env.test.example)、[`admin/.env.prod.example`](../admin/.env.prod.example)。字段结构一致：根路径、history 路由、同域 `/api/v1`、关闭开发代理与 source map。存储前缀分别为 leave-admin-test、leave-admin-prod，首次使用生产前缀需要重新登录。

在 admin 目录执行，已有环境文件时先核对差异，避免覆盖：

```bash
cp .env.prod.example .env.prod
pnpm install --frozen-lockfile
pnpm typecheck
pnpm exec oxlint
pnpm exec eslint .
node --test tests/*.test.mjs
pnpm build:prod
```

| 命令 | Vite mode | 配置 | 产物 | 本地预览 |
| --- | --- | --- | --- | --- |
| `pnpm build` / `pnpm build:prod` | prod | .env.prod | admin/dist/ | `pnpm preview:prod` |
| `pnpm build:test` | test | .env.test | admin/dist-test/ | `pnpm preview:test` |

CI 从干净检出复制所选模板构建，并仅打包所选目录。Vite 还读取 .env、.env.local 等本地覆盖文件，需核对本地设置；VITE_* 会进入浏览器，不能包含凭据。参见 [Vite 环境变量与模式](https://vite.dev/guide/env-and-mode)。静态部署的 API 由 Nginx 代理，本地 preview 仅预览产物，不能替代站点联调。环境文件和构建产物继续忽略，不提交 Git，也不放入 APK 输入目录。

后端本地 Linux/amd64 交叉构建入口保持 `bash deploy/build-linux.sh`，输出 bin/Leave、bin/migrate，关闭 CGO。CI 的 Go 版本取自 go.mod，admin 使用 Node 22、pnpm 10。发布包包含所需 migration、模块产物、脚本、release.json 和 SHA-256 清单，不含服务器 YAML、.env 或上传文件。

## 首次生产 admin 建站

生产 admin 站点通过 Nginx 的 `80` 端口按域名提供服务，CD 使用相同端口检查同域 API。新环境首次建站时，在宝塔新增 `admin.leave.mhspace.dpdns.org` 静态站点，监听端口设为 `80`，根目录设为 `/www/wwwroot/Leave/admin/current`，合入 [`nginx.admin.prod.conf`](nginx.admin.prod.conf) 的 location 配置，保留面板生成的证书与验证目录配置，并将 DNS 指向生产服务器。已有站点无需重复创建。

模板将同域 `/api/v1/` 转发到 127.0.0.1:8000，保留接口前缀；头像使用生产 uploads 目录，history 路由回退 index.html，首页不缓存，缺失 assets 返回 404。站点根目录可等待首次发布创建，API 代理必须先就绪。

```bash
nginx -t
# 检查通过后在宝塔重载 Nginx，再核验同域代理：
curl --fail -H 'Host: admin.leave.mhspace.dpdns.org' http://127.0.0.1:80/api/v1/health
```

日常发布不会创建站点、修改 DNS 或改写 Nginx。admin 发布先启动后端并确认其健康，再检查站点的同域健康接口；检查失败时保留已启动的后端，不切换 admin。这样旧后端已停止或上次重启失败时，重试仍可恢复后端。上述首次服务器准备需另行执行，源码构建通过不代表已上线。

## 工作流与发布顺序

`.github/workflows/ci.yml` 在 master 分支 push 时执行完整 CI，并通过 workflow_call 供测试、生产发布复用相同检查。CI 执行 Go 格式、vet、单元及 MySQL 集成测试、Ruby 元数据测试、Bash 部署夹具测试、App 版本 SQL 测试及 Linux 编译；admin 发布继续执行类型检查、不自动修复的 lint、现有 Node 测试及所选环境构建。

CI 使用 Runner 内独立的 MySQL 8 service，并配置 ADMIN_TEST_MYSQL_DSN、MYSQL_TEST_DSN 和 DEPLOY_TEST_MYSQL=1。集成测试创建、清理自己的临时 schema，不使用业务库或服务器配置。本地缺少独立 MySQL 时，相关测试仍会跳过，不能将跳过视为通过。

```text
标签与版本校验、所需 APK 真实版本校验
    ↓
后端检查/构建；admin 入口同时检查并构建对应环境
    ↓
下载产物，生成 SHA-256 发布包并上传独立运行目录
    ↓
校验发布包、环境配置、current 和所有所选版本
    ↓
核验当前 PID，安装不可变后端版本（暂不切换 current）
    ↓
同步并执行 migration
    ↓
再次核验 PID，停止旧后端，原子切换 current
    ↓
调用对应宝塔脚本，核验 PID、程序、用户和健康接口
    ↓
admin 入口先检查同域健康接口，再发布 admin；或发布 APK / 结束后端发布
```

测试、生产分别使用 leave-test-deploy、leave-prod-deploy 并发组，queue=max，不取消正在运行的发布；服务器各环境再用 flock 避免同时迁移或替换。新后端不得继承部署锁文件描述符。

同名 migration 内容相同则复用，不同则拒绝；仅新增不存在的文件，不删除历史。迁移在旧后端运行期间执行，必须兼容旧版本；删除、重命名旧程序仍使用的字段需分阶段发布。迁移失败时旧服务继续运行，新版本文件可保留供重试。

## 历史密码迁移

历史明文密码通过独立的 `password-migrate` 命令手动转换为 Argon2id，不通过 HTTP 接口，也不随 CI/CD 自动执行。构建、上传、预览、停服迁移和恢复步骤见 [密码迁移说明](password-migration.md)。此数据转换需要暂停账号写入，与可兼容旧服务的表结构迁移分开执行。

## Secrets 与生产发布操作

部署 job 按标签引用 GitHub Environment test 或 prod。沿用四项 Secrets：SERVER_HOST、SERVER_PORT、SERVER_USERNAME、SERVER_SSH_KEY。可以继续使用仓库级 Secrets，也可在各环境配置同名 Secrets（不同服务器使用各自环境配置）。Actions 仍使用密钥，不将服务器密码写入工作流。环境限制在 GitHub 设置中维护，参见 [GitHub 部署环境](https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments)。

1. 确认服务器 YAML、.env、宝塔项目及 SSH Secrets；admin 入口先完成站点与 DNS 准备。
2. 将代码、release.yaml、模板提交到待发布版本。App 还需在同一版本提交真实 APK、manifest 和发布说明。
3. 先用对应测试标签验证，再给同一提交创建生产标签。当前配置为后端 1、admin 1，对应 back-v1、admin-v1，App 为 app-v0.0.10；App 按该提交 manifest 版本生成标签。
4. 推送标签，观察 metadata、backend、admin（按需）和 deploy。标签校验后发生的临时故障可重新运行 Actions，无需移动标签。
5. 发布后检查公网首页、history 子页刷新、登录及真实业务 API；App 检查下载和更新。脚本的本机 HTTP 检查不代替浏览器或业务验证。

服务端调用为 `bash deploy-release.sh RUN_ID-RUN_ATTEMPT prod`；省略环境只进入 test。非法环境名、跨环境发布包均拒绝。发布脚本测试仅用临时夹具和 mock，不部署业务服务。

## 版本冲突、重试与恢复

产物不覆盖：不存在则创建；已有且文件完整、SHA-256 相同则复用；不同或不完整则在迁移和停服前中止。后端比较 Leave 和版本元数据，不按整个仓库提交号或组合标签判断相同。迁移工具由发布包 SHA-256 清单校验，部署时更新 scripts/migrate。现有后端元数据还必须属于同一环境。相同产物重试仍迁移、重启及健康检查。App 数据按 platform+version_code 幂等更新，其他已发布 Android 版本归档。

每次发布保留独立运行目录、SHA256SUMS、release.json、status。任何前置失败都不执行后续操作，后端健康失败不发布 admin/App；后续模块失败保留已成功部署的后端，不自动回滚数据库。

Actions 重试生成新尝试目录，优先重新运行失败 job 以复用原构建产物。重新构建可能改变 admin 构建时间，导致同版本字节不同，此时应复用原产物或递增版本。已有旧目录缺合法 release.json 或后端文件不完整时不自动补齐。

恢复程序或页面时将各自 current 指向本环境保留的旧版本，再通过对应宝塔项目重启。恢复旧生产目录 Leave_old 时，需同步核对宝塔启动路径及持久文件路径；程序恢复不撤销 migration 或客户端版本记录，数据库恢复需结合备份、已执行迁移及独立授权处理。
