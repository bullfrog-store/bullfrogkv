package raft_conn

import (
	"bullfrogkv/logger"
	"bullfrogkv/raftstore/raftstorepb"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

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
		rm := raftmsg(msg)
		logger.Debugf("receive msg from %d, msg: %+v", msg.FromPeer, rm.String())
		s.msgc <- rm
	}
}

func raftmsg(msg *raftstorepb.RaftMsgReq) raftpb.Message {
	return *msg.Message
}
