package server

import (
	"bullfrogkv/raftstore"
	"github.com/gin-gonic/gin"
)

type RaftEngine struct {
	engine *raftstore.RaftStore
}

func NewRaftEngine() *RaftEngine {
	return &RaftEngine{engine: raftstore.NewRaftStore()}
}

func Router(engine *RaftEngine) *gin.Engine {
	ge := gin.New()
	ge.POST("/set", engine.putKVHandle)
	ge.GET("/get", engine.getKVHandle)
	ge.POST("/delete", engine.delKVHandle)
	return ge
}
