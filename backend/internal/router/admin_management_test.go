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
