package controller

import (
	"backend/internal/dao/mysql"
	"backend/internal/models"
	"backend/internal/response"
	"backend/internal/service"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProfileHandler 完善个人信息
func ProfileHandler(c *gin.Context) {
	// 参数绑定
	p := new(models.ParamProfile)
	if err := c.ShouldBindJSON(p); err != nil {
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
		switch {
		case errors.Is(err, mysql.ErrorFinishProfile):
			response.Error(c, response.CodeFinishData)
			return
		case errors.Is(err, mysql.ErrorUserExist):
			response.Error(c, response.CodeUserExist)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
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
		switch {
		case errors.Is(err, mysql.ErrorUserNotExist):
			response.Error(c, response.CodeNeedLogin)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
	}
	response.Success(c, data)
}

// RecordHandler 记录用户行为
func RecordHandler(c *gin.Context) {
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
