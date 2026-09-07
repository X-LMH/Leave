package validator

import "testing"

func TestIsMainlandMobile(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{phone: "13800138000", valid: true},
		{phone: "19900138000", valid: true},
		{phone: "12800138000", valid: false},
		{phone: "1380013800", valid: false},
		{phone: "138001380000", valid: false},
		{phone: "1380013800a", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			if got := IsMainlandMobile(tt.phone); got != tt.valid {
				t.Fatalf("IsMainlandMobile(%q) = %v, want %v", tt.phone, got, tt.valid)
			}
		})
	}
}
