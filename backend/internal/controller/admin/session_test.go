package admin_test

import (
	"backend/internal/config"
	admincontroller "backend/internal/controller/admin"
	appcontroller "backend/internal/controller/app"
	redisdao "backend/internal/dao/redis"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	appservice "backend/internal/service/app"
	"backend/internal/session"
	"backend/internal/testutil/sessiontest"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestLogoutEndpointsAndIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := sessiontest.NewSessionRedis()
	s := session.NewStore(redisdao.NewDAO(r, "default"), config.SessionConfig{})
	admin := admincontroller.NewAuthHandler(adminservice.NewAuthService(nil, s, zap.NewNop()), zap.NewNop(), nil)
	app := appcontroller.NewUserHandler(appservice.NewUserService(nil, s, config.StorageConfig{}, zap.NewNop()), zap.NewNop(), nil)
	engine := gin.New()
	engine.POST("/admin/auth/logout", admin.LogoutHandler)
	engine.POST("/logout", app.LogoutHandler)
	engine.GET("/missing-id", admin.MeHandler)
	engine.GET("/missing-role", func(c *gin.Context) {
		c.Set(request.CtxStuID, "admin")
	}, admin.MeHandler)
	engine.GET("/student", func(c *gin.Context) {
		c.Set(request.CtxStuID, "student")
		c.Set(request.CtxRole, models.RoleStudent)
	}, admin.MeHandler)
	engine.GET("/admin/auth/me", func(c *gin.Context) {
		c.Set(request.CtxStuID, "admin")
		c.Set(request.CtxRole, models.RoleAdmin)
		c.Next()
	}, admin.MeHandler)
	call := func(method, path, token string, code response.Code) response.Response {
		t.Helper()
		req := httptest.NewRequest(method, path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		var body response.Response
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if w.Code != code.HTTPStatus() || body.Code != code {
			t.Fatalf("HTTP %d, code %d, want %d", w.Code, body.Code, code)
		}
		return body
	}
	for _, path := range []string{"/admin/auth/logout", "/logout"} {
		token, err := s.Create(context.Background(), &models.User{
			ID:          1,
			StudentID:   "admin",
			Role:        models.RoleAdmin,
			AuthVersion: 1,
		}, session.ClientApp)
		if err != nil {
			t.Fatal(err)
		}
		r.Err = errors.New("isolated Redis unavailable")
		call("POST", path, token, response.CodeAuthUnavailable)
		r.Err = nil
		if _, err := s.Get(context.Background(), token); err != nil {
			t.Fatal("failed logout removed session")
		}
		call("POST", path, token, response.CodeSuccess)
		call("POST", path, token, response.CodeSuccess)
		if _, err := s.Get(context.Background(), token); !errors.Is(err, session.ErrInvalid) {
			t.Fatal("logout did not revoke token")
		}
		call("POST", path, "", response.CodeInvalidToken)
	}
	call("GET", "/missing-id", "", response.CodeNeedLogin)
	call("GET", "/missing-role", "", response.CodeNeedLogin)
	call("GET", "/student", "", response.CodeNeedLogin)
	identity := call("GET", "/admin/auth/me", "", response.CodeSuccess)
	data := identity.Data.(map[string]interface{})
	if len(data) != 2 || data["student_id"] != "admin" || data["role"] != models.RoleAdmin {
		t.Fatal("identity contract", data)
	}
}
