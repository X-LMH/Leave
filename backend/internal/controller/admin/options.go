package admin

import (
	"backend/internal/response"
	"backend/internal/service/common"
	"errors"

	"github.com/gin-gonic/gin"
)

type OptionsHandler struct {
	options *common.OptionsService
}

func NewOptionsHandler(options *common.OptionsService) *OptionsHandler {
	return &OptionsHandler{
		options: options,
	}
}

func (o *OptionsHandler) GetClassOptionsHandler(c *gin.Context) {
	data, err := o.options.GetClassOptions(false)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func (o *OptionsHandler) GetApartmentOptionsHandler(c *gin.Context) {
	data, err := o.options.GetApartmentOptions(c.Query("gender"), false)
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

func (o *OptionsHandler) GetLeaveTypeOptionsHandler(c *gin.Context) {
	data, err := o.options.GetLeaveTypeOptions(false)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}
