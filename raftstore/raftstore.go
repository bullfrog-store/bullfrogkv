package raftstore

import (
	"bullfrogkv/raftstore/internal"
	"bullfrogkv/raftstore/raftstorepb"
)

var peerMap = map[uint64]string{
	1: "127.0.0.1:6060", // rpc port
	2: "127.0.0.1:6061",
	3: "127.0.0.1:6062",
}

type RaftStore struct {
	pr *peer
}

func NewRaftStore(storeId uint64, dataPath string) *RaftStore {
	return &RaftStore{pr: newPeer(storeId, dataPath)}
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
