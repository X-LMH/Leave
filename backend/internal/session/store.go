// Package session manages opaque, server-side login sessions.
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"backend/internal/config"
	redisdao "backend/internal/dao/redis"
	"backend/internal/models"
)

// 登录来源独立于账号角色，管理员也可以通过 App 建立会话。
const (
	AuthorizationHeader = "Authorization"
	ClientApp           = "app"
	ClientWeb           = "web"
)

var (
	ErrInvalid             = errors.New("invalid session")
	ErrUnavailable         = errors.New("authentication unavailable")
	ErrRegistrationSession = errors.New("account created but session unavailable")
)

type Record struct {
	UserID      uint      `json:"user_id"`
	StudentID   string    `json:"student_id"`
	Role        string    `json:"role"`
	AuthVersion uint64    `json:"auth_version"`
	Client      string    `json:"client"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type Store struct {
	dao        *redisdao.DAO
	studentTTL time.Duration
	adminTTL   time.Duration
}

func NewStore(dao *redisdao.DAO, cfg config.SessionConfig) *Store {
	if cfg.StudentTTL <= 0 {
		cfg.StudentTTL = 720 * time.Hour
	}
	if cfg.AdminTTL <= 0 {
		cfg.AdminTTL = 8 * time.Hour
	}
	return &Store{
		dao:        dao,
		studentTTL: cfg.StudentTTL,
		adminTTL:   cfg.AdminTTL,
	}
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func (s *Store) Create(ctx context.Context, user *models.User, client string) (string, error) {
	if user.ID == 0 || user.StudentID == "" || user.AuthVersion == 0 ||
		(user.Role != models.RoleStudent && user.Role != models.RoleAdmin) ||
		(client != ClientApp && client != ClientWeb) {
		return "", ErrInvalid
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", errors.Join(ErrUnavailable, err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	ttl := s.studentTTL
	if user.Role == models.RoleAdmin {
		ttl = s.adminTTL
	}
	now := time.Now()
	record := Record{
		UserID:      user.ID,
		StudentID:   user.StudentID,
		Role:        user.Role,
		AuthVersion: user.AuthVersion,
		Client:      client,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return "", errors.Join(ErrUnavailable, err)
	}
	// SET 的零/负时长不能被误用为永久存储。
	remaining := time.Until(record.ExpiresAt)
	if remaining <= 0 {
		return "", errors.Join(ErrUnavailable, errors.New("session expired before storage"))
	}
	if err := s.dao.SaveSession(ctx, hashToken(token), payload, remaining); err != nil {
		return "", errors.Join(ErrUnavailable, err)
	}
	return token, nil
}

func validToken(token string) bool {
	if len(token) != 43 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == token
}

func (s *Store) Get(ctx context.Context, token string) (*Record, error) {
	if !validToken(token) {
		return nil, ErrInvalid
	}
	raw, err := s.dao.GetSession(ctx, hashToken(token))
	if errors.Is(err, redisdao.ErrSessionNotFound) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	var record Record
	if json.Unmarshal(raw, &record) != nil || record.UserID == 0 || record.StudentID == "" || record.AuthVersion == 0 ||
		(record.Role != models.RoleStudent && record.Role != models.RoleAdmin) ||
		(record.Client != ClientApp && record.Client != ClientWeb) || record.CreatedAt.IsZero() ||
		!record.ExpiresAt.After(record.CreatedAt) || !record.ExpiresAt.After(time.Now()) {
		return nil, ErrInvalid
	}
	return &record, nil
}

func (s *Store) Revoke(ctx context.Context, token string) error {
	if !validToken(token) {
		return ErrInvalid
	}
	if err := s.dao.DeleteSession(ctx, hashToken(token)); err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	return nil
}

// Bearer accepts one credential only, never silently trims or combines tokens.
func Bearer(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || !validToken(parts[1]) {
		return "", ErrInvalid
	}
	return parts[1], nil
}
