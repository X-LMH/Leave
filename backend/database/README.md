# 数据库变更规范

本目录用于维护项目数据库的表结构、初始化数据和结构变更记录。

```text
database/
├── README.md
├── tables/      # 当前完整表结构
├── seed/        # 初始化数据和测试数据
└── migrations/ # 数据库结构增量变更
```

## 目录职责

### `tables/`

保存当前版本的完整表结构。每个主要业务表对应一个 SQL 文件，例如：

```text
tables/users.sql
tables/app_versions.sql
```

当表结构发生变化时，需要同步更新对应的 `tables/*.sql` 文件，使其始终反映当前最终结构。

`tables/` 用于：

- 初始化全新的开发数据库；
- 查看当前表的完整结构；
- 作为数据库结构的基线参考。

### `seed/`

保存初始化数据、基础数据或测试数据，例如：

```text
seed/leave_types.sql
seed/app_versions.sql
```

数据初始化脚本和表结构脚本分开维护。除非数据本身就是系统必要的基础配置，否则不要把业务数据写入 `tables/`。

### `migrations/`

保存数据库从旧结构升级到新结构的增量变更。

每次结构变更使用一组 `up` 和 `down` 文件：

```text
migrations/
├── 000001_add_app_version_channel.up.sql
└── 000001_add_app_version_channel.down.sql
```

- `up.sql`：执行升级。
- `down.sql`：回滚本次升级。

当前项目的迁移命令入口位于 `cmd/migrate`，它只会按文件名顺序执行 `.up.sql` 文件。

## 执行迁移与初始化

已有数据库升级时，在 `backend/` 目录运行：

```text
go run ./cmd/migrate -config ./config/config.yaml
```

也可以在项目根目录直接运行：

```text
make migrate
```

如果迁移因执行失败而处于 dirty 状态，再次运行 `make migrate` 会自动将状态回退到上一个版本并重试该迁移。自动重试前应确认上一次失败没有留下无法重复执行的部分变更。

迁移程序支持同一文件内的多条 SQL 语句。`.down.sql` 文件不由该命令自动执行；需要回滚时，应先确认影响范围和备份，再通过数据库客户端执行对应的回滚脚本，并将 `schema_migrations` 状态恢复到上一个版本。

新建开发数据库时，先按 `tables/` 中的完整表结构初始化，再执行 `seed/` 中的基础数据脚本；不要对已经按完整表结构初始化的数据库重复执行历史 migration。

## Migration 命名规范

文件名格式：

```text
编号_简短描述.up.sql
编号_简短描述.down.sql
```

示例：

```text
000001_create_audit_logs.up.sql
000001_create_audit_logs.down.sql
000002_add_user_nickname.up.sql
000002_add_user_nickname.down.sql
```

命名要求：

1. 编号必须唯一并递增。
2. `up` 和 `down` 必须使用相同编号和描述。
3. 描述使用小写英文和下划线，准确表达变更内容。
4. 一个 migration 尽量只完成一个相关的数据库变更。
5. 已经执行过的 migration 不允许修改；发现问题时新增修复 migration。

## 新增或修改表结构的流程

以给 `app_versions` 增加 `channel` 字段为例：

### 1. 修改完整表结构

先更新：

```text
tables/app_versions.sql
```

使其包含新的 `channel` 字段。

### 2. 新增 migration 文件

```text
migrations/000001_add_app_version_channel.up.sql
migrations/000001_add_app_version_channel.down.sql
```

`up.sql` 示例：

```sql
ALTER TABLE `app_versions`
    ADD COLUMN `channel` VARCHAR(32) NOT NULL DEFAULT 'default'
    COMMENT '安装包渠道'
    AFTER `platform`;
```

`down.sql` 示例：

```sql
ALTER TABLE `app_versions`
    DROP COLUMN `channel`;
```

### 3. 本地验证

至少验证以下内容：

```text
1. 在备份或测试数据库上执行 up。
2. 检查表结构和相关数据。
3. 执行 down，确认能够回滚。
4. 再次执行 up，确认可以重新升级。
5. 运行后端测试。
```

### 4. 一起提交代码

表结构、migration 和依赖它们的业务代码应尽量在同一个提交或同一个功能分支中维护：

```bash
git add backend/database backend/internal
git commit -m "db: add app version channel"
```

## Up 和 Down 的编写要求

### Up 要求

- 新增字段应考虑默认值或允许为空，避免影响已有数据；
- 添加唯一索引、非空约束前，先确认现有数据满足约束；
- 大表结构变更需要评估锁表时间和执行时长；
- 破坏性修改应拆成多个阶段完成。

### Down 要求

- 尽量提供与 `up` 对应的回滚语句；
- 删除字段或表会导致数据丢失，生产环境执行前必须确认；
- 数据转换类 migration 可能无法完全恢复原始数据，需要备份或补偿脚本；
- 不要把 `down` 当作生产数据恢复方案。

## 兼容性和发布顺序

涉及后端代码和数据库结构时，优先使用向后兼容的变更顺序：

```text
1. 先新增字段、索引或表。
2. 发布同时兼容旧结构和新结构的后端。
3. 迁移或补齐历史数据。
4. 确认旧版本代码不再使用后，再删除旧字段或旧表。
```

不要在旧版本后端仍可能运行时直接删除字段或修改字段含义。

## 基线数据库和历史数据库

如果数据库已经存在，不要把当前完整表结构当作新的 migration 重复执行。应将当前数据库状态视为基线，之后的第一次变更从下一个编号开始：

```text
现有数据库状态：基线
下一次变更：000001_xxx.up.sql / 000001_xxx.down.sql
```

如果未来需要为全新环境提供一键初始化脚本，可以使用 `tables/` 初始化表结构，再执行 `seed/` 初始化基础数据。

## 提交前检查清单

- [ ] 是否同步更新了对应的 `tables/*.sql`？
- [ ] 是否新增了对应的 `up.sql` 和 `down.sql`？
- [ ] migration 编号是否唯一且递增？
- [ ] 是否在测试数据库执行过升级？
- [ ] 是否验证过回滚或明确记录了不可逆原因？
- [ ] 是否考虑旧版本后端的兼容性？
- [ ] 是否避免把生产密码、连接串或敏感数据提交到仓库？
- [ ] 是否已经运行相关测试？

生产环境的 migration 应由发布流程统一执行，不建议通过手工修改线上数据库完成结构变更。
