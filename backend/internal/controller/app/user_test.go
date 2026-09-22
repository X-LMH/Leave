package app

import "testing"

func TestValidStudentID(t *testing.T) {
	tests := []struct {
		name      string
		studentID string
		want      bool
	}{
		{name: "twelve digits", studentID: "202600010001", want: true},
		{name: "surrounding spaces are ignored", studentID: " 202600010001 ", want: true},
		{name: "too short", studentID: "20260001000", want: false},
		{name: "too long", studentID: "2026000100011", want: false},
		{name: "contains letters", studentID: "20260001000a", want: false},
		{name: "contains punctuation", studentID: "20260001000-", want: false},
		{name: "empty", studentID: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validStudentID(tt.studentID); got != tt.want {
				t.Fatalf("validStudentID(%q) = %v, want %v", tt.studentID, got, tt.want)
			}
		})
	}
}
