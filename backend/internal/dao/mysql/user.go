package mysql

import (
	"backend/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (s *DAO) CreateUser(student *models.User) (err error) {
	if err = s.db.Create(student).Error; err != nil {
		if IsDuplicateKeyError(err) {
			return ErrorUserExist
		}
		return err
	}
	return nil
}

// EnsureUserActive 确认账号仍存在且处于启用状态，供鉴权阶段撤销失效令牌。
func (s *DAO) EnsureUserActive(studentID string) error {
	var user models.User
	if err := s.db.Select("status").Where("student_id = ?", studentID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorUserNotExist
		}
		return err
	}
	if user.Status != models.UserStatusActive {
		return ErrorUserDisabled
	}
	return nil
}

// UpdateAppUsage records the latest app entry time, device and version.
func (s *DAO) UpdateAppUsage(studentID string, deviceName string, appVersion string) error {
	return s.db.Model(&models.User{}).Where("student_id = ?", studentID).Updates(map[string]any{
		"last_seen_at":     time.Now(),
		"last_seen_device": deviceName,
		"app_version":      appVersion,
	}).Error
}

func (s *DAO) FindUserByStudentID(studentID string) (*models.User, error) {
	user := new(models.User)
	err := s.db.Model(&models.User{}).Where("student_id = ?", studentID).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrorUserNotExist
		}
		return nil, err
	}
	return user, nil
}
