package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是所有 JSON 响应的统一结构。
type Response struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success 写出固定为 200 的成功响应。
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: CodeSuccess.Message(),
		Data:    data,
	})
}

// Error 根据业务码写出语义正确的 HTTP 错误响应。未知业务码不会暴露给客户端。
func Error(c *gin.Context, code Code) {
	meta, ok := metaByCode[code]
	if !ok {
		code = CodeServerBusy
		meta = metaByCode[CodeServerBusy]
	}
	c.JSON(meta.httpStatus, Response{
		Code:    code,
		Message: meta.message,
	})
}
