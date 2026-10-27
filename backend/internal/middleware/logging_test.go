package middleware

import (
	requestlog "backend/internal/logger"
	"backend/internal/response"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestGinAccessLoggingCorrelationAndPrivacy(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	log := zap.New(core)
	router := gin.New()
	var access bytes.Buffer
	previousWriter := gin.DefaultWriter
	gin.DefaultWriter = &access
	t.Cleanup(func() { gin.DefaultWriter = previousWriter })
	router.Use(RequestID(), AccessLogger(), Recovery(log), CORSMiddleware())
	router.POST("/items/:id", func(c *gin.Context) {
		log.With(requestlog.RequestFields(c)...).Info("业务操作完成")
		response.Success(c, gin.H{"ok": true})
	})
	var previousID string
	for _, path := range []string{"/items/private-id?token=query-secret", "/unknown-private-id?token=query-secret"} {
		entries.TakeAll()
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader("body-secret"))
		request.Header.Set("Authorization", "Bearer header-secret")
		request.Header.Set("X-Request-ID", "client-secret")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		id := recorder.Header().Get("X-Request-ID")
		if len(id) < hex.EncodedLen(8) || id == previousID || id == "client-secret" {
			t.Fatalf("request ID should be server-generated and unique: %q", id)
		}
		previousID = id
		output := access.String()
		if strings.Count(output, "[GIN]") != 1 || !strings.Contains(output, "request_id="+id) || !strings.Contains(output, fmt.Sprint(recorder.Code)) {
			t.Fatalf("unexpected Gin access log: %s", output)
		}
		assertNoSecrets(t, output)
		if entries.FilterMessage("HTTP 请求完成").Len() != 0 {
			t.Fatal("Zap must not duplicate access logs")
		}
		for _, entry := range entries.All() {
			if entry.ContextMap()["request_id"] != id {
				t.Fatalf("business and access logs must share ID: %#v", entry)
			}
			assertNoSecrets(t, fmt.Sprint(entry.ContextMap()))
		}
		if path[1] == 'i' && !strings.Contains(output, "/items/:id") {
			t.Fatalf("want route template: %s", output)
		}
		if recorder.Code == http.StatusOK {
			var body response.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.RequestID != id {
				t.Fatalf("JSON response must share request ID: %s, err=%v", recorder.Body.String(), err)
			}
		}
		access.Reset()
		if !strings.Contains(recorder.Header().Get("Access-Control-Expose-Headers"), "X-Request-ID") {
			t.Fatal("request ID header must be accessible to web clients")
		}
	}
}

func TestRecoveryDoesNotAppendJSONToStartedResponse(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	log := zap.New(core)
	router := gin.New()
	router.Use(RequestID(), Recovery(log))
	router.GET("/stream", func(c *gin.Context) {
		c.String(http.StatusOK, "partial content")
		panic("panic-secret")
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/stream", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "partial content" || entries.FilterLevelExact(zapcore.ErrorLevel).Len() != 1 {
		t.Fatalf("started response must stay intact: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestRecoveryLogsStackWithoutPanicValue(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	log := zap.New(core)
	router := gin.New()
	router.Use(RequestID(), Recovery(log))
	router.GET("/panic", func(c *gin.Context) { panic("panic-secret") })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic?token=query-secret", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", recorder.Code)
	}
	panics := entries.FilterLevelExact(zapcore.ErrorLevel).All()
	if len(panics) != 1 || panics[0].ContextMap()["stack"] == "" || panics[0].ContextMap()["stack"] == nil {
		t.Fatalf("want one panic with stack: %#v", panics)
	}
	var body response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.RequestID == "" || body.RequestID != recorder.Header().Get("X-Request-ID") {
		t.Fatalf("panic response must carry request ID: %s, err=%v", recorder.Body.String(), err)
	}
	if panics[0].ContextMap()["request_id"] != body.RequestID {
		t.Fatal("panic log and response must share request ID")
	}
	for _, entry := range entries.All() {
		assertNoSecrets(t, fmt.Sprint(entry.ContextMap()))
	}
}

func TestRecoveryDoesNotWriteToBrokenConnection(t *testing.T) {
	for _, cause := range []error{syscall.EPIPE, syscall.ECONNRESET, http.ErrAbortHandler} {
		t.Run(cause.Error(), func(t *testing.T) {
			core, entries := observer.New(zapcore.DebugLevel)
			log := zap.New(core)
			router := gin.New()
			router.Use(RequestID(), Recovery(log))
			router.GET("/broken", func(c *gin.Context) {
				panic(fmt.Errorf("header-secret: %w", cause))
			})
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/broken", nil))
			if recorder.Code != http.StatusOK || recorder.Body.Len() != 0 || len(entries.FilterLevelExact(zapcore.WarnLevel).All()) != 1 || len(entries.FilterLevelExact(zapcore.ErrorLevel).All()) != 0 {
				t.Fatalf("broken connection should abort without 500 response: %#v", entries.All())
			}
			for _, entry := range entries.All() {
				assertNoSecrets(t, fmt.Sprint(entry.ContextMap()))
			}
		})
	}
}

func assertNoSecrets(t *testing.T, output string) {
	t.Helper()
	for _, secret := range []string{"private-id", "query-secret", "header-secret", "body-secret", "client-secret", "panic-secret"} {
		if strings.Contains(output, secret) {
			t.Fatalf("log contains sensitive data %q: %s", secret, output)
		}
	}
}
