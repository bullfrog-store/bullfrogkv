package main

import (
	"bullfrogkv/raftstore"
	"github.com/gin-gonic/gin"
)

type raftEngine struct {
	engine *raftstore.RaftStore
}

func newRaftEngine(storeId uint64, dataPath string) *raftEngine {
	return &raftEngine{engine: raftstore.NewRaftStore(storeId, dataPath)}
}

func router(kve *raftEngine) *gin.Engine {
	ge := gin.New()
	ge.POST("/set", kve.putKVHandle)
	ge.GET("/get", kve.getKVHandle)
	ge.POST("/delete", kve.delKVHandle)
	return ge
}
