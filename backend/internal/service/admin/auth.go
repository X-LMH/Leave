package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/session"
	passwordutil "backend/internal/utils/password"
	"context"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService struct {
	logger *zap.Logger
	dao    *mysql.DAO
	tokens *session.Store
}

func NewAuthService(dao *mysql.DAO, tokens *session.Store, logger *zap.Logger) *AuthService {
	return &AuthService{
		logger: logger,
		dao:    dao,
		tokens: tokens,
	}
}

func (a *AuthService) AdminLogin(ctx context.Context, req *dto.AdminLoginRequest) (*dto.AdminLoginResponse, error) {
	userName := strings.TrimSpace(req.UserName)
	user, err := a.dao.FindUserByStudentID(userName)
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

	token, err := a.tokens.Create(ctx, user, session.ClientWeb)
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

func (a *AuthService) ChangePassword(studentID string, input *dto.PasswordChangeRequest) error {
	if err := validatePasswordChange(input); err != nil {
		return err
	}
	return a.dao.Transaction(func(tx *gorm.DB) error {
		// Lock through verification and update so an old password can only succeed once.
		user, err := a.dao.GetUserForUpdate(tx, studentID)
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
		return a.dao.UpdateUserPassword(tx, user.ID, hash)
	})
}

func (a *AuthService) Logout(ctx context.Context, token string) error {
	return a.tokens.Revoke(ctx, token)
}
