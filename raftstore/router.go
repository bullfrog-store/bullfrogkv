package raftstore

import (
	"bullfrogkv/raftstore/raft_conn"
	"bullfrogkv/raftstore/raftstorepb"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

// router is the route control center
type router struct {
	addr       string
	raftClient *raft_conn.RaftClient
	raftServer *raft_conn.RaftServer
}

func newRouter(addr string, sender chan<- raftpb.Message) *router {
	r := &router{
		addr:       addr,
		raftServer: raft_conn.NewRaftServer(sender),
		raftClient: raft_conn.NewRaftClient(),
	}
	return r
}

func (r *router) sendRaftMessage(msgs []raftpb.Message) {
	for i := range msgs {
		addr, ok := peerMap[msgs[i].To]
		if !ok {
			// TODO: Handling address does not exist
			continue
		}
		conn, err := r.raftClient.GetClientConn(addr)
		if err != nil {
			// TODO: Handling get grpc connection failure
			continue
		}
		peerMsg := &raftstorepb.RaftMsgReq{
			Message:  &msgs[i],
			FromPeer: msgs[i].From,
			ToPeer:   msgs[i].To,
		}
		err = conn.Send(peerMsg)
		if err != nil {
			// TODO: Handling grpc send failure
			continue
		}
	}
}
