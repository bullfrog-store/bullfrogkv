package raftstore

type RaftStore struct {
	pr *peer
}

func NewRaftStore() *RaftStore {
	return nil
}

func (rs *RaftStore) Set(key []byte, value []byte) error {
	panic("implement me")
}

func (rs *RaftStore) Get(key []byte) ([]byte, error) {
	panic("implement me")
}

func (rs *RaftStore) Delete(key []byte) error {
	panic("implement me")
}

