package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClassesHandler returns a paginated class list for the admin platform.
func GetClassesHandler(c *gin.Context) {
	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	query := dto.ClassListQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		College:   strings.TrimSpace(c.Query("college")),
		Major:     strings.TrimSpace(c.Query("major")),
		ClassName: strings.TrimSpace(c.Query("class_name")),
	}
	if value := c.Query("is_enabled"); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		query.IsEnabled = &enabled
	}

	data, err := adminservice.GetAdminClasses(query)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

func CreateClassHandler(c *gin.Context) {
	p := new(dto.ClassRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.CreateAdminClass(p)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func UpdateClassHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	p := new(dto.ClassRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.UpdateAdminClass(p, id)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func DeleteClassHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	if err := adminservice.DeleteAdminClass(id); err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, nil)
}
