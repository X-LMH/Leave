package middleware

import (
	"backend/internal/dao/mysql"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/utils/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware jwt认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求头中的Authorization
		authHeader := c.GetHeader("Authorization")

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
		claims, err := jwt.ParseToken(parts[1]) // parts[1]是提取出的令牌字符串
		if err != nil {
			response.Error(c, response.CodeInvalidToken)
			c.Abort()
			return
		}
		if err := mysql.EnsureUserActive(claims.StudentID); err != nil {
			response.Error(c, response.CodeNeedLogin)
			c.Abort()
			return
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
