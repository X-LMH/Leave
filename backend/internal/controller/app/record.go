package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecordHandler struct {
	logger *zap.Logger
	record *appservice.RecordService
}

func NewRecordHandler(record *appservice.RecordService, logger *zap.Logger) *RecordHandler {
	return &RecordHandler{
		logger: logger,
		record: record,
	}
}

// CreateRecordHandler 创建请假记录。
func (r *RecordHandler) CreateRecordHandler(c *gin.Context) {
	log := r.logger.With(logutil.OperationFields(c, "app.record.create")...)

	req := new(dto.RecordCreateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	record, err := r.record.CreateRecord(studentID, req)
	if err != nil {
		switch {
		case errors.Is(err, appservice.ErrorInvalidRecord), errors.Is(err, appservice.ErrorLeaveTypeUnavailable):
			response.Error(c, response.CodeInvalidParam)
		case errors.Is(err, appservice.ErrorProfileIncomplete):
			response.Error(c, response.CodeProfileIncomplete)
		default:
			log.Error("创建请假记录失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	log.Info("创建请假记录成功",
		zap.Uint("record_id", record.ID),
		zap.String("actor_id", studentID),
	)
	response.Success(c, record)
}

func (r *RecordHandler) GetRecordHandler(c *gin.Context) {
	log := r.logger.With(logutil.OperationFields(c, "app.record.get")...)

	recordID, ok := recordIDFromContext(c)
	if !ok {
		return
	}
	log = log.With(zap.Int("record_id", recordID))
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	record, err := r.record.GetRecord(studentID, recordID)
	if err != nil {
		if errors.Is(err, mysql.ErrorRecordNotExist) {
			response.Error(c, response.CodeRecordNotExist)
		} else {
			log.Error("查询请假记录失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	response.Success(c, record)
}

func (r *RecordHandler) GetRecordsLIstHandler(c *gin.Context) {
	log := r.logger.With(logutil.OperationFields(c, "app.record.list")...)

	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	query := dto.RecordListQuery{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}
	leaveType := queryInt(c, "leave_type", 0)
	if leaveType < 0 {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	query.LeaveType = uint(leaveType)
	records, err := r.record.GetRecordsList(studentID, query)
	if err != nil {
		log.Error("查询请假记录列表失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, records)
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}
	return parsed
}

func (r *RecordHandler) DeleteRecordHandler(c *gin.Context) {
	log := r.logger.With(logutil.OperationFields(c, "app.record.delete")...)

	recordID, ok := recordIDFromContext(c)
	if !ok {
		return
	}
	log = log.With(zap.Int("record_id", recordID))
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	if err := r.record.DeleteRecord(studentID, recordID); err != nil {
		if errors.Is(err, mysql.ErrorRecordNotExist) {
			response.Error(c, response.CodeRecordNotExist)
		} else {
			log.Error("删除请假记录失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	log.Info("删除请假记录成功",
		zap.String("actor_id", studentID),
	)
	response.Success(c, nil)
}

func recordIDFromContext(c *gin.Context) (int, bool) {
	recordID, err := strconv.Atoi(c.Param("id"))
	if err != nil || recordID <= 0 {
		response.Error(c, response.CodeInvalidParam)
		return 0, false
	}
	return recordID, true
}
