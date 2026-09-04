package models

import (
	"gorm.io/gorm"
	"time"
)

type Profile struct {
	ID          uint           `gorm:"column:id;primaryKey;autoIncrement"`
	StudentID   string         `gorm:"column:student_id"`
	ClassID     uint           `gorm:"column:class_id"`
	Name        string         `gorm:"column:name"`
	Phone       string         `gorm:"column:phone"`
	Gender      string         `gorm:"column:gender"`
	ParentName  string         `gorm:"column:parent_name"`
	ParentPhone string         `gorm:"column:parent_phone"`
	Apartment   string         `gorm:"column:apartment"`
	ApartmentID string         `gorm:"column:apartment_id"`
	TeacherName string         `gorm:"column:teacher_name"`
	CreateAt    time.Time      `gorm:"column:create_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Profile) TableName() string { return "profiles" }
