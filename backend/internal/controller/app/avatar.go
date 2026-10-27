package app

import (
	logutil "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AvatarHandler struct {
	logger *zap.Logger
	avatar *appservice.AvatarService
}

func NewAvatarHandler(avatar *appservice.AvatarService, logger *zap.Logger) *AvatarHandler {
	return &AvatarHandler{
		logger: logger,
		avatar: avatar,
	}
}

const maxAvatarSize = 5 << 20

func (a *AvatarHandler) UploadAvatarHandler(c *gin.Context) {
	log := a.logger.With(logutil.OperationFields(c, "app.avatar.upload")...)

	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	// 在解析前限制整个请求；为 multipart 头部预留 64 KiB。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarSize+(64<<10))
	upload, err := c.FormFile("file")
	if c.Request.MultipartForm != nil {
		defer func() {
			if err := c.Request.MultipartForm.RemoveAll(); err != nil {
				log.Warn("清理头像上传临时文件失败", logutil.Error(err))
			}
		}()
	}
	if err != nil || len(c.Request.MultipartForm.File["file"]) != 1 {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	if upload.Size > maxAvatarSize {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	content, err := upload.Open()
	if err != nil {
		log.Error("上传头像失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	defer func() {
		if err := content.Close(); err != nil {
			log.Warn("关闭头像上传文件失败", logutil.Error(err))
		}
	}()

	data, err := a.avatar.UploadAvatar(studentID, content)
	if err != nil {
		switch {
		case errors.Is(err, appservice.ErrorInvalidAvatar):
			response.Error(c, response.CodeInvalidParam)
		case errors.Is(err, appservice.ErrorProfileIncomplete):
			response.Error(c, response.CodeProfileIncomplete)
		default:
			log.Error("上传头像失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	log.Info("更新头像成功", zap.String("actor_id", studentID))
	response.Success(c, data)
}
