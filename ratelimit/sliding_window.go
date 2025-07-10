package ratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type SlidingWindow struct {
	client *redis.Client
	key    string
	limit  int64         // 窗口内允许的最大请求数
	window time.Duration // 窗口大小
}

func NewSlidingWindow(client *redis.Client, key string, limit int64, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		client: client,
		key:    key,
		limit:  limit,
		window: window,
	}
}

func (sw *SlidingWindow) Allow(ctx context.Context) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - sw.window.Milliseconds()

	// 使用Redis有序集合实现滑动窗口
	script := `
	local key = KEYS[1]
	local now = tonumber(ARGV[1])
	local window_start = tonumber(ARGV[2])
	local limit = tonumber(ARGV[3])
	
	-- 移除窗口外的记录
	redis.call("ZREMRANGEBYSCORE", key, 0, window_start)
	
	-- 获取当前窗口内的请求数
	local current = redis.call("ZCARD", key)
	
	if current < limit then
		-- 添加当前请求
		redis.call("ZADD", key, now, now)
		-- 设置过期时间避免内存泄漏
		redis.call("EXPIRE", key, (ARGV[4] / 1000) + 1)
		return 1
	else
		return 0
	end
	`

	result, err := sw.client.Eval(ctx, script, []string{sw.key},
		now, windowStart, sw.limit, sw.window.Milliseconds()).Int64()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
