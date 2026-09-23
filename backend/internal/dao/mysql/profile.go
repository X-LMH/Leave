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

func GetApartmentByID(apartmentID uint) (*models.Apartment, error) {
	apartment := new(models.Apartment)
	if err := db.Where("id = ?", apartmentID).First(apartment).Error; err != nil {
		return nil, err
	}
	return apartment, nil
}

func GetApartments(gender string) ([]*models.Apartment, error) {
	apartments := make([]*models.Apartment, 0)
	query := db.Select("id", "name", "gender").Where("is_enabled = ?", 1)
	if gender != "" {
		query = query.Where("gender = ?", gender)
	}
	err := query.Order("sort_order ASC, id ASC").Find(&apartments).Error
	return apartments, err
}
