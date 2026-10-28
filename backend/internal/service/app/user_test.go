package app

import (
	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	"testing"

	"go.uber.org/zap"
)

func TestNewLoginResponse(t *testing.T) {
	tokens := jwt.NewManager("test-registration-session", 1)
	users := NewUserService(nil, tokens, config.StorageConfig{BaseURL: "https://example.com/uploads"}, zap.NewNop())
	for _, role := range []string{models.RoleStudent, models.RoleAdmin} {
		for _, completed := range []bool{false, true} {
			user := dto.User{StudentID: "202410201016"}
			if completed {
				user.Name = "张三"
				user.ClassName = "软件工程1班"
				user.AvatarURL = "avatars/student.jpg"
			}
			data, err := users.newLoginResponse(user, role, completed)
			if err != nil {
				t.Fatal(err)
			}
			claims, err := tokens.ParseToken(data.Token)
			if err != nil {
				t.Fatalf("returned token is unusable: %v", err)
			}
			if claims.StudentID != user.StudentID || claims.Role != role {
				t.Fatalf("wrong session identity: %#v", claims)
			}
			expected := user
			expected.Role = role
			expected.AvatarURL = "https://example.com/uploads/avatars/default.png"
			if completed {
				expected.AvatarURL = "https://example.com/uploads/avatars/student.jpg"
			}
			if data.User != expected || data.ProfileCompleted != completed {
				t.Fatalf("incorrect session data: %#v", data)
			}
		}
	}
}
