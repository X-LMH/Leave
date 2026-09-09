package service

import (
	"backend/internal/dto"
	"backend/internal/models"
	"testing"
	"time"
)

func TestValidateAndNormalizeRecordRequest(t *testing.T) {
	leaveSchool := false
	start := time.Date(2026, time.September, 7, 8, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	request := &dto.RecordCreateRequest{
		LeaveTypeID:    1,
		StartTime:      start,
		EndTime:        start.Add(30 * time.Minute),
		IsLeaveSchool:  &leaveSchool,
		LeaveReason:    " 身体不适 ",
		AffectedCourse: " 高等数学 ",
		TravelWay:      " 步行 ",
		Duration:       1,
		AppliedAt:      start.Add(-time.Hour),
		ApprovedAt:     start,
	}
	if !validateAndNormalizeRecordRequest(request) {
		t.Fatal("expected valid record request")
	}
	if request.LeaveReason != "身体不适" || request.AffectedCourse != "高等数学" || request.TravelWay != "步行" {
		t.Fatalf("request fields were not normalized: %#v", request)
	}

	request.EndTime = request.StartTime
	if validateAndNormalizeRecordRequest(request) {
		t.Fatal("expected equal start and end time to be invalid")
	}

	request.EndTime = request.StartTime.Add(time.Hour)
	request.ApprovedAt = request.AppliedAt.Add(-time.Nanosecond)
	if validateAndNormalizeRecordRequest(request) {
		t.Fatal("expected approval before application to be invalid")
	}

	request.ApprovedAt = request.EndTime
	request.Duration = 0
	if validateAndNormalizeRecordRequest(request) {
		t.Fatal("expected missing or zero duration to be invalid")
	}
}

func TestToRecordResponse(t *testing.T) {
	approvedAt := time.Date(2026, time.September, 7, 9, 0, 0, 0, time.UTC)
	record := &models.Record{
		ID: 7, StudentID: "20260001", Name: "张三", ParentName: "张父", ParentPhone: "13900139000", LeaveTypeID: 2, LeaveTypeName: "病假-本科生", College: "计算机学院", Major: "软件工程", ClassName: "软工 1 班", Duration: 3,
		LeaveReason: "就医", ApprovedAt: &approvedAt,
	}
	response := toRecordResponse(record)
	if response.ID != record.ID || response.StudentID != record.StudentID || response.Name != record.Name || response.ParentName != record.ParentName || response.ParentPhone != record.ParentPhone || response.ClassInfo.ClassName != record.ClassName || response.LeaveType != record.LeaveTypeName || response.ApprovedAt != approvedAt {
		t.Fatalf("record response does not preserve detail data: %#v", response)
	}
}
