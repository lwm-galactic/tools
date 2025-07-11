package registry

// Strategy 负载均衡策略
type Strategy func([]*Service) *Service

// RoundRobinStrategy 轮询
func RoundRobinStrategy() Strategy {
	var index int
	return func(instances []*Service) *Service {
		if len(instances) == 0 {
			return nil
		}
		instance := instances[index%len(instances)]
		index++
		return instance
	}
}
