package service

import (
	"backend/internal/dto"
	"strings"
	"testing"
)

func TestNormalizeAdminLeaveType(t *testing.T) {
	zero := uint64(0)
	disabled := false
	p := &dto.AdminLeaveTypeRequest{Name: "  病假  ", SortOrder: &zero, IsEnabled: &disabled}
	if err := normalizeLeaveType(p); err != nil || p.Name != "病假" || *p.IsEnabled {
		t.Fatalf("valid zero/false rejected: %+v %v", p, err)
	}
	max := uint64(4294967295)
	overflow := max + 1
	for _, tc := range []struct {
		name    string
		sort    *uint64
		enabled *bool
		valid   bool
	}{
		{strings.Repeat("假", 32), &max, &disabled, true},
		{strings.Repeat("假", 33), &zero, &disabled, false},
		{"  ", &zero, &disabled, false},
		{"病假", nil, &disabled, false},
		{"病假", &overflow, &disabled, false},
		{"病假", &zero, nil, false},
	} {
		err := normalizeLeaveType(&dto.AdminLeaveTypeRequest{Name: tc.name, SortOrder: tc.sort, IsEnabled: tc.enabled})
		if (err == nil) != tc.valid {
			t.Errorf("valid=%v err=%v", tc.valid, err)
		}
	}
}
func TestNormalizeAdminClass(t *testing.T) {
	disabled := false
	p := &dto.AdminClassRequest{College: "  学院 ", Major: " 专业 ", ClassName: " 班级 ", IsEnabled: &disabled}
	if err := normalizeClass(p); err != nil || p.College != "学院" || p.Major != "专业" || p.ClassName != "班级" {
		t.Fatalf("normalize: %+v %v", p, err)
	}
	for _, tc := range []struct {
		college, major, name string
		enabled              *bool
		valid                bool
	}{
		{strings.Repeat("院", 100), strings.Repeat("专", 100), strings.Repeat("班", 64), &disabled, true},
		{strings.Repeat("院", 101), "专业", "班级", &disabled, false},
		{"学院", strings.Repeat("专", 101), "班级", &disabled, false},
		{"学院", "专业", strings.Repeat("班", 65), &disabled, false},
		{"学院", "专业", " ", &disabled, false},
		{"学院", "专业", "班级", nil, false},
	} {
		err := normalizeClass(&dto.AdminClassRequest{College: tc.college, Major: tc.major, ClassName: tc.name, IsEnabled: tc.enabled})
		if (err == nil) != tc.valid {
			t.Errorf("valid=%v err=%v", tc.valid, err)
		}
	}
}
