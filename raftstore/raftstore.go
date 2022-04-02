package raftstore

import (
	"bullfrogkv/raftstore/internal"
	"bullfrogkv/raftstore/raftstorepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"net"
	"time"
)

var peerMap = map[uint64]string{
	1: "127.0.0.1:6060", // rpc port
	2: "127.0.0.1:6061",
	3: "127.0.0.1:6062",
}

var enforcementPolicy = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second,
	PermitWithoutStream: true,
}

var serverParameters = keepalive.ServerParameters{
	MaxConnectionIdle: 15 * time.Second,
	Time:              5 * time.Second,
	Timeout:           1 * time.Second,
}

type RaftStore struct {
	pr *peer
}

func NewRaftStore(storeId uint64, dataPath string) *RaftStore {
	rs := &RaftStore{pr: newPeer(storeId, dataPath)}
	go rs.serveGrpc(storeId)
	return rs
}

func (rs *RaftStore) serveGrpc(id uint64) {
	lis, err := net.Listen("tcp", peerMap[id])
	if err != nil {
		panic(err)
	}
	log.Printf("%d listen %v success\n", id, peerMap[id])
	g := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(enforcementPolicy), grpc.KeepaliveParams(serverParameters))
	raftstorepb.RegisterMessageServer(g, rs.pr.router.raftServer)
	g.Serve(lis)
}

func (rs *RaftStore) Set(key []byte, value []byte) error {
	header := &raftstorepb.RaftRequestHeader{
		Term: rs.pr.term(),
	}
	req := &raftstorepb.Request{
		CmdType: raftstorepb.CmdType_Put,
		Put: &raftstorepb.PutRequest{
			Key:   key,
			Value: value,
		},
	}
	cmd := internal.NewRaftCmdRequest(header, req)
	return rs.pr.propose(cmd)
}

func (rs *RaftStore) Get(key []byte) ([]byte, error) {
	cb := rs.pr.linearizableRead(key)
	resp := cb.WaitResp()
	return resp.GetResponse().Get.Value, nil
}

func (rs *RaftStore) Delete(key []byte) error {
	header := &raftstorepb.RaftRequestHeader{
		Term: rs.pr.term(),
	}
	req := &raftstorepb.Request{
		CmdType: raftstorepb.CmdType_Delete,
		Delete: &raftstorepb.DeleteRequest{
			Key: key,
		},
	}
	cmd := internal.NewRaftCmdRequest(header, req)
	return rs.pr.propose(cmd)
}
