package mysql

import (
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func GetUserForUpdate(tx *gorm.DB, studentID string) (*models.User, error) {
	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("student_id = ?", studentID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrorUserNotExist
		}
		return nil, err
	}
	return &user, nil
}

func UpdateUserPassword(tx *gorm.DB, id uint, password string) error {
	// Failed SQL must not write the new password into database logs.
	return tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
		Model(&models.User{}).Where("id = ?", id).Update("password", password).Error
}
