package controller

import (
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/service"
	"backend/internal/utils/validator"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// ProfileHandler 完善个人信息
func ProfileHandler(c *gin.Context) {
	// 参数绑定
	p := new(dto.ProfileRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	if !validateAndNormalizeProfile(p) {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	// 业务处理
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	if err := service.Profile(p, studentID); err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}

	response.Success(c, nil)
}

func GetProfileHandler(c *gin.Context) {
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	data, err := service.GetProfile(studentID)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

// GetClassesHandler returns classes for the profile form selector.
func GetClassesHandler(c *gin.Context) {
	data, err := service.GetClasses()
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

// GetApartmentsHandler returns enabled apartments for the profile form selector.
func GetApartmentsHandler(c *gin.Context) {
	data, err := service.GetApartments(c.Query("gender"))
	if err != nil {
		if errors.Is(err, service.ErrorInvalidApartmentGender) {
			response.Error(c, response.CodeInvalidParam)
			return
		}
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, data)
}

// validateAndNormalizeProfile 规范化并校验个人资料请求。
func validateAndNormalizeProfile(p *dto.ProfileRequest) bool {
	p.Name = strings.TrimSpace(p.Name)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Gender = strings.TrimSpace(p.Gender)
	p.ParentName = strings.TrimSpace(p.ParentName)
	p.ParentPhone = strings.TrimSpace(p.ParentPhone)
	p.DormitoryNumber = strings.TrimSpace(p.DormitoryNumber)
	p.TeacherName = strings.TrimSpace(p.TeacherName)

	if p.Name == "" || p.ParentName == "" || p.TeacherName == "" || p.ClassID == 0 {
		return false
	}
	if p.Gender != models.GenderMale && p.Gender != models.GenderFemale {
		return false
	}
	return validator.IsMainlandMobile(p.Phone) && validator.IsMainlandMobile(p.ParentPhone)
}
