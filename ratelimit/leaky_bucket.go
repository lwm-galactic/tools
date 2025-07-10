package ratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type LeakyBucket struct {
	client   *redis.Client
	key      string
	capacity int64   // 桶容量
	rate     float64 // 漏出速率(个/秒)
}

func NewLeakyBucket(client *redis.Client, key string, capacity int64, rate float64) *LeakyBucket {
	return &LeakyBucket{
		client:   client,
		key:      key,
		capacity: capacity,
		rate:     rate,
	}
}

func (lb *LeakyBucket) Allow(ctx context.Context) (bool, time.Duration, error) {
	now := time.Now().UnixNano()
	script := `
	local key = KEYS[1]
	local now = tonumber(ARGV[1])
	local capacity = tonumber(ARGV[2])
	local rate = tonumber(ARGV[3])
	
	local last_time = redis.call("HGET", key, "last_time")
	local water = redis.call("HGET", key, "water")
	
	if not last_time or not water then
		-- 初始化桶
		redis.call("HMSET", key, "last_time", now, "water", 1)
		return {1, 0}
	else
		-- 计算漏出水量
		last_time = tonumber(last_time)
		water = tonumber(water)
		local elapsed = (now - last_time) / 1e9
		local leaked = elapsed * rate
		water = math.max(water - leaked, 0)
		
		-- 检查是否有空间
		if water < capacity then
			water = water + 1
			redis.call("HMSET", key, "last_time", now, "water", water)
			return {1, 0}
		else
			-- 计算需要等待的时间
			local wait_time = (water - capacity + 1) / rate
			return {0, wait_time}
		end
	end
	`

	result, err := lb.client.Eval(ctx, script, []string{lb.key},
		now, lb.capacity, lb.rate).Float64Slice()
	if err != nil {
		return false, 0, err
	}

	allowed := result[0] == 1
	waitSeconds := result[1]
	return allowed, time.Duration(waitSeconds * float64(time.Second)), nil
}
