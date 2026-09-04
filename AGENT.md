尽量不要`pages/compare/compare`页面呈现出来的样式

## uni-app x / uvue CSS 约束

- uvue 仅支持 class 选择器，不要使用 `page`、`body`、`html` 等标签选择器。
- 不要使用 `100vh`、`100vw` 等视口单位；`min-height` 等属性使用数字或 `px`/`rpx` 值。

添加一个新样式之前，务必要把之前的相同样样式删掉
