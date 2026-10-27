package admin

import (
	"backend/internal/dto"
	logutil "backend/internal/logger"
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
	log := l.logger.With(logutil.OperationFields(c, "admin.leave_type.get_leave_types")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

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
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (l *LeaveTypeHandler) CreateLeaveTypeHandler(c *gin.Context) {
	log := l.logger.With(logutil.OperationFields(c, "admin.leave_type.create_leave_type")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	p := new(dto.LeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := l.leaveType.CreateAdminLeaveType(p)
	if err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("请假类型新增成功", zap.Uint("target_id", data.ID))
	response.Success(c, data)
}

func (l *LeaveTypeHandler) UpdateLeaveTypeHandler(c *gin.Context) {
	log := l.logger.With(logutil.OperationFields(c, "admin.leave_type.update_leave_type")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	p := new(dto.LeaveTypeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := l.leaveType.UpdateAdminLeaveType(p, id)
	if err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("请假类型更新成功")
	response.Success(c, data)
}

func (l *LeaveTypeHandler) DeleteLeaveTypeHandler(c *gin.Context) {
	log := l.logger.With(logutil.OperationFields(c, "admin.leave_type.delete_leave_type")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	if err := l.leaveType.DeleteAdminLeaveType(id); err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("请假类型删除成功")
	response.Success(c, nil)
}
