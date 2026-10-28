package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/request"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"backend/internal/testutil"
	"backend/internal/utils/jwt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	drivermysql "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestLoginLogsSystemFailureAndExpectedRejection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name      string
		queryErr  error
		wantCode  response.Code
		wantLevel zapcore.Level
	}{
		{
			name:      "database failure is redacted",
			queryErr:  &drivermysql.MySQLError{Number: 1406, Message: "sensitive database value"},
			wantCode:  response.CodeServerBusy,
			wantLevel: zapcore.ErrorLevel,
		},
		{
			name:      "unknown account is an expected rejection",
			queryErr:  gorm.ErrRecordNotFound,
			wantCode:  response.CodeInvalidPassword,
			wantLevel: zapcore.InfoLevel,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 查询回调直接返回错误，不连接数据库，也不执行 SQL。
			db, err := gorm.Open(gormmysql.New(gormmysql.Config{
				DSN:                       "tcp(127.0.0.1:1)/logging_test",
				SkipInitializeWithVersion: true,
			}), &gorm.Config{DisableAutomaticPing: true, Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			connection, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = connection.Close() })
			if err := db.Callback().Query().Before("gorm:query").Register("test:query_error", func(tx *gorm.DB) {
				tx.AddError(tc.queryErr)
			}); err != nil {
				t.Fatal(err)
			}
			core, logs := observer.New(zapcore.DebugLevel)
			log := zap.New(core)
			handler := NewUserHandler(appservice.NewUserService(mysql.NewDAO(db, log), jwt.NewManager("test-secret", 1), config.StorageConfig{}, log), log, testutil.NewRateLimiter(t))
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Set(request.CtxRequestID, "login-test-request")
				c.Next()
			})
			engine.POST("/login", handler.LoginHandler)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"student_id":"202610260001","password":"private-password"}`))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)
			var result response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Code != tc.wantCode {
				t.Fatalf("response code=%v, want %v", result.Code, tc.wantCode)
			}
			entries := logs.All()
			if len(entries) != 1 || entries[0].Level != tc.wantLevel {
				t.Fatalf("logs=%+v", entries)
			}
			fields := entries[0].ContextMap()
			if fields["operation"] != "app.login" || fields["request_id"] != "login-test-request" {
				t.Fatalf("fields=%v", fields)
			}
			operationFields := 0
			for _, field := range entries[0].Context {
				if field.Key == "operation" {
					operationFields++
				}
			}
			if operationFields != 1 {
				t.Fatalf("operation must be attached once, got %d", operationFields)
			}
			encoded, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			for _, sensitive := range []string{"202610260001", "private-password", "sensitive database value"} {
				if strings.Contains(string(encoded), sensitive) {
					t.Fatalf("log contains sensitive input: %s", encoded)
				}
			}
		})
	}
}
