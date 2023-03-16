package raftstore

import (
	"bullfrogkv/logger"
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
	for _, msg := range msgs {
		addr, ok := peerMap[msg.To]
		if !ok {
			logger.Warningf("address of peer %d not found!", msg.To)
			continue
		}
		peerMsg := &raftstorepb.RaftMsgReq{
			Message:  &msg,
			FromPeer: msg.From,
			ToPeer:   msg.To,
		}
		logger.Debugf("node %d send msg: %+v", peerMsg.FromPeer, peerMsg.Message.String())

		conn, ok := r.raftClient.GetRaftConn(addr)
		if !ok {
			go r.raftClient.DialAndSend(addr, peerMsg)
			continue
		}

		err := conn.Send(peerMsg)
		if err != nil {
			go r.raftClient.DialAndSend(addr, peerMsg)
		}
	}
}
