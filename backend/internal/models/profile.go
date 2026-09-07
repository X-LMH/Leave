package models

import (
	"time"

	"gorm.io/gorm"
)

type Profile struct {
	ID              uint           `gorm:"column:id;primaryKey;autoIncrement"`
	StudentID       string         `gorm:"column:student_id"`
	ClassID         uint           `gorm:"column:class_id"`
	Name            string         `gorm:"column:name"`
	Phone           string         `gorm:"column:phone"`
	Gender          string         `gorm:"column:gender"` // GenderMale 或 GenderFemale
	ParentName      string         `gorm:"column:parent_name"`
	ParentPhone     string         `gorm:"column:parent_phone"`
	ApartmentID     uint           `gorm:"column:apartment_id"`
	DormitoryNumber string         `gorm:"column:dormitory_number"`
	TeacherName     string         `gorm:"column:teacher_name"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Profile) TableName() string { return "profiles" }
