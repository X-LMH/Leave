package app

import (
	"testing"
	"time"

	"backend/internal/dto"
	"backend/internal/models"
)

func TestValidateAndNormalizeRecordRequest(t *testing.T) {
	leaveSchool := false
	start := time.Date(2026, time.September, 7, 8, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	now := start.Add(24 * time.Hour)
	request := &dto.RecordCreateRequest{
		LeaveTypeID:    1,
		StartTime:      start,
		EndTime:        start.Add(30 * time.Minute),
		IsLeaveSchool:  &leaveSchool,
		LeaveReason:    " 身体不适 ",
		AffectedCourse: " 高等数学 ",
		TravelWay:      " 步行 ",
		AppliedAt:      start.Add(-time.Hour),
		ApprovedAt:     start,
	}
	if !validateAndNormalizeRecordRequest(request, now) {
		t.Fatal("expected valid record request")
	}
	if request.LeaveReason != "身体不适" || request.AffectedCourse != "高等数学" || request.TravelWay != "步行" {
		t.Fatalf("request fields were not normalized: %#v", request)
	}

	request.EndTime = request.StartTime
	if validateAndNormalizeRecordRequest(request, now) {
		t.Fatal("expected equal start and end time to be invalid")
	}

	request.EndTime = request.StartTime.Add(time.Hour)
	request.ApprovedAt = request.AppliedAt.Add(-time.Nanosecond)
	if validateAndNormalizeRecordRequest(request, now) {
		t.Fatal("expected approval before application to be invalid")
	}

	request.ApprovedAt = request.EndTime
	if !validateAndNormalizeRecordRequest(request, now) {
		t.Fatal("expected valid request without a duration snapshot")
	}
}

func TestValidateRecordApprovalTime(t *testing.T) {
	now := time.Date(2026, time.September, 7, 8, 0, 0, 0, time.UTC)
	appliedAt := now.Add(-time.Hour)
	tests := []struct {
		name       string
		appliedAt  time.Time
		approvedAt time.Time
		want       bool
	}{
		{name: "approval before application", appliedAt: appliedAt, approvedAt: appliedAt.Add(-time.Nanosecond)},
		{name: "approval equals application", appliedAt: appliedAt, approvedAt: appliedAt},
		{name: "approval just after application", appliedAt: appliedAt, approvedAt: appliedAt.Add(time.Nanosecond), want: true},
		{name: "approval just before now", appliedAt: appliedAt, approvedAt: now.Add(-time.Nanosecond), want: true},
		{name: "approval equals now", appliedAt: appliedAt, approvedAt: now, want: true},
		{name: "approval in future", appliedAt: appliedAt, approvedAt: now.Add(time.Nanosecond)},
		{name: "application and approval in future", appliedAt: now.Add(time.Hour), approvedAt: now.Add(2 * time.Hour)},
		{name: "missing application", approvedAt: now},
		{name: "missing approval", appliedAt: appliedAt},
		{name: "same instant in another timezone", appliedAt: appliedAt, approvedAt: now.In(time.FixedZone("CST", 8*60*60)), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leaveSchool := false
			request := &dto.RecordCreateRequest{
				LeaveTypeID:   1,
				StartTime:     now.Add(time.Hour),
				EndTime:       now.Add(2 * time.Hour),
				IsLeaveSchool: &leaveSchool,
				LeaveReason:   "就医",
				AppliedAt:     tt.appliedAt,
				ApprovedAt:    tt.approvedAt,
			}
			if got := validateAndNormalizeRecordRequest(request, now); got != tt.want {
				t.Fatalf("validateAndNormalizeRecordRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDurationInHours(t *testing.T) {
	start := time.Date(2026, time.September, 30, 23, 30, 0, 0, time.UTC)
	tests := []struct {
		name  string
		end   time.Time
		days  uint
		hours uint
	}{
		{name: "thirty minutes", end: start.Add(30 * time.Minute), days: 0, hours: 1},
		{name: "twenty three hours", end: start.Add(23 * time.Hour), days: 0, hours: 23},
		{name: "one day", end: start.Add(24 * time.Hour), days: 1, hours: 0},
		{name: "one day and ten minutes", end: start.Add(24*time.Hour + 10*time.Minute), days: 1, hours: 1},
		{name: "across month", end: start.Add(49 * time.Hour), days: 2, hours: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if days, hours := durationParts(start, test.end); days != test.days || hours != test.hours {
				t.Fatalf("expected %d days %d hours, got %d days %d hours", test.days, test.hours, days, hours)
			}
		})
	}
}

func TestToRecordResponse(t *testing.T) {
	approvedAt := time.Date(2026, time.September, 7, 9, 0, 0, 0, time.UTC)
	record := &models.Record{
		ID: 7, StudentID: "20260001", Name: "张三", ParentName: "张父", ParentPhone: "13900139000", LeaveTypeID: 2, LeaveTypeName: "病假-本科生", College: "计算机学院", Major: "软件工程", ClassName: "软工 1 班", StartTime: approvedAt, EndTime: approvedAt.Add(27 * time.Hour),
		LeaveReason: "就医", ApprovedAt: &approvedAt,
	}
	response := toRecordResponse(record)
	if response.ID != record.ID || response.StudentID != record.StudentID || response.ApplicantInfo.Name != record.Name || response.ApplicantInfo.ParentName != record.ParentName || response.ApplicantInfo.ParentPhone != record.ParentPhone || response.ClassInfo.ClassName != record.ClassName || response.LeaveType != record.LeaveTypeName || response.DurationDays != 1 || response.DurationHours != 3 || response.ApprovedAt != approvedAt {
		t.Fatalf("record response does not preserve detail data: %#v", response)
	}
}
