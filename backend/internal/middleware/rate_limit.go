package middleware

import (
	"backend/internal/dto"
	requestlog "backend/internal/logger"
	"backend/internal/request"
	"backend/internal/response"
	"context"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

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

// RateLimiter shares Redis counters across processes and HTTP handlers.
type RateLimiter struct {
	client *redis.Client
	rules  map[string]map[string]rateLimitRule
	logger *zap.Logger
}

func NewRateLimiter(client *redis.Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		client: client,
		rules:  defaultRateLimits(),
		logger: logger,
	}
}

func rateLimitKey(side, operation, dimension, identity string) string {
	identity = strings.TrimSpace(identity)
	if dimension == DimensionIP {
		if addr, err := netip.ParseAddr(identity); err == nil {
			identity = addr.Unmap().String()
		}
	}
	return fmt.Sprintf("leave:%s:%s:%s:%s", side, operation, dimension, identity)
}

// Check consumes one dimension. Earlier dimensions are never refunded.
func (l *RateLimiter) Check(c *gin.Context, side, operation, dimension, identity string) bool {
	rule, ok := l.rules[operation][dimension]
	log := l.logger.With(requestlog.RequestFields(c)...).With(
		zap.String("operation", side+"."+operation),
		zap.String("dimension", dimension),
	)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
	defer cancel()

	var result []int64
	var err error
	if !ok || rule.Limit <= 0 || rule.WindowSeconds <= 0 {
		err = fmt.Errorf("missing or invalid rate limit rule: operation=%q dimension=%q", operation, dimension)
	} else {
		// EVAL avoids a separate script-cache miss round trip. The shared client disables retries.
		result, err = l.client.Eval(ctx, counterScript,
			[]string{rateLimitKey(side, operation, dimension, identity)},
			rule.Limit, rule.WindowSeconds*1000).Int64Slice()
	}
	if err != nil || len(result) != 2 || (result[0] != 0 && result[0] != 1) || result[1] < 0 {
		if err == nil {
			err = fmt.Errorf("invalid rate limit script result")
		}
		log.Error("限流检查失败", requestlog.Error(err))
		response.Error(c, response.CodeRateLimitUnavailable)
		c.Abort()
		return false
	}
	if result[0] == 1 {
		return true
	}
	seconds := max((result[1]+999)/1000, 1)
	c.Header("Retry-After", strconv.FormatInt(seconds, 10))
	log.Info("请求触发限流", zap.Int64("retry_after", seconds))
	response.ErrorWithData(c, response.CodeRateLimited, dto.RateLimitResponse{RetryAfter: seconds})
	c.Abort()
	return false
}

func (l *RateLimiter) IP(side, operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if l.Check(c, side, operation, DimensionIP, c.ClientIP()) {
			c.Next()
		}
	}
}

func (l *RateLimiter) Account(side, operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := request.GetCurrentStuID(c)
		if err != nil {
			response.Error(c, response.CodeNeedLogin)
			c.Abort()
			return
		}
		if l.Check(c, side, operation, DimensionAccount, identity) {
			c.Next()
		}
	}
}
