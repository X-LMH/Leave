package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/service"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func managementError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrorInvalidAdminInput):
		response.Error(c, response.CodeInvalidParam)
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, response.CodeManagementNotFound)
	case mysql.IsDuplicateKeyError(err):
		response.Error(c, response.CodeManagementDuplicate)
	case errors.Is(err, mysql.ErrorClassInUse):
		response.Error(c, response.CodeClassInUse)
	default:
		response.Error(c, response.CodeServerBusy)
	}
}
func managementID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || value == 0 {
		response.Error(c, response.CodeInvalidParam)
		return 0, false
	}
	return uint(value), true
}
func enabledFilter(c *gin.Context) (*bool, bool) {
	value := c.Query("is_enabled")
	if value == "" {
		return nil, true
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		response.Error(c, response.CodeInvalidParam)
		return nil, false
	}
	return &enabled, true
}
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
	query := dto.AdminLeaveTypeListQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		Name:      strings.TrimSpace(c.Query("name")),
		IsEnabled: enabled,
	}
	data, err := service.GetAdminLeaveTypes(query)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func CreateLeaveTypeHandler(c *gin.Context) {
	p := new(dto.AdminLeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := service.CreateAdminLeaveType(p)
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
	p := new(dto.AdminLeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := service.UpdateAdminLeaveType(p, id)
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
	if err := service.DeleteAdminLeaveType(id); err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, nil)
}

func CreateClassHandler(c *gin.Context) {
	p := new(dto.AdminClassRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := service.CreateAdminClass(p)
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
	p := new(dto.AdminClassRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := service.UpdateAdminClass(p, id)
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
	if err := service.DeleteAdminClass(id); err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, nil)
}
