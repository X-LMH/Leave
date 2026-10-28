package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestRateLimitRedisIntegration(t *testing.T) {
	addr := os.Getenv("RATE_LIMIT_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set RATE_LIMIT_TEST_REDIS_ADDR for independent loopback Redis Lua/concurrency/TTL coverage")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		t.Fatal("Redis integration tests require a numeric loopback address")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, MaxRetries: -1, ContextTimeoutEnabled: true})
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	limiter := NewRateLimiter(client, zap.NewNop())
	rules := limiter.rules
	rules[OperationLogin][DimensionAccount] = rateLimitRule{Limit: 2, WindowSeconds: 1}
	rules[OperationLogin][DimensionIP] = rateLimitRule{Limit: 2, WindowSeconds: 1}
	identity := fmt.Sprintf("rate-limit-test-%x", randomBytes(t))
	var keys []string
	for _, entry := range []struct{ side, operation, dimension, id string }{
		{"app", "login", "account", identity},
		{"app", "login", "account", identity + "-other"},
		{"admin", "login", "account", identity},
		{"app", "register", "account", identity},
		{"app", "login", "ip", identity},
		{"app", "login", "ip", identity + "-other"},
		{"app", "login", "account", identity + "-concurrent"},
	} {
		keys = append(keys, rateLimitKey(entry.side, entry.operation, entry.dimension, entry.id))
	}
	t.Cleanup(func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		if err := client.Del(cleanupCtx, keys...).Err(); err != nil {
			t.Error(err)
		}
	})
	check := func(side, operation, dimension, id, ip, token string) bool {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/test", nil)
		c.Request.RemoteAddr = net.JoinHostPort(ip, "1234")
		c.Request.Header.Set("Authorization", "Bearer "+token)
		return limiter.Check(c, side, operation, dimension, id)
	}
	for i := 0; i < 2; i++ {
		if !check("app", "login", "account", identity, "192.0.2.1", "one") {
			t.Fatal("early rejection")
		}
	}
	before, err := client.PTTL(ctx, keys[0]).Result()
	if err != nil || before <= 0 || before > time.Second {
		t.Fatalf("TTL=%v error=%v", before, err)
	}
	if check("app", "login", "account", identity, "192.0.2.2", "two") {
		t.Fatal("IP/token change bypassed account quota")
	}
	after, err := client.PTTL(ctx, keys[0]).Result()
	if err != nil || after > before {
		t.Fatalf("rejection extended TTL: %v -> %v (%v)", before, after, err)
	}
	count, err := client.Get(ctx, keys[0]).Int()
	if err != nil || count != 2 {
		t.Fatalf("rejection changed counter: %d (%v)", count, err)
	}
	for _, entry := range []struct{ side, operation, dimension, id string }{
		{"app", "login", "account", identity + "-other"},
		{"admin", "login", "account", identity},
		{"app", "register", "account", identity},
	} {
		if !check(entry.side, entry.operation, entry.dimension, entry.id, "192.0.2.1", "one") {
			t.Fatal("independent quota rejected")
		}
	}
	for i := 0; i < 2; i++ {
		if !check("app", "login", "ip", identity, "192.0.2.1", fmt.Sprint(i)) {
			t.Fatal("early IP rejection")
		}
	}
	if check("app", "login", "ip", identity, "192.0.2.1", "new-account") {
		t.Fatal("account change bypassed IP quota")
	}
	if !check("app", "login", "ip", identity+"-other", "192.0.2.2", "one") {
		t.Fatal("independent IP rejected")
	}
	time.Sleep(time.Second + 20*time.Millisecond)
	if !check("app", "login", "account", identity, "192.0.2.1", "one") {
		t.Fatal("expired window did not recover")
	}
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if check("app", "login", "account", identity+"-concurrent", "192.0.2.1", "one") {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 2 {
		t.Fatalf("concurrent allowed=%d, want 2", allowed.Load())
	}
}

func randomBytes(t *testing.T) []byte {
	t.Helper()
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		t.Fatal(err)
	}
	return value
}
