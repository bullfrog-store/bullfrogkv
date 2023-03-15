package main

import (
	"bullfrogkv/config"
	"bullfrogkv/logger"
	"bullfrogkv/server"
	"flag"
	"os"
)

const (
	envKeyBullfrogHome = "BULLFROG_HOME"
)

const (
	confFlag    = "conf"
	confAbsPath = "/nodes/conf.toml"
)

func fatalfInitError(err error) {
	logger.Fatalf("Encounter an unexpected error in the initialization stage, err: %s", err.Error())
}

func errorfRuntimeError(err error) {
	logger.Errorf("Encounter an unexpected error in the runtime stage, err: %s", err.Error())
}

func main() {
	// Start command:
	//     1. cd ${BULLFROG_HOME}
	//     2. go build main.go
	//     3.
	//       ./main -conf=./nodes/node1.toml
	//       ./main -conf=./nodes/node2.toml
	//       ./main -conf=./nodes/node3.toml

	var err error
	homepath, ok := os.LookupEnv(envKeyBullfrogHome)
	if !ok {
		if homepath, err = os.Getwd(); err != nil {
			fatalfInitError(err)
		}
		if err = os.Setenv(envKeyBullfrogHome, homepath); err != nil {
			fatalfInitError(err)
		}
	}

	confpath := flag.String(confFlag, homepath+confAbsPath, "the default config file path of server")
	flag.Parse()

	if err = config.LoadConfigFile(*confpath); err != nil {
		fatalfInitError(err)
	}

	srv, err := server.New()
	if err != nil {
		fatalfInitError(err)
	}

	ginsrv := server.Router(srv)
	if err = ginsrv.Run(config.GlobalConfig.RouteConfig.ServeAddr); err != nil {
		errorfRuntimeError(err)
	}
}
