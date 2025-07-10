package ratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type TokenBucket struct {
	client   *redis.Client
	key      string
	capacity int64   // 桶容量
	rate     float64 // 令牌生成速率(个/秒)
}

func NewTokenBucket(client *redis.Client, key string, capacity int64, rate float64) *TokenBucket {
	return &TokenBucket{
		client:   client,
		key:      key,
		capacity: capacity,
		rate:     rate,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, tokens int64) (bool, error) {
	now := time.Now().UnixNano()
	// 使用Lua脚本保证原子性
	script := `
	local key = KEYS[1]
	local now = tonumber(ARGV[1])
	local tokens_requested = tonumber(ARGV[2])
	local capacity = tonumber(ARGV[3])
	local rate = tonumber(ARGV[4])
	
	local last_time = redis.call("HGET", key, "last_time")
	local tokens_available = redis.call("HGET", key, "tokens")
	
	if not last_time or not tokens_available then
		-- 初始化桶
		redis.call("HMSET", key, "last_time", now, "tokens", capacity)
		last_time = now
		tokens_available = capacity
	else
		-- 计算新增令牌
		last_time = tonumber(last_time)
		tokens_available = tonumber(tokens_available)
		local elapsed = (now - last_time) / 1e9
		local new_tokens = elapsed * rate
		tokens_available = math.min(tokens_available + new_tokens, capacity)
	end
	
	-- 检查是否有足够令牌
	if tokens_available >= tokens_requested then
		tokens_available = tokens_available - tokens_requested
		redis.call("HMSET", key, "last_time", now, "tokens", tokens_available)
		return 1
	else
		return 0
	end
	`

	result, err := tb.client.Eval(ctx, script, []string{tb.key},
		now, tokens, tb.capacity, tb.rate).Int64()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
