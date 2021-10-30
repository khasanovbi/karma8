package server

import "time"

type HTTPConfig struct {
	Addr string `config:"addr" yaml:"addr"`
}

type BalancerConfig struct {
	HostToWeight map[string]int `config:"hosts" yaml:"hosts"`
}

type PGConfig struct {
	Host     string `config:"host" yaml:"host"`
	Port     uint16 `config:"port" yaml:"port"`
	DB       string `config:"db" yaml:"db"`
	User     string `config:"user" yaml:"user"`
	Password string `config:"password" yaml:"password"`
}

type Config struct {
	HTTP            HTTPConfig     `config:"http" yaml:"http"`
	Balancer        BalancerConfig `config:"balancer" yaml:"balancer"`
	PG              PGConfig       `config:"pg" yaml:"pg"`
	ShutdownTimeout time.Duration  `config:"shutdown_timeout" yaml:"shutdown_timeout"`
	MinChunkSize    int64          `config:"min_chunk_size" yaml:"min_chunk_size"`
	HostSplitCount  int            `config:"host_split_count" yaml:"host_split_count"`
}
