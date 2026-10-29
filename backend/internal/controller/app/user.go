package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/middleware"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	limiter *middleware.RateLimiter
	logger  *zap.Logger
	user    *appservice.UserService
}

func NewUserHandler(user *appservice.UserService, logger *zap.Logger, limiter *middleware.RateLimiter) *UserHandler {
	return &UserHandler{
		limiter: limiter,
		logger:  logger,
		user:    user,
	}
}

// RegisterHandler 用户注册
func (u *UserHandler) RegisterHandler(c *gin.Context) {
	log := u.logger.With(logutil.OperationFields(c, "app.register")...)

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
		strings.TrimSpace(p.Password) == "" || p.Password != p.RePassword ||
		utf8.RuneCountInString(strings.TrimSpace(p.DeviceName)) > 255 ||
		len(p.AppVersion) > 64 {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	if !u.limiter.Check(c, middleware.SideApp, middleware.OperationRegister, middleware.DimensionAccount, p.StudentID) {
		return
	}
	// 业务逻辑
	data, err := u.user.Register(p)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorUserExist):
			response.Error(c, response.CodeUserExist)
			return
		default:
			log.Error("注册失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
			return
		}
	}
	// 返回响应
	log.Info("注册成功",
		zap.String("actor_id", strings.TrimSpace(p.StudentID)),
	)
	response.Success(c, data)
}

// LoginHandler 用户登录
func (u *UserHandler) LoginHandler(c *gin.Context) {
	log := u.logger.With(logutil.OperationFields(c, "app.login")...)

	// 参数绑定
	p := new(dto.LoginRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		response.Error(c, response.CodeInvalidParam)
		return
	}
	if !validLoginAccount(p.StudentID) ||
		len(p.Password) == 0 ||
		utf8.RuneCountInString(strings.TrimSpace(p.DeviceName)) > 255 ||
		len(p.AppVersion) > 64 {

		// 参数错误
		response.Error(c, response.CodeInvalidParam)
		return
	}

	if !u.limiter.Check(c, middleware.SideApp, middleware.OperationLogin, middleware.DimensionAccount, p.StudentID) {
		return
	}
	// 业务逻辑
	data, err := u.user.Login(p)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorUserNotExist), errors.Is(err, mysql.ErrorInvalidPassword):
			// 不暴露账号是否存在，避免被用于枚举注册用户。
			log.Info("登录被拒绝",
				zap.String("reason", "invalid_credentials"),
			)
			response.Error(c, response.CodeInvalidPassword)
			return
		default:
			log.Error("登录失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
			return
		}
	}

	log.Info("登录成功",
		zap.String("actor_id", strings.TrimSpace(p.StudentID)),
	)
	response.Success(c, data)
}

// LogoutHandler 退出登录。
// 当前使用无状态 JWT，客户端在收到成功响应后清除本地 Token。
func (u *UserHandler) LogoutHandler(c *gin.Context) {
	response.Success(c, nil)
}

func (u *UserHandler) UpdateAppInfoHandler(c *gin.Context) {
	log := u.logger.With(logutil.OperationFields(c, "app.client_info.update")...)

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
	if err := u.user.UpdateAppInfo(p, studentID); err != nil {
		log.Error("更新客户端信息失败", logutil.Error(err))
		response.Error(c, response.CodeServerBusy)
		return
	}
	response.Success(c, nil)
}

// ChangePasswordHandler 修改密码
func (u *UserHandler) ChangePasswordHandler(c *gin.Context) {
	log := u.logger.With(logutil.OperationFields(c, "app.password.change")...)

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
	if err := u.user.ChangePassword(p, studentID); err != nil {
		switch {
		case errors.Is(err, mysql.ErrorNotRightPassword):
			response.Error(c, response.CodeNotRightPassword)
			return
		default:
			log.Error("修改密码失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
			return
		}
	}
	log.Info("修改密码成功",
		zap.String("actor_id", studentID),
	)
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

// 登录兼容管理员账号；注册仍只接受十二位学号。
func validLoginAccount(account string) bool {
	length := utf8.RuneCountInString(strings.TrimSpace(account))
	return length > 0 && length <= 22
}
