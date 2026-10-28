package middleware

import (
	requestlog "backend/internal/logger"
	"backend/internal/response"
	"errors"
	"net/http"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccessLogger 统一通过 Zap 记录请求，使终端与按日文件使用同一套字段。
func AccessLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		logger.With(requestlog.RequestFields(c)...).Info("HTTP 请求完成",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(started)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// Recovery 避免将 panic 值或请求内容写入日志，其中可能含有密码等私有数据。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if err, ok := recovered.(error); ok && (errors.Is(err, syscall.EPIPE) ||
					errors.Is(err, syscall.ECONNRESET) ||
					errors.Is(err, http.ErrAbortHandler)) {
					// 与 Gin Recovery 一致，断开的连接不能继续写响应。
					logger.With(requestlog.RequestFields(c)...).Warn("请求连接已断开")
					c.Abort()
					return
				}
				logger.With(requestlog.RequestFields(c)...).Error("请求发生 panic", zap.Stack("stack"))
				if !c.Writer.Written() {
					response.Error(c, response.CodeServerBusy)
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
