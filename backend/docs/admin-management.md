# 管理端请假原因与班级接口

接口前缀为 `/api/v1/admin`。所有接口要求 `Authorization: Bearer <管理员 Token>`，不受 App 版本校验限制。登录接口沿用 `/api/v1/admin/auth/login`。

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| GET | `/leave-types` | 请假原因分页列表 |
| POST | `/leave-types` | 新增请假原因 |
| PUT | `/leave-types/:id` | 编辑请假原因 |
| DELETE | `/leave-types/:id` | 软删除请假原因 |
| GET | `/classes` | 班级分页列表 |
| POST | `/classes` | 新增班级 |
| PUT | `/classes/:id` | 编辑班级 |
| DELETE | `/classes/:id` | 删除未被学生资料引用的班级 |

## 列表

两类列表接受 `page`（默认 1）、`page_size`（默认 20，最大 100）和可选 `is_enabled=true/false`。

请假原因支持 `name` 模糊查询，按 `sort_order ASC, id ASC` 排序。班级支持 `college`、`major`、`class_name` 模糊查询，按学院、专业、班级名称、ID 升序排列。

统一响应为 `{ "code": 0, "message": "success", "data": { "items": [], "total": 0, "page": 1, "page_size": 20 } }`。空结果的 `items` 为数组。

## 新增与编辑

请假原因请求体：

```json
{ "name": "病假-本科生", "sort_order": 0, "is_enabled": false }
```

班级请求体：

```json
{ "college": "计算机学院", "major": "软件工程", "class_name": "软工24-1", "is_enabled": true }
```

所有文本保存前去除首尾空白。原因名称为 1–32 个字符；学院、专业为 1–100 个字符；班级名称为 1–64 个字符。`sort_order` 必须提供，范围为 0–4294967295；`is_enabled` 必须提供，允许 false。

新增和编辑成功均返回最新记录。原因记录包含 `id`、`name`、`sort_order`、`is_enabled`、`created_at`、`updated_at`；班级记录包含 `id`、`college`、`major`、`class_name`、`is_enabled`、`created_at`、`updated_at`。

删除成功返回 `{ "code": 0, "message": "success", "data": null }`。原因软删除后不在列表和客户端选项中出现，其名称仍受唯一约束。班级被任何学生资料（包括软删除资料）引用时禁止删除，可通过编辑停用。已有请假记录中的快照不随字典编辑变化。

## 错误

| HTTP 状态 | 业务码 | 含义 |
| --- | --- | --- |
| 400 | 2001 | 参数不合法、缺少状态或排序、无效 ID |
| 401 | 1004 / 1005 | 未登录、Token 无效或无管理员权限 |
| 404 | 2004 | 管理记录不存在或原因已经删除 |
| 409 | 2005 | 名称或学院＋专业＋班级名称重复 |
| 409 | 2006 | 班级被学生资料引用 |
| 500 | 9001 | 服务异常 |

管理端遇到受保护接口的 HTTP 401 会清理登录并跳转登录页，不调用 Token 刷新接口。其他失败保留编辑内容或删除确认框，通过统一请求模块提示。

## 迁移与验证

增量迁移 `000009_add_profiles_class_index` 为 `profiles.class_id` 添加索引；完整表结构已同步。按 `database/README.md` 的迁移流程发布，勿向已有数据库重复执行完整建表 SQL。

在 `backend/` 执行 `go test ./...`。MySQL 集成测试默认跳过；使用隔离的本地 MySQL 实例时，设置 `ADMIN_TEST_MYSQL_DSN` 后执行 `go test ./internal/dao/mysql -run TestAdminManagementIntegration -v`。DSN 必须使用 TCP loopback（localhost、127.0.0.1 或 ::1）。测试账号需要创建和删除数据库的权限；测试创建独立的 `leave_admin_test_<时间戳>` 数据库并在结束时删除，不读取应用配置中的数据库。

在 `admin/` 执行 `node --test tests/management.test.mjs`，验证真实页面逻辑和 API 适配器处理远程分页、筛选、乱序响应、失败保存、重复提交及删除末页的行为；另执行 `pnpm typecheck`、相关文件 lint 和 `pnpm build`。

页面已移除本地模拟数据来源。旧浏览器模拟数据不会被上传或导入数据库，初次对接展示数据库中已有的数据。
