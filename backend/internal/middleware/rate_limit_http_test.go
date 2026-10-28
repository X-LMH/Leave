package middleware_test

import (
	admincontroller "backend/internal/controller/admin"
	appcontroller "backend/internal/controller/app"
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/router"
	adminservice "backend/internal/service/admin"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Intercept Redis commands to test HTTP wiring; this does not execute Lua.
type evalHook struct {
	run func(context.Context, *redis.Cmd) error
}

func (h evalHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) { return next(ctx, network, addr) }
}
func (h evalHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (h evalHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if cmd.Name() != "eval" {
			return errors.New("unexpected Redis command")
		}
		return h.run(ctx, cmd.(*redis.Cmd))
	}
}
func testLimiter(t *testing.T, run func(context.Context, *redis.Cmd) error, logger *zap.Logger) *middleware.RateLimiter {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1})
	t.Cleanup(func() { _ = client.Close() })
	client.AddHook(evalHook{run: run})
	return middleware.NewRateLimiter(client, logger)
}
func perform(r http.Handler, method, path, body, ip string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = net.JoinHostPort(ip, "1234")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimitHTTPResponsesAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		result    []interface{}
		err       error
		cancelled bool
		timeout   bool
		status    int
		code      int
		retry     string
	}{
		{name: "allowed", result: []interface{}{int64(1), int64(60000)}, status: 204},
		{name: "round up", result: []interface{}{int64(0), int64(1001)}, status: 429, code: 2008, retry: "2"},
		{name: "minimum retry", result: []interface{}{int64(0), int64(0)}, status: 429, code: 2008, retry: "1"},
		{name: "Redis error", err: errors.New("connection lost"), status: 503, code: 9002},
		{name: "timeout", timeout: true, status: 503, code: 9002},
		{name: "empty result", status: 503, code: 9002},
		{name: "invalid decision", result: []interface{}{int64(2), int64(60000)}, status: 503, code: 9002},
		{name: "bad ttl", result: []interface{}{int64(0), int64(-1)}, status: 503, code: 9002},
		{name: "cancelled", cancelled: true, status: 503, code: 9002},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core, logs := observer.New(zap.DebugLevel)
			limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > 500*time.Millisecond {
					t.Error("missing bounded context")
				}
				if tc.timeout {
					<-ctx.Done()
					return ctx.Err()
				}
				if tc.cancelled {
					return ctx.Err()
				}
				cmd.SetVal(tc.result)
				return tc.err
			}, zap.New(core))
			r := gin.New()
			r.Use(middleware.CORSMiddleware(), limiter.IP("app", "login"))
			reached := false
			r.GET("/test", func(c *gin.Context) { reached = true; c.Status(204) })
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.0.2.1:1234"
			if tc.cancelled {
				ctx, cancel := context.WithCancel(req.Context())
				cancel()
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || reached != (tc.status == 204) {
				t.Fatalf("status=%d reached=%v body=%s", w.Code, reached, w.Body)
			}
			if w.Header().Get("Retry-After") != tc.retry {
				t.Fatal(w.Header())
			}
			if !strings.Contains(w.Header().Get("Access-Control-Expose-Headers"), "Retry-After") {
				t.Fatal(w.Header())
			}
			if tc.code != 0 {
				var body struct {
					Code int
					Data struct {
						RetryAfter int `json:"retry_after"`
					}
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Code != tc.code {
					t.Fatal(w.Body)
				}
				if tc.status == 429 && body.Data.RetryAfter < 1 {
					t.Fatal(w.Body)
				}
				if logs.Len() != 1 {
					t.Fatal(logs.All())
				}
				encoded, _ := json.Marshal(logs.All()[0].ContextMap())
				if strings.Contains(string(encoded), "leave:") || strings.Contains(string(encoded), "192.0.2.1") {
					t.Fatal(string(encoded))
				}
			}
		})
	}
}

func TestRouteLimitOrderAndSpecialOperations(t *testing.T) {
	for _, tc := range []struct {
		method, path, side, operation, dimension string
		authFail                                 bool
	}{
		{"GET", "/api/v1/home", "app", "general", "account", false},
		{"POST", "/api/v1/record", "app", "record-create", "account", false},
		{"POST", "/api/v1/password", "app", "password-change", "account", false},
		{"POST", "/api/v1/feedback", "app", "feedback-create", "account", false},
		{"POST", "/api/v1/profile/avatar", "app", "avatar-upload", "account", false},
		{"GET", "/api/v1/admin/records/123", "admin", "general", "account", false},
		{"POST", "/api/v1/admin/auth/password", "admin", "password-change", "account", false},
		{"GET", "/api/v1/home", "app", "general", "account", true},
	} {
		t.Run(fmt.Sprintf("%s%s/%s/unauth=%v", tc.method, tc.path, tc.operation, tc.authFail), func(t *testing.T) {
			var keys []string
			limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
				key := cmd.Args()[3].(string)
				keys = append(keys, key)
				allowed := int64(1)
				if strings.HasPrefix(key, "leave:"+tc.side+":"+tc.operation+":"+tc.dimension+":") {
					allowed = 0
				}
				cmd.SetVal([]interface{}{allowed, int64(60000)})
				return nil
			}, zap.NewNop())
			deps := router.Dependencies{RateLimiter: limiter, Logger: zap.NewNop(),
				AuthMiddleware: func(c *gin.Context) {
					if len(keys) != 1 {
						t.Error("IP must precede auth")
					}
					if tc.authFail {
						c.AbortWithStatus(401)
						return
					}
					c.Set(request.CtxStuID, "202610270001")
					c.Set(request.CtxRole, models.RoleAdmin)
					c.Next()
				},
				AppVersionMiddleware: func(c *gin.Context) {
					if len(keys) != 1 {
						t.Error("IP must precede version")
					}
					c.Next()
				},
			}
			r := router.SetupRouter(deps)
			w := perform(r, tc.method, tc.path, "", "192.0.2.1")
			want := 429
			if tc.authFail {
				want = 401
				if len(keys) != 1 {
					t.Fatal(keys)
				}
			}
			if w.Code != want {
				t.Fatalf("status=%d keys=%v body=%s", w.Code, keys, w.Body)
			}
			if !tc.authFail {
				expected := 2
				if tc.operation != "general" {
					expected = 3
				}
				if len(keys) != expected {
					t.Fatalf("expected %d checks, got %v", expected, keys)
				}
			}
			if !tc.authFail && !strings.HasPrefix(keys[len(keys)-1], "leave:"+tc.side+":"+tc.operation+":"+tc.dimension+":") {
				t.Fatal(keys)
			}
		})
	}
}

func TestLoginRegistrationValidationAndAccountLimits(t *testing.T) {
	for _, tc := range []struct {
		side, operation, body string
		status, calls         int
	}{
		{"app", "login", `{}`, 400, 1},
		{"app", "login", `{"student_id":"202610270001","password":"secret"}`, 429, 2},
		{"app", "register", `{"student_id":"bad","password":"secret","re_password":"secret"}`, 400, 1},
		{"app", "register", `{"student_id":"202610270001","password":"secret","re_password":"secret"}`, 429, 2},
		{"admin", "login", `{}`, 400, 1},
		{"admin", "login", `{"user_name":"admin","password":"secret"}`, 429, 2},
	} {
		t.Run(tc.side+tc.operation+tc.body, func(t *testing.T) {
			var keys []string
			limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
				keys = append(keys, cmd.Args()[3].(string))
				allowed := int64(1)
				if len(keys) > 1 {
					allowed = 0
				}
				cmd.SetVal([]interface{}{allowed, int64(5000)})
				return nil
			}, zap.NewNop())
			deps := router.Dependencies{
				RateLimiter: limiter, Logger: zap.NewNop(),
				AppUser:              appcontroller.NewUserHandler(nil, zap.NewNop(), limiter),
				AdminAuth:            admincontroller.NewAuthHandler(loginAccountService(t, "admin", nil), zap.NewNop(), limiter),
				AuthMiddleware:       func(c *gin.Context) { t.Fatal("public login must not authenticate") },
				AppVersionMiddleware: func(c *gin.Context) { c.Next() },
			}
			path := "/api/v1/auth/" + tc.operation
			if tc.side == "admin" {
				path = "/api/v1/admin/auth/" + tc.operation
			}
			w := perform(router.SetupRouter(deps), "POST", path, tc.body, "192.0.2.1")
			if w.Code != tc.status || len(keys) != tc.calls {
				t.Fatalf("status=%d keys=%v body=%s", w.Code, keys, w.Body)
			}
			for _, key := range keys {
				if !strings.HasPrefix(key, "leave:"+tc.side+":"+tc.operation+":") {
					t.Fatal(keys)
				}
			}
		})
	}
}

func TestRateLimitClientIPBehindLocalProxy(t *testing.T) {
	for _, path := range []string{"/api/v1/home", "/api/v1/admin/overview"} {
		for _, tc := range []struct {
			name, peer, realIP, forwardedFor, want string
		}{
			{"local proxy", "127.0.0.1", "198.51.100.10", "203.0.113.99", "198.51.100.10"},
			{"another client", "127.0.0.1", "198.51.100.11", "", "198.51.100.11"},
			{"IPv6 proxy", "::1", "2001:db8::10", "", "2001:db8::10"},
			{"IPv6 client", "127.0.0.1", "2001:db8::11", "", "2001:db8::11"},
			{"untrusted peer", "192.0.2.1", "198.51.100.10", "203.0.113.99", "192.0.2.1"},
			{"untrusted IPv6 peer", "2001:db8::1", "198.51.100.10", "", "2001:db8::1"},
			{"missing header", "127.0.0.1", "", "203.0.113.99", "127.0.0.1"},
			{"invalid header", "127.0.0.1", "invalid", "203.0.113.99", "127.0.0.1"},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				var keys []string
				limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
					keys = append(keys, cmd.Args()[3].(string))
					cmd.SetVal([]interface{}{int64(0), int64(60000)})
					return nil
				}, zap.NewNop())
				r := router.SetupRouter(router.Dependencies{RateLimiter: limiter, Logger: zap.NewNop()})
				req := httptest.NewRequest("GET", path, nil)
				req.RemoteAddr = net.JoinHostPort(tc.peer, "1234")
				req.Header.Set("X-Real-IP", tc.realIP)
				req.Header.Set("X-Forwarded-For", tc.forwardedFor)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				side := "app"
				if strings.Contains(path, "/admin/") {
					side = "admin"
				}
				want := "leave:" + side + ":general:ip:" + tc.want
				if w.Code != 429 || len(keys) != 1 || keys[0] != want {
					t.Fatalf("status=%d keys=%v want=%s", w.Code, keys, want)
				}
			})
		}
	}
}

func TestIPRejectionAndUntrustedHeaders(t *testing.T) {
	var keys []string
	limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
		keys = append(keys, cmd.Args()[3].(string))
		cmd.SetVal([]interface{}{int64(0), int64(60000)})
		return nil
	}, zap.NewNop())
	deps := router.Dependencies{RateLimiter: limiter, Logger: zap.NewNop(),
		AppVersions:          appcontroller.NewAppVersionHandler(nil, "", zap.NewNop()),
		AuthMiddleware:       func(c *gin.Context) { t.Fatal("must stop before auth") },
		AppVersionMiddleware: func(c *gin.Context) { t.Fatal("must stop before version") },
	}
	r := router.SetupRouter(deps)
	if w := perform(r, "GET", "/api/v1/home", "", "192.0.2.1"); w.Code != 429 {
		t.Fatal(w.Code)
	}
	// Different forwarding headers from an untrusted peer must not change the key.
	req := httptest.NewRequest("GET", "/api/v1/home", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.100")
	r.ServeHTTP(httptest.NewRecorder(), req)
	if keys[0] != keys[1] {
		t.Fatal("untrusted forwarding header changed quota")
	}
	count := len(keys)
	for _, entry := range []struct{ method, path string }{{"GET", "/api/v1/app-version"}, {"GET", "/api/v1/app-download"}, {"HEAD", "/api/v1/app-download"}} {
		if w := perform(r, entry.method, entry.path, "", "192.0.2.1"); w.Code != 400 {
			t.Fatalf("exempt %s: %d", entry.path, w.Code)
		}
	}
	if w := perform(r, "GET", "/api/v1/health", "", "192.0.2.1"); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := perform(r, "OPTIONS", "/api/v1/home", "", "192.0.2.1"); w.Code != 204 {
		t.Fatal(w.Code)
	}
	if len(keys) != count {
		t.Fatal("exempt routes consumed quota")
	}
}

// Supply a query result without connecting to a database.
func loginAccountService(t *testing.T, account string, queryErr error) *adminservice.AuthService {
	t.Helper()
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       "tcp(127.0.0.1:1)/rate_limit_test",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
		if queryErr != nil {
			tx.AddError(queryErr)
			return
		}
		*tx.Statement.Dest.(*models.User) = models.User{StudentID: account}
		tx.RowsAffected = 1
	}); err != nil {
		t.Fatal(err)
	}
	return adminservice.NewAuthService(mysql.NewDAO(db, zap.NewNop()), nil, zap.NewNop())
}

func TestAdminLoginCaseSensitiveAccountsHaveSeparateLimits(t *testing.T) {
	var keys []string
	limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
		keys = append(keys, cmd.Args()[3].(string))
		cmd.SetVal([]interface{}{int64(0), int64(5000)})
		return nil
	}, zap.NewNop())
	handler := admincontroller.NewAuthHandler(loginAccountService(t, "admin", nil), zap.NewNop(), limiter)
	r := gin.New()
	r.POST("/login", handler.LoginHandler)
	for _, account := range []string{"admin", "Admin", "ADMIN", "ádmin", "a\u0301dmin", " admin "} {
		body, err := json.Marshal(map[string]string{"user_name": account, "password": "secret"})
		if err != nil {
			t.Fatal(err)
		}
		w := perform(r, "POST", "/login", string(body), "192.0.2.1")
		if w.Code != 429 {
			t.Fatalf("%q: status=%d body=%s", account, w.Code, w.Body)
		}
	}
	if len(keys) != 6 {
		t.Fatal(keys)
	}
	if keys[0] != keys[5] {
		t.Fatal("surrounding whitespace must share quota")
	}
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			if keys[i] == keys[j] {
				t.Fatal("distinct accounts shared quota")
			}
		}
	}
}

func TestAdminLoginLimitBeforeAccountLookup(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		calls  int
	}{
		{"unknown", gorm.ErrRecordNotFound, 429, 1},
		{"database failure", errors.New("database unavailable"), 429, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
				calls++
				cmd.SetVal([]interface{}{int64(0), int64(5000)})
				return nil
			}, zap.NewNop())
			handler := admincontroller.NewAuthHandler(loginAccountService(t, "", tc.err), zap.NewNop(), limiter)
			r := gin.New()
			r.POST("/login", handler.LoginHandler)
			w := perform(r, "POST", "/login", `{"user_name":"missing","password":"secret"}`, "192.0.2.1")
			if w.Code != tc.status || calls != tc.calls {
				t.Fatalf("status=%d calls=%d", w.Code, calls)
			}
		})
	}
}

func TestRateLimitRuleSelection(t *testing.T) {
	for _, tc := range []struct {
		name      string
		operation string
		dimension string
		limit     int64
		window    int64
	}{
		{"general IP", middleware.OperationGeneral, middleware.DimensionIP, 600, 60000},
		{"general account", middleware.OperationGeneral, middleware.DimensionAccount, 60, 60000},
		{"login IP", middleware.OperationLogin, middleware.DimensionIP, 60, 60000},
		{"login account", middleware.OperationLogin, middleware.DimensionAccount, 10, 60000},
		{"password account", middleware.OperationPasswordChange, middleware.DimensionAccount, 5, 600000},
		{"record account", middleware.OperationRecordCreate, middleware.DimensionAccount, 5, 60000},
		{"avatar account", middleware.OperationAvatarUpload, middleware.DimensionAccount, 3, 60000},
		{"unsupported dimension", middleware.OperationLogin, "unknown", 0, 0},
		{"unsupported operation", "unknown", middleware.DimensionAccount, 0, 0},
		{"account-only operation", middleware.OperationPasswordChange, middleware.DimensionIP, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			limiter := testLimiter(t, func(ctx context.Context, cmd *redis.Cmd) error {
				calls++
				args := cmd.Args()
				if args[4] != tc.limit || args[5] != tc.window {
					t.Fatalf("unexpected rule: %v", args[4:])
				}
				if !strings.Contains(args[3].(string), ":"+tc.dimension+":") {
					t.Fatal("Redis key dimension changed")
				}
				cmd.SetVal([]interface{}{int64(1), tc.window})
				return nil
			}, zap.NewNop())
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", nil)
			allowed := limiter.Check(c, middleware.SideApp, tc.operation, tc.dimension, "test-account")
			if tc.limit == 0 {
				if allowed || calls != 0 || w.Code != 503 {
					t.Fatalf("invalid rule allowed=%v calls=%d status=%d", allowed, calls, w.Code)
				}
			} else if !allowed || calls != 1 {
				t.Fatalf("valid rule allowed=%v calls=%d", allowed, calls)
			}
		})
	}
}
