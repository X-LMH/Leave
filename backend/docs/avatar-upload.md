# 上传头像

`POST /api/v1/profile/avatar`，沿用现有版本校验和 JWT 登录校验。

请求使用 `multipart/form-data`，文件字段为 `file`，一次上传一个文件。
支持 JPEG、PNG，最大 5 MiB，宽高分别不超过 4096 像素。
服务端根据文件内容校验图片并生成随机文件名，不使用客户端文件名。
用户需先完善个人资料，否则返回业务码 `3002`。

成功响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "avatar_url": "http://localhost:8081/api/v1/uploads/avatars/abc123.jpg"
  }
}
```

数据库 `profiles.avatar_url` 保存 `avatars/abc123.jpg` 形式的相对路径。
创建个人资料时头像字段为空字符串，仅在用户上传头像后保存相对路径。
查询个人资料的 `GET /api/v1/profile` 返回拼接后的 `avatar_url`，历史空头像路径也使用默认头像。
更新个人资料不会覆盖头像；上传头像更新成功后删除旧头像，保留共享默认头像；数据库更新失败则清理新文件。

## 配置与部署

配置 `storage.root_dir` 为持久化文件根目录，`storage.base_url` 为文件访问前缀。
默认头像图片需放在 `storage.root_dir` 下的 `avatars/default.png`。
本地应从 `backend` 目录启动（`go run ./cmd/server`），配置 `root_dir: "../uploads"` 指向项目根目录的 `uploads`，Nginx 应挂载此目录。
本地配置已提供；服务器需参照 `config.example.yaml` 补充这两个配置项。
存储根目录或访问前缀为空时，配置初始化失败，程序终止启动。
文件访问由 Nginx 提供，上传目录需要与 Nginx 的 alias 或 Docker 挂载目录一致。
移动设备访问时，本地配置的 `localhost` 应替换为设备能访问的开发机地址。
Nginx 上传请求的 `client_max_body_size` 应大于 5 MiB（例如 `6m`），以容纳 multipart 头部。

已有数据库需执行 `000007`（新增头像字段）和 `000008`（更新注释）migration；
已执行的 migration 不要重复执行。本接口不新增表结构变更。
当前不会自动执行 migration，也不会自动修改 Nginx 或服务器配置。
