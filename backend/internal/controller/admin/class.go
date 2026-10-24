package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ClassHandler struct {
	logger *zap.Logger
	class  *adminservice.ClassService
}

func NewClassHandler(class *adminservice.ClassService, logger *zap.Logger) *ClassHandler {
	return &ClassHandler{
		logger: logger,
		class:  class,
	}
}

// GetClassesHandler returns a paginated class list for the admin platform.
func (c *ClassHandler) GetClassesHandler(ctx *gin.Context) {
	pagination, ok := request.ParsePagination(ctx)
	if !ok {
		response.Error(ctx, response.CodeInvalidParam)
		return
	}

	query := dto.ClassListQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		College:   strings.TrimSpace(ctx.Query("college")),
		Major:     strings.TrimSpace(ctx.Query("major")),
		ClassName: strings.TrimSpace(ctx.Query("class_name")),
	}
	if value := ctx.Query("is_enabled"); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			response.Error(ctx, response.CodeInvalidParam)
			return
		}
		query.IsEnabled = &enabled
	}

	data, err := c.class.GetAdminClasses(query)
	if err != nil {
		response.Error(ctx, response.CodeServerBusy)
		return
	}
	response.Success(ctx, data)
}

func (c *ClassHandler) CreateClassHandler(ctx *gin.Context) {
	p := new(dto.ClassRequest)
	if err := ctx.ShouldBindJSON(p); err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		return
	}
	data, err := c.class.CreateAdminClass(p)
	if err != nil {
		managementError(ctx, err)
		return
	}
	response.Success(ctx, data)
}

func (c *ClassHandler) UpdateClassHandler(ctx *gin.Context) {
	id, ok := managementID(ctx)
	if !ok {
		return
	}
	p := new(dto.ClassRequest)
	if err := ctx.ShouldBindJSON(p); err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		return
	}
	data, err := c.class.UpdateAdminClass(p, id)
	if err != nil {
		managementError(ctx, err)
		return
	}
	response.Success(ctx, data)
}

func (c *ClassHandler) DeleteClassHandler(ctx *gin.Context) {
	id, ok := managementID(ctx)
	if !ok {
		return
	}
	if err := c.class.DeleteAdminClass(id); err != nil {
		managementError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
