package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/service"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// RegisterHandler 用户注册
func RegisterHandler(c *gin.Context) {
	// 参数绑定
	//response.Error(c, response.CodeServiceFix)
	//return

	p := new(dto.RegisterRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	if !validStudentID(p.StudentID) ||
		len(p.Password) < 6 || len(p.Password) > 30 || len(p.RePassword) == 0 ||
		strings.TrimSpace(p.Password) == "" || p.Password != p.RePassword {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	// 业务逻辑
	if err := service.Register(p); err != nil {
		switch {
		case errors.Is(err, mysql.ErrorUserExist):
			response.Error(c, response.CodeUserExist)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
	}
	// 返回响应
	response.Success(c, nil)
}

// LoginHandler 用户登录
func LoginHandler(c *gin.Context) {
	// 参数绑定
	p := new(dto.LoginRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	if !validStudentID(p.StudentID) ||
		len(p.Password) == 0 ||
		utf8.RuneCountInString(strings.TrimSpace(p.DeviceName)) > 255 ||
		len(p.AppVersion) > 64 {

		// 参数错误
		response.Error(c, response.CodeInvalidParam)
		return
	}

	// 业务逻辑
	data, err := service.Login(p)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorUserNotExist), errors.Is(err, mysql.ErrorInvalidPassword):
			// 不暴露账号是否存在，避免被用于枚举注册用户。
			response.Error(c, response.CodeInvalidPassword)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
	}

	response.Success(c, data)
}

// LogoutHandler 退出登录。
// 当前使用无状态 JWT，客户端在收到成功响应后清除本地 Token。
func LogoutHandler(c *gin.Context) {
	response.Success(c, nil)
}

func UpdateAppInfoHandler(c *gin.Context) {
	p := new(dto.ClientInfoRequest)
	if err := c.ShouldBindJSON(p); err != nil || strings.TrimSpace(p.AppVersion) == "" || utf8.RuneCountInString(strings.TrimSpace(p.DeviceName)) > 255 {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	if err := service.UpdateAppInfo(p, studentID); err != nil {
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, nil)
}

// ChangePasswordHandler 修改密码
func ChangePasswordHandler(c *gin.Context) {
	// 参数校验
	p := new(dto.PasswordChangeRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	if len(p.Password) == 0 || len(p.NewPassword) < 6 || len(p.NewPassword) > 30 || len(p.RePassword) == 0 || strings.TrimSpace(p.NewPassword) == "" || p.NewPassword != p.RePassword {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	// 业务处理
	studentID, err := request.GetCurrentStuID(c)
	if err != nil {
		response.Error(c, response.CodeNeedLogin)
		return
	}
	if err := service.ChangePassword(p, studentID); err != nil {
		switch {
		case errors.Is(err, mysql.ErrorNotRightPassword):
			response.Error(c, response.CodeNotRightPassword)
			return
		default:
			response.Error(c, response.CodeServerBusy)
			return
		}
	}
	response.Success(c, nil)
}

// validStudentID 校验学号是否为 12 位数字。
func validStudentID(studentID string) bool {
	studentID = strings.TrimSpace(studentID)
	if len(studentID) != 12 {
		return false
	}
	for _, r := range studentID {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
