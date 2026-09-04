package models

import (
	"gorm.io/gorm"
	"time"
)

type Record struct {
	ID             uint           `gorm:"column:id;primaryKey;autoIncrement"`
	StudentID      string         `gorm:"column:student_id"`
	Name           string         `gorm:"column:name"`
	LeaveTypeID    uint           `gorm:"column:leave_type_id"`
	StartTime      time.Time      `gorm:"column:start_time"`
	EndTime        time.Time      `gorm:"column:end_time"`
	Duration       uint           `gorm:"column:duration"`
	AffectedCourse string         `gorm:"column:affected_course"`
	IsLeaveSchool  bool           `gorm:"column:is_leave_school"` // LeaveSchoolNo 或 LeaveSchoolYes
	LeaveReason    string         `gorm:"column:leave_reason"`
	TravelWay      string         `gorm:"column:travel_way"`
	AppliedAt      time.Time      `gorm:"column:applied_at"`
	ApprovedAt     *time.Time     `gorm:"column:approved_at"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Record) TableName() string { return "records" }
