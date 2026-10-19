package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/request"
	appservice "backend/internal/service/app"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseFeedbackSubmission(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fields []string
		valid  bool
	}{
		{"no attachments", nil, true},
		{"one", []string{"attachment_0"}, true},
		{"three", []string{"attachment_0", "attachment_1", "attachment_2"}, true},
		{"four", []string{"attachment_0", "attachment_1", "attachment_2", "attachment_3"}, false},
		{"duplicate", []string{"attachment_0", "attachment_0"}, false},
		{"wrong field", []string{"file"}, false},
		{"noncontiguous", []string{"attachment_1"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			_ = writer.WriteField("content", "some feedback")
			_ = writer.WriteField("is_anonymous", "false")
			for _, field := range tc.fields {
				part, err := writer.CreateFormFile(field, "a.png")
				if err != nil {
					t.Fatal(err)
				}
				_, _ = fmt.Fprint(part, "image")
			}
			_ = writer.Close()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/feedback", &body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())
			defer func() {
				if c.Request.MultipartForm != nil {
					_ = c.Request.MultipartForm.RemoveAll()
				}
			}()
			req, uploads, err := parseFeedbackSubmission(c)
			if tc.valid {
				if err != nil || len(uploads) != len(tc.fields) || req.IsAnonymous {
					t.Fatalf("req=%v count=%d err=%v", req, len(uploads), err)
				}
			} else if err == nil {
				t.Fatal("accepted malformed files")
			}
		})
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/feedback", strings.NewReader(`{"content":"legacy feedback","is_anonymous":true}`))
	c.Request.Header.Set("Content-Type", "application/json")
	req, uploads, err := parseFeedbackSubmission(c)
	if err != nil || !req.IsAnonymous || len(uploads) != 0 {
		t.Fatalf("legacy JSON: %v %v", req, err)
	}
}

func TestFeedbackRequestBodyLimit(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(request.CtxStuID, "student")
	c.Request = httptest.NewRequest("POST", "/feedback", strings.NewReader(`{"content":"`+strings.Repeat("x", appservice.MaxFeedbackAttachments*appservice.MaxFeedbackAttachmentSize+(128<<10))+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	NewFeedbackHandler(appservice.NewFeedbackService(dao, settings.Storage)).CreateFeedbackHandler(c)
	// Handler must reject before reaching the database.
	recorder := c.Writer
	if recorder.Status() != 400 {
		t.Fatalf("status=%d", recorder.Status())
	}
}
