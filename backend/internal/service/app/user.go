package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	passwordutil "backend/internal/utils/password"
	"strings"
	"time"

	"gorm.io/gorm"
)

func Register(req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	studentID := strings.TrimSpace(req.StudentID)
	// 先准备会话，再一次性写入账号，避免签发失败时留下已创建的账号。
	data, err := newLoginResponse(dto.User{StudentID: studentID}, models.RoleStudent, false)
	if err != nil {
		return nil, err
	}
	hash, err := passwordutil.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	student := &models.User{
		StudentID:      studentID,
		Password:       hash,
		Role:           models.RoleStudent,
		Status:         models.UserStatusActive,
		LastSeenAt:     &now,
		LastSeenDevice: strings.TrimSpace(req.DeviceName),
		AppVersion:     strings.TrimSpace(req.AppVersion),
	}
	if err := mysql.CreateUser(student); err != nil {
		return nil, err
	}
	return data, nil
}

func Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	studentID := strings.TrimSpace(req.StudentID)

	user, err := mysql.FindUserByStudentID(studentID)
	if err != nil {
		return nil, err
	}

	if user.Status != models.UserStatusActive {
		return nil, mysql.ErrorInvalidPassword
	}
	matched, err := passwordutil.Verify(user.Password, req.Password)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, mysql.ErrorInvalidPassword
	}

	if err := mysql.UpdateAppUsage(
		user.StudentID,
		strings.TrimSpace(req.DeviceName),
		strings.TrimSpace(req.AppVersion),
	); err != nil {
		return nil, err
	}

	loginUser, profileCompleted, err := buildLoginUser(user.StudentID)
	if err != nil {
		return nil, err
	}

	return newLoginResponse(loginUser, user.Role, profileCompleted)
}

func newLoginResponse(user dto.User, role string, profileCompleted bool) (*dto.LoginResponse, error) {
	token, err := jwt.GenerateToken(user.StudentID, role)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		Token:            token,
		User:             user,
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
	return mysql.Transaction(func(tx *gorm.DB) error {
		user, err := mysql.GetUserForUpdate(tx, studentID)
		if err != nil {
			return err
		}
		matched, err := passwordutil.Verify(user.Password, p.Password)
		if err != nil {
			return err
		}
		if !matched {
			return mysql.ErrorNotRightPassword
		}
		hash, err := passwordutil.Hash(p.NewPassword)
		if err != nil {
			return err
		}
		return mysql.UpdateUserPassword(tx, user.ID, hash)
	})
}

func UpdateAppInfo(p *dto.ClientInfoRequest, studentID string) error {
	return mysql.UpdateAppUsage(studentID, strings.TrimSpace(p.DeviceName), strings.TrimSpace(p.AppVersion))
}
