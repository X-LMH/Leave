package app

import (
	"backend/internal/request"
	"backend/internal/response"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadAvatarRejectsInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		loggedIn bool
		field    string
		count    int
		size     int
		wantCode response.Code
	}{
		{name: "requires login", wantCode: response.CodeNeedLogin},
		{name: "missing file", loggedIn: true, wantCode: response.CodeInvalidParam},
		{name: "wrong field", loggedIn: true, field: "avatar", count: 1, size: 10, wantCode: response.CodeInvalidParam},
		{name: "empty file", loggedIn: true, field: "file", count: 1, wantCode: response.CodeInvalidParam},
		{name: "multiple files", loggedIn: true, field: "file", count: 2, size: 10, wantCode: response.CodeInvalidParam},
		{name: "invalid image content", loggedIn: true, field: "file", count: 1, size: 10, wantCode: response.CodeInvalidParam},
		{name: "file size limit", loggedIn: true, field: "file", count: 1, size: maxAvatarSize + 1, wantCode: response.CodeInvalidParam},
		{name: "request size limit", loggedIn: true, field: "file", count: 1, size: maxAvatarSize + (128 << 10), wantCode: response.CodeInvalidParam},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			for i := 0; i < tt.count; i++ {
				part, err := writer.CreateFormFile(tt.field, "avatar.png")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := part.Write(make([]byte, tt.size)); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/profile/avatar", &body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())
			if tt.loggedIn {
				c.Set(request.CtxStuID, "student")
			}
			UploadAvatarHandler(c)
			var result response.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Code != tt.wantCode || recorder.Code != tt.wantCode.HTTPStatus() {
				t.Fatalf("got status=%d code=%d, want code=%d", recorder.Code, result.Code, tt.wantCode)
			}
		})
	}
}

func TestUploadAvatarRejectsNonMultipartBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(request.CtxStuID, "student")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/profile/avatar", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")
	UploadAvatarHandler(c)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", recorder.Code)
	}
}
