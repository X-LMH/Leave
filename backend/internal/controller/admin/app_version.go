package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAppVersionsHandler(c *gin.Context) {
	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.GetAdminAppVersions(dto.AppVersionListQuery{
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

func GetCurrentAppVersionHandler(c *gin.Context) {
	data, err := adminservice.GetAdminCurrentAppVersion()
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func GetAppVersionHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.GetAdminAppVersion(id)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}
