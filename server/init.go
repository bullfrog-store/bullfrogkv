package main

import (
	"bullfrogkv/storage"
	"github.com/gin-gonic/gin"
)

func newKVEngine(kvPath, raftPath string) *kvEngine {
	return &kvEngine{kve: storage.NewEngines(kvPath, raftPath)}
}

func router(kve *kvEngine) *gin.Engine {
	ge := gin.New()
	ge.PUT("/", kve.putKVHandle)
	ge.GET("/", kve.getKVHandle)
	ge.DELETE("/", kve.delKVHandle)
	return ge
}
