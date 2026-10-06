package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/service"
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

	query := dto.AdminClassListQuery{
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

	data, err := service.GetAdminClasses(query)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}
