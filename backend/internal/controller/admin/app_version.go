package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AppVersionHandler struct {
	appVersion *adminservice.AppVersionService
}

func NewAppVersionHandler(appVersion *adminservice.AppVersionService) *AppVersionHandler {
	return &AppVersionHandler{
		appVersion: appVersion,
	}
}

func (a *AppVersionHandler) GetAppVersionsHandler(c *gin.Context) {
	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := a.appVersion.GetAdminAppVersions(dto.AppVersionListQuery{
		Keyword:  c.Query("keyword"),
		Platform: c.Query("platform"),
		Status:   c.Query("status"),
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func (a *AppVersionHandler) GetCurrentAppVersionHandler(c *gin.Context) {
	data, err := a.appVersion.GetAdminCurrentAppVersion()
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func (a *AppVersionHandler) GetAppVersionHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := a.appVersion.GetAdminAppVersion(id)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}
