package admin

import (
	logutil "backend/internal/logger"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/request"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/gorm"
)

func TestManagementErrorLogsOnlyUnexpectedFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name   string
		err    error
		status int
		logs   int
	}{
		{"invalid input", adminservice.ErrorInvalidAdminInput, http.StatusBadRequest, 0},
		{"not found", gorm.ErrRecordNotFound, http.StatusNotFound, 0},
		{"duplicate", gorm.ErrDuplicatedKey, http.StatusConflict, 0},
		{"unexpected", errors.New("database unavailable"), http.StatusInternalServerError, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			core, entries := observer.New(zap.InfoLevel)
			r := gin.New()
			r.GET("/items/:id", func(c *gin.Context) {
				c.Set(request.CtxStuID, "123456789012")
				c.Set(request.CtxRequestID, "test-request")
				log := zap.New(core).With(logutil.OperationFields(c, "admin.test.read")...)
				if actorID, err := request.GetCurrentStuID(c); err == nil {
					log = log.With(zap.String("actor_id", actorID))
				}
				managementError(c, tt.err, log)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/items/3?password=secret", nil))
			if w.Code != tt.status || entries.Len() != tt.logs {
				t.Fatalf("status=%d logs=%d; want %d %d", w.Code, entries.Len(), tt.status, tt.logs)
			}
			if tt.logs == 0 {
				return
			}
			fields := entries.All()[0].ContextMap()
			if fields["request_id"] != "test-request" || fields["operation"] != "admin.test.read" || fields["route"] != "/items/:id" {
				t.Fatalf("missing request context: %v", fields)
			}
			if fields["actor_id"] != "123456789012" || fields["error"] != "database unavailable" {
				t.Fatalf("unexpected diagnostic fields: %v", fields)
			}
			for _, value := range fields {
				if value == "secret" {
					t.Fatalf("sensitive input included: %v", fields)
				}
			}
		})
	}
}
