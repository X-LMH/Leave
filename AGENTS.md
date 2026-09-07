## 注意

- 尽量不要改变 `frontend/pages/leave/detail.uvue` 页面呈现出来的样式。
- 目前还处于前后端分别开发的阶段，不要擅自对接前后端。

## uni-app x / uvue CSS 约束

- uvue 仅支持 class 选择器，不要使用 `page`、`body`、`html` 等标签选择器。
- 不要使用 `100vh`、`100vw` 等视口单位；`min-height` 等属性使用数字或 `px`/`rpx` 值。
- 添加新样式前，务必删除已有的相同样式。

## 代码格式

- 编写代码时保持规范的缩进与换行；不要将多段逻辑或大量属性持续堆叠在同一行。

## DTO 迁移

`backend/internal/models/params.go` 正在渐进迁移到`backend/internal/dto`。
修改或新增相关代码时，将涉及的参数定义一并迁移到`backend/internal/dto`，并按需更新导入和引用；不要仅为此进行一次性全量迁移。
