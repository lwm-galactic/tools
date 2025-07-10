package redis

import (
	"encoding/base64"
	"errors"
	"github.com/buger/jsonparser"
	"github.com/redis/go-redis/v9"
	"strings"
)

// RedisCluster is a storage manager that uses the redis database.
type RedisCluster struct {
	KeyPrefix    string
	HashKeys     bool
	IsCache      bool
	hashIdentity ExtractHashIdentity
}

// ErrRedisIsDown is returned when we can't communicate with redis.
var ErrRedisIsDown = errors.New("storage: Redis is either down or ws not configured")

type defaultExtractHashIdentity struct{}

// B64JSONPrefix stand for `{"` in base64.
const B64JSONPrefix = "ey"

func (defaultExtractHashIdentity) ExtractHashIdentity(token string) string {
	// Legacy tokens not b64 and not JSON records
	if strings.HasPrefix(token, B64JSONPrefix) {
		if jsonToken, err := base64.StdEncoding.DecodeString(token); err == nil {
			hashAlgo, _ := jsonparser.GetString(jsonToken, "h")

			return hashAlgo
		}
	}

	return ""
}

// Connected returns true if we are connected to redis.
func Connected() bool {
	if v := redisUp.Load(); v != nil {
		return v.(bool)
	}

	return false
}

func (r *RedisCluster) SetExtractHashIdentity(hashIdentity ExtractHashIdentity) {
	r.hashIdentity = hashIdentity
}

// Connect will establish a connection this is always true because we are dynamically using redis.
func (r *RedisCluster) Connect() bool {
	return true
}
func (r *RedisCluster) hashKey(in string) string {
	if !r.HashKeys {
		// Not hashing? Return the raw key
		return in
	}

	return HashStr(in, r.hashIdentity)
}

func (r *RedisCluster) fixKey(keyName string) string {
	return r.KeyPrefix + r.hashKey(keyName)
}

func (r *RedisCluster) cleanKey(keyName string) string {
	return strings.Replace(keyName, r.KeyPrefix, "", 1)
}

func (r *RedisCluster) up() error {
	if !Connected() {
		return ErrRedisIsDown
	}

	return nil
}

func (r *RedisCluster) singleton() redis.UniversalClient {
	return singleton(r.IsCache)
}
