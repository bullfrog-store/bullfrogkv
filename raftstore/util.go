package raftstore

import "go.etcd.io/etcd/raft/v3/raftpb"

func isEmptyConfState(cs raftpb.ConfState) bool {
	return len(cs.Voters) == 0 && len(cs.VotersOutgoing) == 0 &&
		len(cs.Learners) == 0 && len(cs.LearnersNext) == 0
}