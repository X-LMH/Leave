package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestQueryInt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		url      string
		key      string
		fallback int
		want     int
	}{
		{name: "missing uses fallback", url: "/records", key: "page", fallback: 3, want: 3},
		{name: "valid integer", url: "/records?leave_type=2", key: "leave_type", want: 2},
		{name: "negative integer is preserved", url: "/records?leave_type=-1", key: "leave_type", want: -1},
		{name: "invalid integer returns sentinel", url: "/records?leave_type=abc", key: "leave_type", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, tt.url, nil)

			if got := queryInt(c, tt.key, tt.fallback); got != tt.want {
				t.Fatalf("queryInt(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestRecordIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		id         string
		wantID     int
		wantOK     bool
		wantStatus int
	}{
		{name: "positive ID", id: "12", wantID: 12, wantOK: true, wantStatus: http.StatusOK},
		{name: "zero ID", id: "0", wantOK: false, wantStatus: http.StatusBadRequest},
		{name: "negative ID", id: "-1", wantOK: false, wantStatus: http.StatusBadRequest},
		{name: "non numeric ID", id: "abc", wantOK: false, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			gotID, gotOK := recordIDFromContext(c)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Fatalf("recordIDFromContext() = %d, %v; want %d, %v", gotID, gotOK, tt.wantID, tt.wantOK)
			}
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
