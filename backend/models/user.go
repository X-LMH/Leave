package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint           `gorm:"column:id;primaryKey;autoIncrement"`
	StudentID   string         `gorm:"column:student_id"`
	Password    string         `gorm:"column:password"`
	Role        string         `gorm:"column:role"`
	Status      uint8          `gorm:"column:status"`
	LastLoginAt *time.Time     `gorm:"column:last_login_at"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (User) TableName() string { return "users" }
