package middleware

import (
	"backend/internal/request"
	"crypto/rand"

	"github.com/gin-gonic/gin"
)

// RequestID 为请求生成独立 ID，统一响应和日志只读取该值。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := rand.Text()
		c.Set(request.CtxRequestID, id)
		c.Header(request.HeaderRequestID, id)
		c.Next()
	}
}
