package controller

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/response"
	"backend/internal/service"
	"backend/internal/utils/validator"
	"errors"
	"fmt"
	"strconv"
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
	studentID, err := GetCurrentStuID(c)
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
	studentID, err := GetCurrentStuID(c)
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

// CreateRecordHandler 记录用户行为
func CreateRecordHandler(c *gin.Context) {
	p := new(models.ParamRecord)

	if err := c.ShouldBindJSON(p); err != nil {
		fmt.Println(p)
		response.Error(c, response.CodeInvalidParam)
		return
	}

	studentID, err := GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	err = service.CreateRecord(studentID, p)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorUserNotExist):
			response.Error(c, response.CodeNeedLogin)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
	}
	response.Success(c, nil)
}

func GetRecordHandler(c *gin.Context) {
	req := c.Param("id")
	recordID, err := strconv.Atoi(req)
	if err != nil {
		fmt.Println(req)
		response.Error(c, response.CodeInvalidParam)
		return
	}

	record, err := service.GetRecord(recordID)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorRecordNotExist):
			response.Error(c, response.CodeRecordNotExist)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
	}

	response.Success(c, record)
}

func GetRecordsLIstHandler(c *gin.Context) {
	studentID, err := GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}

	res, err := service.GetRecordsList(studentID)
	if err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, res)
}

func DeleteRecordHandler(c *gin.Context) {
	req := c.Param("id")
	recordID, err := strconv.Atoi(req)
	if err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	err = service.DeleteRecord(recordID)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorRecordNotExist):
			response.Error(c, response.CodeRecordNotExist)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
	}
	response.Success(c, nil)
}
