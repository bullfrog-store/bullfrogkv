package config

import (
	"github.com/BurntSushi/toml"
)

// default value for config
const (
	defaultLogLevel = "info"

	defaultDataPath = "/tmp/bullfrog"

	defaultElectionTick       = 10
	defaultHeartbeatTick      = 1
	defaultMaxSizePerMsg      = 1024 * 1024
	defaultMaxInflightMsgs    = 256
	defaultLogGCCountLimit    = 100
	defaultCompactCheckPeriod = 100 // check compaction per 10s (CompactCheckPeriod * 100ms)
)

type Config struct {
	CommonConfig CommonConfig `toml:"common"`
	RouteConfig  RouteConfig  `toml:"route"`
	StoreConfig  StoreConfig  `toml:"store"`
	RaftConfig   RaftConfig   `toml:"raft"`
}

type CommonConfig struct {
	LogLevel string `toml:"log_level"`
	// NodeCount (MUST FILL)
	NodeCount int `toml:"node_count"`
}

type RouteConfig struct {
	// ServeAddr (MUST FILL)
	ServeAddr string `toml:"serve_addr"`
	// GrpcAddrs (MUST FILL)
	GrpcAddrs []string `toml:"grpc_addrs"`
}

type StoreConfig struct {
	// StoreId (MUST FILL)
	StoreId  uint64 `toml:"store_id"`
	DataPath string `toml:"data_path"`
}

type RaftConfig struct {
	ElectionTick       int    `toml:"election_tick"`
	HeartbeatTick      int    `toml:"heartbeat_tick"`
	MaxSizePerMsg      uint64 `toml:"max_size_per_msg"`
	MaxInflightMsgs    int    `toml:"max_inflight_msgs"`
	LogGCCountLimit    int    `toml:"log_gc_count_limit"`
	CompactCheckPeriod int    `toml:"compact_check_period"`
}

func LoadConfigFile(path string) *Config {
	c := &Config{}
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		panic(err)
	}
	ensureDefault(c)
	return c
}

func validate(c *Config) {
	if c.CommonConfig.NodeCount <= 0 {
		panic("NodeCount cannot equal or less than 0")
	}

	if len(c.RouteConfig.ServeAddr) == 0 {
		panic("ServeAddr cannot be empty")
	}
	if len(c.RouteConfig.GrpcAddrs) != c.CommonConfig.NodeCount {
		panic("the number of GrpcAddrs must equal NodeCount")
	}
	for _, addr := range c.RouteConfig.GrpcAddrs {
		if c.RouteConfig.ServeAddr == addr {
			panic("GrpcAddrs is overlapping with ServeAddr")
		}
	}

	if c.StoreConfig.StoreId <= 0 {
		panic("StoreId cannot equal or less than 0")
	}
}

func ensureDefault(c *Config) {
	validate(c)

	if len(c.CommonConfig.LogLevel) == 0 {
		c.CommonConfig.LogLevel = defaultLogLevel
	}

	if len(c.StoreConfig.DataPath) == 0 {
		c.StoreConfig.DataPath = defaultDataPath
	}

	if c.RaftConfig.ElectionTick == 0 {
		c.RaftConfig.ElectionTick = defaultElectionTick
	}
	if c.RaftConfig.HeartbeatTick == 0 {
		c.RaftConfig.HeartbeatTick = defaultHeartbeatTick
	}
	if c.RaftConfig.MaxSizePerMsg == 0 {
		c.RaftConfig.MaxSizePerMsg = defaultMaxSizePerMsg
	}
	if c.RaftConfig.MaxInflightMsgs == 0 {
		c.RaftConfig.MaxInflightMsgs = defaultMaxInflightMsgs
	}
	if c.RaftConfig.LogGCCountLimit == 0 {
		c.RaftConfig.LogGCCountLimit = defaultLogGCCountLimit
	}
	if c.RaftConfig.CompactCheckPeriod == 0 {
		c.RaftConfig.CompactCheckPeriod = defaultCompactCheckPeriod
	}
}
