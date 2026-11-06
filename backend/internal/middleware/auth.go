package middleware

import (
	"errors"

	"backend/internal/dao/mysql"
	requestlog "backend/internal/logger"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/session"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminAuthMiddleware 仅允许管理员访问受保护的管理端接口。
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := request.GetCurrentRole(c)
		if err != nil || role != models.RoleAdmin {
			response.Error(c, response.CodeNeedLogin)
			c.Abort()
			return
		}
		c.Next()
	}
}

// SessionAuthMiddleware 校验服务端会话，再以数据库当前状态建立身份上下文。
func SessionAuthMiddleware(sessions *session.Store, dao *mysql.DAO, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(session.AuthorizationHeader) == "" {
			response.Error(c, response.CodeNeedLogin)
			c.Abort()
			return
		}
		token, err := session.Bearer(c.GetHeader(session.AuthorizationHeader))
		if err != nil {
			abortSessionAuth(c, err, logger)
			return
		}
		record, err := sessions.Get(c.Request.Context(), token)
		if err != nil {
			abortSessionAuth(c, err, logger)
			return
		}
		user, err := dao.FindAuthUser(c.Request.Context(), record.UserID)
		if errors.Is(err, mysql.ErrorUserNotExist) {
			err = session.ErrInvalid
		}
		if err != nil {
			abortSessionAuth(c, err, logger)
			return
		}
		if user.Status != models.UserStatusActive || user.AuthVersion != record.AuthVersion ||
			user.StudentID != record.StudentID || user.Role != record.Role {
			abortSessionAuth(c, session.ErrInvalid, logger)
			return
		}
		c.Set(request.CtxStuID, user.StudentID)
		c.Set(request.CtxRole, user.Role)
		c.Next()
	}
}

// 依赖故障不能误报登录失效，否则客户端会丢弃仍然有效的会话。
func abortSessionAuth(c *gin.Context, err error, logger *zap.Logger) {
	if errors.Is(err, session.ErrInvalid) {
		response.Error(c, response.CodeInvalidToken)
	} else {
		logger.With(requestlog.RequestFields(c)...).Error("认证依赖不可用",
			zap.String("operation", "auth.session"),
			requestlog.Error(err),
		)
		response.Error(c, response.CodeAuthUnavailable)
	}
	c.Abort()
}
