package balancer

import (
	"context"
	"github.com/smallnest/weighted"
	"karma8"
	"sync"
)

type weightedRoundRobinBalancer struct {
	roundRobinWeighted weighted.SW
	mutex              sync.Mutex
}

func (m *weightedRoundRobinBalancer) GetHosts(ctx context.Context, count int) ([]string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	hosts := make([]string, 0, count)

	// NOTE: As alternative we could use random choice.
	seen := make(map[string]struct{}, count)
	for len(hosts) < count {
		host := m.roundRobinWeighted.Next().(string)
		_, ok := seen[host]
		if !ok {
			hosts = append(hosts, host)
			seen[host] = struct{}{}
		}
	}
	return hosts, nil
}

func NewWeightedRoundRobinBalancer(hostToWeight map[string]int) karma8.Balancer {
	// smooth round robin
	roundRobinWeighted := weighted.SW{}
	for host, weight := range hostToWeight {
		roundRobinWeighted.Add(host, weight)
	}

	return &weightedRoundRobinBalancer{
		roundRobinWeighted: roundRobinWeighted,
	}
}
