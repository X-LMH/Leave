package admin

import (
	"backend/internal/dto"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func validRecordInput() dto.AdminRecordUpdateRequest {
	leaving := true
	start := time.Date(2026, 10, 10, 8, 0, 0, 0, time.Local)
	return dto.AdminRecordUpdateRequest{
		LeaveTypeID:    1,
		StartTime:      start,
		EndTime:        start.Add(time.Hour),
		IsLeaveSchool:  &leaving,
		LeaveReason:    "  就医  ",
		AffectedCourse: "  课程  ",
		TravelWay:      "  公交  ",
		Destination:    "  医院  ",
		AppliedAt:      start.Add(-time.Hour),
		ApprovedAt:     start,
	}
}

func TestNormalizeAdminRecord(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*dto.AdminRecordUpdateRequest)
	}{
		{"missing type", func(p *dto.AdminRecordUpdateRequest) { p.LeaveTypeID = 0 }},
		{"missing school flag", func(p *dto.AdminRecordUpdateRequest) { p.IsLeaveSchool = nil }},
		{"blank reason", func(p *dto.AdminRecordUpdateRequest) { p.LeaveReason = "  " }},
		{"missing start", func(p *dto.AdminRecordUpdateRequest) { p.StartTime = time.Time{} }},
		{"missing end", func(p *dto.AdminRecordUpdateRequest) { p.EndTime = time.Time{} }},
		{"missing application", func(p *dto.AdminRecordUpdateRequest) { p.AppliedAt = time.Time{} }},
		{"missing approval", func(p *dto.AdminRecordUpdateRequest) { p.ApprovedAt = time.Time{} }},
		{"equal times", func(p *dto.AdminRecordUpdateRequest) { p.EndTime = p.StartTime }},
		{"reversed times", func(p *dto.AdminRecordUpdateRequest) { p.EndTime = p.StartTime.Add(-time.Second) }},
		{"early approval", func(p *dto.AdminRecordUpdateRequest) { p.ApprovedAt = p.AppliedAt.Add(-time.Second) }},
		{"missing destination", func(p *dto.AdminRecordUpdateRequest) { p.Destination = " " }},
		{"long course", func(p *dto.AdminRecordUpdateRequest) { p.AffectedCourse = strings.Repeat("课", 256) }},
		{"long travel", func(p *dto.AdminRecordUpdateRequest) { p.TravelWay = strings.Repeat("车", 33) }},
		{"long destination", func(p *dto.AdminRecordUpdateRequest) { p.Destination = strings.Repeat("地", 256) }},
		{"long text bytes", func(p *dto.AdminRecordUpdateRequest) { p.LeaveReason = strings.Repeat("事", 21846) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := validRecordInput()
			tc.change(&p)
			if !errors.Is(normalizeAdminRecord(&p), ErrorInvalidAdminInput) {
				t.Fatal("expected invalid input")
			}
		})
	}
	p := validRecordInput()
	p.AffectedCourse = strings.Repeat("课", 255)
	p.TravelWay = strings.Repeat("车", 32)
	p.Destination = strings.Repeat("地", 255)
	p.LeaveReason = strings.Repeat("事", 21845)
	p.ApprovedAt = p.AppliedAt
	if err := normalizeAdminRecord(&p); err != nil {
		t.Fatal(err)
	}
	p = validRecordInput()
	if err := normalizeAdminRecord(&p); err != nil {
		t.Fatal(err)
	}
	if p.LeaveReason != "就医" || p.AffectedCourse != "课程" || p.TravelWay != "公交" || p.Destination != "医院" {
		t.Fatal("text was not trimmed")
	}
	*p.IsLeaveSchool = false
	p.TravelWay = strings.Repeat("车", 33)
	if err := normalizeAdminRecord(&p); err != nil {
		t.Fatal(err)
	}
	if p.TravelWay != "" || p.Destination != "" {
		t.Fatal("non-school travel fields must be cleared")
	}
}

func TestAdminRecordHistoricalResponse(t *testing.T) {
	row := &models.Record{
		ID:            42,
		StudentID:     "20260001",
		Name:          "历史姓名",
		Gender:        "female",
		College:       "原学院",
		Major:         "原专业",
		ClassName:     "原班级",
		ParentName:    "原家长",
		ParentPhone:   "13800138000",
		TeacherName:   "原辅导员",
		LeaveTypeID:   99,
		LeaveTypeName: "历史原因",
		CreatedAt:     time.Now(),
	}
	result := adminRecordResponse(row)
	if result.Name != row.Name || result.College != row.College || result.Major != row.Major || result.ClassName != row.ClassName || result.ParentName != row.ParentName || result.ParentPhone != row.ParentPhone || result.TeacherName != row.TeacherName || result.LeaveTypeName != row.LeaveTypeName || result.CreatedAt != row.CreatedAt {
		t.Fatal("historical snapshot changed")
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"approved_at":null`) {
		t.Fatalf("nullable approval lost: %s", data)
	}
}
