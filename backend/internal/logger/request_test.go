package logger

import (
	"backend/internal/request"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestOperationFieldsKeepRequestAndIdentityIsolated(t *testing.T) {
	core, entries := observer.New(zap.InfoLevel)
	base := zap.New(core)
	router := gin.New()
	router.GET("/items/:id", func(c *gin.Context) {
		c.Set(request.CtxRequestID, c.Param("id"))
		c.Set(request.CtxStuID, "authenticated-student")
		log := base.With(OperationFields(c, "item.read")...)
		if c.Param("id") == "first" {
			log = log.With(zap.String("actor_id", "explicit-actor"))
		}
		log.Info("操作完成")
	})
	for _, id := range []string{"first", "second"} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items/"+id+"?token=private-token", nil))
	}
	base.Info("后台任务")
	logs := entries.All()
	for i, id := range []string{"first", "second"} {
		fields := logs[i].ContextMap()
		if fields["request_id"] != id || fields["operation"] != "item.read" || fields["route"] != "/items/:id" {
			t.Fatalf("unexpected request fields: %v", fields)
		}
		for _, field := range logs[i].Context {
			if strings.Contains(field.String, "private-token") || strings.Contains(field.String, "authenticated-student") {
				t.Fatalf("unexpected automatic identity or query: %v", field)
			}
		}
	}
	if logs[0].ContextMap()["actor_id"] != "explicit-actor" || logs[1].ContextMap()["actor_id"] != nil || len(logs[2].Context) != 0 {
		t.Fatal("child logger fields must not leak into another request or the injected logger")
	}
}
