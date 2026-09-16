package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name      string
		url       string
		want      Pagination
		wantValid bool
	}{
		{name: "defaults", url: "/records", want: Pagination{Page: DefaultPage, PageSize: DefaultPageSize}, wantValid: true},
		{name: "custom values", url: "/records?page=2&page_size=50", want: Pagination{Page: 2, PageSize: 50}, wantValid: true},
		{name: "maximum page size", url: "/records?page_size=100", want: Pagination{Page: 1, PageSize: 100}, wantValid: true},
		{name: "invalid page", url: "/records?page=0", wantValid: false},
		{name: "page size too large", url: "/records?page_size=101", wantValid: false},
		{name: "negative page size", url: "/records?page_size=-1", wantValid: false},
		{name: "non numeric page", url: "/records?page=x", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, tt.url, nil)

			got, gotValid := ParsePagination(c)
			if gotValid != tt.wantValid || got != tt.want {
				t.Fatalf("ParsePagination() = %#v, %v; want %#v, %v", got, gotValid, tt.want, tt.wantValid)
			}
		})
	}
}
