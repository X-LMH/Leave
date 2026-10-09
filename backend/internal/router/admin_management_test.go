package router

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestManagementRoutesRequireLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := SetupRouter()
	for _, path := range []string{"/api/v1/admin/classes", "/api/v1/admin/leave-types"} {
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
	r := SetupRouter()
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/v1/admin/app-versions"},
		{"GET", "/api/v1/admin/app-versions/current"},
		{"GET", "/api/v1/admin/app-versions/1"},
		{"GET", "/api/v1/admin/students"},
		{"GET", "/api/v1/admin/students/1"},
		{"PUT", "/api/v1/admin/students/1"},
		{"PUT", "/api/v1/admin/students/1/status"},
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
