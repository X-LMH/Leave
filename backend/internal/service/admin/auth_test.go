package admin

import (
	"backend/internal/dto"
	"errors"
	"strings"
	"testing"
)

func TestValidatePasswordChange(t *testing.T) {
	for _, tc := range []struct {
		name     string
		old      string
		password string
		confirm  string
		valid    bool
	}{
		{"six bytes", "old", "123456", "123456", true},
		{"thirty bytes", "old", strings.Repeat("a", 30), strings.Repeat("a", 30), true},
		{"five bytes", "old", "12345", "12345", false},
		{"thirty one bytes", "old", strings.Repeat("a", 31), strings.Repeat("a", 31), false},
		{"unicode six bytes", "old", "密码", "密码", true},
		{"unicode thirty bytes", "old", strings.Repeat("密", 10), strings.Repeat("密", 10), true},
		{"unicode too long", "old", strings.Repeat("密", 11), strings.Repeat("密", 11), false},
		{"empty old", "", "123456", "123456", false},
		{"empty new", "old", "", "", false},
		{"empty confirm", "old", "123456", "", false},
		{"spaces", "old", "      ", "      ", false},
		{"unicode spaces", "old", "\u3000\u3000", "\u3000\u3000", false},
		{"unicode next line", "old", "\u0085\u0085\u0085", "\u0085\u0085\u0085", false},
		{"BOM is not whitespace", "old", "\ufeff\ufeff", "\ufeff\ufeff", true},
		{"mismatch", "old", "123456", "654321", false},
		{"unchanged", "123456", "123456", "123456", false},
		{"preserve spaces", "old", " pass ", " pass ", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := &dto.PasswordChangeRequest{
				Password:    tc.old,
				NewPassword: tc.password,
				RePassword:  tc.confirm,
			}
			err := validatePasswordChange(input)
			if (err == nil) != tc.valid || (err != nil && !errors.Is(err, ErrorInvalidAdminInput)) {
				t.Fatalf("validation result: %v", err)
			}
			if input.NewPassword != tc.password {
				t.Fatal("password was normalized")
			}
		})
	}
	if !errors.Is(validatePasswordChange(nil), ErrorInvalidAdminInput) {
		t.Fatal("nil input must be rejected")
	}
}
