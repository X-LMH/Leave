package app

import (
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxAvatarSize = 5 << 20

func UploadAvatarHandler(c *gin.Context) {
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
				log.Printf("清理头像上传临时文件失败: %v", err)
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
		response.Error(c, response.CodeServerBusy)
		return
	}
	defer func() {
		if err := content.Close(); err != nil {
			log.Printf("关闭头像上传文件失败: %v", err)
		}
	}()

	data, err := appservice.UploadAvatar(studentID, content)
	if err != nil {
		switch {
		case errors.Is(err, appservice.ErrorInvalidAvatar):
			response.Error(c, response.CodeInvalidParam)
		case errors.Is(err, appservice.ErrorProfileIncomplete):
			response.Error(c, response.CodeProfileIncomplete)
		default:
			log.Printf("上传头像失败 student_id=%s: %v", studentID, err)
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	response.Success(c, data)
}
