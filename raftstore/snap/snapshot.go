package snap

import "go.etcd.io/etcd/raft/v3/raftpb"

type SnapshotStateType int

const (
	SnapshotGenerating SnapshotStateType = iota
	SnapshotRelaxed
	SnapshotApplying
)

type SnapshotState struct {
	StateType SnapshotStateType
	Receiver  chan *raftpb.Snapshot
}
