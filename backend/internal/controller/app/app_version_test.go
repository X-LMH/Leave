package app

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOpenAPKMissingFile(t *testing.T) {
	_, _, err := openAPK(t.TempDir(), "missing.apk")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing APK: %v", err)
	}
}

func TestIsSafeAPKFileName(t *testing.T) {
	for _, test := range []struct {
		name string
		want bool
	}{
		{name: "leave-android-2-0.0.2.apk", want: true},
		{name: "", want: false},
		{name: "../leave.apk", want: false},
		{name: "dir/leave.apk", want: false},
		{name: `dir\\leave.apk`, want: false},
		{name: "leave.zip", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := isSafeAPKFileName(test.name); got != test.want {
				t.Fatalf("isSafeAPKFileName(%q) = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

func TestServeAPKSupportsFullRangeAndHeadResponses(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "Leave-*.apk")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err := file.WriteString("abcdef"); err != nil {
		t.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name        string
		method      string
		rangeHeader string
		status      int
		body        string
	}{
		{name: "full", method: http.MethodGet, status: http.StatusOK, body: "abcdef"},
		{name: "range", method: http.MethodGet, rangeHeader: "bytes=1-3", status: http.StatusPartialContent, body: "bcd"},
		{name: "invalid range", method: http.MethodGet, rangeHeader: "bytes=10-20", status: http.StatusRequestedRangeNotSatisfiable, body: "invalid range: failed to overlap\n"},
		{name: "head", method: http.MethodHead, status: http.StatusOK, body: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := file.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, "/api/v1/app-download?platform=android", nil)
			if test.rangeHeader != "" {
				request.Header.Set("Range", test.rangeHeader)
			}
			context, _ := gin.CreateTestContext(recorder)
			context.Request = request
			serveAPK(context, file, info)

			if recorder.Code != test.status || recorder.Body.String() != test.body {
				t.Fatalf("status/body = %d/%q, want %d/%q", recorder.Code, recorder.Body.String(), test.status, test.body)
			}
			if test.status != http.StatusRequestedRangeNotSatisfiable {
				if got := recorder.Header().Get("Content-Type"); got != "application/vnd.android.package-archive" {
					t.Fatalf("Content-Type = %q", got)
				}
			}
			if got := recorder.Header().Get("Content-Disposition"); !strings.Contains(got, ".apk") {
				t.Fatalf("Content-Disposition = %q", got)
			}
			if test.status != http.StatusRequestedRangeNotSatisfiable {
				if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
					t.Fatalf("Cache-Control = %q", got)
				}
			}
		})
	}
}
