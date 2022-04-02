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

func router(engine *raftEngine) *gin.Engine {
	ge := gin.New()
	ge.POST("/set", engine.putKVHandle)
	ge.GET("/get", engine.getKVHandle)
	ge.POST("/delete", engine.delKVHandle)
	return ge
}
