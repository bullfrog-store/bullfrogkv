package server

import (
	"bullfrogkv/raftstore"
	"github.com/gin-gonic/gin"
)

// Bullfrog request paths
const (
	pathSet    = "/set"
	pathGet    = "/get"
	pathDelete = "/delete"
)

type BullfrogServer struct {
	engine *raftstore.RaftStore
}

func New() (*BullfrogServer, error) {
	var err error
	srv := &BullfrogServer{}

	if srv.engine, err = raftstore.New(); err != nil {
		return nil, err
	}
	return srv, nil
}

func Router(srv *BullfrogServer) *gin.Engine {
	ginsrv := gin.New()
	ginsrv.GET(pathSet, srv.handlerSet)
	ginsrv.GET(pathGet, srv.handlerGet)
	ginsrv.GET(pathDelete, srv.handlerDelete)
	return ginsrv
}
