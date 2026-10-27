package app

import (
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FeedbackHandler struct {
	logger   *zap.Logger
	feedback *appservice.FeedbackService
}

func NewFeedbackHandler(feedback *appservice.FeedbackService, logger *zap.Logger) *FeedbackHandler {
	return &FeedbackHandler{
		logger:   logger,
		feedback: feedback,
	}
}

// CreateFeedbackHandler 接收文字反馈及可选图片，兼容旧版 JSON 请求。
func (f *FeedbackHandler) CreateFeedbackHandler(c *gin.Context) {
	log := f.logger.With(logutil.OperationFields(c, "app.feedback.create")...)

	// 解析前限制请求总大小，三张图片之外预留 multipart 表单开销。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body,
		appservice.MaxFeedbackAttachments*appservice.MaxFeedbackAttachmentSize+(64<<10))
	defer func() {
		if c.Request.MultipartForm != nil {
			if err := c.Request.MultipartForm.RemoveAll(); err != nil {
				log.Warn("清理反馈上传临时文件失败", logutil.Error(err))
			}
		}
	}()
	req, uploads, err := parseFeedbackSubmission(c)
	if err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	feedback, err := f.feedback.CreateFeedback(studentID, req, uploads...)
	if err != nil {
		if errors.Is(err, appservice.ErrorInvalidFeedback) {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		log.Error("提交反馈失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	log.Info("提交反馈成功",
		zap.Uint("feedback_id", feedback.ID),
		zap.Bool("anonymous", req.IsAnonymous),
	)
	response.Success(c, feedback)
}

func parseFeedbackSubmission(c *gin.Context) (*dto.FeedbackCreateRequest, []*multipart.FileHeader, error) {
	req := new(dto.FeedbackCreateRequest)
	if c.ContentType() != "multipart/form-data" {
		return req, nil, c.ShouldBindJSON(req)
	}
	if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
		return nil, nil, err
	}
	form := c.Request.MultipartForm
	if len(form.Value["content"]) != 1 || len(form.Value["is_anonymous"]) != 1 {
		return nil, nil, appservice.ErrorInvalidFeedback
	}
	anonymous, err := strconv.ParseBool(form.Value["is_anonymous"][0])
	if err != nil {
		return nil, nil, err
	}
	req.Content = form.Value["content"][0]
	req.IsAnonymous = anonymous
	if len(form.File) > appservice.MaxFeedbackAttachments {
		return nil, nil, appservice.ErrorInvalidFeedback
	}
	uploads := make([]*multipart.FileHeader, 0, len(form.File))
	// 文件字段连续编号，同时拒绝重复字段、未知字段和超量文件。
	for index := 0; index < len(form.File); index++ {
		files := form.File[fmt.Sprintf("attachment_%d", index)]
		if len(files) != 1 || files[0].Size > appservice.MaxFeedbackAttachmentSize {
			return nil, nil, appservice.ErrorInvalidFeedback
		}
		uploads = append(uploads, files[0])
	}
	return req, uploads, nil
}
