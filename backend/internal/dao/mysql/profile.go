package mysql

import (
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func FinishProfile(s *models.Profile) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "student_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"class_id":         s.ClassID,
			"name":             s.Name,
			"phone":            s.Phone,
			"gender":           s.Gender,
			"parent_name":      s.ParentName,
			"parent_phone":     s.ParentPhone,
			"apartment_id":     s.ApartmentID,
			"dormitory_number": s.DormitoryNumber,
			"teacher_name":     s.TeacherName,
		}),
	}).Create(s).Error
}

func GetProfileByStuID(studentID string) (data *models.Profile, err error) {
	data = new(models.Profile)
	if err = db.
		Where("student_id = ?", studentID).
		First(data).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return
	}
	return
}

func GetClassByID(classID uint) (*models.Class, error) {
	class := new(models.Class)
	if err := db.Where("id = ?", classID).First(class).Error; err != nil {
		return nil, err
	}
	return class, nil
}

func GetApartmentByID(apartmentID uint) (*models.Apartment, error) {
	apartment := new(models.Apartment)
	if err := db.Where("id = ?", apartmentID).First(apartment).Error; err != nil {
		return nil, err
	}
	return apartment, nil
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
