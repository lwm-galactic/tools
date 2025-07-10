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