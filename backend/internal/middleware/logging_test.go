package middleware

import (
	"backend/internal/config"
	requestlog "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestAccessLoggingCorrelationAndPrivacy(t *testing.T) {
	core, entries := observer.New(zapcore.DebugLevel)
	log := zap.New(core)
	router := gin.New()
	var access bytes.Buffer
	previousWriter := gin.DefaultWriter
	gin.DefaultWriter = &access
	t.Cleanup(func() { gin.DefaultWriter = previousWriter })
	router.Use(RequestID(), AccessLogger(log), Recovery(log), CORSMiddleware())
	router.POST("/items/:id", func(c *gin.Context) {
		log.With(requestlog.RequestFields(c)...).Info("业务操作完成")
		response.Success(c, gin.H{"ok": true})
	})
	router.GET("/panic", func(c *gin.Context) { panic("panic-secret") })
	var previousID string
	for _, tc := range []struct {
		method string
		path   string
		route  string
		status int
	}{
		{http.MethodPost, "/items/private-id?token=query-secret", "/items/:id", 200},
		{http.MethodGet, "/unknown-private-id?token=query-secret", "unmatched", 404},
		{http.MethodGet, "/panic?token=query-secret", "/panic", 500},
		{http.MethodOptions, "/items/private-id", "unmatched", 204},
	} {
		t.Run(tc.method+tc.route, func(t *testing.T) {
			entries.TakeAll()
			access.Reset()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("body-secret"))
			req.RemoteAddr = "192.0.2.1:1234"
			req.Header.Set("Authorization", "Bearer header-secret")
			req.Header.Set("X-Request-ID", "client-secret")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			id := w.Header().Get("X-Request-ID")
			if len(id) < hex.EncodedLen(8) || id == previousID || id == "client-secret" {
				t.Fatalf("invalid request ID: %q", id)
			}
			previousID = id
			logs := entries.FilterMessage("HTTP 请求完成").All()
			if w.Code != tc.status || len(logs) != 1 {
				t.Fatalf("status=%d access logs=%v", w.Code, logs)
			}
			fields := logs[0].ContextMap()
			if fields["route"] != tc.route || fields["method"] != tc.method || fields["status"] != int64(tc.status) || fields["client_ip"] != "192.0.2.1" || fields["latency"] == nil {
				t.Fatalf("unexpected access fields: %v", fields)
			}
			for _, entry := range entries.All() {
				if entry.ContextMap()["request_id"] != id {
					t.Fatal("request and business logs must share ID")
				}
				assertNoSecrets(t, fmt.Sprint(entry.ContextMap()))
			}
			if strings.Contains(access.String(), "[GIN]") {
				t.Fatal("duplicate Gin access log")
			}
			if !strings.Contains(w.Header().Get("Access-Control-Expose-Headers"), "X-Request-ID") {
				t.Fatal("request ID must be accessible to web clients")
			}
		})
	}
}

func TestAccessLoggingAuthenticatedActor(t *testing.T) {
	core, entries := observer.New(zapcore.InfoLevel)
	log := zap.New(core)
	router := gin.New()
	router.Use(RequestID(), AccessLogger(log), Recovery(log))
	router.GET("/:scenario", func(c *gin.Context) {
		scenario := c.Param("scenario")
		if scenario == "anonymous" {
			c.Status(http.StatusOK)
			return
		}
		if scenario == "unauthorized" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(request.CtxStuID, "202610300001")
		c.Set(request.CtxRole, "student")
		if scenario == "panic" {
			panic("panic-secret")
		}
		if scenario == "forbidden" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Status(http.StatusOK)
	})
	for _, scenario := range []string{"authenticated", "anonymous", "forbidden", "unauthorized", "panic"} {
		t.Run(scenario, func(t *testing.T) {
			entries.TakeAll()
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/"+scenario+"?student_id=untrusted-actor", nil)
			req.Header.Set("Authorization", "Bearer header-secret")
			router.ServeHTTP(recorder, req)
			logs := entries.FilterMessage("HTTP 请求完成").All()
			if len(logs) != 1 {
				t.Fatalf("want one access log, got %d", len(logs))
			}
			fields := logs[0].ContextMap()
			if scenario == "anonymous" || scenario == "unauthorized" {
				if fields["actor_id"] != nil || fields["actor_role"] != nil {
					t.Fatalf("unauthenticated request must not inherit identity: %v", fields)
				}
			} else if fields["actor_id"] != "202610300001" || fields["actor_role"] != "student" {
				t.Fatalf("missing authenticated identity: %v", fields)
			}
			assertNoSecrets(t, fmt.Sprint(fields))
			if strings.Contains(fmt.Sprint(fields), "untrusted-actor") {
				t.Fatal("must not use identity from request parameters")
			}
		})
	}
	log.Info("后台任务")
	if len(entries.FilterMessage("后台任务").All()[0].Context) != 0 {
		t.Fatal("access fields must not modify the shared logger")
	}
}

func TestAccessLoggingWritesDailyFile(t *testing.T) {
	dir := t.TempDir()
	log, err := requestlog.NewLogger(config.LogConfig{
		Level:    "info",
		Encoding: "console",
		Dir:      dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(RequestID(), AccessLogger(log))
	router.GET("/items/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	router.POST("/items/:id", func(c *gin.Context) { c.Status(http.StatusCreated) })
	ids := make(map[string]string)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/items/private-id?token=query-secret", strings.NewReader("body-secret"))
		router.ServeHTTP(w, req)
		ids[method] = w.Header().Get("X-Request-ID")
	}
	// 按实际产生的文件读取，避免测试恰好跨越午夜时依赖单个日期。
	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		assertNoSecrets(t, string(content))
		for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
			var entry map[string]any
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatal(err)
			}
			method, _ := entry["method"].(string)
			if entry["msg"] != "HTTP 请求完成" || entry["route"] != "/items/:id" || ids[method] == "" || entry["request_id"] != ids[method] {
				t.Fatalf("unexpected file entry: %v", entry)
			}
			if _, err := time.ParseDuration(entry["latency"].(string)); err != nil {
				t.Fatal(err)
			}
			delete(ids, method)
			count++
		}
	}
	if count != 2 || len(ids) != 0 {
		t.Fatalf("want GET and POST once each, count=%d remaining=%v", count, ids)
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
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["request_id"] != nil {
		t.Fatalf("panic JSON response must omit request ID: %s, err=%v", recorder.Body.String(), err)
	}
	id := recorder.Header().Get("X-Request-ID")
	if id == "" || panics[0].ContextMap()["request_id"] != id {
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
