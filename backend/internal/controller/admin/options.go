package admin

import (
	logutil "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/service/common"
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OptionsHandler struct {
	logger  *zap.Logger
	options *common.OptionsService
}

func NewOptionsHandler(options *common.OptionsService, logger *zap.Logger) *OptionsHandler {
	return &OptionsHandler{
		logger:  logger,
		options: options,
	}
}

func (o *OptionsHandler) GetClassOptionsHandler(c *gin.Context) {
	log := o.logger.With(logutil.OperationFields(c, "admin.options.get_class_options")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	data, err := o.options.GetClassOptions(false)
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (o *OptionsHandler) GetApartmentOptionsHandler(c *gin.Context) {
	log := o.logger.With(logutil.OperationFields(c, "admin.options.get_apartment_options")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	data, err := o.options.GetApartmentOptions(c.Query("gender"), false)
	if err != nil {
		if errors.Is(err, common.ErrorInvalidApartmentGender) {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (o *OptionsHandler) GetLeaveTypeOptionsHandler(c *gin.Context) {
	log := o.logger.With(logutil.OperationFields(c, "admin.options.get_leave_type_options")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	data, err := o.options.GetLeaveTypeOptions(false)
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}
