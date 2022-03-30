package raftstore

import "go.etcd.io/etcd/raft/v3/raftpb"

// router is the route control center
type router struct {
	addr          string
	raftMsgSender chan<- raftpb.Message
}

func newRouter(addr string, sender chan<- raftpb.Message) *router {
	return &router{
		addr:          addr,
		raftMsgSender: sender,
	}
}

func (r *router) sendRaftMessage(msgs []raftpb.Message) {
	// TODO: grpc stream send messages
}
