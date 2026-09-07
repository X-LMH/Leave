package models

import "time"

type Apartment struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name"`
	Gender    string    `gorm:"column:gender"` // GenderMale 或 GenderFemale
	SortOrder uint      `gorm:"column:sort_order"`
	IsEnabled uint8     `gorm:"column:is_enabled"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Apartment) TableName() string { return "apartments" }
