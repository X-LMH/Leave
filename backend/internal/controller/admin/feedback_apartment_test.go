package admin

import (
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/request"
	"backend/internal/response"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
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
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/", CreateApartmentHandler)
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
	gin.SetMode(gin.TestMode)
	for _, route := range []struct {
		method, path string
		handler      gin.HandlerFunc
	}{
		{"GET", "/feedbacks", GetFeedbacksHandler},
		{"GET", "/feedbacks/:id", GetFeedbackHandler},
		{"PATCH", "/feedbacks/:id/handling", UpdateFeedbackHandlingHandler},
		{"GET", "/feedbacks/:id/attachments/:index", GetFeedbackAttachmentHandler},
		{"DELETE", "/feedbacks/:id", DeleteFeedbackHandler},
		{"GET", "/apartments", GetApartmentsHandler},
		{"POST", "/apartments", CreateApartmentHandler},
		{"PUT", "/apartments/:id", UpdateApartmentHandler},
		{"DELETE", "/apartments/:id", DeleteApartmentHandler},
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
