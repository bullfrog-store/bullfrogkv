package raftstore

import (
	"bullfrogkv/storage"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

// peerStorage implements the interfaces of ETCD
// raft.Node for storing raft persistent states.
type peerStorage struct {
	engines *storage.Engines
}

func newPeerStorage() *peerStorage {
	return nil
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
