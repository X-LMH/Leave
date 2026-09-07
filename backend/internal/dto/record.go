package dto

import "time"

// RecordCreateRequest is the JSON body accepted when creating a leave record.
// All time values use RFC3339.
type RecordCreateRequest struct {
	LeaveTypeID    uint      `json:"leave_type_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Duration       uint      `json:"duration"`
	AffectedCourse string    `json:"affected_course"`
	IsLeaveSchool  *bool     `json:"is_leave_school"`
	LeaveReason    string    `json:"leave_reason"`
	TravelWay      string    `json:"travel_way"`
	AppliedAt      time.Time `json:"applied_at"`
	ApprovedAt     time.Time `json:"approved_at"`
}

// RecordResponse contains all information needed by the leave-record detail view.
type RecordResponse struct {
	ID             uint             `json:"id"`
	StudentID      string           `json:"student_id"`
	Name           string           `json:"name"`
	ClassInfo      ProfileClassInfo `json:"class_info"`
	LeaveType      string           `json:"leave_type"`
	StartTime      time.Time        `json:"start_time"`
	EndTime        time.Time        `json:"end_time"`
	Duration       uint             `json:"duration"`
	AffectedCourse string           `json:"affected_course"`
	IsLeaveSchool  bool             `json:"is_leave_school"`
	LeaveReason    string           `json:"leave_reason"`
	TravelWay      string           `json:"travel_way"`
	AppliedAt      time.Time        `json:"applied_at"`
	ApprovedAt     time.Time        `json:"approved_at"`
}

// RecordListItem is the compact record representation returned by GET /records.
type RecordListItem struct {
	ID          uint      `json:"id"`
	LeaveType   string    `json:"leave_type"`
	LeaveReason string    `json:"leave_reason"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    uint      `json:"duration"`
}

type RecordListQuery struct {
	Page      int
	PageSize  int
	LeaveType uint
}
