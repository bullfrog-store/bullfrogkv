package raft_conn

import (
	"bullfrogkv/raftstore/raftstorepb"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

// TODO: use NewRaftStore().pr.router.raftServer to register grpc server in main.go
type RaftServer struct {
	Msgs chan<- raftpb.Message
}

func NewPeerServer(sender chan<- raftpb.Message) *RaftServer {
	return &RaftServer{
		Msgs: sender,
	}
}

func (s *RaftServer) RaftMessage(stream raftstorepb.Message_RaftMessageServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		message := getMessage(msg)
		s.Msgs <- message
	}
}

func getMessage(m *raftstorepb.RaftMsgReq) raftpb.Message {
	return *m.Message
}
