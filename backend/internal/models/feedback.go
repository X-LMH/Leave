package models

import (
	"time"

	"gorm.io/gorm"
)

// Feedback is a text submission from a signed-in user. Anonymous submissions do not retain a student ID.
type Feedback struct {
	ID          uint           `gorm:"column:id;primaryKey;autoIncrement"`
	StudentID   *string        `gorm:"column:student_id"`
	IsAnonymous bool           `gorm:"column:is_anonymous"`
	Content     string         `gorm:"column:content"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Feedback) TableName() string { return "feedbacks" }
