package balancer

import (
	"context"
	"go.uber.org/atomic"
	"karma8"
)

type roundRobinBalancer struct {
	hosts        []string
	indexPointer atomic.Uint32
}

func (m *roundRobinBalancer) GetHosts(ctx context.Context, count int) ([]string, error) {
	index := m.indexPointer.Add(uint32(count))
	hosts := make([]string, 0, count)
	var i uint32 = 0
	for ; i < uint32(count); i++ {
		hosts = append(hosts, m.hosts[(index+i)%uint32(len(m.hosts))])
	}
	return hosts, nil
}

func NewRoundRobinBalancer(hosts []string) karma8.Balancer {
	return &roundRobinBalancer{
		hosts: hosts,
	}
}
