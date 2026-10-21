package app

import (
	"backend/internal/request"
	"backend/internal/response"

	"github.com/gin-gonic/gin"
)

func (r *RecordHandler) GetHomeHandler(c *gin.Context) {
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	home, err := r.record.GetHome(studentID)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, home)
}
