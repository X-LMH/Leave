package app

import (
	"backend/internal/request"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestApartmentOptionsRejectInvalidGender(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(request.CtxStuID, "test") })
	r.GET("/apartments", GetApartmentOptionsHandler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/apartments?gender=unknown", nil))
	if w.Code != 400 {
		t.Fatalf("invalid gender status: %d", w.Code)
	}
}

func TestProfileOptionsRequireStudentIdentity(t *testing.T) {
	for _, handler := range []gin.HandlerFunc{GetClassOptionsHandler, GetApartmentOptionsHandler} {
		r := gin.New()
		r.GET("/", handler)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != 401 {
			t.Fatalf("missing identity status: %d", w.Code)
		}
	}
}
