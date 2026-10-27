package middleware

import (
	requestlog "backend/internal/logger"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	appPlatformHeader    = "X-App-Platform"
	appVersionCodeHeader = "X-App-Version-Code"
)

// AppVersionMiddleware only permits clients at or above the current published version.
func AppVersionMiddleware(versions *appservice.AppVersionService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := strings.TrimSpace(c.GetHeader(appPlatformHeader))
		versionCode, err := strconv.Atoi(strings.TrimSpace(c.GetHeader(appVersionCodeHeader)))
		if platform == "" || err != nil || versionCode <= 0 {
			response.Error(c, response.CodeForceUpdate)
			c.Abort()
			return
		}

		current, isCurrent, err := versions.IsCurrentAppVersion(platform, versionCode)
		if err != nil {
			logger.With(requestlog.RequestFields(c)...).Error("校验客户端版本失败",
				zap.String("operation", "app.check_version"),
				requestlog.Error(err),
			)
			response.Error(c, response.CodeServerBusy)
			c.Abort()
			return
		}
		if !isCurrent {
			response.ErrorWithData(c, response.CodeForceUpdate, current)
			c.Abort()
			return
		}
		c.Next()
	}
}
