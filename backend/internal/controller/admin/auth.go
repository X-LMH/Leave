package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

func LoginHandler(c *gin.Context) {
	p := new(dto.AdminLoginRequest)
	if err := c.ShouldBindJSON(p); err != nil || strings.TrimSpace(p.UserName) == "" || p.Password == "" {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	data, err := adminservice.AdminLogin(p)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorUserNotExist), errors.Is(err, mysql.ErrorInvalidPassword):
			response.Error(c, response.CodeInvalidPassword)
		default:
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	response.Success(c, data)
}

func ChangePasswordHandler(c *gin.Context) {
	input := new(dto.PasswordChangeRequest)
	if err := c.ShouldBindJSON(input); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	if err := adminservice.ChangePassword(studentID, input); err != nil {
		switch {
		case errors.Is(err, adminservice.ErrorInvalidAdminInput):
			response.Error(c, response.CodeInvalidParam)
		case errors.Is(err, mysql.ErrorNotRightPassword):
			response.Error(c, response.CodeNotRightPassword)
		case errors.Is(err, mysql.ErrorUserNotExist), errors.Is(err, mysql.ErrorUserDisabled):
			response.Error(c, response.CodeNeedLogin)
		default:
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	response.Success(c, nil)
}
