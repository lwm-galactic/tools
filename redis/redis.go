package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/lwm-galactic/tools/murmur3"
	"hash"
)

// ErrKeyNotFound is a standard error for when a key is not found in the storage engine.
var ErrKeyNotFound = errors.New("key not found")

// Handler is a standard interface to a storage backend, used by AuthorisationManager to read and write key values to
// the backend.
type Handler interface {
	GetKey(ctx context.Context, key string) (string, error) // Returned string is expected to be a JSON object (user.SessionState)
	GetMultiKey(ctx context.Context, keys []string) ([]string, error)
	GetRawKey(ctx context.Context, key string) (string, error)
	SetKey(ctx context.Context, key string, value string, ttl int64) error // Second input string is expected to be a JSON object (user.SessionState)
	SetRawKey(ctx context.Context, key string, value string, ttl int64) error
	SetExp(ctx context.Context, key string, ttl int64) error // Set key expiration
	GetExp(ctx context.Context, key string) (int64, error)   // Returns expiry of a key
	GetKeys(ctx context.Context, pattern string) []string
	DeleteKey(ctx context.Context, key string) bool
	DeleteAllKeys(ctx context.Context) bool
	DeleteRawKey(ctx context.Context, key string) bool
	Connect(ctx context.Context) bool
	GetKeysAndValues(ctx context.Context) map[string]string
	GetKeysAndValuesWithFilter(ctx context.Context, filter string) map[string]string
	DeleteKeys(ctx context.Context, keys []string) bool
	Decrement(ctx context.Context, key string)
	IncrememntWithExpire(ctx context.Context, key string, ttl int64) int64
	SetRollingWindow(ctx context.Context, key string, per int64, val string, pipeline bool) (int, []interface{})
	GetRollingWindow(ctx context.Context, key string, per int64, pipeline bool) (int, []interface{})
	GetSet(ctx context.Context, key string) (map[string]string, error)
	AddToSet(ctx context.Context, key string, value string)
	GetAndDeleteSet(ctx context.Context, key string) []interface{}
	RemoveFromSet(ctx context.Context, key string, value string)
	DeleteScanMatch(ctx context.Context, pattern string) bool
	GetKeyPrefix() string
	AddToSortedSet(ctx context.Context, key string, value string, score float64)
	GetSortedSetRange(ctx context.Context, key string, scoreFrom string, scoreTo string) ([]string, []float64, error)
	RemoveSortedSetRange(ctx context.Context, key string, scoreFrom string, scoreTo string) error
	GetListRange(ctx context.Context, key string, from int64, to int64) ([]string, error)
	RemoveFromList(ctx context.Context, key string, value string) error
	AppendToSet(ctx context.Context, key string, value string)
	Exists(ctx context.Context, key string) (bool, error)
}

const defaultHashAlgorithm = "murmur64"

// Defines algorithm constant.
var (
	HashSha256    = "sha256"
	HashMurmur32  = "murmur32"
	HashMurmur64  = "murmur64"
	HashMurmur128 = "murmur128"
)

func hashFunction(algorithm string) (hash.Hash, error) {
	switch algorithm {
	case HashSha256:
		return sha256.New(), nil
	case HashMurmur64:
		return murmur3.New64(), nil
	case HashMurmur128:
		return murmur3.New128(), nil
	case "", HashMurmur32:
		return murmur3.New32(), nil
	default:
		return murmur3.New32(), fmt.Errorf("unknown key hash function: %s. Falling back to murmur32", algorithm)
	}
}

type ExtractHashIdentity interface {
	ExtractHashIdentity(string) string
}

// HashStr return hash the give string and return.
func HashStr(in string, hashIdentity ExtractHashIdentity) string {
	h, _ := hashFunction(hashIdentity.ExtractHashIdentity(in))
	_, _ = h.Write([]byte(in))

	return hex.EncodeToString(h.Sum(nil))
}
