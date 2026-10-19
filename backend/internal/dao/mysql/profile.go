package mysql

import (
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *DAO) SaveAppProfile(tx *gorm.DB, profile *models.Profile) error {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "student_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"class_id":         profile.ClassID,
			"name":             profile.Name,
			"phone":            profile.Phone,
			"gender":           profile.Gender,
			"parent_name":      profile.ParentName,
			"parent_phone":     profile.ParentPhone,
			"apartment_id":     profile.ApartmentID,
			"dormitory_number": profile.DormitoryNumber,
			"teacher_name":     profile.TeacherName,
		}),
	}).Create(profile).Error
}

func (s *DAO) GetProfileByStuID(studentID string) (data *models.Profile, err error) {
	data = new(models.Profile)
	if err = s.db.
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

func (s *DAO) UpdateProfileAvatar(studentID, key string) error {
	result := s.db.Model(&models.Profile{}).
		Where("student_id = ?", studentID).
		Update("avatar_url", key)
	if result.Error != nil {
		return result.Error
	}
	// 新头像使用随机路径，未更新任何行意味着用户资料不存在。
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *DAO) GetApartmentByID(apartmentID uint) (*models.Apartment, error) {
	apartment := new(models.Apartment)
	if err := s.db.Where("id = ?", apartmentID).First(apartment).Error; err != nil {
		return nil, err
	}
	return apartment, nil
}

func (s *DAO) GetApartmentOptions(gender string, enabledOnly bool) ([]*models.Apartment, error) {
	apartments := make([]*models.Apartment, 0)
	query := s.db.Select("id", "name", "gender", "is_enabled")
	if enabledOnly {
		query = query.Where("is_enabled = ?", 1)
	}
	if gender != "" {
		query = query.Where("gender = ?", gender)
	}
	err := query.Order("sort_order ASC, id ASC").Find(&apartments).Error
	return apartments, err
}
