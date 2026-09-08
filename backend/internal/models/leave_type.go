package models

import (
	"gorm.io/gorm"
	"time"
)

const (
	LeaveTypeDisabled = false
	LeaveTypeEnabled  = true
)

type LeaveType struct {
	ID        uint           `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string         `gorm:"column:name"`
	SortOrder uint           `gorm:"column:sort_order"`
	IsEnabled bool           `gorm:"column:is_enabled"` // LeaveTypeDisabled 或 LeaveTypeEnabled
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (LeaveType) TableName() string { return "leave_types" }
