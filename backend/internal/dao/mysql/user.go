package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

func CreateUser(student *models.User) (err error) {
	if err = db.Create(student).Error; err != nil {
		if IsDuplicateKeyError(err) {
			return ErrorUserExist
		}
		return err
	}
	return nil
}

// EnsureUserActive 确认账号仍存在且处于启用状态，供鉴权阶段撤销失效令牌。
func EnsureUserActive(studentID string) error {
	var user models.User
	if err := db.Select("status").Where("student_id = ?", studentID).First(&user).Error; err != nil {
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

// UpdateLoginInfo 记录账号最近一次成功登录的时间、设备名称和应用版本。
func UpdateLoginInfo(studentID string, deviceName string, appVersion string) error {
	return db.Model(&models.User{}).Where("student_id = ?", studentID).Updates(map[string]interface{}{
		"last_login_at":          time.Now(),
		"last_login_device":      deviceName,
		"last_login_app_version": appVersion,
	}).Error
}

func FindUserByStudentID(studentID string) (*models.User, error) {
	user := new(models.User)
	err := db.Model(&models.User{}).Where("student_id = ?", studentID).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrorUserNotExist
		}
		return nil, err
	}
	return user, nil
}

func ChangePassword(p *dto.PasswordChangeRequest, studentID string) (err error) {
	var student = new(models.User)
	// 查询用户
	if err = db.Where("student_id = ?", studentID).First(&student).Error; err != nil {
		return err
	}
	// 原密码错误
	if student.Password != p.Password {
		return ErrorNotRightPassword
	}
	// 修改
	if err = db.
		Model(&models.User{}).
		Where("student_id = ?", studentID).
		Update("password", p.NewPassword).
		Error; err != nil {
		return err
	}
	return nil
}
