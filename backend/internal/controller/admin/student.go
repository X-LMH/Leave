package admin

import (
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StudentHandler struct {
	logger  *zap.Logger
	student *adminservice.StudentService
}

func NewStudentHandler(student *adminservice.StudentService, logger *zap.Logger) *StudentHandler {
	return &StudentHandler{
		logger:  logger,
		student: student,
	}
}

func (s *StudentHandler) GetStudentsHandler(c *gin.Context) {
	log := s.logger.With(logutil.OperationFields(c, "admin.student.get_students")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	query := dto.StudentQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		StudentID: strings.TrimSpace(c.Query("student_id")),
		Name:      strings.TrimSpace(c.Query("name")),
	}
	if value := c.Query("class_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 32)
		if err != nil || id == 0 {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		query.ClassID = uint(id)
	}
	if value := c.Query("status"); value != "" {
		status, err := strconv.ParseUint(value, 10, 8)
		if err != nil || status > 1 {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		v := uint8(status)
		query.Status = &v
	}
	data, err := s.student.GetAdminStudents(query)
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (s *StudentHandler) GetStudentHandler(c *gin.Context) {
	log := s.logger.With(logutil.OperationFields(c, "admin.student.get_student")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	data, err := s.student.GetAdminStudent(id)
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (s *StudentHandler) UpdateStudentHandler(c *gin.Context) {
	log := s.logger.With(logutil.OperationFields(c, "admin.student.update_student")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	input := new(dto.ProfileRequest)
	if err := c.ShouldBindJSON(input); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := s.student.UpdateAdminStudent(id, input)
	if err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("学生资料更新成功")
	response.Success(c, data)
}

func (s *StudentHandler) UpdateStudentStatusHandler(c *gin.Context) {
	log := s.logger.With(logutil.OperationFields(c, "admin.student.update_student_status")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	input := new(dto.StudentStatusRequest)
	if err := c.ShouldBindJSON(input); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := s.student.UpdateAdminStudentStatus(id, input)
	if err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("学生账号状态更新成功", zap.Uint8("status", *input.Status))
	response.Success(c, data)
}

func (s *StudentHandler) ResetStudentPasswordHandler(c *gin.Context) {
	log := s.logger.With(logutil.OperationFields(c, "admin.student.reset_student_password")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	if err := s.student.ResetAdminStudentPassword(id); err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("重置密码成功")
	response.Success(c, nil)
}
