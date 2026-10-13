# Leave 部署说明

本项目采用单台 Linux 服务器、宝塔和 Nginx 部署，不要求使用 Docker。App 安装包由 HBuilderX 本地打包，GitHub Actions 负责后端构建及发布。

## 当前状态与调整范围

生产环境 `/www/wwwroot/Leave` 仍使用原有根目录 `current`、`releases/` 结构。现有 `.github/workflows/deploy.yml` 在推送 `v*` 标签或手动触发时，执行 App + 后端组合发布；管理端尚未接入 CD。

下文的模块目录是新的部署布局，已在 `/www/wwwroot/Leave_test` 中准备。现有工作流及发布、重启脚本尚未适配该布局，不能直接用它们发布到测试目录。目前已准备目录、部分已有文件和测试配置，并将 `leave_list_local` 的表结构及数据复制到独立测试库；后端已支持宝塔从程序目录启动时读取根目录配置；测试服务由用户在宝塔面板启动和停止，不由助手启动。未运行 migration 工具，未修改生产服务。

## 新服务器目录

生产与测试环境分别使用独立根目录：

- 生产：`/www/wwwroot/Leave`
- 测试：`/www/wwwroot/Leave_test`

两个环境采用相同的目标布局，以测试目录为例：

```text
/www/wwwroot/Leave_test/
├── backend/
│   ├── current -> releases/v1/
│   └── releases/
│       └── v1/
│           ├── Leave
│           └── app-version.sql          # 仅 App 发布时生成并保存
├── admin/
│   ├── current -> releases/v1/           # 管理端产物就绪后创建
│   └── releases/
│       └── v1/
│           ├── index.html
│           └── assets/
├── packages/
│   └── android/
│       └── leave-android-<versionCode>-<versionName>.apk
├── config/
│   └── config.test.yaml                 # 测试配置；生产使用 config.server.yaml
├── .env.example                         # 环境变量模板
├── .env                                 # 配置测试环境后创建
├── database/
│   └── migrations/
├── uploads/
│   └── avatars/
├── backups/
│   └── mysql/
└── scripts/
    ├── migrate
    ├── run-migrations.sh
    └── publish-version.sh
```

`backend/current` 与 `admin/current` 分别指向各自的发布版本。发布目录保存不可变的产物，配置、数据库迁移、上传文件、APK 和备份保存在根目录的共享位置，不随版本切换迁移。

后端和管理端版本采用整数形式 `v1`、`v2`、`v3`，测试初始版本为 `v1`；各自的版本目录只包含版本号，不包含 tag 前缀。App 使用三段版本号，例如 `v1.0.0`，新测试 App 从 `1.0.0` 开始。tag 前缀表示发布入口，例如 `admin-v1`、`app-v1.0.0`、`back-v1`。

后端和管理端的版本分别管理，App 的三段版本不能直接用作后端目录名。组合发布时需明确本次后端版本；重复发布同一后端版本只能复用相同程序，不允许用不同程序覆盖已有版本。这些规则待后续 CD 改造落实，本阶段不修改工作流。之前导入的生产历史版本 `v0.0.10` 保留原名，新测试版本统一采用上述规则。

App 安装包继续使用 `manifest.json` 中的 `versionCode` 和 `versionName` 命名。标签用于标识发布，不能替代 Android 安装包本身的版本信息。

## 测试目录已准备的文件

从当前生产版本 `v0.0.10` 复制以下文件，并逐一核对 SHA-256：

- 后端程序：`backend/releases/v0.0.10/Leave`。
- APK：`packages/android/leave-android-10-0.0.10.apk`。
- 生产现有的 16 个 migration 文件：`database/migrations/`。
- 默认头像：`uploads/avatars/default.png`。
- 可复用工具：`scripts/migrate`、`scripts/run-migrations.sh`、`scripts/publish-version.sh`。

已将当前本地后端交叉编译为 Linux/amd64（`CGO_ENABLED=0`），上传后按新命名规则整理为 `backend/releases/v1/Leave`，`backend/current` 指向 `releases/v1`。程序及 YAML 上传前后 SHA-256 一致，目录改名后程序校验值未改变；程序为 amd64 ELF，可执行权限为 `0755`。之前导入的 `v0.0.10` 后端保留在历史目录中。管理端已创建空目录 `admin/releases/v1/`，尚无构建产物，不创建 `admin/current`。

测试配置使用从本地上传的 `config/config.test.yaml`，端口为 `9000`，数据库/用户为 `leave_list_test`；APK 目录为 `/www/wwwroot/Leave_test/packages/android`，上传目录为 `/www/wwwroot/Leave_test/uploads`。资源地址暂为 `https://test.example.invalid/api/v1/uploads`，需要在测试 Nginx 站点配置完成后替换。旧 YAML 已替换，旧 `config.test.yaml.example` 已删除，测试 YAML 与 `.env` 权限为 `root:www 0640`，允许宝塔配置的 `www` 用户读取。测试目录不保留 `config.server.yaml`；生产配置文件名保持不变。

按用户要求将本地 `.env` 上传到测试根目录，保留本地的 `JWT_SECRET`，将 `MYSQL_PASSWORD` 替换为测试库密码；`.env.example` 保持空值模板。测试配置使用 `127.0.0.1:3306` 连接 `leave_list_test`，已验证登录。

复制前确认 `leave_list_test` 没有表。从 `leave_list_local` 使用一致性事务导出表结构及数据，再导入测试库；源库只读，未操作生产库 `leave_list`。复制后逐表核对表列表、行数和扩展校验值，9 张表共 189 行全部一致。导出文件保存在测试目录 `backups/mysql/leave_list_local-20261013-183233.sql`，权限为 `0600`；携带源库凭据的临时上传文件已删除。

后端配置不是通过 `*` 通配符选择。服务和迁移工具共用 `FindConfigFile`，在工作目录的 `config/` 下依次选择 `config.local.yaml`、`config.test.yaml`、`config.server.yaml`，找到第一个文件后读取；保留旧发布布局的 `../../config/config.test.yaml`、`../../config/config.server.yaml` 回退，并支持新布局的 `../../../config/config.test.yaml`、`../../../config/config.server.yaml`。不匹配 `.example` 或其他 YAML 文件。

宝塔面板会将程序目录作为工作目录，并在保存时重新生成启动脚本。本布局下 `backend/current` 解析为 `backend/releases/v1`，后端可向上三级读取环境根目录的 YAML 和 `.env`，不要求手动修改宝塔脚本中的 `cd`。从环境根目录直接运行也保持支持。新后端程序已上传并支持测试 YAML；`scripts/migrate` 仍为之前复制的工具，调用时需通过 `-config` 显式指定测试配置。

本阶段按用户要求不修改构建、CD 工作流和部署脚本。已有部署脚本仍默认使用 `config.server.yaml`；独立调用迁移、客户端版本脚本时可显式传入 `CONFIG_FILE` 指向测试文件。迁移程序的 `-config` 参数仍可显式指定文件。后续适配测试 CD 时再统一配置路径和新发布目录，不直接执行现有组合发布流程。

没有复制真实用户头像、历史发布程序、数据库备份、历史版本发布 SQL，以及依赖原目录和生产宝塔项目的 `deploy-release.sh`、`switch-and-restart.sh`。复制的数据库迁移和客户端版本发布工具未执行；测试后端交由用户在宝塔面板管理。

## 计划采用的三个 CD 入口

以下是已确定的发布方案，尚未写入工作流：

| 标签示例 | 发布范围 | 执行顺序 |
| --- | --- | --- |
| `admin-v1` | 后端 + admin | 后端检查、构建、迁移及部署成功后，发布管理端 |
| `app-v1.0.0` | 后端 + App | 后端检查、构建、迁移及部署成功后，发布 APK 并更新客户端版本记录 |
| `back-v1` | 仅后端 | 后端检查、构建、迁移及部署 |

三个入口复用同一套后端部署逻辑。同一环境的生产发布必须串行执行，不能按不同标签各自并发切换同一个后端。三者联合更新时，分别触发 admin 和 App 入口，接受后端重复部署。

只有 App 入口读取 `frontend/manifest.json`、选择 APK，并根据 `deploy/releases/android.yaml` 生成 `app-version.sql`。其他入口不要求 APK，也不更新客户端版本记录。后端目录命名不再依赖 App 版本。

目前复制的 APK 仍为真实版本 `0.0.10`，不通过重命名将其标记为 `1.0.0`。测试 App 发布 `1.0.0` 时，需要更新 manifest 并重新打包，本阶段不修改 App 版本或安装包。

## 测试环境运行前的配置

测试后端由宝塔面板管理，环境配置要求及待完成事项如下：

1. 检查 `config/config.test.yaml` 和 `.env`。测试数据库必须与生产数据库隔离；使用已导入数据并验证登录的 `leave_list_test`，不要将目标配置改为源库 `leave_list_local`。
2. 确认测试端口未被占用，配置测试域名及存储资源地址。
3. 在宝塔中创建独立 Go 项目 `Leave_test`。启动文件使用 `/www/wwwroot/Leave_test/backend/current/Leave`，面板可按默认方式从程序目录启动，后端会读取环境根目录的测试配置与 `.env`。
4. 测试项目使用独立 PID 文件和启动脚本，部署逻辑不能回退到生产项目 `Leave` 的默认值。
5. 管理端发布后，将测试站点根目录设为 `/www/wwwroot/Leave_test/admin/current`，配置 HTTPS、API 反向代理；采用 history 路由时配置页面回退。头像、APK 路径根据实际后端接口及现有 Nginx 映射配置。
6. 适配工作流和切换脚本到 `backend/current`、`backend/releases/` 新路径，再验证后端健康接口 `/api/v1/health` 和管理端访问。

测试配置、密钥和运行产物不提交到 Git。本地文档记录目录和流程，不保存服务器密码或真实配置内容。

## 当前工作流与脚本

现有 CI 为 `.github/workflows/ci.yml`，负责后端检查、测试和编译。现有 CD 为 `.github/workflows/deploy.yml`，先复用 CI，再上传 APK、后端程序、迁移工具、migration、客户端版本 SQL 和部署脚本。

目前服务器侧 `deploy-release.sh` 依次执行数据库迁移、客户端版本记录写入、后端切换与宝塔重启；发布目录仍为 `$APP_ROOT/releases/v$VERSION_NAME`，启动文件仍为 `$APP_ROOT/current/Leave`。这些路径和 App 绑定关系都需要后续改造，不能仅改变 `APP_ROOT` 就视为适配完成。

当前工作流使用 GitHub Secrets：`SERVER_HOST`、`SERVER_PORT`、`SERVER_USERNAME`、`SERVER_SSH_KEY`、`SERVER_APP_ROOT`。新流程需要区分测试和生产的连接信息、根目录及宝塔项目配置。

数据库 migration 为累积历史，新变更使用新文件，不修改已发布 migration。`publish-version.sh` 将旧客户端版本标记为 `archived`，插入新的 `published` 版本；这是 App 发布操作，不属于表结构 migration。

## 回滚边界

后端和管理端分别切换自己的 `current` 到兼容的旧版本；后端切换后通过对应宝塔项目重启并检查健康接口。程序回滚不会自动回滚数据库 migration 或客户端版本记录，数据库与接口变更必须考虑旧程序、旧 App 的兼容性。

复制文件与建立软链接不代表测试环境已上线。测试数据库复制和配置已完成。宝塔项目 `Leave_test` 的启动文件应为 `/www/wwwroot/Leave_test/backend/current/Leave`，宝塔可从 `/www/wwwroot/Leave_test/backend/current` 启动。修复版后端支持新布局下向上三级读取 YAML 和 `.env`，并已验证 `www` 用户可读取这两个文件；助手此前启动的测试进程已停止，后续启动和停止由用户在宝塔面板操作。此前本机健康接口曾返回 HTTP 200，但这不代表当前服务运行，也不代表公网访问或完整业务流程已验证。Nginx 站点、资源访问地址及三入口 CD 尚未完成。
