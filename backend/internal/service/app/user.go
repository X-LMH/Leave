package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/session"
	"backend/internal/utils/file"
	passwordutil "backend/internal/utils/password"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService struct {
	logger  *zap.Logger
	dao     *mysql.DAO
	tokens  *session.Store
	storage config.StorageConfig
}

func NewUserService(dao *mysql.DAO, tokens *session.Store, storage config.StorageConfig, logger *zap.Logger) *UserService {
	return &UserService{
		logger:  logger,
		dao:     dao,
		tokens:  tokens,
		storage: storage,
	}
}

func (u *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	studentID := strings.TrimSpace(req.StudentID)
	// 写入账号前准备资料；账号创建成功后才建立可用会话。
	data, err := u.newLoginResponse(dto.User{StudentID: studentID}, models.RoleStudent, false)
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
		AuthVersion:    1,
		Password:       hash,
		Role:           models.RoleStudent,
		Status:         models.UserStatusActive,
		LastSeenAt:     &now,
		LastSeenDevice: strings.TrimSpace(req.DeviceName),
		AppVersion:     strings.TrimSpace(req.AppVersion),
	}
	if err := u.dao.CreateUser(student); err != nil {
		return nil, err
	}
	token, err := u.tokens.Create(ctx, student, session.ClientApp)
	if err != nil {
		return nil, errors.Join(session.ErrRegistrationSession, err)
	}
	data.Token = token
	return data, nil
}

func (u *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	studentID := strings.TrimSpace(req.StudentID)

	user, err := u.dao.FindUserByStudentID(studentID)
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

	if err := u.dao.UpdateAppUsage(
		user.StudentID,
		strings.TrimSpace(req.DeviceName),
		strings.TrimSpace(req.AppVersion),
	); err != nil {
		return nil, err
	}

	loginUser, profileCompleted, err := u.buildLoginUser(user.StudentID)
	if err != nil {
		return nil, err
	}

	data, err := u.newLoginResponse(loginUser, user.Role, profileCompleted)
	if err != nil {
		return nil, err
	}
	token, err := u.tokens.Create(ctx, user, session.ClientApp)
	if err != nil {
		return nil, err
	}
	data.Token = token
	return data, nil
}

func (u *UserService) newLoginResponse(user dto.User, role string, profileCompleted bool) (*dto.LoginResponse, error) {
	// 登录与注册沿用资料接口的头像地址规则。
	avatarPath := user.AvatarURL
	if avatarPath == "" {
		avatarPath = service.DefaultAvatarPath
	}
	avatarURL, err := file.AccessURL(u.storage.BaseURL, avatarPath)
	if err != nil {
		return nil, err
	}
	user.AvatarURL = avatarURL
	user.Role = role
	return &dto.LoginResponse{
		User:             user,
		ProfileCompleted: profileCompleted,
	}, nil
}

func (u *UserService) buildLoginUser(studentID string) (dto.User, bool, error) {
	loginUser := dto.User{
		StudentID: studentID,
	}

	profile, err := u.dao.GetProfileByStuID(studentID)
	if err != nil {
		return dto.User{}, false, err
	}
	if profile == nil {
		return loginUser, false, nil
	}

	loginUser.Name = profile.Name
	loginUser.AvatarURL = profile.AvatarURL

	class, err := u.dao.GetClassByID(profile.ClassID)
	if err != nil {
		return dto.User{}, false, err
	}
	if class != nil {
		loginUser.ClassName = class.ClassName
	}

	return loginUser, true, nil
}

func (u *UserService) ChangePassword(p *dto.PasswordChangeRequest, studentID string) error {
	return u.dao.Transaction(func(tx *gorm.DB) error {
		user, err := u.dao.GetUserForUpdate(tx, studentID)
		if err != nil {
			return err
		}
		if user.Status != models.UserStatusActive {
			return mysql.ErrorUserDisabled
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
		return u.dao.UpdateUserPassword(tx, user.ID, hash)
	})
}

func (u *UserService) UpdateAppInfo(p *dto.ClientInfoRequest, studentID string) error {
	return u.dao.UpdateAppUsage(studentID, strings.TrimSpace(p.DeviceName), strings.TrimSpace(p.AppVersion))
}

func (u *UserService) Logout(ctx context.Context, token string) error {
	return u.tokens.Revoke(ctx, token)
}
