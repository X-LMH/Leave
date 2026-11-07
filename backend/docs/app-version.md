# 客户端版本与更新说明

版本查询沿用 `GET /api/v1/app-version?platform=android&version_code=16`。版本号、筛选历史版本和 APK 下载接口行为不变。

`updates` 按 `version_code` 降序返回，最新版本在上、旧版本在下。App 对历史缓存也按相同规则排序。

每个 `updates` 元素保留 `version_code`、`version_name`、`release_notes`，新增 `release_notes_grouped`：

```json
{
  "version_code": 17,
  "version_name": "1.2.0",
  "release_notes": ["新增功能", "体验优化", "问题修复", "更新后需重新登录"],
  "release_notes_grouped": {
    "features": ["新增功能"],
    "improvements": ["体验优化"],
    "bug_fixes": ["问题修复"],
    "notices": ["更新后需重新登录"]
  }
}
```

`release_notes` 始终为字符串数组，按新增功能、优化、问题修复、更新须知的顺序展开，供旧 App、首页摘要和表格预览使用。分类对象中的空分类返回空数组。历史数据库数组返回 `release_notes_grouped: null`，新版显示为“更新内容”。全部为空时显示“该版本暂无更新说明”，不推断修复或优化内容。

强制更新错误的 `data` 使用同一个 `AppVersionResponse`，其 `updates` 同样包含上述两个字段。App 本地缓存解析沿用相同兼容规则；缺少新字段的旧缓存仍可读。安装包下载不解析更新说明，不因说明格式错误而改变下载行为。

管理端版本列表、当前版本、详情及概览的版本对象同样返回分类字段；旧的数组字段继续保留。非法 JSON 在管理端沿用错误返回，App 查询记录日志并返回空说明；不静默把未知对象当成有效更新。

存储继续使用 `app_versions.release_notes JSON`，新版本保存分类对象，旧数组不批量重写。migration 000014 仅更新列注释，回滚不会转换对象数据；回退到不支持对象的旧后端前需要另行处理兼容性。
