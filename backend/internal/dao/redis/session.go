package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrInvalidSessionTTL = errors.New("session TTL must be positive")
)

// SessionClient 只要求会话存储使用的命令，便于进行隔离测试。
type SessionClient interface {
	Set(context.Context, string, any, time.Duration) *redis.StatusCmd
	Get(context.Context, string) *redis.StringCmd
	Del(context.Context, ...string) *redis.IntCmd
}

// DAO 管理 Redis 中的会话数据，不决定用户身份、角色或有效期规则。
type DAO struct {
	client SessionClient
	prefix string
}

func NewDAO(client SessionClient, namespace string) *DAO {
	if namespace == "" {
		namespace = "default"
	}
	return &DAO{
		client: client,
		prefix: "leave:" + namespace + ":auth:session:",
	}
}

func (d *DAO) SaveSession(ctx context.Context, tokenHash string, payload []byte, ttl time.Duration) error {
	// Redis SET 的零时长会生成永久 Key，存储层也必须拒绝无效 TTL。
	if ttl <= 0 {
		return ErrInvalidSessionTTL
	}
	// 在调用方上下文上限制 Redis 操作时长，保留原有取消信号和更早的截止时间。
	redisCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return d.client.Set(redisCtx, d.prefix+tokenHash, payload, ttl).Err()
}

func (d *DAO) GetSession(ctx context.Context, tokenHash string) ([]byte, error) {
	redisCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	payload, err := d.client.Get(redisCtx, d.prefix+tokenHash).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	return payload, err
}

func (d *DAO) DeleteSession(ctx context.Context, tokenHash string) error {
	redisCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return d.client.Del(redisCtx, d.prefix+tokenHash).Err()
}
