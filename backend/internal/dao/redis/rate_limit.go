package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// counterScript 通过 Lua 原子执行固定窗口限流，避免并发请求超额放行。
// 首次请求创建计数器并设置窗口过期时间，后续请求只递增计数，达到上限后拒绝放行。
// 返回是否放行（1/0）及窗口剩余时间（毫秒），供调用方处理限流响应。
const counterScript = `
local count = tonumber(redis.call('GET', KEYS[1]) or '0')
local ttl = redis.call('PTTL', KEYS[1])
if count >= tonumber(ARGV[1]) then
 return {0, ttl}
end
if count == 0 then
 redis.call('SET', KEYS[1], 1, 'PX', ARGV[2])
 ttl = tonumber(ARGV[2])
else
 redis.call('INCR', KEYS[1])
end
return {1, ttl}
`

// ConsumeRateLimit 原子读取和更新计数器；限流规则及 HTTP 响应由调用方处理。
func ConsumeRateLimit(ctx context.Context, client *redis.Client, key string, limit, windowMilliseconds int64) ([]int64, error) {
	// EVAL 避免脚本缓存未命中时多一次往返，客户端禁用重试以避免重复扣减。
	return client.Eval(ctx, counterScript, []string{key}, limit, windowMilliseconds).Int64Slice()
}
