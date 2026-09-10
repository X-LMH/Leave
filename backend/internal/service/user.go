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

	if req.Password != user.Password || user.Status != models.UserStatusActive {
		return nil, mysql.ErrorInvalidPassword
	}

	if err := mysql.UpdateAppUsage(
		user.StudentID,
		strings.TrimSpace(req.DeviceName),
		strings.TrimSpace(req.AppVersion),
	); err != nil {
		return nil, err
	}

	token, err := jwt.GenerateToken(user.StudentID, user.Role)
	if err != nil {
		return nil, err
	}

	loginUser, profileCompleted, err := buildLoginUser(user.StudentID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:            token,
		User:             loginUser,
		ProfileCompleted: profileCompleted,
	}, nil
}

func buildLoginUser(studentID string) (dto.User, bool, error) {
	loginUser := dto.User{
		StudentID: studentID,
	}

	profile, err := mysql.GetProfileByStuID(studentID)
	if err != nil {
		return dto.User{}, false, err
	}
	if profile == nil {
		return loginUser, false, nil
	}

	loginUser.Name = profile.Name

	class, err := mysql.GetClassByID(profile.ClassID)
	if err != nil {
		return dto.User{}, false, err
	}
	if class != nil {
		loginUser.ClassName = class.ClassName
	}

	return loginUser, true, nil
}

func ChangePassword(p *dto.PasswordChangeRequest, studentID string) error {
	return mysql.ChangePassword(p, studentID)
}

func UpdateAppInfo(p *dto.ClientInfoRequest, studentID string) error {
	return mysql.UpdateAppUsage(studentID, strings.TrimSpace(p.DeviceName), strings.TrimSpace(p.AppVersion))
}
