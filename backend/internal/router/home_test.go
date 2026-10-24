package router

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHomeRequiresLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := testDependencies(mysql.NewDAO(nil, zap.NewNop()), config.Config{})
	deps.AppVersionMiddleware = func(c *gin.Context) { c.Next() }
	r := SetupRouter(deps)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/home?student_id=someone-else", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("home without login: status=%d body=%s", w.Code, w.Body.String())
	}
}
