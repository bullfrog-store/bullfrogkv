package snap

import "go.etcd.io/etcd/raft/v3/raftpb"

type SnapshotStateType int

const (
	SnapshotGenerating SnapshotStateType = iota
	SnapshotRelaxed
	SnapshotApplying
	SnapshotToGen
	SnapshotApplied
)

type SnapshotState struct {
	StateType SnapshotStateType
	Receiver  chan *raftpb.Snapshot
}
