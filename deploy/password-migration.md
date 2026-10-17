# 历史密码迁移

`backend/cmd/password-migrate` 是独立 Go 命令行工具，直接复用密码 utils 将 `users.password` 中的明文转换为 Argon2id，不依赖运行中的后端或 HTTP 接口，也不属于表结构 migration。

## 构建与上传

在 `backend/` 下构建服务器使用的 Linux amd64 可执行文件（与现有部署平台一致）：

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o password-migrate ./cmd/password-migrate
```

将生成的 `password-migrate` 上传到目标环境根目录的 `scripts/` 下。只上传可执行文件；YAML 和 `.env` 沿用服务器文件，不随工具覆盖。服务器不需要安装 Go。此工具不自动纳入 CI/CD，不自动运行。

## 预览与执行

以下是生产环境示例；测试环境使用 `/www/wwwroot/Leave_test` 和 `config/config.test.yaml`。

```bash
cd /www/wwwroot/Leave
chmod +x scripts/password-migrate

# 默认只读取和统计；确认输出的主机、库名与待迁移数量。
./scripts/password-migrate -config config/config.server.yaml -env .env

# 手动暂停旧后端及其他账号写入，完成数据库备份后执行。
./scripts/password-migrate -config config/config.server.yaml -env .env -apply

# 再次预览，应显示待迁移为 0，然后切换到支持哈希的后端并启动。
./scripts/password-migrate -config config/config.server.yaml -env .env
```

只要求 MySQL 配置，不需要 JWT 或上传目录设置。`MYSQL_PASSWORD` 环境变量优先于 YAML；`.env` 不覆盖进程已有环境变量。默认读取当前目录 `.env`，不存在时使用已有环境变量或 YAML，也可以用 `-env ""` 跳过文件加载。必须显式指定 `-config`，避免自动选择错误环境。

## 行为与恢复

- 同时处理学生和管理员账号，保留原密码内容及首尾空格，不修改其他字段。
- 已有合法 Argon2id 哈希自动跳过，可重复执行。空密码、损坏的 Argon2id 或以 `$` 开头的未知格式会使整个迁移终止，需人工检查；本工具针对历史明文数据，不负责其他算法迁移。
- 执行模式在一个事务中锁定账号并逐条写入、回读验证，提交前失败全部回滚；输出只有目标地址、库名、数量和必要的账号 ID，不输出密码或哈希。适合当前规模的数据，账号量较大时应先在备份库评估耗时和锁持有时间。
- 迁移前备份并暂停账号写入，迁移后使用支持哈希验证的后端。旧版明文比较后端不能登录迁移后的账号，也不要在迁移完成后继续用旧版注册或改密。
- 哈希不可还原为明文；若要回滚应用到旧版，需要停服并从迁移前备份恢复密码数据。提交时连接异常会报告结果未知，应先预览检查再处理，不假定已经回滚。

本地 `leave_list_local` 的 26 个账号此前已完成迁移；本说明不表示测试或生产服务器已经执行。
