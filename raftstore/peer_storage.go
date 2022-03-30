package raftstore

import (
	"bullfrogkv/storage"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type peerStorage struct {
	engine *storage.Engines
}

func newPeerStorage(path string) *peerStorage {
	ps := &peerStorage{
		engine: storage.NewEngines(path+storage.KvPath, path+storage.RaftPath),
	}
	return ps
}

func (ps *peerStorage) InitialState() (raftpb.HardState, raftpb.ConfState, error) {
	panic("implement me")
}

func (ps *peerStorage) Entries(lo, hi, maxSize uint64) ([]raftpb.Entry, error) {
	panic("implement me")
}

func (ps *peerStorage) Term(i uint64) (uint64, error) {
	panic("implement me")
}

func (ps *peerStorage) LastIndex() (uint64, error) {
	panic("implement me")
}

func (ps *peerStorage) FirstIndex() (uint64, error) {
	panic("implement me")
}

func (ps *peerStorage) Snapshot() (raftpb.Snapshot, error) {
	panic("implement me")
}

func (ps *peerStorage) AppliedIndex() uint64 {
	// TODO: last applied index
	return 0
}
