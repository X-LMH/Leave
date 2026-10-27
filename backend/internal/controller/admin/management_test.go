package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	commonservice "backend/internal/service/common"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestManagementRejectsInvalidRequests(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil, zap.NewNop())
	gin.SetMode(gin.TestMode)
	studentHandler := NewStudentHandler(adminservice.NewStudentService(dao, settings.Storage, zap.NewNop()), zap.NewNop())
	appVersionHandler := NewAppVersionHandler(adminservice.NewAppVersionService(dao, zap.NewNop()), zap.NewNop())
	leaveTypeHandler := NewLeaveTypeHandler(adminservice.NewLeaveTypeService(dao, zap.NewNop()), zap.NewNop())
	recordHandler := NewRecordHandler(adminservice.NewRecordService(dao, zap.NewNop()), zap.NewNop())
	classHandler := NewClassHandler(adminservice.NewClassService(dao, zap.NewNop()), zap.NewNop())
	for _, tc := range []struct {
		handler            gin.HandlerFunc
		method, path, body string
	}{
		{appVersionHandler.GetAppVersionsHandler, "GET", "/items?page=0", ""},
		{appVersionHandler.GetAppVersionsHandler, "GET", "/items?page_size=101", ""},
		{appVersionHandler.GetAppVersionsHandler, "GET", "/items?platform=unknown", ""},
		{appVersionHandler.GetAppVersionsHandler, "GET", "/items?status=draft", ""},
		{appVersionHandler.GetAppVersionHandler, "GET", "/items/0", ""},
		{appVersionHandler.GetAppVersionHandler, "GET", "/items/-1", ""},
		{appVersionHandler.GetAppVersionHandler, "GET", "/items/bad", ""},
		{appVersionHandler.GetAppVersionHandler, "GET", "/items/18446744073709551616", ""},
		{NewOptionsHandler(commonservice.NewOptionsService(dao, zap.NewNop()), zap.NewNop()).GetApartmentOptionsHandler, "GET", "/items?gender=unknown", ""},
		{studentHandler.GetStudentsHandler, "GET", "/items?page=0", ""},
		{recordHandler.GetRecordsHandler, "GET", "/items?page=0", ""},
		{recordHandler.GetRecordsHandler, "GET", "/items?page_size=101", ""},
		{recordHandler.GetRecordsHandler, "GET", "/items?class_name=test", ""},
		{recordHandler.GetRecordsHandler, "GET", "/items?leave_type_id=-1", ""},
		{recordHandler.GetRecordsHandler, "GET", "/items?is_leave_school=bad", ""},
		{recordHandler.GetRecordsHandler, "GET", "/items?start_time_from=bad", ""},
		{recordHandler.GetRecordsHandler, "GET", "/items?start_time_from=2026-10-11T00:00:00Z&start_time_to=2026-10-10T00:00:00Z", ""},
		{recordHandler.GetRecordHandler, "GET", "/items/0", ""},
		{recordHandler.GetRecordHandler, "GET", "/items/bad", ""},
		{recordHandler.UpdateRecordHandler, "PUT", "/items/1", `{}`},
		{recordHandler.UpdateRecordHandler, "PUT", "/items/1", `{"start_time":"bad"}`},
		{studentHandler.GetStudentsHandler, "GET", "/items?page_size=101", ""},
		{studentHandler.GetStudentsHandler, "GET", "/items?class_id=0", ""},
		{studentHandler.GetStudentsHandler, "GET", "/items?class_id=bad", ""},
		{studentHandler.GetStudentsHandler, "GET", "/items?status=2", ""},
		{studentHandler.GetStudentsHandler, "GET", "/items?status=-1", ""},
		{studentHandler.GetStudentHandler, "GET", "/items/0", ""},
		{studentHandler.UpdateStudentHandler, "PUT", "/items/1", `{}`},
		{studentHandler.UpdateStudentStatusHandler, "PUT", "/items/1", `{}`},
		{studentHandler.UpdateStudentStatusHandler, "PUT", "/items/1", `{"status":2}`},
		{studentHandler.UpdateStudentStatusHandler, "PUT", "/items/1", `{"status":-1}`},
		{studentHandler.UpdateStudentStatusHandler, "PUT", "/items/1", `{"status":1.5}`},
		{leaveTypeHandler.GetLeaveTypesHandler, "GET", "/items?page=0", ""},
		{leaveTypeHandler.GetLeaveTypesHandler, "GET", "/items?is_enabled=bad", ""},
		{classHandler.GetClassesHandler, "GET", "/items?page_size=101", ""},
		{leaveTypeHandler.CreateLeaveTypeHandler, "POST", "/items", `{"name":"a","sort_order":-1,"is_enabled":true}`},
		{leaveTypeHandler.CreateLeaveTypeHandler, "POST", "/items", `{"name":"a","sort_order":0}`},
		{leaveTypeHandler.CreateLeaveTypeHandler, "POST", "/items", `{"name":"a","sort_order":1.5,"is_enabled":true}`},
		{classHandler.CreateClassHandler, "POST", "/items", `{"college":"a","major":"b","class_name":"c"}`},
		{classHandler.UpdateClassHandler, "PUT", "/items/0", `{}`},
		{leaveTypeHandler.DeleteLeaveTypeHandler, "DELETE", "/items/nope", ""},
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
		r.GET("/", func(c *gin.Context) { managementError(c, tc.err, zap.NewNop()) })
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
