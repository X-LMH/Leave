# Leave 测试环境 CI/CD

第一版只发布测试环境，不响应生产标签。生产服务器配置、Nginx、数据库和宝塔项目保持原样。App 继续由 HBuilderX 本地打包，后端和 admin 在 GitHub Runner 从标签对应的提交构建。

## 发布入口与版本

| 标签 | 发布内容 | 版本来源 |
| --- | --- | --- |
| `back-test-v1` | 后端 | `release.yaml` 的 `backend.version` |
| `admin-test-v1` | 后端成功后发布 admin | 后端、admin 各自的配置版本 |
| `app-test-v1.0.0` | 后端成功后发布 APK、更新客户端版本记录 | 后端版本来自配置，App 版本来自 manifest |

后端和 admin 使用正整数版本，目录名为 `v1`、`v2`；App 使用三段版本号及递增的 versionCode。tag 版本必须与对应模块的版本来源一致。组合发布的后端版本可以与 admin 版本不同；仅后端或 admin 发布不要求 manifest 或 APK 存在。

版本校验读取标签指向的提交：`back-test-vN` 只比较 `backend.version`，`admin-test-vN` 只比较 `admin.version`，`app-test-vX.Y.Z` 只比较 manifest 的 `versionName`。admin 和 App 仍读取后端配置版本用于部署后端，但不将自己的标签版本与后端版本比较。

不一致时，Actions 在 metadata 阶段报错并停止，后续构建和部署不会执行。错误会列出标签版本、配置字段和配置版本，例如 `back-test-v2（标签版本 2）与 deploy/release.yaml 的 backend.version=1 不一致，发布已停止`。如配置版本写错，修改配置并提交后，为正确提交重新创建符合版本的标签；如标签写错，创建与原提交配置一致的正确标签。不要只修改工作区后重新运行旧标签：重试仍读取旧标签指向的提交，也不要为修正标签随意覆盖已经发布的远端标签。

`deploy/release.yaml` 只维护版本、APK 目录和发布说明，不重复维护固定构建参数或 App 版本。旧 `releases/android.yaml` 已移除。App 版本仍读取 `frontend/manifest.json` 的 `versionName`、`versionCode`。

App 发布时，配置的 APK 目录直接包含的 `.apk` 文件必须恰好一个，文件名可以自定义，不能按修改时间挑选。CI 使用 Android SDK `apkanalyzer` 核对安装包真实版本，再命名为 `leave-android-<versionCode>-<versionName>.apk`。已有 `0.0.10` 安装包不能通过重命名变成 `1.0.0`。

版本产物不覆盖：发布前检查所有选定模块，版本不存在则创建；已有且文件完整、SHA-256 相同则复用；不同或不完整则在迁移、停服之前中止，保持服务运行。后端比较 Leave 和 migrate，元数据仅用于确认版本，不以整个仓库提交号或组合标签判断相同。相同产物重试仍执行 migration、宝塔重启和健康检查。App 数据按 `platform + version_code` 更新，不重复插入记录；其他已发布 Android 版本标记为 archived。

## 测试服务器目录

```text
/www/wwwroot/Leave_test/
├── backend/
│   ├── current -> releases/v1/
│   └── releases/v1/
│       ├── Leave
│       ├── migrate
│       └── release.json
├── admin/
│   ├── current -> releases/v1/
│   └── releases/v1/
│       ├── index.html
│       └── assets/
├── packages/android/                 # 按版本保存 APK
├── config/config.test.yaml           # 服务器维护，不随发布覆盖
├── .env                              # MYSQL_PASSWORD、JWT_SECRET
├── database/migrations/              # 累积历史，不修改已上传的同名文件
├── uploads/                          # 持久上传文件
├── scripts/migrate                   # 最新迁移工具
├── backups/
│   ├── mysql/
│   └── deploy/<运行ID-尝试次数-时间>/ # 保留的历史备份（新发布不再覆盖版本）
└── .deploy/
    ├── deploy.lock
    └── <运行ID-尝试次数>/
        ├── bundle.tar.gz
        ├── bundle/                   # 校验和、元数据、产物及本次发布脚本
        └── status                    # 当前阶段或失败阶段
```

协调脚本将根目录固定为 `Leave_test`，不使用 `SERVER_APP_ROOT`，不允许回退到生产。测试 YAML 必须指向 loopback MySQL 的 `leave_list_test`，APK 和 uploads 存储路径必须位于测试根目录。配置和 `.env` 需让宝塔 `www` 用户可读，例如 `root:www 0640`；不要将真实凭据提交到 Git。

测试宝塔 Go 项目为 `Leave_test`：启动文件为 `/www/wwwroot/Leave_test/backend/current/Leave`，PID 文件为 `/var/tmp/gopids/Leave_test.pid`，启动脚本为 `/www/server/go_project/vhost/scripts/Leave_test.sh`。后端支持从程序目录向上三级读取测试 YAML 和 `.env`。

发布核验 PID 对应的程序及用户后停止该进程，替换文件，再以 `www` 用户执行宝塔自动生成的脚本。不会修改该脚本或另建服务启动流程；面板继续管理同一个项目。不按端口批量杀进程，PID 不符、停止超时或端口仍被占用都中止发布。

## admin 构建与站点

可提交的完整模板为 [`admin/.env.test.example`](../admin/.env.test.example)。CI 将它复制为 `.env.test`，按锁文件安装依赖，执行类型检查、不自动修复的 lint、现有 Node 测试及 `pnpm build:test`。

本地在 `admin/` 下可以执行：

```bash
cp .env.test.example .env.test
pnpm build:test
```

已有 `.env.test` 时先检查差异，避免覆盖本地配置。产物为 `admin/dist-test/`，由 CI 上传到服务器管理端版本目录，不放入 APK 输入目录，不提交 Git。`.env`、`.env.test` 和产物继续忽略，无需删除 `.gitignore`。`VITE_*` 会进入浏览器产物，只能填写公开设置。

复用现有测试站点 `admin-test.leave.mhspace.dpdns.org`，根目录为 `/www/wwwroot/Leave_test/admin/current`，同域 `/api/v1` 由 Nginx 代理到测试后端。发布不重建站点、不修改 Nginx。发布后通过本机 Host 请求核对首页内容、入口 JS/CSS 和同域健康接口；这不代替公网、浏览器或完整业务验证。

## 工作流与发布顺序

`.github/workflows/ci.yml` 保留 master 分支 CI，并供测试 CD 复用：后端格式、vet、测试、Linux/amd64 编译，以及发布脚本、隔离 MySQL SQL 测试。`ADMIN_TEST_MYSQL_DSN` 指向 Runner 内的 MySQL，不能使用业务配置或服务器数据库。Go 版本读取 go.mod，CGO 关闭；admin 使用 Node 22、pnpm 10。

`.github/workflows/deploy.yml` 仅匹配三个测试标签。旧 `v*` 及生产入口已停用；版本校验通过后发生的临时故障可通过 Actions 页面重新运行，不需要移动或重建 Git tag。标签与配置不一致则按上文修正。

```text
读取并校验标签、配置及所需 APK
    ↓
后端 CI；admin 入口同时进行管理端检查和构建
    ↓
下载产物，生成带 SHA-256 校验和的发布包
    ↓
串行上传到测试运行目录
    ↓
校验文件、测试配置及 current 链接；检查所有所选版本是否可复用
    ↓
预检当前后端 PID；准备并安装后端版本文件（尚不切换 current）
    ↓
同步并执行 migration
    ↓
再次核验 PID，停止旧后端，切换 current
    ↓
调用测试宝塔脚本，核验 PID、运行程序和健康接口
    ↓
按入口发布 admin / APK，或结束后端发布
```

同名 migration 内容相同则复用，不同则拒绝；仅新增不存在的文件，不删除历史文件。所有测试部署共用 `leave-test-deploy` 并发组，`queue: max`、不取消正在运行的发布；服务器另用 flock，避免不同发布同时替换文件或执行数据库操作。

迁移之前先核验当前进程身份、准备并安装不可变的后端文件；此时旧服务和 current 保持原样。迁移失败时，新版本目录可以保留供相同产物重试，旧服务继续运行。停服前再次核验 PID，避免预检之后进程变化导致误停。迁移在旧后端运行期间执行，因此数据库变更必须兼容旧版本；删除或重命名仍被旧程序使用的字段应分阶段发布。

## GitHub Secrets 与发布准备

沿用四项 Secrets：`SERVER_HOST`、`SERVER_PORT`、`SERVER_USERNAME`、`SERVER_SSH_KEY`。当前脚本使用已有的特权 SSH 账号，随后以 www 启动宝塔项目。SSH 密码不写进工作流；密钥必须允许对应账号登录。

推送标签前：

1. 提交本次代码、release.yaml、环境模板；App 发布还需将唯一 APK、manifest 和发布说明提交到同一版本。
2. 检查对应模块的 tag 版本和版本来源一致。
3. 确认测试配置、宝塔项目和 Nginx 站点已就绪，四项 Secrets 已配置。
4. 推送对应测试标签，观察 Actions 的检查和发布结果。

当前仓库的 App 版本仍是 `0.0.10`，可用对应标签测试现有包；准备好 `1.0.0` 的真实测试包后，再更新 manifest 并替换 APK。脚本不会修改 App 版本或自动打包。

## 失败、重试与恢复

每次发布使用独立运行目录，保留 SHA-256 清单、release.json 和 status，便于定位失败阶段。任何前置步骤失败都不继续执行后续操作。后端健康检查失败不发布 admin 或 App；后续模块失败保留已成功部署的后端，不自动回滚数据库。

Actions 重试会生成新的尝试目录；优先重新运行失败的 job，以复用已构建的产物。版本文件一致时复用并继续迁移、重启及健康检查。同名 migration 保持一致，App 版本记录幂等更新。重新构建若产生不同字节（例如 admin 构建时间变化），也会拒绝同版本发布，需使用原产物或递增版本。已有旧目录缺少 migrate、合法 release.json 或文件不完整时不自动补齐，应递增版本重新发布。

恢复程序或页面时将 current 指向保留的旧版本，后端仍通过测试宝塔项目重启。程序恢复不会撤销 migration 或 App 版本记录，不能将数据库回滚等同于文件恢复。暂不自动清理历史 APK、备份和失败运行目录。

## 验证边界

本次代码实现不推送标签、不上传服务器文件、不执行真实发布。实际 GitHub Actions、SSH Secrets、三种测试发布及公网/浏览器回归仍需后续验证；本地临时目录和隔离 MySQL 测试不能替代这些验证。
