# 许可证与第三方声明

## Leave 自有代码

Copyright (C) 2026 avery-my

Leave 自有代码采用 **GNU General Public License v3.0 only**，标识为 `GPL-3.0-only`。完整条款见根目录 [LICENSE](../LICENSE)。此声明适用于本仓库中由 Leave 作者拥有版权、且未单独注明其他许可证的原创内容；不改变第三方代码及素材原有的授权，也不追溯改变此前已授予的权利。

Leave is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, version 3 only.

Leave is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with Leave. If not, see <https://www.gnu.org/licenses/>.

## 第三方内容

| 内容 | 原始许可与处理方式 |
| --- | --- |
| Soybean Admin 管理端基础代码 | MIT，Copyright (c) 2021 Soybean；保留 [admin/LICENSE](../admin/LICENSE)，不得替换为 Leave 的版权声明 |
| Go 与管理端依赖 | 以各依赖随附的 LICENSE、NOTICE 和源码声明为准；`go.mod`、`go.sum`、`package.json`、`pnpm-lock.yaml` 记录依赖，不是许可证清单 |
| uni-app x 运行时、SDK、图片、字体等 | 保留各自授权与分发要求；不得仅因出现在仓库或构建产物中就视为 GPL 授权 |

本文件不是完整第三方依赖许可证审计。对来源不明确的复制代码和素材，应在发布前核实授权；不具备授权的内容应取得许可、替换或移除。新增第三方内容时，记录来源、许可证及必须保留的声明。

## 贡献与分发

- 提交原创代码时，确认有权按本项目许可证提供该贡献；不要求转让版权。第三方内容需注明来源与原许可证，不能冒充原创重新授权。
- 新增原创源码文件可使用相应语言的注释声明 `Copyright (C) 2026 avery-my`（其他贡献者填写自己的署名与实际年份），并附 `SPDX-License-Identifier: GPL-3.0-only`；不要覆盖第三方文件已有的头部声明。
- 向他人提供 APK 或其他二进制版本时，应依 GPL 提供相应版本的完整对应源码、构建及安装所需材料和许可证声明，不能只链接会持续变化的 `master`。
- 现有 Release 的源码压缩包可以作为源码入口，但应核实与安装包实际构建提交一致、材料完整；GitHub 自动生成压缩包本身不等于已完成分发合规检查。过去已发布的 APK 不因这次增加许可证而自动完成上述核对。

采用方式参考 [GNU 官方指南](https://www.gnu.org/licenses/gpl-howto.html)。本仓库使用 `LICENSE` 保存许可全文，与部分 GNU 项目使用 `COPYING` 的文件名区别不影响许可选择。
