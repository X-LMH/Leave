package controller

import (
	"backend/internal/response"
	"backend/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCurrentAppVersionHandler returns the current published version for a platform.
func GetCurrentAppVersionHandler(c *gin.Context) {
	platform := strings.TrimSpace(c.Query("platform"))
	if platform == "" {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	version, err := service.GetCurrentAppVersion(platform)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, version)
}
