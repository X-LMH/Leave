package logger

import (
	"backend/internal/request"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OperationFields 返回请求和操作字段，供原生 Zap Logger.With 使用。
// 操作者和目标等业务字段由调用方按需添加。
func OperationFields(c *gin.Context, operation string) []zap.Field {
	return append(RequestFields(c), zap.String("operation", operation))
}

// RequestFields 只读取请求元数据，不生成 ID，也不自动附加身份信息。
func RequestFields(c *gin.Context) []zap.Field {
	route := c.FullPath()
	if route == "" {
		route = "unmatched"
	}
	return []zap.Field{
		zap.String("request_id", request.GetRequestID(c)),
		zap.String("method", c.Request.Method),
		zap.String("route", route),
	}
}
