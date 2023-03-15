package raftstore

import (
	"bullfrogkv/raftstore/internal"
	"bullfrogkv/storage"
	"errors"
	"time"
)

var (
	ErrLostReadResponse = errors.New("it lost read response, retry read")
)

type RaftStore struct {
	pr *peer
}

func New() (*RaftStore, error) {
	var err error
	rs := &RaftStore{}

	if rs.pr, err = newPeer(); err != nil {
		return nil, err
	}
	return rs, nil
}

func (rs *RaftStore) Set(key, value []byte) error {
	cmd := internal.NewPutCmdRequest(key, value)
	// TODO: wait success message
	return rs.pr.propose(cmd)
}

func (rs *RaftStore) Get(key []byte) ([]byte, error) {
	cb := rs.pr.linearizableRead(key)
	resp := cb.WaitRespWithTimeout(time.Second)
	if resp == nil {
		return nil, ErrLostReadResponse
	}
	value := resp.GetResponse().Get.Value
	if value == nil {
		return nil, storage.ErrNotFound
	}
	return value, nil
}

func (rs *RaftStore) Delete(key []byte) error {
	cmd := internal.NewDeleteCmdRequest(key)
	// TODO: wait success message
	return rs.pr.propose(cmd)
}
