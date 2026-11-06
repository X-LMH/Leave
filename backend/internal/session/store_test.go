package session_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/config"
	redisdao "backend/internal/dao/redis"
	"backend/internal/models"
	"backend/internal/session"
	"backend/internal/testutil/sessiontest"
	"backend/internal/utils/jwt"
)

func TestSessionLifecycle(t *testing.T) {
	ctx := context.Background()
	r := sessiontest.NewSessionRedis()
	s := session.NewStore(redisdao.NewDAO(r, "isolated"), config.SessionConfig{})
	user := &models.User{
		ID:          7,
		StudentID:   "student",
		Role:        models.RoleStudent,
		AuthVersion: 3,
	}
	first, err := s.Create(ctx, user, session.ClientApp)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Create(ctx, user, session.ClientApp)
	if err != nil || first == second {
		t.Fatal("independent sessions required")
	}
	digest := sha256.Sum256([]byte(first))
	key := "leave:isolated:auth:session:" + hex.EncodeToString(digest[:])
	if r.Values[key] == "" || strings.Contains(r.Values[key], first) {
		t.Fatal("raw token persisted or wrong key")
	}
	if ttl := time.Until(r.Expiry[key]); ttl < 719*time.Hour || ttl > 720*time.Hour {
		t.Fatal("student TTL", ttl)
	}
	record, err := s.Get(ctx, first)
	if err != nil || record.UserID != 7 || record.StudentID != "student" || record.AuthVersion != 3 || record.Client != session.ClientApp {
		t.Fatal(record, err)
	}
	if err := s.Revoke(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err := s.Revoke(ctx, first); err != nil {
		t.Fatal("logout must be idempotent", err)
	}
	if _, err := s.Get(ctx, first); !errors.Is(err, session.ErrInvalid) {
		t.Fatal("revoked token accepted", err)
	}
	if _, err := s.Get(ctx, second); err != nil {
		t.Fatal("other session revoked", err)
	}
	user.Role = models.RoleAdmin
	admin, err := s.Create(ctx, user, session.ClientWeb)
	if err != nil {
		t.Fatal(err)
	}
	record, err = s.Get(ctx, admin)
	if err != nil || record.ExpiresAt.Sub(record.CreatedAt) != 8*time.Hour {
		t.Fatal("admin TTL", err)
	}
}

func TestSessionRejectsInvalidDataAndUnavailableStorage(t *testing.T) {
	ctx := context.Background()
	r := sessiontest.NewSessionRedis()
	s := session.NewStore(redisdao.NewDAO(r, "default"), config.SessionConfig{})
	user := &models.User{
		ID:          1,
		StudentID:   "admin",
		Role:        models.RoleAdmin,
		AuthVersion: 1,
	}

	token, err := s.Create(ctx, user, session.ClientApp)
	if err != nil {
		t.Fatal(err)
	}
	var key, original string
	for k, value := range r.Values {
		key, original = k, value
	}
	for _, payload := range []string{"{", "{}", `{"user_id":1}`} {
		r.Values[key] = payload
		if _, err := s.Get(ctx, token); !errors.Is(err, session.ErrInvalid) {
			t.Fatal("corrupt data accepted", err)
		}
	}
	var expired session.Record
	if err := json.Unmarshal([]byte(original), &expired); err != nil {
		t.Fatal(err)
	}
	expired.CreatedAt = time.Now().Add(-2 * time.Hour)
	expired.ExpiresAt = time.Now().Add(-time.Hour)
	payload, err := json.Marshal(expired)
	if err != nil {
		t.Fatal(err)
	}
	r.Values[key] = string(payload)
	if _, err := s.Get(ctx, token); !errors.Is(err, session.ErrInvalid) {
		t.Fatal("expired data accepted", err)
	}
	r.Values[key] = original
	r.Expiry[key] = time.Now().Add(-time.Second)
	if _, err := s.Get(ctx, token); !errors.Is(err, session.ErrInvalid) {
		t.Fatal("expired Redis key accepted", err)
	}
	legacy, err := jwt.NewManager("test-only", 1).GenerateToken("admin", "admin")
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{"", "bad", legacy, token + "extra"} {
		if _, err := s.Get(ctx, invalid); !errors.Is(err, session.ErrInvalid) {
			t.Fatal("invalid token accepted")
		}
		if _, err := session.Bearer("Bearer " + invalid); !errors.Is(err, session.ErrInvalid) {
			t.Fatal("invalid header accepted")
		}
	}
	r.Err = errors.New("Redis offline")
	if _, err := s.Get(ctx, token); !errors.Is(err, session.ErrUnavailable) {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, user, session.ClientApp); !errors.Is(err, session.ErrUnavailable) {
		t.Fatal(err)
	}
	if err := s.Revoke(ctx, token); !errors.Is(err, session.ErrUnavailable) {
		t.Fatal(err)
	}
}

func TestShortSessionsNeverCreatePersistentKeys(t *testing.T) {
	r := sessiontest.NewSessionRedis()
	s := session.NewStore(redisdao.NewDAO(r, "default"), config.SessionConfig{
		StudentTTL: time.Nanosecond,
	})
	token, err := s.Create(context.Background(), &models.User{
		ID:          1,
		StudentID:   "student",
		Role:        models.RoleStudent,
		AuthVersion: 1,
	}, session.ClientApp)
	// Windows 时钟可能尚未前进，短会话允许成功或因到期被拒绝，但不能永久存储。
	if err != nil {
		if token != "" || !errors.Is(err, session.ErrUnavailable) || len(r.Values) != 0 {
			t.Fatal("expired session was persisted", err)
		}
		return
	}
	if token == "" || len(r.Values) != 1 {
		t.Fatal("missing short session")
	}
	for _, expiresAt := range r.Expiry {
		if expiresAt.IsZero() {
			t.Fatal("session persisted without TTL")
		}
	}
}
