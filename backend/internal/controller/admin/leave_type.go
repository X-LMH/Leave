package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"strings"
)

func GetLeaveTypesHandler(c *gin.Context) {
	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	enabled, ok := enabledFilter(c)
	if !ok {
		return
	}
	query := dto.LeaveTypeListQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		Name:      strings.TrimSpace(c.Query("name")),
		IsEnabled: enabled,
	}
	data, err := adminservice.GetAdminLeaveTypes(query)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func CreateLeaveTypeHandler(c *gin.Context) {
	p := new(dto.LeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.CreateAdminLeaveType(p)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func UpdateLeaveTypeHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	p := new(dto.LeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.UpdateAdminLeaveType(p, id)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func DeleteLeaveTypeHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	if err := adminservice.DeleteAdminLeaveType(id); err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, nil)
}
