package controller

import (
	"backend/dao/mysql"
	"backend/models"
	"backend/service"
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
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 业务处理
	studentID, err := GetCurrentStuID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	if err := service.Profile(p, studentID); err != nil {
		if errors.Is(err, mysql.ErrorFinishProfile) {
			ResponseError(c, CodeFinishData)
			return
		}
		if errors.Is(err, mysql.ErrorUserExist) {
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, nil)
}

func GetProfileHandler(c *gin.Context) {
	studentID, err := GetCurrentStuID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	data, err := service.GetProfile(studentID)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeNeedLogin)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// RecordHandler 记录用户行为
func RecordHandler(c *gin.Context) {
	p := new(models.ParamRecord)

	if err := c.ShouldBindJSON(p); err != nil {
		fmt.Println(p)
		ResponseError(c, CodeInvalidParam)
		return
	}

	studentID, err := GetCurrentStuID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}

	err = service.CreateRecord(studentID, p)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeNeedLogin)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

func GetRecordHandler(c *gin.Context) {
	req := c.Param("id")
	recordID, err := strconv.Atoi(req)
	if err != nil {
		fmt.Println(req)
		ResponseError(c, CodeInvalidParam)
		return
	}

	record, err := service.GetRecord(recordID)
	if err != nil {
		if errors.Is(err, mysql.ErrorRecordNotExist) {
			ResponseError(c, CodeRecordNotExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, record)
}

func GetRecordsLIstHandler(c *gin.Context) {
	studentID, err := GetCurrentStuID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
	}

	res, err := service.GetRecordsList(studentID)
	if err != nil {
		ResponseError(c, CodeServerBusy)
	}
	ResponseSuccess(c, res)
}

func DeleteRecordHandler(c *gin.Context) {
	req := c.Param("id")
	recordID, err := strconv.Atoi(req)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	err = service.DeleteRecord(recordID)
	if err != nil {
		if errors.Is(err, mysql.ErrorRecordNotExist) {
			ResponseError(c, CodeRecordNotExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}
