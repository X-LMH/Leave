package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"strings"

	"go.uber.org/zap"
)

type LeaveTypeHandler struct {
	logger    *zap.Logger
	leaveType *adminservice.LeaveTypeService
}

func NewLeaveTypeHandler(leaveType *adminservice.LeaveTypeService, logger *zap.Logger) *LeaveTypeHandler {
	return &LeaveTypeHandler{
		logger:    logger,
		leaveType: leaveType,
	}
}

func (l *LeaveTypeHandler) GetLeaveTypesHandler(c *gin.Context) {
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
	data, err := l.leaveType.GetAdminLeaveTypes(query)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func (l *LeaveTypeHandler) CreateLeaveTypeHandler(c *gin.Context) {
	p := new(dto.LeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := l.leaveType.CreateAdminLeaveType(p)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func (l *LeaveTypeHandler) UpdateLeaveTypeHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	p := new(dto.LeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := l.leaveType.UpdateAdminLeaveType(p, id)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func (l *LeaveTypeHandler) DeleteLeaveTypeHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	if err := l.leaveType.DeleteAdminLeaveType(id); err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, nil)
}
