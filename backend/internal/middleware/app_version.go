package middleware

import (
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	appPlatformHeader    = "X-App-Platform"
	appVersionCodeHeader = "X-App-Version-Code"
)

// AppVersionMiddleware only permits clients at or above the current published version.
func AppVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := strings.TrimSpace(c.GetHeader(appPlatformHeader))
		versionCode, err := strconv.Atoi(strings.TrimSpace(c.GetHeader(appVersionCodeHeader)))
		if platform == "" || err != nil || versionCode <= 0 {
			response.Error(c, response.CodeForceUpdate)
			c.Abort()
			return
		}

		current, isCurrent, err := appservice.IsCurrentAppVersion(platform, versionCode)
		if err != nil {
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
