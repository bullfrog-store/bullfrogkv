package raftstore

import "go.etcd.io/etcd/raft/v3/raftpb"

// router is the route control center
type router struct {
	raftMsgSender chan<- raftpb.Message
}

func newRouter(sender chan<- raftpb.Message) *router {
	return &router{
		raftMsgSender: sender,
	}
}

func (r *router) sendRaftMessage(msgs []raftpb.Message) {
	// TODO: grpc stream send messages
}
