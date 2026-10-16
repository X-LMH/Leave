package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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
	if value := c.Query("is_handled"); value != "" {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return query, false
		}
		query.IsHandled = &v
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

func UpdateFeedbackHandlingHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	req := new(dto.FeedbackHandlingRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.UpdateAdminFeedbackHandling(id, req)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func GetFeedbackAttachmentHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	image, err := adminservice.OpenAdminFeedbackAttachment(id, index)
	if err != nil {
		managementError(c, err)
		return
	}
	defer func() {
		if err := image.Close(); err != nil {
			log.Printf("关闭反馈附件失败: %v", err)
		}
	}()
	info, err := image.Stat()
	if err != nil || !info.Mode().IsRegular() {
		response.Error(c, response.CodeServerBusy)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), image)
}
