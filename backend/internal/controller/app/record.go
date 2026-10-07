package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateRecordHandler 创建请假记录。
func CreateRecordHandler(c *gin.Context) {
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

	record, err := appservice.CreateRecord(studentID, req)
	if err != nil {
		switch {
		case errors.Is(err, appservice.ErrorInvalidRecord), errors.Is(err, appservice.ErrorLeaveTypeUnavailable):
			response.Error(c, response.CodeInvalidParam)
		case errors.Is(err, appservice.ErrorProfileIncomplete):
			response.Error(c, response.CodeProfileIncomplete)
		default:
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	response.Success(c, record)
}

func GetRecordHandler(c *gin.Context) {
	recordID, ok := recordIDFromContext(c)
	if !ok {
		return
	}
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	record, err := appservice.GetRecord(studentID, recordID)
	if err != nil {
		if errors.Is(err, mysql.ErrorRecordNotExist) {
			response.Error(c, response.CodeRecordNotExist)
		} else {
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	response.Success(c, record)
}

func GetRecordsLIstHandler(c *gin.Context) {
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
	records, err := appservice.GetRecordsList(studentID, query)
	if err != nil {
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

func DeleteRecordHandler(c *gin.Context) {
	recordID, ok := recordIDFromContext(c)
	if !ok {
		return
	}
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	if err := appservice.DeleteRecord(studentID, recordID); err != nil {
		if errors.Is(err, mysql.ErrorRecordNotExist) {
			response.Error(c, response.CodeRecordNotExist)
		} else {
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
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
