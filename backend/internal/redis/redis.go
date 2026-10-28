package redis

import (
	"backend/internal/config"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Open 创建独立客户端，启动时检查连接，调用方负责关闭。
func Open(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		MaxRetries:            -1, // Do not replay counter scripts after a lost reply.
		Addr:                  net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Username:              cfg.Username,
		Password:              cfg.Password,
		DB:                    cfg.DB,
		DialTimeout:           5 * time.Second,
		ReadTimeout:           3 * time.Second,
		WriteTimeout:          3 * time.Second,
		ContextTimeoutEnabled: true,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Join(fmt.Errorf("redis ping failed: %w", err), client.Close())
	}
	return client, nil
}
