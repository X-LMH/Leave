package app

import (
	logutil "backend/internal/logger"
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
	log := o.logger.With(logutil.OperationFields(c, "app.options.classes")...)

	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	data, err := o.options.GetProfileClassOptions(studentID)
	if err != nil {
		log.Error("查询班级选项失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

func (o *OptionsHandler) GetApartmentOptionsHandler(c *gin.Context) {
	log := o.logger.With(logutil.OperationFields(c, "app.options.apartments")...)

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
		log.Error("查询公寓选项失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

func (o *OptionsHandler) GetLeaveTypeOptionsHandler(c *gin.Context) {
	log := o.logger.With(logutil.OperationFields(c, "app.options.leave_types")...)

	data, err := o.commonOptions.GetLeaveTypeOptions(true)
	if err != nil {
		log.Error("查询请假类型选项失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}
