package middleware

import (
	"backend/internal/dao/mysql"
	requestlog "backend/internal/logger"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/session"
	"backend/internal/utils/jwt"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JWTAuthMiddleware jwt认证中间件
func JWTAuthMiddleware(tokens *jwt.Manager, dao *mysql.DAO, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求头中的Authorization
		authHeader := c.GetHeader(session.AuthorizationHeader)

		// 不含有请求头
		if authHeader == "" {
			response.Error(c, response.CodeNeedLogin)
			c.Abort() // 终止请求链，不再执行后续处理
			return
		}

		// 分析请求头
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			// 格式错误，返回401
			response.Error(c, response.CodeInvalidToken)
			c.Abort()
			return
		}

		// 解析JWT令牌
		claims, err := tokens.ParseToken(parts[1]) // parts[1]是提取出的令牌字符串
		if err != nil {
			response.Error(c, response.CodeInvalidToken)
			c.Abort()
			return
		}
		if err := dao.EnsureUserActive(claims.StudentID); err != nil {
			if !errors.Is(err, mysql.ErrorUserNotExist) && !errors.Is(err, mysql.ErrorUserDisabled) {
				logger.With(requestlog.RequestFields(c)...).Error("查询认证用户状态失败",
					zap.String("operation", "auth.ensure_user_active"),
					requestlog.Error(err),
				)
			}
			response.Error(c, response.CodeNeedLogin)
			c.Abort()
			return
		}
		if claims.Role == models.RoleAdmin {
			// 角色可能在令牌签发后变更，管理员声明须与数据库当前角色一致。
			user, err := dao.FindUserByStudentID(claims.StudentID)
			if err != nil && !errors.Is(err, mysql.ErrorUserNotExist) {
				logger.With(requestlog.RequestFields(c)...).Error("查询管理员身份失败",
					zap.String("operation", "auth.find_admin"),
					requestlog.Error(err),
				)
			}
			if err != nil || user.Role != models.RoleAdmin || user.Status != models.UserStatusActive {
				response.Error(c, response.CodeNeedLogin)
				c.Abort()
				return
			}
		}

		// 令牌验证通过，将用户信息存入上下文
		c.Set(request.CtxStuID, claims.StudentID)
		c.Set(request.CtxRole, claims.Role)

		// 继续执行后续的处理函数（如业务逻辑）
		c.Next()
	}
}

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
