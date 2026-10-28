package dto

import "time"

type HomeStatistics struct {
	TotalCount int64 `json:"total_count"`
	MonthCount int64 `json:"month_count"`
}

type HomeRecord struct {
	ID            uint      `json:"id"`
	LeaveType     string    `json:"leave_type"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	DurationDays  uint      `json:"duration_days"`
	DurationHours uint      `json:"duration_hours"`
	IsLeaveSchool bool      `json:"is_leave_school"`
}

type HomeUser struct {
	Role      string `json:"role"`
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	ClassName string `json:"classname"`
	AvatarURL string `json:"avatar_url"`
}

type HomeResponse struct {
	User         HomeUser       `json:"user"`
	Statistics   HomeStatistics `json:"statistics"`
	LatestRecord *HomeRecord    `json:"latest_record"`
}
