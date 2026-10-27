package admin

import (
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"strings"

	"go.uber.org/zap"
)

type ApartmentHandler struct {
	logger    *zap.Logger
	apartment *adminservice.ApartmentService
}

func NewApartmentHandler(apartment *adminservice.ApartmentService, logger *zap.Logger) *ApartmentHandler {
	return &ApartmentHandler{
		logger:    logger,
		apartment: apartment,
	}
}

func (a *ApartmentHandler) GetApartmentsHandler(c *gin.Context) {
	log := a.logger.With(logutil.OperationFields(c, "admin.apartment.get_apartments")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	enabled, ok := enabledFilter(c)
	if !ok {
		return
	}
	data, err := a.apartment.GetAdminApartments(dto.ApartmentListQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		Name:      strings.TrimSpace(c.Query("name")),
		Gender:    strings.TrimSpace(c.Query("gender")),
		IsEnabled: enabled,
	})
	if err != nil {
		managementError(c, err, log)
		return
	}
	response.Success(c, data)
}

func (a *ApartmentHandler) CreateApartmentHandler(c *gin.Context) {
	log := a.logger.With(logutil.OperationFields(c, "admin.apartment.create_apartment")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	p := new(dto.ApartmentRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := a.apartment.CreateAdminApartment(p)
	if err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("公寓新增成功", zap.Uint("target_id", data.ID))
	response.Success(c, data)
}

func (a *ApartmentHandler) UpdateApartmentHandler(c *gin.Context) {
	log := a.logger.With(logutil.OperationFields(c, "admin.apartment.update_apartment")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	p := new(dto.ApartmentRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := a.apartment.UpdateAdminApartment(id, p)
	if err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("公寓更新成功")
	response.Success(c, data)
}

func (a *ApartmentHandler) DeleteApartmentHandler(c *gin.Context) {
	log := a.logger.With(logutil.OperationFields(c, "admin.apartment.delete_apartment")...)
	if actorID, err := request.GetCurrentStuID(c); err == nil {
		log = log.With(zap.String("actor_id", actorID))
	}

	id, ok := managementID(c)
	if !ok {
		return
	}
	log = log.With(zap.Uint("target_id", id))
	if err := a.apartment.DeleteAdminApartment(id); err != nil {
		managementError(c, err, log)
		return
	}
	log.Info("公寓删除成功")
	response.Success(c, nil)
}
