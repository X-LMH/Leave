package app

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestApartmentOptionsRejectInvalidGender(t *testing.T) {
	r := gin.New()
	r.GET("/apartments", GetApartmentOptionsHandler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/apartments?gender=unknown", nil))
	if w.Code != 400 {
		t.Fatalf("invalid gender status: %d", w.Code)
	}
}
