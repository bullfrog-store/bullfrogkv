package meta

import (
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/storage"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func GetRaftLocalState(engines *storage.Engines) (*raftstorepb.RaftLocalState, error) {
	raftState := &raftstorepb.RaftLocalState{
		HardState: &raftpb.HardState{},
	}
	value, err := engines.ReadRaft(RaftLocalStateKey())
	if err == storage.ErrNotFound {
		return raftState, nil
	}
	if err != nil {
		return nil, err
	}
	if err = proto.Unmarshal(value, raftState); err != nil {
		return nil, err
	}
	return raftState, nil
}

func InitRaftLocalState(engines *storage.Engines) *raftstorepb.RaftLocalState {
	raftState, err := GetRaftLocalState(engines)
	if err != nil {
		panic("raftLocalState parse failed")
	}
	return raftState
}

func GetRaftApplyState(engines *storage.Engines) (*raftstorepb.RaftApplyState, error) {
	applyState := &raftstorepb.RaftApplyState{
		TruncatedState: &raftstorepb.RaftTruncatedState{},
	}
	value, err := engines.ReadKV(RaftApplyStateKey())
	if err == storage.ErrNotFound {
		return applyState, nil
	}
	if err != nil {
		return nil, err
	}
	if err = proto.Unmarshal(value, applyState); err != nil {
		return nil, err
	}
	return applyState, nil
}

func InitRaftApplyState(engines *storage.Engines) *raftstorepb.RaftApplyState {
	applyState, err := GetRaftApplyState(engines)
	if err != nil {
		panic("raftApplyState parse failed")
	}
	return applyState
}
