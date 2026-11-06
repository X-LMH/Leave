# 认证与会话

App、Web 管理端使用 32 字节随机数编码的 Base64URL Token，通过 `Authorization: Bearer <token>` 发送。服务仅使用 Redis 会话认证；旧 JWT 实现、配置和签名依赖已移除。

## 存储与鉴权

Redis Key 为 `leave:<app.namespace>:auth:session:<SHA256(token)>`；Value 保存 `user_id`、`student_id`、`role`、`auth_version`、`client`、`created_at`、`expires_at`。原始 Token、密码和个人资料不写入会话或日志。会话存储操作有 3 秒超时，不做内存/MySQL/JWT 回退。

`app.namespace`、`app.port`、`session.student_ttl` 和 `session.admin_ttl` 必须在配置文件中明确填写，不由配置加载器补默认值。不同环境需配置不同 namespace；模板中学生会话为 `720h`（30 天），管理员会话为 `8h`。均为固定有效期，不自动续期，不使用 Refresh Token。多次登录创建独立会话。

鉴权先查询 Redis，再按主键读取 MySQL 中账号的必要字段，校验存在、启用状态、认证版本、账号标识及角色。Redis 内的账号和角色是登录时快照，权限由数据库当前状态决定。管理员 App 会话可以访问管理接口。

## 接口

- 原有 App 注册/登录返回结构不变，包含 `token`、`user`、`profile_completed`。
- `POST /api/v1/admin/auth/login` 返回 `{token}`。
- `GET /api/v1/admin/auth/me` 返回 `{student_id, role}`，需要有效管理员会话。Web 登录与刷新通过该接口恢复身份。
- `POST /api/v1/logout`、`POST /api/v1/admin/auth/logout` 删除当前会话；合法 Token 已过期/已删除时也成功，非法 Bearer 格式或非法 Token 返回 `1005`。退出不经过普通会话有效性校验，但保留所属路由的限流和 App 版本规则。
- 会话无效返回 `1004/1005`、HTTP 401；依赖故障返回 `9003`、HTTP 503，客户端保留登录状态。
- 注册账号创建成功但会话写入失败返回 `1008`、HTTP 503，提示“注册成功，请重新登录”；App 切回登录模式并保留账号，清空密码，不自动再次注册。

客户端只允许对应当前 Token 的认证失效响应清理状态，旧请求不能注销新会话。Web 身份查询与退出请求固定携带调用时捕获的 Token，请求拦截器不覆盖显式凭据；旧会话清理先完成路由和标签重置，跳转前再次确认没有新登录。主动退出仅在服务端撤销成功后清理当前端状态，失败时提示重试；本地状态清理本身不调用退出接口。

## 账号级失效

`users.auth_version` 初始为 1。修改或重置密码时与密码更新原子递增；禁用时与状态更新在同一事务递增。重新启用不恢复旧会话。任何新增角色变更流程必须在同一事务递增 `auth_version`，不能只更新角色字段。当前没有账号角色编辑入口。

会话使用验证密码时读取的认证版本，禁止旧密码验证后重新读取新版本再签发会话。改密成功后全部旧会话（含当前会话）失效，两端返回登录页。TTL 到期会自动清理 Redis 记录，不增加撤销字段、Token 索引或定时会话清理任务。

已完成鉴权的在途请求不由此机制强行取消，之后的请求会被拒绝。

## 升级与回滚

先应用 `000013_add_user_auth_version.up.sql`，再启动新服务；旧 JWT 需要重新登录。Redis 数据丢失后需要重新登录。回滚字段会丢失认证版本；重新升级前应更换 namespace 或清除对应旧会话，避免旧版本会话恢复有效。执行 migration、访问服务器及部署须另行授权。

## 验证

`go test ./...`、`go build ./...` 在 backend 目录执行。会话及鉴权单元测试使用隔离命令模拟器；真实会话集成测试需 `SESSION_TEST_REDIS_ADDR`，既有路由限流集成需 `RATE_LIMIT_TEST_REDIS_ADDR`，均仅允许专用的数字 loopback 地址。数据库集成沿用 `ADMIN_TEST_MYSQL_DSN` 等已有隔离测试配置，不使用业务库。未配置时跳过，不表示真实基础设施验证通过。

Web 运行 `pnpm typecheck`、`pnpm lint`、`pnpm build` 及相关 `node --test`；App 请求逻辑的 Node 模拟测试放在 admin/tests/app-auth.test.mjs，不能代替 HBuilderX Android 编译与真机验证。

## 代码职责

`internal/dao/redis/redis.go` 提供客户端初始化及连接检查；`cmd/server` 负责调用、依赖注入和退出时关闭。`internal/dao/redis/session.go` 负责会话 Key、保存、查询、删除及存储操作超时；`rate_limit.go` 负责限流 Lua 数据操作。`internal/session` 保留随机 Token 生成、哈希、Bearer 解析、会话序列化与校验、角色对应有效期和认证错误映射。限流规则、身份规则及 HTTP 响应继续由原有业务层和中间件处理。
