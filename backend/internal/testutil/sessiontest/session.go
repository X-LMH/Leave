package sessiontest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"backend/internal/config"
	redisdao "backend/internal/dao/redis"
	"backend/internal/session"

	"github.com/redis/go-redis/v9"
)

// SessionRedis is an isolated command fake, never a connection to a business Redis.
type SessionRedis struct {
	mu     sync.Mutex
	Values map[string]string
	Expiry map[string]time.Time
	Err    error
}

func NewSessionRedis() *SessionRedis {
	return &SessionRedis{
		Values: map[string]string{},
		Expiry: map[string]time.Time{},
	}
}

func (r *SessionRedis) Set(_ context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return redis.NewStatusResult("", r.Err)
	}
	switch v := value.(type) {
	case []byte:
		r.Values[key] = string(v)
	case string:
		r.Values[key] = v
	default:
		return redis.NewStatusResult("", errors.New("unsupported fake value"))
	}
	if ttl <= 0 {
		r.Expiry[key] = time.Time{}
	} else {
		r.Expiry[key] = time.Now().Add(ttl)
	}
	return redis.NewStatusResult("OK", nil)
}

func (r *SessionRedis) Get(_ context.Context, key string) *redis.StringCmd {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return redis.NewStringResult("", r.Err)
	}
	value, ok := r.Values[key]
	if !ok || (!r.Expiry[key].IsZero() && !r.Expiry[key].After(time.Now())) {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(value, nil)
}

func (r *SessionRedis) Del(_ context.Context, keys ...string) *redis.IntCmd {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return redis.NewIntResult(0, r.Err)
	}
	var count int64
	for _, key := range keys {
		if _, ok := r.Values[key]; ok {
			count++
		}
		delete(r.Values, key)
		delete(r.Expiry, key)
	}
	return redis.NewIntResult(count, nil)
}

var sessionStores sync.Map

// Repeated route assembly within one test shares the same isolated sessions.
func NewStore(t *testing.T) *session.Store {
	t.Helper()
	store := session.NewStore(redisdao.NewDAO(NewSessionRedis(), "default"), config.SessionConfig{})
	actual, loaded := sessionStores.LoadOrStore(t, store)
	if !loaded {
		t.Cleanup(func() { sessionStores.Delete(t) })
	}
	return actual.(*session.Store)
}
