package models

import (
	"gorm.io/gorm"
	"time"
)

const (
	LeaveSchoolNo  = false
	LeaveSchoolYes = true
)

type Record struct {
	ID             uint           `gorm:"column:id;primaryKey;autoIncrement"`
	StudentID      string         `gorm:"column:student_id"`
	Name           string         `gorm:"column:name"`
	Gender         string         `gorm:"column:gender"`
	ParentName     string         `gorm:"column:parent_name"`
	ParentPhone    string         `gorm:"column:parent_phone"`
	TeacherName    string         `gorm:"column:teacher_name"`
	LeaveTypeID    uint           `gorm:"column:leave_type_id"`
	LeaveTypeName  string         `gorm:"column:leave_type_name"`
	College        string         `gorm:"column:college"`
	Major          string         `gorm:"column:major"`
	ClassName      string         `gorm:"column:class_name"`
	StartTime      time.Time      `gorm:"column:start_time"`
	EndTime        time.Time      `gorm:"column:end_time"`
	AffectedCourse string         `gorm:"column:affected_course"`
	IsLeaveSchool  bool           `gorm:"column:is_leave_school"` // LeaveSchoolNo 或 LeaveSchoolYes
	LeaveReason    string         `gorm:"column:leave_reason"`
	TravelWay      string         `gorm:"column:travel_way"`
	Destination    string         `gorm:"column:destination"`
	AppliedAt      time.Time      `gorm:"column:applied_at"`
	ApprovedAt     *time.Time     `gorm:"column:approved_at"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Record) TableName() string { return "records" }
