package session_test

import (
	"backend/internal/config"
	redisdao "backend/internal/dao/redis"
	"backend/internal/models"
	"backend/internal/session"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisSessionIntegration(t *testing.T) {
	addr := os.Getenv("SESSION_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set SESSION_TEST_REDIS_ADDR for isolated loopback Redis session coverage")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("test Redis requires a numeric loopback address")
	}
	client := redis.NewClient(&redis.Options{
		Addr:                  addr,
		MaxRetries:            -1,
		ContextTimeoutEnabled: true,
	})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	namespace := fmt.Sprintf("session-test-%d", time.Now().UnixNano())
	s := session.NewStore(redisdao.NewDAO(client, namespace), config.SessionConfig{
		StudentTTL: time.Minute,
	})
	token, err := s.Create(ctx, &models.User{
		ID:          1,
		StudentID:   "isolated",
		Role:        models.RoleStudent,
		AuthVersion: 1,
	}, session.ClientApp)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(token))
	key := "leave:" + namespace + ":auth:session:" + hex.EncodeToString(digest[:])
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
		defer cleanupCancel()
		if err := client.Del(cleanupCtx, key).Err(); err != nil {
			t.Error(err)
		}
	})
	ttl, err := client.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl <= 0 || ttl > time.Minute {
		t.Fatal("Redis TTL", ttl)
	}
	if _, err := s.Get(ctx, token); err != nil {
		t.Fatal(err)
	}
	if err := s.Revoke(ctx, token); err != nil {
		t.Fatal(err)
	}
	if err := s.Revoke(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, token); !errors.Is(err, session.ErrInvalid) {
		t.Fatal("logout replay accepted", err)
	}
}
