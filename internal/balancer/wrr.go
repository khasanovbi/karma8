package balancer

import (
	"context"
	"go.uber.org/atomic"
	"karma8"
)

type weightedRoundRobinBalancer struct {
	hosts        []string
	count        uint32
	indexPointer atomic.Uint32
}

func (m *weightedRoundRobinBalancer) GetHosts(ctx context.Context) ([]string, error) {
	index := m.indexPointer.Add(m.count)
	hosts := make([]string, 0, m.count)
	var i uint32 = 0
	for ; i < m.count; i++ {
		hosts = append(hosts, m.hosts[(index+i)%uint32(len(m.hosts))])
	}
	return hosts, nil
}

func NewWeightedRoundRobinBalancer(hosts []string, count uint32) karma8.Balancer {
	return &weightedRoundRobinBalancer{
		hosts: hosts,
		count: count,
	}
}
