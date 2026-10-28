// Package testutil provides isolated infrastructure for integration tests.
package testutil

import (
	"backend/internal/middleware"
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// NewRateLimiter requires a dedicated loopback Redis, never a development tunnel.
func NewRateLimiter(t *testing.T) *middleware.RateLimiter {
	t.Helper()
	addr := os.Getenv("RATE_LIMIT_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set RATE_LIMIT_TEST_REDIS_ADDR for isolated loopback Redis coverage")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("test Redis requires a numeric loopback address")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, MaxRetries: -1, ContextTimeoutEnabled: true})
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal("test Redis unavailable: ", err)
	}
	hook := &isolatedKeys{prefix: fmt.Sprintf("test:%s:%d:", t.Name(), time.Now().UnixNano()), keys: make(map[string]struct{})}
	client.AddHook(hook)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for key := range hook.keys {
			if err := client.Del(ctx, key).Err(); err != nil {
				t.Error(err)
			}
		}
	})
	return middleware.NewRateLimiter(client, zap.NewNop())
}

// Prefix actual EVAL keys so concurrent tests cannot consume each other's quota.
type isolatedKeys struct {
	prefix string
	mu     sync.Mutex
	keys   map[string]struct{}
}

func (h *isolatedKeys) DialHook(next redis.DialHook) redis.DialHook { return next }
func (h *isolatedKeys) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (h *isolatedKeys) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if cmd.Name() == "eval" {
			args := cmd.Args()
			key := h.prefix + args[3].(string)
			args[3] = key
			h.mu.Lock()
			h.keys[key] = struct{}{}
			h.mu.Unlock()
		}
		return next(ctx, cmd)
	}
}
