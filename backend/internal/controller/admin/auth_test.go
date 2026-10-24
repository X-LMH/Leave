package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"backend/internal/utils/jwt"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestChangePasswordValidationAndRole(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil, zap.NewNop())
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name string
		role string
		body string
		code response.Code
	}{
		{"malformed JSON", models.RoleAdmin, `{`, response.CodeInvalidParam},
		{"empty fields", models.RoleAdmin, `{}`, response.CodeInvalidParam},
		{"short password", models.RoleAdmin, `{"password":"old","new_password":"12345","re_password":"12345"}`, response.CodeInvalidParam},
		{"mismatch", models.RoleAdmin, `{"password":"old","new_password":"123456","re_password":"654321"}`, response.CodeInvalidParam},
		{"student denied", models.RoleStudent, `{"password":"old","new_password":"123456","re_password":"123456"}`, response.CodeNeedLogin},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			// Set the context normally populated by JWT middleware; no database needed.
			r.Use(func(c *gin.Context) {
				c.Set(request.CtxStuID, "test-admin")
				c.Set(request.CtxRole, tc.role)
			})
			r.POST("/password", middleware.AdminAuthMiddleware(), NewAuthHandler(adminservice.NewAuthService(dao, jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays), zap.NewNop()), zap.NewNop()).ChangePasswordHandler)
			req := httptest.NewRequest("POST", "/password", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			var result response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Code != tc.code || w.Code != tc.code.HTTPStatus() || result.Data != nil {
				t.Fatalf("unexpected response: HTTP %d, code %d", w.Code, result.Code)
			}
		})
	}
}
