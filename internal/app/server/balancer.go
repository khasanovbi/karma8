package server

import (
	"karma8"
	"karma8/internal/balancer"
)

func newBalancer(conf *BalancerConfig) karma8.Balancer {
	return balancer.NewWeightedRoundRobinBalancer(conf.HostToWeight)
}
