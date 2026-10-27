package middleware

import (
	requestlog "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const requestRouteKey = "log_request_route"

// AccessLogger 保留 Gin 的终端显示风格，附加请求 ID，不再写入 Zap。
func AccessLogger() gin.HandlerFunc {
	accessLogger := gin.LoggerWithConfig(gin.LoggerConfig{
		SkipQueryString: true,
		Formatter: func(p gin.LogFormatterParams) string {
			var statusColor, methodColor, latencyColor, reset string
			if p.IsOutputColor() {
				statusColor = p.StatusCodeColor()
				methodColor = p.MethodColor()
				latencyColor = p.LatencyColor()
				reset = p.ResetColor()
			}
			switch {
			case p.Latency > time.Minute:
				p.Latency = p.Latency.Truncate(10 * time.Second)
			case p.Latency > time.Second:
				p.Latency = p.Latency.Truncate(10 * time.Millisecond)
			case p.Latency > time.Millisecond:
				p.Latency = p.Latency.Truncate(10 * time.Microsecond)
			}
			// 使用路由模板，避免将路径参数或未匹配地址中的个人信息打印出来。
			return fmt.Sprintf("[GIN] %s |%s %3d %s|%s %8v %s| %15s |%s %-7s %s %q | request_id=%v\n",
				p.TimeStamp.Format("2006/01/02 - 15:04:05"),
				statusColor, p.StatusCode, reset,
				latencyColor, p.Latency, reset, p.ClientIP,
				methodColor, p.Method, reset, p.Keys[requestRouteKey], p.Keys[request.CtxRequestID],
			)
		},
	})
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		c.Set(requestRouteKey, route)
		accessLogger(c)
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
