package admin

import (
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"strings"
)

func GetApartmentsHandler(c *gin.Context) {
	pagination, ok := request.ParsePagination(c)
	if !ok {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	enabled, ok := enabledFilter(c)
	if !ok {
		return
	}
	data, err := adminservice.GetAdminApartments(dto.ApartmentListQuery{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		Name:      strings.TrimSpace(c.Query("name")),
		Gender:    strings.TrimSpace(c.Query("gender")),
		IsEnabled: enabled,
	})
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func CreateApartmentHandler(c *gin.Context) {
	p := new(dto.ApartmentRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.CreateAdminApartment(p)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func UpdateApartmentHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	p := new(dto.ApartmentRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	data, err := adminservice.UpdateAdminApartment(id, p)
	if err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, data)
}

func DeleteApartmentHandler(c *gin.Context) {
	id, ok := managementID(c)
	if !ok {
		return
	}
	if err := adminservice.DeleteAdminApartment(id); err != nil {
		managementError(c, err)
		return
	}
	response.Success(c, nil)
}
