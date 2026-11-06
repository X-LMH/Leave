package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/dao/mysql"
	redisdao "backend/internal/dao/redis"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/session"
	"backend/internal/testutil/sessiontest"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type authContextKey struct{}

func TestSessionAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       "tcp(127.0.0.1:1)/isolated_auth_test",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Error(err)
		}
	})
	user := models.User{
		ID:          1,
		StudentID:   "admin",
		Role:        models.RoleAdmin,
		Status:      models.UserStatusActive,
		AuthVersion: 1,
	}
	var queryErr error
	if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
		if tx.Statement.Context.Value(authContextKey{}) != "auth-test" {
			t.Error("authentication query lost request context")
		}
		if queryErr != nil {
			tx.AddError(queryErr)
			return
		}
		*tx.Statement.Dest.(*models.User) = user
		tx.RowsAffected = 1
	}); err != nil {
		t.Fatal(err)
	}
	r := sessiontest.NewSessionRedis()
	s := session.NewStore(redisdao.NewDAO(r, "default"), config.SessionConfig{})
	ctx := context.WithValue(context.Background(), authContextKey{}, "auth-test")
	token, err := s.Create(ctx, &user, session.ClientApp)
	if err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	engine.GET("/admin", SessionAuthMiddleware(s, mysql.NewDAO(db, zap.NewNop()), zap.NewNop()), AdminAuthMiddleware(), func(c *gin.Context) {
		id, err := request.GetCurrentStuID(c)
		if err != nil {
			t.Fatal(err)
		}
		if id != user.StudentID {
			t.Error("incorrect identity context")
		}
		response.Success(c, nil)
	})
	call := func(credential string, expected response.Code) {
		t.Helper()
		req := httptest.NewRequest("GET", "/admin", nil).WithContext(ctx)
		if credential != "" {
			req.Header.Set("Authorization", "Bearer "+credential)
		}
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		var body response.Response
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Code != expected || w.Code != expected.HTTPStatus() {
			t.Fatalf("got HTTP %d code %d, want %d", w.Code, body.Code, expected)
		}
	}
	call("", response.CodeNeedLogin)
	call(token, response.CodeSuccess) // An admin App session can access management APIs.
	// 固定的旧 JWT 样例，仅验证会话认证拒绝旧格式，不再依赖 JWT 签发实现。
	const legacy = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdHVkZW50X2lkIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjQxMDI0NDQ4MDB9.BVnNuiNNK6y7MOXEWcPynVITgpJZItJRmoA6KNImAMw"
	call(legacy, response.CodeInvalidToken)
	user.Status = models.UserStatusDisabled
	user.AuthVersion++
	call(token, response.CodeInvalidToken)
	user.Status = models.UserStatusActive
	call(token, response.CodeInvalidToken) // Re-enable cannot resurrect old sessions.
	oldSnapshot := user
	user.AuthVersion++ // Password change wins after password verification, before session write.
	stale, err := s.Create(ctx, &oldSnapshot, session.ClientApp)
	if err != nil {
		t.Fatal(err)
	}
	call(stale, response.CodeInvalidToken)
	current, err := s.Create(ctx, &user, session.ClientWeb)
	if err != nil {
		t.Fatal(err)
	}
	call(current, response.CodeSuccess)
	user.Role = models.RoleStudent
	call(current, response.CodeInvalidToken)
	student, err := s.Create(ctx, &user, session.ClientApp)
	if err != nil {
		t.Fatal(err)
	}
	call(student, response.CodeNeedLogin)
	user.Role = models.RoleAdmin
	queryErr = errors.New("database unavailable")
	call(current, response.CodeAuthUnavailable)
	queryErr = gorm.ErrRecordNotFound
	call(current, response.CodeInvalidToken)
	queryErr = nil
	r.Err = errors.New("Redis unavailable")
	call(current, response.CodeAuthUnavailable)
	r.Err = nil
	call(current, response.CodeSuccess)
	if err := s.Revoke(ctx, current); err != nil {
		t.Fatal(err)
	}
	call(current, response.CodeInvalidToken)
}
