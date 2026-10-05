---
name: bugfix-uniapp-usb-localhost
description: 排查 Leave 项目 uni-app x Android 真机 USB 调试时无法访问电脑本地后端、提示网络失败的问题。检查 8080 后端与 ADB 反向转发，并从手机侧验证连接；适用于本项目本地调试，默认保留现有配置。
---

# Bugfix：uni-app x USB 本地后端访问

## 项目定位

本 skill 属于 Leave 项目。排查前读取项目 AGENTS.md，并核对实际配置与进程：

- `frontend/utils/api.uts`：开发 API 地址，当前为 `http://127.0.0.1:8080/api/v1`。
- `backend/config/config.local.yaml`：本地后端端口和图片访问前缀；当前 API 使用 8080，图片服务使用 8081。
- `backend/cmd/server/main.go`：后端启动入口；从 `backend` 目录启动本地服务。
- `/api/v1/health`：后端健康检查；`deploy/nginx.local.conf`：本地图片服务配置。

以上是本次排障时的项目状态，不替代后续代码和运行状态。

## 核心经验

Android 的 localhost 通常指向手机，但存在 ADB 反向转发时，手机访问对应本地端口可以经调试连接到达电脑。不要仅凭客户端使用 127.0.0.1 就认定请求地址错误。

USB 调试连接和目标端口的反向转发有效时，本地后端访问不需要热点或同一 Wi-Fi；这不代表其他互联网请求也无需网络。HBuilderX 的调试通道正常，不代表业务后端端口也已转发。

## 排查与修复

1. 只读检查当前开发请求地址及端口、后端实际监听、手机连接方式。以代码和运行状态为准，不把 8080 固定为所有项目的端口。
2. 从电脑访问后端健康接口。失败时先排查后端进程、端口或启动错误；健康接口成功仍不能证明手机可访问。
3. 优先使用 PATH 中的 adb；找不到时，在实际 HBuilderX 安装目录的 `plugins/launcher-tools/tools/adbs/` 查找。此电脑曾使用 `D:\Code\IDE\HBuilderX\plugins\launcher-tools\tools\adbs\adb.exe`，路径只是线索。
4. 执行 `adb devices`，确认目标设备状态为 `device`。`unauthorized` 需用户在手机授权；多设备时确认目标并始终使用 `-s`，不要自动选择第一台或记录固定设备序列号。
5. 执行 `adb -s <serial> reverse --list`，检查业务端口的准确映射。只有 8000/8001 等调试端口的映射不足以访问 8080 后端。
6. 用户已要求修复连接且允许操作当前调试设备时，为所需端口添加反向转发。若用户只要求解释或查看，先提供诊断，不执行设备状态变更。若该手机端口已映射到其他目标，先查明用途，避免直接覆盖。

PowerShell 示例：使用实际 adb 路径、设备序列号和业务端口替换变量。

```powershell
$adbPath = 'D:\Code\IDE\HBuilderX\plugins\launcher-tools\tools\adbs\adb.exe'
$deviceSerial = '<当前目标设备序列号>'
& $adbPath -s $deviceSerial reverse tcp:8080 tcp:8080
& $adbPath -s $deviceSerial reverse --list
```

左侧端口在手机，右侧端口在电脑。修复 localhost 访问时通常将两侧业务端口设为相同值。

## 验证与失败处理

- 确认新映射存在，再从手机侧请求同一健康接口，检查 HTTP 状态和响应内容。手机浏览器或设备中的 curl 可用于检查；未验证 App 本身时，不宣称 App 功能已全部恢复。
- 若设备只有 `toybox nc`，可通过 `adb shell -T toybox nc` 发送 HTTP 请求。发送后保持 stdin 打开至收到响应，设置连接和读取超时；部分环境中直接管道发送后 stdin EOF 会导致提前退出。空输出或退出码 0 不算成功证据。
- 手机侧健康检查成功后，请用户重新打开或重试 App。仍失败时查看该次请求的真实 URL、状态码和日志，再区分版本校验、鉴权、响应解析或运行包问题，不循环重复改地址。
- 没有可用设备或 USB 授权时，明确报告缺少的条件；不要伪造真机验证结果。Wi-Fi 直连场景需另行检查电脑 LAN 地址与入站规则，不能套用 USB 已连通的结论。

## 变更边界与报告

- 默认保留项目现有请求地址和配置。用户说“不改配置”时，不修改前端地址、后端存储 URL、manifest、热点设置或防火墙。
- 只操作本次需要的设备和端口，不执行 `reverse --remove-all`，也不重启整个 ADB 服务来处理单个端口问题。
- 区分业务 API 与图片等独立服务。转发 8080 不会启动 8081 上的图片服务器；只有相应服务已运行且需求涉及该端口时，才处理它。
- 报告最终保留的操作和验证证据。例如“新增目标手机 8080 的 ADB 反向转发，手机侧健康接口返回 200；未改配置或防火墙”。
- 提醒手机重启、ADB 服务重启或调试连接重建后，映射可能需要重新建立。需要撤销时仅删除本次新增映射：`adb -s <serial> reverse --remove tcp:8080`。

## 本次案例

2026-10-05，电脑后端监听 `0.0.0.0:8080` 且本机健康检查正常；USB 手机已连接，但仅有 8000、8001 的反向转发。添加 `tcp:8080 → tcp:8080` 后，手机侧访问 `127.0.0.1:8080/api/v1/health` 返回 HTTP 200 与 `status: ok`。先前尝试的热点 IP 配置已撤销。已验证的是手机到后端健康接口的连接，不是所有 App 页面和头像服务。

参考：[Android 官方本地开发服务器访问说明](https://developer.android.com/develop/ui/views/layout/webapps/access-local-server?hl=en)。
