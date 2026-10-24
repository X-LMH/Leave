package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestFeedbackQueryValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, query := range []string{"page=0", "page_size=101", "is_anonymous=other", "is_handled=other"} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/?"+query, nil)
		if _, ok := parseFeedbackQuery(c); ok {
			t.Fatalf("accepted %s", query)
		}
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?is_anonymous=false", nil)
	q, ok := parseFeedbackQuery(c)
	if !ok || q.IsAnonymous == nil || *q.IsAnonymous || q.Page != 1 || q.PageSize != 20 {
		t.Fatalf("valid query: %+v", q)
	}
	for _, value := range []string{"false", "true", ""} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/?is_handled="+value, nil)
		q, ok := parseFeedbackQuery(c)
		if !ok || (value == "" && q.IsHandled != nil) ||
			(value != "" && (q.IsHandled == nil || *q.IsHandled != (value == "true"))) {
			t.Fatalf("handling filter %q: %+v", value, q)
		}
	}
}
func TestApartmentRequestRejectsInvalidWireValues(t *testing.T) {
	dao := mysql.NewDAO(nil, zap.NewNop())
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/", NewApartmentHandler(adminservice.NewApartmentService(dao, zap.NewNop()), zap.NewNop()).CreateApartmentHandler)
	for _, body := range []string{
		`{"name":"楼","gender":"male","sort_order":-1,"is_enabled":true}`,
		`{"name":"楼","gender":"male","sort_order":4294967296,"is_enabled":true}`,
		`{"name":"楼","gender":"male","sort_order":0.5,"is_enabled":true}`,
		`{"name":"楼","gender":"male","sort_order":0}`,
		`{"name":"楼","gender":"male","is_enabled":true}`,
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != response.CodeInvalidParam.HTTPStatus() {
			t.Fatalf("%s: %d", body, w.Code)
		}
	}
}

func TestFeedbackApartmentRoutesRejectStudentRole(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil, zap.NewNop())
	gin.SetMode(gin.TestMode)
	feedbackHandler := NewFeedbackHandler(adminservice.NewFeedbackService(dao, settings.Storage, zap.NewNop()), zap.NewNop())
	apartmentHandler := NewApartmentHandler(adminservice.NewApartmentService(dao, zap.NewNop()), zap.NewNop())
	for _, route := range []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"GET", "/feedbacks", feedbackHandler.GetFeedbacksHandler},
		{"GET", "/feedbacks/:id", feedbackHandler.GetFeedbackHandler},
		{"PATCH", "/feedbacks/:id/handling", feedbackHandler.UpdateFeedbackHandlingHandler},
		{"GET", "/feedbacks/:id/attachments/:index", feedbackHandler.GetFeedbackAttachmentHandler},
		{"DELETE", "/feedbacks/:id", feedbackHandler.DeleteFeedbackHandler},
		{"GET", "/apartments", apartmentHandler.GetApartmentsHandler},
		{"POST", "/apartments", apartmentHandler.CreateApartmentHandler},
		{"PUT", "/apartments/:id", apartmentHandler.UpdateApartmentHandler},
		{"DELETE", "/apartments/:id", apartmentHandler.DeleteApartmentHandler},
	} {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(request.CtxRole, models.RoleStudent) }, middleware.AdminAuthMiddleware())
		r.Handle(route.method, route.path, route.handler)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(route.method, strings.ReplaceAll(route.path, ":id", "1"), nil))
		if w.Code != 401 {
			t.Fatalf("%s %s: %d", route.method, route.path, w.Code)
		}
	}
}
