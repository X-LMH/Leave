package mysql

import (
	"backend/models"
	"errors"

	"gorm.io/gorm"
)

func CheckUserExist(studentID string) (err error) {
	var exist int64
	if err = db.Model(&models.User{}).Where("student_id = ?", studentID).Count(&exist).Error; err != nil {
		return err
	}
	if exist > 0 {
		return ErrorUserExist
	}
	return nil
}

func CreateUser(student *models.User) (err error) {
	if err = db.Create(student).Error; err != nil {
		return ErrorRegister
	}
	return nil
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

func ChangePassword(p *models.ParamPassword, studentID string) (err error) {
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
