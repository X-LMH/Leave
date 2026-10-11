package app

import (
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"backend/internal/service/common"
	"errors"

	"github.com/gin-gonic/gin"
)

func GetClassOptionsHandler(c *gin.Context) {
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	data, err := appservice.GetProfileClassOptions(studentID)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

func GetApartmentOptionsHandler(c *gin.Context) {
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	data, err := appservice.GetProfileApartmentOptions(studentID, c.Query("gender"))
	if err != nil {
		if errors.Is(err, common.ErrorInvalidApartmentGender) {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

func GetLeaveTypeOptionsHandler(c *gin.Context) {
	data, err := common.GetLeaveTypeOptions(true)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}
