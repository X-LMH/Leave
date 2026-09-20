package request

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetCurrentStuID(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		set       bool
		want      string
		wantError error
	}{
		{name: "valid student ID", value: "202600010001", set: true, want: "202600010001"},
		{name: "missing", wantError: ErrorUserNotLogin},
		{name: "wrong type", value: 202600010001, set: true, wantError: ErrorUserNotLogin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			if tt.set {
				c.Set(CtxStuID, tt.value)
			}

			got, err := GetCurrentStuID(c)
			if got != tt.want || err != tt.wantError {
				t.Fatalf("GetCurrentStuID() = %q, %v; want %q, %v", got, err, tt.want, tt.wantError)
			}
		})
	}
}

func TestGetCurrentRole(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)
	if _, err := GetCurrentRole(c); err != ErrorUserPermissionDenied {
		t.Fatalf("missing role error = %v, want %v", err, ErrorUserPermissionDenied)
	}

	c.Set(CtxRole, 1)
	if _, err := GetCurrentRole(c); err != ErrorUserPermissionDenied {
		t.Fatalf("wrong role type error = %v, want %v", err, ErrorUserPermissionDenied)
	}

	c.Set(CtxRole, "student")
	role, err := GetCurrentRole(c)
	if role != "student" || err != nil {
		t.Fatalf("GetCurrentRole() = %q, %v", role, err)
	}
}
