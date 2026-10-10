package dto

import "time"

type AdminRecordQuery struct {
	Page          int
	PageSize      int
	StudentID     string
	Name          string
	ClassFilter   *RecordClassFilter
	LeaveTypeID   uint
	IsLeaveSchool *bool
	StartTimeFrom *time.Time
	StartTimeTo   *time.Time
}

type RecordClassFilter struct {
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"class_name"`
}

type AdminRecordListItem struct {
	ID            uint      `json:"id"`
	StudentID     string    `json:"student_id"`
	Name          string    `json:"name"`
	ClassName     string    `json:"class_name"`
	LeaveTypeID   uint      `json:"leave_type_id"`
	LeaveTypeName string    `json:"leave_type_name"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsLeaveSchool bool      `json:"is_leave_school"`
	AppliedAt     time.Time `json:"applied_at"`
}

type AdminRecord struct {
	AdminRecordListItem
	Gender         string     `json:"gender"`
	College        string     `json:"college"`
	Major          string     `json:"major"`
	ParentName     string     `json:"parent_name"`
	ParentPhone    string     `json:"parent_phone"`
	TeacherName    string     `json:"teacher_name"`
	AffectedCourse string     `json:"affected_course"`
	LeaveReason    string     `json:"leave_reason"`
	TravelWay      string     `json:"travel_way"`
	Destination    string     `json:"destination"`
	ApprovedAt     *time.Time `json:"approved_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type AdminRecordListResponse struct {
	Items    []*AdminRecordListItem `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// AdminRecordUpdateRequest deliberately excludes immutable applicant snapshots.
type AdminRecordUpdateRequest struct {
	LeaveTypeID    uint      `json:"leave_type_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	AffectedCourse string    `json:"affected_course"`
	IsLeaveSchool  *bool     `json:"is_leave_school"`
	LeaveReason    string    `json:"leave_reason"`
	TravelWay      string    `json:"travel_way"`
	Destination    string    `json:"destination"`
	AppliedAt      time.Time `json:"applied_at"`
	ApprovedAt     time.Time `json:"approved_at"`
}
