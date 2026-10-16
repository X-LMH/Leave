package admin

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFeedbackHandlingRejectsMissingOrInvalidState(t *testing.T) {
	for _, body := range []string{`{}`, `{"is_handled":null}`, `{"is_handled":"false"}`, `{"is_handled":0}`, `not json`} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Request = httptest.NewRequest("PATCH", "/", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		UpdateFeedbackHandlingHandler(c)
		if w.Code != 400 {
			t.Fatalf("body=%s status=%d", body, w.Code)
		}
	}
}

func TestFeedbackAttachmentRejectsInvalidIndex(t *testing.T) {
	for _, index := range []string{"-1", "abc", "99999999999999999999999999"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "index", Value: index}}
		c.Request = httptest.NewRequest("GET", "/", nil)
		GetFeedbackAttachmentHandler(c)
		if w.Code != 400 {
			t.Fatalf("index=%s status=%d", index, w.Code)
		}
	}
}
