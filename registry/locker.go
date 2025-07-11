package registry

import (
	"context"
	"github.com/lwm-galactic/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"time"
)

// Locker 分布式锁接口
type Locker interface {
	// Acquire 加锁，带上下文用于超时/取消控制，返回是否成功和错误信息
	Acquire(ctx context.Context, key string) (bool, error)
	// Release 解锁
	Release(ctx context.Context) error
}

type EtcdLocker struct {
	session *concurrency.Session
	mutex   *concurrency.Mutex
	client  *clientv3.Client
	key     string
}

func NewEtcdLocker(client *clientv3.Client, key string, ttl int) (*EtcdLocker, error) {
	session, err := concurrency.NewSession(client, concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}

	mutex := concurrency.NewMutex(session, key)
	return &EtcdLocker{
		session: session,
		mutex:   mutex,
		client:  client,
		key:     key,
	}, nil
}

func (l *EtcdLocker) Acquire(ctx context.Context, key string) (bool, error) {
	logger.Infof("Trying to acquire lock for %s\n", key)

	// 使用上下文控制超时
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := l.mutex.Lock(ctxWithTimeout)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (l *EtcdLocker) Release(ctx context.Context) error {
	return l.mutex.Unlock(context.Background())
}
