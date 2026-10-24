package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/request"
	appservice "backend/internal/service/app"
	commonservice "backend/internal/service/common"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestApartmentOptionsRejectInvalidGender(t *testing.T) {
	dao := mysql.NewDAO(nil, zap.NewNop())
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(request.CtxStuID, "test") })
	r.GET("/apartments", NewOptionsHandler(appservice.NewOptionsService(dao, commonservice.NewOptionsService(dao, zap.NewNop()), zap.NewNop()), commonservice.NewOptionsService(dao, zap.NewNop()), zap.NewNop()).GetApartmentOptionsHandler)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/apartments?gender=unknown", nil))
	if w.Code != 400 {
		t.Fatalf("invalid gender status: %d", w.Code)
	}
}

func TestProfileOptionsRequireStudentIdentity(t *testing.T) {
	dao := mysql.NewDAO(nil, zap.NewNop())
	optionsHandler := NewOptionsHandler(appservice.NewOptionsService(dao, commonservice.NewOptionsService(dao, zap.NewNop()), zap.NewNop()), commonservice.NewOptionsService(dao, zap.NewNop()), zap.NewNop())
	for _, handler := range []gin.HandlerFunc{optionsHandler.GetClassOptionsHandler, optionsHandler.GetApartmentOptionsHandler} {
		r := gin.New()
		r.GET("/", handler)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != 401 {
			t.Fatalf("missing identity status: %d", w.Code)
		}
	}
}
