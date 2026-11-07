# 功能实现状态

核查日期：2026-11-07。下表来自页面、请求封装和后端路由的静态核对，表示源码存在调用链，不代表完成部署、Android 编译或真机验收。后续功能变更应同步更新本表。

## Android 客户端

| 功能 | 源码状态 | 核对入口与边界 |
| --- | --- | --- |
| 注册、登录、退出、修改密码 | 已接入请求 | `pages/auth/login.uvue`、`utils/api.uts`；使用服务端会话 |
| 首页与记录列表 | 已接入请求 | `components/HomeTabContent.uvue`、`components/RecordsTabContent.uvue`；支持列表分页、删除 |
| 假条生成与详情 | 已接入请求 | `pages/leave/apply.uvue`、`pages/leave/detail.uvue`；创建记录并按 ID 获取详情 |
| 个人资料、头像 | 已接入请求 | `pages/profile/edit.uvue`、`pages/profile/info.uvue`；上传文件还依赖静态服务 |
| 反馈、版本检查与更新 | 已接入请求 | `pages/profile/feedback.uvue`、`pages/update/index.uvue`；安装行为需 Android 验收 |
| 管理员入口 | 已接入请求 | `utils/admin-api.uts`、`pages/admin/`；概览、列表、基础资料维护、反馈处理及学生状态等按页面提供能力 |
| 外观设置 | 本地功能 | `utils/app-theme.uts`；不是后端联调功能 |
| 旧提交页面 | 本地旧入口 | `pages/submit/index.uvue` 仍使用查询参数跳转；详情页当前按记录 ID 加载，不应作为已联调提交入口 |

## Web 管理端与后端

| 功能 | 源码状态 | 核对入口与边界 |
| --- | --- | --- |
| 登录、会话身份、修改密码 | 已接入请求 | `admin/src/service/api/auth.ts` 与 `/admin/auth/*` |
| 看板、学生查询与维护 | 已接入请求 | `admin/src/views/home/`、`student/`；后端提供统计、资料、状态与重置密码接口 |
| 请假记录查询、详情与编辑 | 已接入请求 | `admin/src/views/leave-record/` 与 `/admin/records`；未提供管理端删除、作废或导出路由，不能按功能范围文档推定已实现 |
| 班级、公寓、假期类型 | 已接入请求 | `admin/src/service/api/management.ts`、`apartment.ts` 与相应管理路由 |
| 反馈查看、处理与删除 | 已接入请求 | `admin/src/views/feedback/` 与 `/admin/feedbacks` |
| 版本列表、当前版本与详情 | 已接入查询 | `admin/src/views/app-version/`；管理路由只提供 GET，版本发布沿用 `deploy/` 脚本 |
| 真实学校审批流 | 未实现，不属于当前定位 | 记录中的通过时间是模拟数据，不代表审批人执行过审批 |

后端路由以 `backend/internal/router/route.go` 为准；App 请求分别核对 `frontend/utils/api.uts` 与 `frontend/utils/admin-api.uts`。此表只列已核查的主要功能，不据此宣称所有页面均已接入；发现 Mock 或占位功能时应逐项记录，不能仅凭接口存在标为完成。

`docs/admin-feature-scope.md` 是业务范围参考，不是当前完成清单。运行准备见根 README，部署和版本发布见 [部署说明](../deploy/README.md)。
