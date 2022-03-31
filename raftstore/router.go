package raftstore

import (
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/storage/peer_storage"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

// router is the route control center
type router struct {
	addr          string
	raftMsgSender chan<- raftpb.Message
	peerClient    *peer_storage.PeerClient
}

func newRouter(addr string, sender chan<- raftpb.Message) *router {
	return &router{
		addr:          addr,
		raftMsgSender: sender,
		peerClient:    peer_storage.NewPeerClient(),
	}
}

func (r *router) sendRaftMessage(msgs []raftpb.Message) {
	for i := range msgs {
		to_peer := msgs[i].To
		to_addr, ok := peerMap[to_peer]
		if !ok {
			// TODO: Handling address does not exist
			continue
		}
		conn, err := r.peerClient.GetPeerConn(to_addr)
		if err != nil {
			// TODO: Handling get grpc connection failure
			continue
		}
		peerMsg := &raftstorepb.RaftMessage{
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
