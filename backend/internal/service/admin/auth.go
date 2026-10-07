package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	"strings"
)

func AdminLogin(req *dto.AdminLoginRequest) (*dto.AdminLoginResponse, error) {
	userName := strings.TrimSpace(req.UserName)
	user, err := mysql.FindUserByStudentID(userName)
	if err != nil {
		return nil, err
	}
	if user.Role != models.RoleAdmin || user.Status != models.UserStatusActive || req.Password != user.Password {
		return nil, mysql.ErrorInvalidPassword
	}

	token, err := jwt.GenerateToken(user.StudentID, user.Role)
	if err != nil {
		return nil, err
	}
	return &dto.AdminLoginResponse{Token: token}, nil
}
