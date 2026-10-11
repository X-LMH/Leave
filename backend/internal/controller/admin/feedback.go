package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func parseFeedbackQuery(c *gin.Context) (dto.FeedbackListQuery, bool) {
	pagination, ok := request.ParsePagination(c)
	if !ok {
		return dto.FeedbackListQuery{}, false
	}
	query := dto.FeedbackListQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		Keyword:   strings.TrimSpace(c.Query("keyword")),
		StudentID: strings.TrimSpace(c.Query("student_id")),
	}
	if value := c.Query("is_anonymous"); value != "" {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return query, false
		}
		query.IsAnonymous = &v
	}
	return query, true
}

func GetFeedbacksHandler(c *gin.Context) {
	query, ok := parseFeedbackQuery(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.GetAdminFeedbacks(query)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func GetFeedbackHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	data, err := adminservice.GetAdminFeedback(id)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func DeleteFeedbackHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	if err := adminservice.DeleteAdminFeedback(id); err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, nil)
}
