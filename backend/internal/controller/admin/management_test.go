package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
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
		{GetAppVersionsHandler, "GET", "/items?page=0", ""},
		{GetAppVersionsHandler, "GET", "/items?page_size=101", ""},
		{GetAppVersionsHandler, "GET", "/items?platform=unknown", ""},
		{GetAppVersionsHandler, "GET", "/items?status=draft", ""},
		{GetAppVersionHandler, "GET", "/items/0", ""},
		{GetAppVersionHandler, "GET", "/items/-1", ""},
		{GetAppVersionHandler, "GET", "/items/bad", ""},
		{GetAppVersionHandler, "GET", "/items/18446744073709551616", ""},
		{GetApartmentOptionsHandler, "GET", "/items?gender=unknown", ""},
		{GetStudentsHandler, "GET", "/items?page=0", ""},
		{GetStudentsHandler, "GET", "/items?page_size=101", ""},
		{GetStudentsHandler, "GET", "/items?class_id=0", ""},
		{GetStudentsHandler, "GET", "/items?class_id=bad", ""},
		{GetStudentsHandler, "GET", "/items?status=2", ""},
		{GetStudentsHandler, "GET", "/items?status=-1", ""},
		{GetStudentHandler, "GET", "/items/0", ""},
		{UpdateStudentHandler, "PUT", "/items/1", `{}`},
		{UpdateStudentStatusHandler, "PUT", "/items/1", `{}`},
		{UpdateStudentStatusHandler, "PUT", "/items/1", `{"status":2}`},
		{UpdateStudentStatusHandler, "PUT", "/items/1", `{"status":-1}`},
		{UpdateStudentStatusHandler, "PUT", "/items/1", `{"status":1.5}`},
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
		{gorm.ErrRecordNotFound, 404}, {gorm.ErrDuplicatedKey, 409}, {mysql.ErrorClassInUse, 409}, {adminservice.ErrorInvalidAdminInput, 400},
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
