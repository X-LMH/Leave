package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/dao/mysql"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestHomeRequiresLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := testDependencies(t, mysql.NewDAO(nil, zap.NewNop()), config.Config{})
	deps.AppVersionMiddleware = func(c *gin.Context) { c.Next() }
	r := SetupRouter(deps)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/home?student_id=someone-else", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("home without login: status=%d body=%s", w.Code, w.Body.String())
	}
}
