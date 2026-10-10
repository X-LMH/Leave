# 后端开发规范

本文件适用于 `backend/`，与根目录 `AGENTS.md` 的共同约定一起遵循。未带模块前缀的路径相对于 `backend/`；带 `backend/`、`frontend/` 等前缀的路径相对于项目根目录。

## 目录职责与修改边界

* 保持现有简单分层：`Handler -> Service -> DAO`。
* `internal/controller/app`、`internal/service/app` 服务客户端，`internal/controller/admin`、`internal/service/admin` 服务管理端；修改前确认所属业务端，不为单端需求修改另一端行为。
* 跨端共享业务规则沿用 `internal/service/common` 等现有公共位置；数据库查询和锁放在 `internal/dao/mysql`，不要将业务决策下沉到 DAO。
* 请求解析、统一响应、路由和中间件分别沿用 `internal/request`、`internal/response`、`internal/router`、`internal/middleware`，保持现有返回码、`message` 和 `data` 约定。
* `cmd/` 保存命令入口，`config/` 保存配置，`database/` 保存 SQL；不要将业务实现堆入入口。

## Go 代码规范

* 使用符合 Go 习惯的写法，错误及时返回。
* 新增或重构结构体时，结构体定义和初始化尽量采用纵向排列，每个字段单独一行；尤其是多字段的 Model、DTO、请求和响应结构，不要为了减少行数挤在一行。
* GORM 查询语句较长、单行过于拥挤时，采用多行纵向排列，链式调用每个方法单独一行；某个方法的字段或参数较多时，再将其字段或参数逐项换行。此规则适用于整个查询，不限于 `Select`。简短、清晰的查询可保留单行写法。
* 不忽略错误，不做无意义的错误包装。
* 避免不必要的指针、接口、类型转换和第三方依赖。
* 编写或修改代码时，为必要的业务规则、边界条件和不明显的实现意图添加简洁注释；注释应解释原因或约束，不重复代码本身。
* 参数校验以实际业务需求为准，不进行脱离场景的过度防御性校验。
* 数据库操作放在 DAO 层，业务逻辑放在 Service 层，HTTP 逻辑放在 Handler 层。
* 关联对象的启用状态、是否允许保留旧选项、性别匹配等业务规则由 Service 校验；DAO 负责查询、加锁和持久化，不混入业务决策。
* 多步写入由 Service 编排事务内的查询、业务校验和保存，DAO 提供接收同一事务的操作方法及必要的事务入口。依赖数据库状态的校验必须与写入处于同一事务，并按并发需求加锁，不能为了分层将校验移到事务外。
* 按实际需求拆分职责，避免为简单操作新增 Repository、事务框架等额外抽象。

## DTO 规范

* 请求参数和响应结构统一放在 `backend/internal/dto`，按业务领域组织文件。
* `backend/internal/models` 用于定义数据模型，不再放置请求参数或响应 DTO。
* 修改接口时，同步更新相关 DTO 和引用，仅调整本次需求涉及的内容。

## 后端分页约定

* 分页请求统一在 controller 中调用 `request.ParsePagination` 解析和校验，非法参数返回参数错误；默认第 1 页、每页 20 条，最多 100 条。
* DAO 分页查询统一使用 `Scopes(Paginate(page, pageSize))`；`Paginate` 仅接收已校验的整数参数，不依赖 Gin 或 HTTP 请求，不重复校验或自动纠正参数。
* 需要返回总数时，先统计符合筛选条件的总数，再对列表查询应用分页 Scope。

## 数据库变更

* 每次修改数据库表结构或字段注释，都必须同步新增对应的 migration（`up.sql` 和 `down.sql`）；优先提供与升级对应的回滚语句。无法完整回滚时，在 migration 中明确记录不可逆原因、数据影响及备份或补偿方案，不将结构回滚等同于原始数据恢复。
* 更新 `backend/database/tables/` 中对应的完整表结构定义，并确保 migration 能完整反映本次变更；不要修改已应用的历史 migration。
* 编写 migration 与执行数据库变更是不同操作；按 `backend/database/README.md` 在独立测试数据库验证，不因修改 SQL 就直接执行部署或操作现有业务数据。
* 客户端版本发布遵循根目录的发布边界，版本数据更新不属于表结构 migration。

## 验证

* Go 命令在 `backend/` 模块根目录运行，工具链和依赖以 `go.mod` 为准。
* 对修改的 Go 文件运行 `gofmt`，运行受影响包的 `go test`；涉及跨包调用或公共结构时扩大测试范围，按需执行 `go test ./...` 和 `go build ./...`。仅编译成功不代表业务验证通过。
* MySQL 集成测试按测试代码配置独立本地环境；需要 `ADMIN_TEST_MYSQL_DSN` 的测试仅允许 loopback TCP 连接，不使用开发业务库或生产库。未配置导致的跳过不算集成测试通过。
* 测试受 Windows 临时目录权限、文件占用或数据库环境影响时，说明失败原因及已通过的检查，不将局部通过表述为全量通过。
