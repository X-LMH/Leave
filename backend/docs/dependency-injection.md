# 后端依赖注入

后端使用手动构造函数注入，保持 `Handler → Service → DAO` 分层，不使用 DI 容器。

## 启动与资源归属

`cmd/server/main.go` 负责启动与资源管理，通过 `cmd/server/dependencies.go` 中的 `buildDependencies` 函数组装依赖：先调用 `config.Load()` 获取配置，再调用 `mysql.Open(cfg.Mysql)` 建立连接，创建 `mysql.NewDAO(db)`、JWT Manager、各业务 Service 和 Handler，并构建鉴权及 App 版本校验中间件，最后将 Handler 和已构建的中间件传给 `router.SetupRouter(router.Dependencies{...})`。

入口拥有数据库连接并负责关闭；DAO、Service 和 Handler 不关闭共享连接。配置在创建对象时按值传入，JWT Manager 同时用于登录签发和中间件验签。迁移命令保留自己的数据库配置读取流程，无需初始化 HTTP 应用或 JWT。

## 新增业务

- DAO 方法放在现有 mysql DAO 上，通过实例持有的连接访问数据库；分页和错误判断等无状态函数保持包函数。
- 按业务领域创建 Service，构造函数只接收实际需要的 DAO、其他 Service、JWT Manager 或配置片段；默认直接使用具体类型。
- Handler 接收相应 Service；中间件的数据库、JWT 和版本查询依赖也显式传入。
- 鉴权和 App 版本校验中间件在启动入口构建，路由通过 `gin.HandlerFunc` 接收，不直接依赖 DAO、JWT Manager 或 Service；无外部依赖的 CORS 和管理员角色校验中间件仍在路由中构建。
- 在启动入口组装新增组件，并更新路由依赖；不要新增全局默认实例或通过 Gin Context 查找服务。

## 事务与测试

Service 通过 DAO 的事务方法编排业务，事务内的 DAO 操作显式接收同一个 `tx *gorm.DB`。状态校验、行锁和写入保持在原事务内，不在回调中改用普通连接。

测试使用独立配置、DAO 和业务对象，不替换全局变量。纯参数校验测试可以注入没有数据库连接的 DAO，但不得进入实际数据库操作。路由测试的 `testDependencies` 显式组装测试依赖，不读取开发配置。

在后端目录运行 `go test ./...`、`go vet ./...` 和 `go build ./...`。MySQL 集成测试沿用 `ADMIN_TEST_MYSQL_DSN` 或 `MYSQL_TEST_DSN`，仅允许 loopback TCP，并创建独立临时数据库；未配置时跳过。DAO 的双数据库隔离与回滚测试属于已有管理端数据库集成套件。
