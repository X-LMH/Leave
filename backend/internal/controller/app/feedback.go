package app

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"

	"github.com/gin-gonic/gin"
)

// CreateFeedbackHandler submits text feedback from the current user.
func CreateFeedbackHandler(c *gin.Context) {
	req := new(dto.FeedbackCreateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	feedback, err := appservice.CreateFeedback(studentID, req)
	if err != nil {
		if errors.Is(err, appservice.ErrorInvalidFeedback) {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, feedback)
}
