package admin

import (
	"backend/internal/response"
	"backend/internal/service/common"
	"errors"

	"github.com/gin-gonic/gin"
)

func GetClassOptionsHandler(c *gin.Context) {
	data, err := common.GetClassOptions(false)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func GetApartmentOptionsHandler(c *gin.Context) {
	data, err := common.GetApartmentOptions(c.Query("gender"), false)
	if err != nil {
		if errors.Is(err, common.ErrorInvalidApartmentGender) {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func GetLeaveTypeOptionsHandler(c *gin.Context) {
	data, err := common.GetLeaveTypeOptions(false)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}
