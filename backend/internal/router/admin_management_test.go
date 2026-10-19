package router

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestManagementRoutesRequireLogin(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil)
	gin.SetMode(gin.TestMode)
	r := SetupRouter(testDependencies(dao, settings))
	for _, path := range []string{"/api/v1/admin/classes", "/api/v1/admin/leave-types", "/api/v1/admin/apartments"} {
		for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
			target := path
			if method == "PUT" || method == "DELETE" {
				target += "/1"
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(method, target, nil))
			if w.Code != 401 {
				t.Errorf("%s %s: %d", method, target, w.Code)
			}
		}
	}
}

func TestStudentRoutesRequireLogin(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil)
	r := SetupRouter(testDependencies(dao, settings))
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/v1/admin/overview"},
		{"POST", "/api/v1/admin/auth/password"},
		{"GET", "/api/v1/admin/app-versions"},
		{"GET", "/api/v1/admin/app-versions/current"},
		{"GET", "/api/v1/admin/app-versions/1"},
		{"GET", "/api/v1/admin/feedbacks"},
		{"GET", "/api/v1/admin/feedbacks/1"},
		{"PATCH", "/api/v1/admin/feedbacks/1/handling"},
		{"GET", "/api/v1/admin/feedbacks/1/attachments/0"},
		{"DELETE", "/api/v1/admin/feedbacks/1"},
		{"GET", "/api/v1/admin/students"},
		{"GET", "/api/v1/admin/records"},
		{"GET", "/api/v1/admin/records/1"},
		{"PUT", "/api/v1/admin/records/1"},
		{"GET", "/api/v1/admin/students/1"},
		{"PUT", "/api/v1/admin/students/1"},
		{"PUT", "/api/v1/admin/students/1/status"},
		{"POST", "/api/v1/admin/students/1/reset-password"},
		{"GET", "/api/v1/admin/class-options"},
		{"GET", "/api/v1/admin/apartment-options"},
		{"GET", "/api/v1/admin/leave-type-options"},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(route.method, route.path, nil))
		if w.Code != 401 {
			t.Errorf("%s %s: %d", route.method, route.path, w.Code)
		}
	}
}
