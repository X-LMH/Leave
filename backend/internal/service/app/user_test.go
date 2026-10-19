package app

import (
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	"testing"
)

func TestNewLoginResponse(t *testing.T) {
	tokens := jwt.NewManager("test-registration-session", 1)
	users := NewUserService(nil, tokens)
	for _, completed := range []bool{false, true} {
		user := dto.User{StudentID: "202410201016"}
		if completed {
			user.Name = "张三"
			user.ClassName = "软件工程1班"
		}
		data, err := users.newLoginResponse(user, models.RoleStudent, completed)
		if err != nil {
			t.Fatal(err)
		}
		claims, err := tokens.ParseToken(data.Token)
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
