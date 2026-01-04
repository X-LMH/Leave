package mysql

import (
	"backend/models"
	"errors"

	"gorm.io/gorm"
)

func CheckExist(studentID string) (exist bool, err error) {
	var count int64
	if err = db.
		Model(&models.Profile{}).
		Where("student_id = ?", studentID).
		Count(&count).
		Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func FinishProfile(s *models.Profile) (err error) {
	var exist bool
	exist, err = CheckExist(s.StudentID)
	if err != nil {
		return
	}
	if exist {
		err = db.Model(&models.Profile{}).Where("student_id = ?", s.StudentID).Updates(s).Error
		return
	}
	if err = db.Create(s).Error; err != nil {
		return ErrorFinishProfile
	}
	return nil
}

func GetProfileByStuID(studentID string) (data *models.Profile, err error) {
	// 检查用户是否存在
	err = db.Model(&models.User{}).Where("student_id = ?", studentID).Error
	if err != nil {
		return nil, ErrorUserNotExist
	}
	exist, err := CheckExist(studentID)
	if err != nil {
		return nil, ErrorUserNotExist
	}
	if !exist {
		return data, nil
	}
	if err = db.
		Model(&models.Profile{}).
		Where("student_id = ?", studentID).
		First(&data).
		Error; err != nil {
		return
	}
	return
}

func InsertRecord(p *models.Record) (err error) {
	err = db.Create(p).Error
	if err != nil {
		return ErrorSubmitRecord
	}
	return nil
}

func GetRecordByID(recordID int) (*models.Record, error) {
	record := new(models.Record)
	err := db.Model(&models.Record{}).Where("id = ?", recordID).First(record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrorRecordNotExist
		}
		return nil, err
	}
	return record, nil
}

func GetRecordsByStuID(studentID string) ([]*models.Record, error) {
	records := make([]*models.Record, 0)
	err := db.Model(&models.Record{}).Where("student_id = ?", studentID).Order("created_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func DeleteRecordByID(recordID int) error {
	err := db.Model(&models.Record{}).Where("id = ?", recordID).Delete(&models.Record{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorRecordNotExist
		}
		return err
	}
	return nil
}
