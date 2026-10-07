package app

import (
	"backend/internal/response"
	"backend/internal/service/common"
	"errors"

	"github.com/gin-gonic/gin"
)

func GetClassOptionsHandler(c *gin.Context) {
	data, err := common.GetClassOptions(true)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

func GetApartmentOptionsHandler(c *gin.Context) {
	data, err := common.GetApartmentOptions(c.Query("gender"), true)
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
