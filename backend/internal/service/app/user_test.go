package app

import (
	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	"testing"
)

func TestNewLoginResponse(t *testing.T) {
	previous := config.Cfg
	t.Cleanup(func() { config.Cfg = previous })
	config.Cfg.JWT.Secret = "test-registration-session"
	config.Cfg.JWT.ExpirationDays = 1
	for _, completed := range []bool{false, true} {
		user := dto.User{StudentID: "202410201016"}
		if completed {
			user.Name = "张三"
			user.ClassName = "软件工程1班"
		}
		data, err := newLoginResponse(user, models.RoleStudent, completed)
		if err != nil {
			t.Fatal(err)
		}
		claims, err := jwt.ParseToken(data.Token)
		if err != nil {
			t.Fatalf("returned token is unusable: %v", err)
		}
		if claims.StudentID != user.StudentID || claims.Role != models.RoleStudent {
			t.Fatalf("wrong session identity: %#v", claims)
		}
		if data.User != user || data.ProfileCompleted != completed {
			t.Fatalf("incorrect session data: %#v", data)
		}
	}
}
