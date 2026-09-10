package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleStudent              = "student"
	RoleAdmin                = "admin"
	UserStatusDisabled uint8 = 0
	UserStatusActive   uint8 = 1
)

type User struct {
	ID             uint           `gorm:"column:id;primaryKey;autoIncrement"`
	StudentID      string         `gorm:"column:student_id"`
	Password       string         `gorm:"column:password"`
	Role           string         `gorm:"column:role"`   // RoleStudent 或 RoleAdmin
	Status         uint8          `gorm:"column:status"` // UserStatusDisabled 或 UserStatusActive
	LastSeenAt     *time.Time     `gorm:"column:last_seen_at"`
	LastSeenDevice string         `gorm:"column:last_seen_device"`
	AppVersion     string         `gorm:"column:app_version"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (User) TableName() string { return "users" }
