package redis

import (
	"context"
	"github.com/lwm-galactic/logger"
	"github.com/lwm-galactic/tools/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/atomic"
	"strconv"
	"time"
)

// 全局变量定义
var (
	singlePool      atomic.Value // 普通连接池
	singleCachePool atomic.Value // 缓存专用连接池
	redisUp         atomic.Value // Redis可用状态
	disableRedis    atomic.Value // 手动禁用开关

	// 熔断和重试相关变量
	retryBackoff   atomic.Int32 // 当前退避指数
	failureCount   atomic.Int32 // 连续失败次数
	lastFailure    atomic.Value // 最近失败时间
	circuitTripped atomic.Bool  // 熔断状态
	lastSuccess    atomic.Value // 最近成功时间
)

// 常量定义
var (
	maxBackoff             = 32 * time.Second // 最大退避时间
	circuitTimeout         = 5 * time.Minute  // 熔断超时时间
	failureThreshold int32 = 5                // 触发熔断的失败次数
)

func singleton(cache bool) redis.UniversalClient {
	if cache {
		v := singleCachePool.Load()
		if v != nil {
			return v.(redis.UniversalClient)
		}

		return nil
	}
	if v := singlePool.Load(); v != nil {
		return v.(redis.UniversalClient)
	}

	return nil
}

// nolint: unparam
func connectSingleton(cache bool, config *Config) bool {
	if singleton(cache) == nil {
		logger.Debug("Connecting to redis cluster")
		if config != nil && config.RetryTime != 0 {
			failureThreshold = config.RetryTime
		}

		if cache {
			singleCachePool.Store(NewRedisClusterPool(cache, config))

			return true
		}
		singlePool.Store(NewRedisClusterPool(cache, config))

		return true
	}

	return true
}

func shouldConnect() bool {
	if v := disableRedis.Load(); v != nil {
		return !v.(bool)
	}

	return true
}

// clusterConnectionIsOpen 检查Redis连接健康状态
func clusterConnectionIsOpen(cluster RedisCluster) bool {
	c := singleton(cluster.IsCache)
	if c == nil {
		return false
	}

	// 使用上下文控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testKey := "redis-test-" + strconv.FormatUint(uuid.MustGet(0), 10)

	// 测试写入和读取
	if err := c.Set(ctx, testKey, "test", time.Second).Err(); err != nil {
		logger.Warnf("Redis SET测试失败: %s", err.Error())
		return false
	}

	if _, err := c.Get(ctx, testKey).Result(); err != nil {
		logger.Warnf("Redis GET测试失败: %s", err.Error())
		return false
	}

	return true
}

// ConnectToRedis 主连接管理函数
func ConnectToRedis(ctx context.Context, config *Config) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	c := []RedisCluster{{}, {IsCache: true}}

	// 初始连接
	initializeConnections(c, config)

	// 主循环
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if !shouldProceed() {
				continue
			}

			if checkConnections(ctx, c, config) {
				// 连接健康
				resetFailureState()
			}
		}
	}
}

// initializeConnections 初始化连接
func initializeConnections(c []RedisCluster, config *Config) {
	for _, v := range c {
		if !connectSingleton(v.IsCache, config) || !clusterConnectionIsOpen(v) {
			redisUp.Store(false)
			recordFailure()
			return
		}
	}
	redisUp.Store(true)
	lastSuccess.Store(time.Now())
}

// shouldProceed 判断是否继续检查
func shouldProceed() bool {
	if !shouldConnect() {
		return false
	}

	// 检查熔断状态
	if circuitTripped.Load() {
		if time.Since(lastFailure.Load().(time.Time)) >= circuitTimeout {
			circuitTripped.Store(false)
			failureCount.Store(0)
			return true
		}
		return false
	}
	return true
}

// checkConnections 检查所有连接
func checkConnections(ctx context.Context, c []RedisCluster, config *Config) bool {
	allHealthy := true

	for _, v := range c {
		select {
		case <-ctx.Done():
			return false
		default:
			if !checkSingleConnection(v, config) {
				allHealthy = false
				// 单个连接失败时立即处理
				handleConnectionFailure()
				break
			}
		}
	}

	return allHealthy
}

// checkSingleConnection 检查单个连接
func checkSingleConnection(v RedisCluster, config *Config) bool {
	if !connectSingleton(v.IsCache, config) {
		return false
	}
	return clusterConnectionIsOpen(v)
}

// handleConnectionFailure 处理连接失败
func handleConnectionFailure() {
	redisUp.Store(false)
	recordFailure()

	// 检查是否需要触发熔断
	if failureCount.Load() >= failureThreshold &&
		time.Since(lastFailure.Load().(time.Time)) < circuitTimeout {
		triggerCircuitBreaker()
		return
	}

	// 执行指数退避
	performBackoff()
}

// recordFailure 记录失败状态
func recordFailure() {
	failureCount.Inc()
	lastFailure.Store(time.Now())
}

// triggerCircuitBreaker 触发熔断
func triggerCircuitBreaker() {
	circuitTripped.Store(true)
	logger.Error("Redis熔断触发，暂停操作5分钟")
	time.Sleep(circuitTimeout)
	circuitTripped.Store(false)
	failureCount.Store(0)
}

// performBackoff 执行指数退避
func performBackoff() {
	backoff := time.Duration(1<<min(retryBackoff.Load(), 5)) * time.Second
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	time.Sleep(backoff)
	retryBackoff.Inc()
}

// resetFailureState 重置失败状态
func resetFailureState() {
	redisUp.Store(true)
	retryBackoff.Store(0)
	failureCount.Store(0)
	lastSuccess.Store(time.Now())
}

// helper functions
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
