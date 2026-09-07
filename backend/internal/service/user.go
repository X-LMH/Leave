package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	"strings"
)

func Register(req *dto.RegisterRequest) (err error) {
	studentID := strings.TrimSpace(req.StudentID)
	student := &models.User{
		StudentID: studentID,
		Password:  req.Password,
		Role:      models.RoleStudent,
		Status:    models.UserStatusActive,
	}
	return mysql.CreateUser(student)
}

func Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	studentID := strings.TrimSpace(req.StudentID)
	user, err := mysql.FindUserByStudentID(studentID)
	if err != nil {
		return nil, err
	}
	if req.Password != user.Password {
		return nil, mysql.ErrorInvalidPassword
	}
	if user.Status != models.UserStatusActive {
		return nil, mysql.ErrorInvalidPassword
	}
	if err := mysql.UpdateLastLoginAt(user.StudentID); err != nil {
		return nil, err
	}
	token, err := jwt.GenerateToken(user.StudentID, user.Role)
	if err != nil {
		return nil, err
	}

	profile, err := mysql.GetProfileByStuID(user.StudentID)
	if err != nil {
		return nil, err
	}
	name := ""
	if profile != nil {
		name = profile.Name
	}
	return &dto.LoginResponse{Token: token, Name: name}, nil
}

func ChangePassword(p *dto.PasswordChangeRequest, studentID string) error {
	return mysql.ChangePassword(p, studentID)
}
