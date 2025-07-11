package registry

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Service struct {
	Name          string
	Addr          string
	Port          int
	TTL           time.Duration
	leaseID       clientv3.LeaseID
	LastHeartbeat time.Time
}

func (s *Service) IsHealthy(now time.Time) bool {
	return now.Sub(s.LastHeartbeat) < s.TTL
}

func (s *Service) buildServerKey(prefix string) string {
	return fmt.Sprintf("%s/%s/%s:%d", prefix, s.Name, s.Addr, s.Port)
}
