package meta

import (
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/storage"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func InitRaftLocalState(engine storage.Engine) (*raftstorepb.RaftLocalState, error) {
	return GetRaftLocalState(engine)
}

func InitRaftApplyState(engine storage.Engine) (*raftstorepb.RaftApplyState, error) {
	return GetRaftApplyState(engine)
}

func InitConfState(engine storage.Engine) (*raftpb.ConfState, error) {
	return GetRaftConfState(engine)
}

func GetRaftLocalState(engines storage.Engine) (*raftstorepb.RaftLocalState, error) {
	raftState := &raftstorepb.RaftLocalState{
		HardState: &raftpb.HardState{},
	}
	value, err := engines.ReadMeta(RaftLocalStateKey())
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

func GetRaftApplyState(engine storage.Engine) (*raftstorepb.RaftApplyState, error) {
	applyState := &raftstorepb.RaftApplyState{
		TruncatedState: &raftstorepb.RaftTruncatedState{},
	}
	value, err := engine.ReadMeta(RaftApplyStateKey())
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

func GetRaftConfState(engine storage.Engine) (*raftpb.ConfState, error) {
	confState := &raftpb.ConfState{}
	value, err := engine.ReadMeta(RaftConfStateKey())
	if err == storage.ErrNotFound {
		return confState, nil
	}
	if err != nil {
		return nil, err
	}
	if err = proto.Unmarshal(value, confState); err != nil {
		return nil, err
	}
	return confState, nil
}
