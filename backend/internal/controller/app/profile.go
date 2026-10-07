package app

import (
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
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
	if err := appservice.Profile(p, studentID); err != nil {
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
	data, err := appservice.GetProfile(studentID)
	if err != nil {
		if errors.Is(err, appservice.ErrorProfileIncomplete) {
			response.Error(c, response.CodeProfileIncomplete)
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
