package config

import (
	"bullfrogkv/logger"
	"errors"
	"github.com/BurntSushi/toml"
	"os"
	"strings"
)

var GlobalConfig *Config

// Default values for config
const (
	// Common config
	defaultServeType = "http"
	defaultLogLevel  = "info"
	defaultLogFile   = "bullfrog.log"

	// Store config
	defaultDataPathPrefix = "./bullfrog/"

	// Raft config
	defaultElectionTick       = 10
	defaultHeartbeatTick      = 1
	defaultMaxSizePerMsg      = 1024 * 1024
	defaultMaxInflightMsgs    = 256
	defaultLogGCCountLimit    = 100
	defaultCompactCheckPeriod = 100 // Check compaction per 10s (CompactCheckPeriod * 100ms)
	defaultSnapshotTryCount   = 5
)

type Config struct {
	CommonConfig CommonConfig `toml:"common"`
	RouteConfig  RouteConfig  `toml:"route"`
	StoreConfig  StoreConfig  `toml:"store"`
	RaftConfig   RaftConfig   `toml:"raft"`
}

type CommonConfig struct {
	ServeType string `toml:"serve_type"` // MUST be filled
	LogLevel  string `toml:"log_level"`
	LogFile   string `toml:"log_file"`
	NodeName  string `toml:"node_name"`
	NodeCount int    `toml:"node_count"` // MUST be filled
}

type RouteConfig struct {
	ServeAddr string   `toml:"serve_addr"` // MUST be filled
	GrpcAddrs []string `toml:"grpc_addrs"` // MUST be filled
}

type StoreConfig struct {
	StoreId  uint64 `toml:"store_id"` // MUST be filled
	DataPath string `toml:"data_path"`
}

type RaftConfig struct {
	ElectionTick       int    `toml:"election_tick"`
	HeartbeatTick      int    `toml:"heartbeat_tick"`
	MaxSizePerMsg      uint64 `toml:"max_size_per_msg"`
	MaxInflightMsgs    int    `toml:"max_inflight_msgs"`
	LogGCCountLimit    uint64 `toml:"log_gc_count_limit"`
	CompactCheckPeriod int    `toml:"compact_check_period"`
	SnapshotTryCount   int    `toml:"snapshot_try_count"`
}

func LoadConfigFile(path string) error {
	GlobalConfig = &Config{}
	if _, err := toml.DecodeFile(path, &GlobalConfig); err != nil {
		return err
	}
	if err := ensureDefault(GlobalConfig); err != nil {
		return err
	}

	// For logger
	nodepath := GlobalConfig.StoreConfig.DataPath
	if err := os.MkdirAll(nodepath, 0775); err != nil {
		return err
	}
	logfile, err := os.Create(nodepath + "/" + GlobalConfig.CommonConfig.LogFile)
	if err != nil {
		return err
	}
	logger.ResetGlobalLogger(logger.NewWithFile(logfile))
	logger.SetLogLevel(GlobalConfig.CommonConfig.LogLevel)
	return nil
}

func validate(c *Config) error {
	// Common config
	if len(c.CommonConfig.ServeType) != 0 {
		srvtyp := strings.ToLower(strings.Trim(c.CommonConfig.ServeType, " "))
		if srvtyp != "terminal" && srvtyp != "http" {
			return errors.New("ServeType must be terminal or http")
		} else {
			c.CommonConfig.ServeType = srvtyp
		}
	}
	if c.CommonConfig.NodeCount <= 0 {
		return errors.New("NodeCount cannot equal or less than 0")
	}

	// RouteConfig
	if len(c.RouteConfig.ServeAddr) == 0 {
		return errors.New("ServeAddr cannot be empty")
	}
	if len(c.RouteConfig.GrpcAddrs) != c.CommonConfig.NodeCount {
		return errors.New("the number of GrpcAddrs must equal NodeCount")
	}
	for _, addr := range c.RouteConfig.GrpcAddrs {
		if c.RouteConfig.ServeAddr == addr {
			return errors.New("some GrpcAddrs is overlapping with ServeAddr")
		}
	}

	// Store config
	if c.StoreConfig.StoreId <= 0 {
		return errors.New("StoreId must > 0")
	}

	return nil
}

func ensureDefault(c *Config) error {
	if err := validate(c); err != nil {
		return err
	}

	// Common config
	if len(c.CommonConfig.ServeType) == 0 {
		c.CommonConfig.ServeType = defaultServeType
	}
	if len(c.CommonConfig.LogLevel) == 0 {
		c.CommonConfig.LogLevel = defaultLogLevel
	}
	if len(c.CommonConfig.LogFile) == 0 {
		c.CommonConfig.LogFile = defaultLogFile
	}
	if len(c.CommonConfig.NodeName) == 0 {
		port := strings.Split(c.RouteConfig.ServeAddr, ":")[1]
		c.CommonConfig.NodeName = port
	}

	// Store config
	if len(c.StoreConfig.DataPath) == 0 {
		c.StoreConfig.DataPath = defaultDataPathPrefix + c.CommonConfig.NodeName
	}

	// Raft config
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
	if c.RaftConfig.SnapshotTryCount == 0 {
		c.RaftConfig.SnapshotTryCount = defaultSnapshotTryCount
	}

	return nil
}
