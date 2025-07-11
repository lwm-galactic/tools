package registry

import (
	"context"
	"fmt"
	"github.com/lwm-galactic/logger"
	"testing"
	"time"
)

func TestNewEtcdRegistry1(t *testing.T) {
	etcdRegistry, err := NewEtcdRegistry(
		WithUsername("root"),
		WithPassword("root"),
	)
	if err != nil {
		logger.Errorf("NewEtcdRegistry err: %v", err)
		return
	}
	svc := Service{
		Name: "Test",
		Addr: "127.0.0.2",
		Port: 1239,
		TTL:  5 * time.Second,
	}
	err = etcdRegistry.Register(context.Background(), &svc)
	if err != nil {
		logger.Errorf("register err: %v", err)
		return
	}
	time.Sleep(10 * time.Second)
	etcdRegistry.Unregister(context.Background(), &svc)
}

func TestNewEtcdRegistry2(t *testing.T) {
	etcdRegistry, err := NewEtcdRegistry(
		WithUsername("root"),
		WithPassword("root"),
	)
	if err != nil {
		logger.Errorf("NewEtcdRegistry err: %v", err)
		return
	}
	svc := Service{
		Name: "Test",
		Addr: "127.0.0.1",
		Port: 1239,
		TTL:  5 * time.Second,
	}
	err = etcdRegistry.Register(context.Background(), &svc)
	if err != nil {
		logger.Errorf("register err: %v", err)
		return
	}
}

func TestEtcd(t *testing.T) {
	etcdRegistry, err := NewEtcdRegistry(
		WithUsername("root"),
		WithPassword("root"),
	)
	if err != nil {
		logger.Errorf("NewEtcdRegistry err: %v", err)
		return
	}
	err = etcdRegistry.Subscribe("Test", callback)
	if err != nil {
		logger.Errorf("subscribe err: %v", err)
		return
	}
	select {}

}

func callback(svc []*Service) {
	for _, s := range svc {
		fmt.Println(*s)
	}
}
