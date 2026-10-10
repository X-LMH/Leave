package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

func parseAdminRecordQuery(c *gin.Context) (dto.AdminRecordQuery, bool) {
	pagination, ok := request.ParsePagination(c)
	if !ok {
		return dto.AdminRecordQuery{}, false
	}
	query := dto.AdminRecordQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		StudentID: strings.TrimSpace(c.Query("student_id")),
		Name:      strings.TrimSpace(c.Query("name")),
	}
	// 保留空字符串快照，区分“未筛选”和“筛选值为空”。
	values := c.Request.URL.Query()
	college, hasCollege := values["college"]
	major, hasMajor := values["major"]
	className, hasClass := values["class_name"]
	if hasCollege || hasMajor || hasClass {
		if !hasCollege || !hasMajor || !hasClass {
			return query, false
		}
		query.ClassFilter = &dto.RecordClassFilter{
			College:   college[0],
			Major:     major[0],
			ClassName: className[0],
		}
	}
	if value := c.Query("leave_type_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 32)
		if err != nil || id == 0 {
			return query, false
		}
		query.LeaveTypeID = uint(id)
	}
	if value := c.Query("is_leave_school"); value != "" {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return query, false
		}
		query.IsLeaveSchool = &v
	}
	for key, target := range map[string]**time.Time{"start_time_from": &query.StartTimeFrom, "start_time_to": &query.StartTimeTo} {
		if value := c.Query(key); value != "" {
			v, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return query, false
			}
			*target = &v
		}
	}
	if query.StartTimeFrom != nil && query.StartTimeTo != nil && !query.StartTimeTo.After(*query.StartTimeFrom) {
		return query, false
	}
	return query, true
}

func GetRecordsHandler(c *gin.Context) {
	query, ok := parseAdminRecordQuery(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.GetAdminRecords(query)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func GetRecordHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	data, err := adminservice.GetAdminRecord(id)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func UpdateRecordHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	input := new(dto.AdminRecordUpdateRequest)
	if err := c.ShouldBindJSON(input); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.UpdateAdminRecord(id, input)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}
