package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	"backend/internal/request"
	"backend/internal/response"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestManagementRejectsInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		handler            gin.HandlerFunc
		method, path, body string
	}{
		{GetLeaveTypesHandler, "GET", "/items?page=0", ""},
		{GetLeaveTypesHandler, "GET", "/items?is_enabled=bad", ""},
		{GetClassesHandler, "GET", "/items?page_size=101", ""},
		{CreateLeaveTypeHandler, "POST", "/items", `{"name":"a","sort_order":-1,"is_enabled":true}`},
		{CreateLeaveTypeHandler, "POST", "/items", `{"name":"a","sort_order":0}`},
		{CreateLeaveTypeHandler, "POST", "/items", `{"name":"a","sort_order":1.5,"is_enabled":true}`},
		{CreateClassHandler, "POST", "/items", `{"college":"a","major":"b","class_name":"c"}`},
		{UpdateClassHandler, "PUT", "/items/0", `{}`},
		{DeleteLeaveTypeHandler, "DELETE", "/items/nope", ""},
	} {
		r := gin.New()
		r.Handle(tc.method, "/items", tc.handler)
		r.Handle(tc.method, "/items/:id", tc.handler)
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s %s: %d", tc.method, tc.path, w.Code)
		}
	}
}
func TestManagementErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{gorm.ErrRecordNotFound, 404}, {gorm.ErrDuplicatedKey, 409}, {mysql.ErrorClassInUse, 409}, {service.ErrorInvalidAdminInput, 400},
	} {
		r := gin.New()
		r.GET("/", func(c *gin.Context) { managementError(c, tc.err) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != tc.status {
			t.Errorf("%v: %d", tc.err, w.Code)
		}
	}
}
func TestAdminAccess(t *testing.T) {
	for _, role := range []string{"student", "admin"} {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(request.CtxRole, role) })
		r.Use(middleware.AdminAuthMiddleware())
		r.GET("/", func(c *gin.Context) { response.Success(c, nil) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		expected := 401
		if role == "admin" {
			expected = 200
		}
		if w.Code != expected {
			t.Errorf("role %s: %d", role, w.Code)
		}
	}
}
