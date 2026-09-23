package models

import "time"

type Class struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement"`
	College   string    `gorm:"column:college"`
	Major     string    `gorm:"column:major"`
	ClassName string    `gorm:"column:class_name"`
	IsEnabled bool      `gorm:"column:is_enabled"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Class) TableName() string { return "classes" }
