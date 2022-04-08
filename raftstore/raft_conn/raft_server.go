package raft_conn

import (
	"bullfrogkv/logger"
	"bullfrogkv/raftstore/raftstorepb"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

// TODO: use NewRaftStore().pr.router.raftServer to register grpc server in main.go
type RaftServer struct {
	msgc chan<- raftpb.Message
}

func NewRaftServer(sender chan<- raftpb.Message) *RaftServer {
	return &RaftServer{
		msgc: sender,
	}
}

func (s *RaftServer) RaftMessage(stream raftstorepb.Message_RaftMessageServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		rm := raftMsg(msg)
		logger.Infof("[grpc] receive msg from %d, msg: %+v", msg.FromPeer, rm.String())
		s.msgc <- rm
	}
}

// raftMsg TODO: rename RaftMsgReq
func raftMsg(msg *raftstorepb.RaftMsgReq) raftpb.Message {
	return *msg.Message
}
