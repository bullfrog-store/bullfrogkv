package raftstore

type peer struct {
	// ETCD Raft Node
	// raftGroup raft.Node

	ps *peerStorage
}
