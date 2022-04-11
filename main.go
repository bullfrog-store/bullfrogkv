package main

import (
	"bullfrogkv/config"
	"bullfrogkv/logger"
	"bullfrogkv/server"
	"flag"
)

func main() {
	// Start command (under bullfrogkv dir):
	//     1. go build main.go
	//     2.
	//       ./main -conf=./conf/node1.toml
	//       ./main -conf=./conf/node2.toml
	//       ./main -conf=./conf/node3.toml

	confPath := flag.String("conf", "/tmp/conf.toml", "the config file of server")
	flag.Parse()

	config.LoadConfigFile(*confPath)
	logger.SetLogLevel(config.GlobalConfig.CommonConfig.LogLevel)

	engine := server.NewRaftEngine()
	ge := server.Router(engine)
	ge.Run(config.GlobalConfig.RouteConfig.ServeAddr)
}
