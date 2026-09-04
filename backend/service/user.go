package service

import (
	"backend/dao/mysql"
	"backend/models"
	"backend/utils/jwt"
)

func Register(p *models.ParamRegister) (err error) {
	// 用户已存在
	if err = mysql.CheckUserExist(p.StudentID); err != nil {
		return err
	}
	student := &models.User{
		StudentID: p.StudentID,
		Password:  p.Password,
	}
	return mysql.CreateUser(student)
}

func Login(p *models.ParamLogin) (string, error) {
	user, err := mysql.FindUserByStudentID(p.StudentID)
	if err != nil {
		return "", err
	}
	if p.Password != user.Password {
		return "", mysql.ErrorInvalidPassword
	}
	return jwt.GenerateToken(user.StudentID, user.Role)
}

func ChangePassword(p *models.ParamPassword, studentID string) error {
	return mysql.ChangePassword(p, studentID)
}
