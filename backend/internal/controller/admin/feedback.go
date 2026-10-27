package admin

import (
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type FeedbackHandler struct {
	logger   *zap.Logger
	feedback *adminservice.FeedbackService
}

func NewFeedbackHandler(feedback *adminservice.FeedbackService, logger *zap.Logger) *FeedbackHandler {
	return &FeedbackHandler{
		logger:   logger,
		feedback: feedback,
	}
}

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

func (f *FeedbackHandler) GetFeedbacksHandler(c *gin.Context) {
	log := f.logger.With(logutil.OperationFields(c, "admin.feedback.get_feedbacks")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	query, ok := parseFeedbackQuery(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := f.feedback.GetAdminFeedbacks(query)
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (f *FeedbackHandler) GetFeedbackHandler(c *gin.Context) {
	log := f.logger.With(logutil.OperationFields(c, "admin.feedback.get_feedback")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	data, err := f.feedback.GetAdminFeedback(id)
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (f *FeedbackHandler) DeleteFeedbackHandler(c *gin.Context) {
	log := f.logger.With(logutil.OperationFields(c, "admin.feedback.delete_feedback")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	if err := f.feedback.DeleteAdminFeedback(id); err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("反馈删除成功")
	response.Success(c, nil)
}

func (f *FeedbackHandler) UpdateFeedbackHandlingHandler(c *gin.Context) {
	log := f.logger.With(logutil.OperationFields(c, "admin.feedback.update_feedback_handling")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	req := new(dto.FeedbackHandlingRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := f.feedback.UpdateAdminFeedbackHandling(id, req)
	if err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("反馈处理状态更新成功", zap.Bool("is_handled", *req.IsHandled))
	response.Success(c, data)
}

func (f *FeedbackHandler) GetFeedbackAttachmentHandler(c *gin.Context) {
	log := f.logger.With(logutil.OperationFields(c, "admin.feedback.get_feedback_attachment")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	image, err := f.feedback.OpenAdminFeedbackAttachment(id, index)
	if err != nil {
		managementError(c, err, log)
		return
	}
	defer func() {
		if err := image.Close(); err != nil {
			log.Warn("关闭反馈附件失败", logutil.Error(err))
		}
	}()
	info, err := image.Stat()
	if err != nil || !info.Mode().IsRegular() {
		log.Error("读取反馈附件信息失败", zap.String("reason", "stat_failed_or_not_regular"), logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), image)
}
