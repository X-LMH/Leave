package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	passwordutil "backend/internal/utils/password"
	"strings"

	"gorm.io/gorm"
)

func AdminLogin(req *dto.AdminLoginRequest) (*dto.AdminLoginResponse, error) {
	userName := strings.TrimSpace(req.UserName)
	user, err := mysql.FindUserByStudentID(userName)
	if err != nil {
		return nil, err
	}
	if user.Role != models.RoleAdmin || user.Status != models.UserStatusActive {
		return nil, mysql.ErrorInvalidPassword
	}
	matched, err := passwordutil.Verify(user.Password, req.Password)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, mysql.ErrorInvalidPassword
	}

	token, err := jwt.GenerateToken(user.StudentID, user.Role)
	if err != nil {
		return nil, err
	}
	return &dto.AdminLoginResponse{Token: token}, nil
}

func validatePasswordChange(input *dto.PasswordChangeRequest) error {
	if input == nil || input.Password == "" || len(input.NewPassword) < 6 || len(input.NewPassword) > 30 ||
		strings.TrimSpace(input.NewPassword) == "" || input.NewPassword != input.RePassword ||
		input.NewPassword == input.Password {
		return ErrorInvalidAdminInput
	}
	return nil
}

func ChangePassword(studentID string, input *dto.PasswordChangeRequest) error {
	if err := validatePasswordChange(input); err != nil {
		return err
	}
	return mysql.Transaction(func(tx *gorm.DB) error {
		// Lock through verification and update so an old password can only succeed once.
		user, err := mysql.GetUserForUpdate(tx, studentID)
		if err != nil {
			return err
		}
		if user.Role != models.RoleAdmin {
			return mysql.ErrorUserNotExist
		}
		if user.Status != models.UserStatusActive {
			return mysql.ErrorUserDisabled
		}
		matched, err := passwordutil.Verify(user.Password, input.Password)
		if err != nil {
			return err
		}
		if !matched {
			return mysql.ErrorNotRightPassword
		}
		hash, err := passwordutil.Hash(input.NewPassword)
		if err != nil {
			return err
		}
		return mysql.UpdateUserPassword(tx, user.ID, hash)
	})
}
