package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/response"
	"backend/internal/service"
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

	data, err := service.AdminLogin(p)
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
