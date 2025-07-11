package registry

import (
	"context"
	"fmt"
	"github.com/lwm-galactic/logger"
	"github.com/lwm-galactic/tools/json"
	"go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"time"
)

type Registry interface {
	Register(ctx context.Context, svc *Service) error
	Unregister(ctx context.Context, svc *Service) error
	Subscribe(name string, callback func([]*Service)) error
	Close() error
}

var (
	registry Registry
	once     sync.Once
)

type EtcdRegistry struct {
	prefix string
	client *clientv3.Client

	// locker Locker
	// subscribers map[string][]func([]*Service)
	// strategies  map[string]Strategy
}

func newEtcdClient(options *Options) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"117.72.74.31:2379"},
		DialTimeout: 5 * time.Second,
		Username:    options.Username,
		Password:    options.Password,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func NewEtcdRegistry(opts ...Option) (Registry, error) {
	if registry == nil {

	}
	var e error
	once.Do(func() {
		opt := &Options{}
		for _, o := range opts {
			o(opt)
		}
		client, err := newEtcdClient(opt)
		if err != nil {
			e = err
		}
		registry = &EtcdRegistry{client: client}
	})
	if e != nil {
		return nil, e
	}

	return registry, nil
}

func (r *EtcdRegistry) registerService(svc *Service) error {
	serviceKey := svc.buildServerKey(r.prefix)

	logger.Debugf("watch key: %s", serviceKey)
	// 步骤一：创建租约
	leaseResp, err := r.client.Grant(context.Background(), int64(svc.TTL.Seconds()))
	if err != nil {
		return err
	}
	svc.leaseID = leaseResp.ID
	// 步骤二：将服务信息序列化为 JSON（也可以用 protobuf）
	serviceValue := fmt.Sprintf(`{"name":"%s","addr":"%s","port":%d}`, svc.Name, svc.Addr, svc.Port)
	// 步骤三：绑定租约并写入 etcd
	_, err = r.client.Put(context.Background(), serviceKey, serviceValue, clientv3.WithLease(svc.leaseID))
	return err
}

// Register 在注册时启动协程定期发送心跳
func (r *EtcdRegistry) Register(ctx context.Context, svc *Service) error {
	/*
		// 获取锁
			ok, err := r.locker.Acquire(ctx, "register_lock_"+svc.Name)
			// 获取锁失败
			if err != nil {

			}
			if !ok {

			}
			defer func(locker Locker, ctx context.Context) {
				err := locker.Release(ctx)
				if err != nil { // 释放锁失败

				}
			}(r.locker, ctx)
	*/
	// 注册服务逻辑
	err := r.registerService(svc)
	if err != nil {
		return err
	}

	// 启动后台心跳协程
	go func() {
		ticker := time.NewTicker(svc.TTL / 2)
		defer ticker.Stop()
		retryCount := 0
		const maxRetries = 3 // 最大重试次数
		for {
			select {
			case <-ticker.C:
				var err2 error
				for ; retryCount < maxRetries; retryCount++ {
					_, err2 = r.client.KeepAliveOnce(context.Background(), svc.leaseID)
					if err2 == nil {
						retryCount = 0 // 成功后重置重试计数器
						break
					}

					logger.Errorf("心跳失败，正在重试 (%d/%d): %v", retryCount+1, maxRetries, err2)
					time.Sleep(time.Second * time.Duration(1<<uint(retryCount))) // 指数退避
				}

				if err2 != nil {
					logger.Errorf("心跳失败超过最大重试次数: %v", err2)
					// 可选：触发服务注销或重启逻辑
					err = r.Unregister(ctx, svc)
					if err != nil {
						logger.Errorf("unregister err: %v", err)
						return
					}
					return
				}
			case <-ctx.Done():
				err = r.Unregister(ctx, svc)
				if err != nil {
					logger.Errorf("unregister err: %v", err)
					return
				}
				return
			}
		}
	}()
	return nil
}

func (r *EtcdRegistry) Unregister(ctx context.Context, svc *Service) error {
	key := svc.buildServerKey(r.prefix)
	_, err := r.client.Delete(ctx, key)
	if err != nil {
		return err
	}

	// 释放租约
	if svc.leaseID != 0 {
		_, err = r.client.Revoke(ctx, svc.leaseID)
		return err
	}
	return nil
}

func (r *EtcdRegistry) Subscribe(serviceName string, callback func([]*Service)) error {
	// 构造监听的前缀键
	watchKey := r.prefix + "/" + serviceName
	logger.Debugf("watch key: %s", watchKey)

	// 初始获取当前服务列表
	initialServices, err := r.discoverServices(watchKey)
	if err != nil {
		return err
	}
	callback(initialServices)

	// 启动监听goroutine
	go r.watchServices(watchKey, callback)

	return nil
}

// watchServices 监听指定前缀的服务变化
func (r *EtcdRegistry) watchServices(watchKey string, callback func([]*Service)) {
	watcher := clientv3.NewWatcher(r.client)
	defer watcher.Close()

	watchChan := watcher.Watch(context.Background(), watchKey, clientv3.WithPrefix())
	var mutex sync.Mutex
	for range watchChan {
		mutex.Lock()
		services, err := r.discoverServices(watchKey)
		if err != nil {
			log.Printf("watch services error: %v", err)
			continue
		}
		callback(services)
		mutex.Unlock()
	}
}

func (r *EtcdRegistry) discoverServices(watchKey string) ([]*Service, error) {

	resp, err := r.client.Get(context.Background(), watchKey, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var services []*Service
	for _, kv := range resp.Kvs {
		var svc Service
		if err := json.Unmarshal(kv.Value, &svc); err != nil {
			continue // 跳过无效数据
		}
		services = append(services, &svc)
	}
	return services, nil
}

func (r *EtcdRegistry) Close() error { return nil }
