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