package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/middleware"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	limiter *middleware.RateLimiter
	logger  *zap.Logger
	auth    *adminservice.AuthService
}

func NewAuthHandler(auth *adminservice.AuthService, logger *zap.Logger, limiter *middleware.RateLimiter) *AuthHandler {
	return &AuthHandler{
		limiter: limiter,
		logger:  logger,
		auth:    auth,
	}
}

func (a *AuthHandler) LoginHandler(c *gin.Context) {
	log := a.logger.With(logutil.OperationFields(c, "admin.auth.login")...)

	p := new(dto.AdminLoginRequest)
	if err := c.ShouldBindJSON(p); err != nil || strings.TrimSpace(p.UserName) == "" || p.Password == "" {
		response.Error(c, response.CodeInvalidParam)
		return
	}

	if !a.limiter.Check(c, middleware.SideAdmin, middleware.OperationLogin, middleware.DimensionAccount, strings.TrimSpace(p.UserName)) {
		return
	}
	data, err := a.auth.AdminLogin(p)
	if err != nil {
		switch {
		case errors.Is(err, mysql.ErrorUserNotExist), errors.Is(err, mysql.ErrorInvalidPassword):
			log.Info("管理员登录被拒绝", zap.String("reason", "invalid_credentials"))
			response.Error(c, response.CodeInvalidPassword)
		default:
			log.Error("账号操作失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	log.Info("管理员登录成功", zap.String("actor_id", strings.TrimSpace(p.UserName)))
	response.Success(c, data)
}

func (a *AuthHandler) ChangePasswordHandler(c *gin.Context) {
	log := a.logger.With(logutil.OperationFields(c, "admin.auth.change_password")...)

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
	log = log.With(zap.String("actor_id", studentID))
	if err := a.auth.ChangePassword(studentID, input); err != nil {
		switch {
		case errors.Is(err, adminservice.ErrorInvalidAdminInput):
			response.Error(c, response.CodeInvalidParam)
		case errors.Is(err, mysql.ErrorNotRightPassword):
			response.Error(c, response.CodeNotRightPassword)
		case errors.Is(err, mysql.ErrorUserNotExist), errors.Is(err, mysql.ErrorUserDisabled):
			response.Error(c, response.CodeNeedLogin)
		default:
			log.Error("账号操作失败", logutil.Error(err))
			response.Error(c, response.CodeServerBusy)
		}
		return
	}
	log.Info("修改密码成功")
	response.Success(c, nil)
}
