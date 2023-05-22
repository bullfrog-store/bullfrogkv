package server

import (
	"bullfrogkv/raftstore"
)

const (
	msgSetSuccess    = "set success"
	msgSetFailure    = "set failure"
	msgGetSuccess    = "get success"
	msgGetFailure    = "get failure"
	msgDeleteSuccess = "delete success"
	msgDeleteFailure = "delete failure"
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

func toBytes(data string) []byte {
	return []byte(data)
}
