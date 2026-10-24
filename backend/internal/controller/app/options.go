package app

import (
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"backend/internal/service/common"
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OptionsHandler struct {
	logger        *zap.Logger
	options       *appservice.OptionsService
	commonOptions *common.OptionsService
}

func NewOptionsHandler(options *appservice.OptionsService, commonOptions *common.OptionsService, logger *zap.Logger) *OptionsHandler {
	return &OptionsHandler{
		logger:        logger,
		options:       options,
		commonOptions: commonOptions,
	}
}

func (o *OptionsHandler) GetClassOptionsHandler(c *gin.Context) {
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	data, err := o.options.GetProfileClassOptions(studentID)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

func (o *OptionsHandler) GetApartmentOptionsHandler(c *gin.Context) {
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	data, err := o.options.GetProfileApartmentOptions(studentID, c.Query("gender"))
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

func (o *OptionsHandler) GetLeaveTypeOptionsHandler(c *gin.Context) {
	data, err := o.commonOptions.GetLeaveTypeOptions(true)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}
