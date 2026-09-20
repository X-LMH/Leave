package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	Success(c, gin.H{"ok": true})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != CodeSuccess || got.Message != "success" {
		t.Fatalf("response metadata = %#v", got)
	}
}

func TestErrorResponseUsesBusinessCodeStatusAndMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	Error(c, CodeNeedLogin)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != CodeNeedLogin || got.Message != CodeNeedLogin.Message() {
		t.Fatalf("response = %#v", got)
	}
}

func TestUnknownErrorCodeFallsBackSafely(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	ErrorWithData(c, Code(99999), "hidden")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	var got Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != CodeServerBusy || got.Message != CodeServerBusy.Message() || got.Data != "hidden" {
		t.Fatalf("fallback response = %#v", got)
	}
}

func TestCodeMetadataUnknownCodeFallsBack(t *testing.T) {
	unknown := Code(-1)
	if unknown.Message() != CodeServerBusy.Message() || unknown.HTTPStatus() != http.StatusInternalServerError {
		t.Fatalf("unknown code metadata = %q, %d", unknown.Message(), unknown.HTTPStatus())
	}
}
