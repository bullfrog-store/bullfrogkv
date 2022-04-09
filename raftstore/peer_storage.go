package raftstore

import (
	"bullfrogkv/config"
	"bullfrogkv/logger"
	"bullfrogkv/raftstore/meta"
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/raftstore/snap"
	"bullfrogkv/storage"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type peerStorage struct {
	engines          *storage.Engines
	raftState        *raftstorepb.RaftLocalState
	applyState       *raftstorepb.RaftApplyState
	confState        *raftpb.ConfState
	snapshotState    *snap.SnapshotState
	snapshotTryCount int
}

func newPeerStorage(path string) *peerStorage {
	// snapshotState and snapshotTryCount need not to init
	ps := &peerStorage{
		engines: storage.NewEngines(path+storage.KvPath, path+storage.MetaPath),
	}
	ps.raftState = meta.InitRaftLocalState(ps.engines)
	ps.applyState = meta.InitRaftApplyState(ps.engines)
	ps.confState = meta.InitConfState(ps.engines)
	ps.snapshotState = &snap.SnapshotState{
		StateType: snap.SnapshotToGen,
	}
	return ps
}

func (ps *peerStorage) InitialState() (raftpb.HardState, raftpb.ConfState, error) {
	hs := raftpb.HardState{}
	cs := raftpb.ConfState{}
	if !raft.IsEmptyHardState(*ps.raftState.HardState) {
		hs = *ps.raftState.HardState
	}
	if !isEmptyConfState(*ps.confState) {
		cs = *ps.confState
	}
	return hs, cs, nil
}

func (ps *peerStorage) Entries(lo, hi, maxSize uint64) ([]raftpb.Entry, error) {
	if err := ps.checkRange(lo, hi); err != nil || lo == hi {
		return nil, err
	}
	entryIndex := lo
	entrySize := int(hi - lo)
	entries := make([]raftpb.Entry, 0, entrySize)
	for i := lo; i < hi; i++ {
		key := meta.RaftLogEntryKey(i)
		val, err := ps.engines.ReadMeta(key)
		if err != nil {
			return nil, err
		}
		var entry raftpb.Entry
		if err := entry.Unmarshal(val); err != nil {
			return nil, err
		}
		// Maybe here has been compacted
		if entry.Index != entryIndex {
			break
		}
		entryIndex++
		entries = append(entries, entry)
	}
	// Here is correct result
	if len(entries) == entrySize {
		return entries, nil
	}
	// We can't fetch enough entries
	return nil, raft.ErrUnavailable
}

func (ps *peerStorage) Term(i uint64) (uint64, error) {
	// If index is truncated log.index
	if i == ps.truncateIndex() {
		return ps.truncateTerm(), nil
	}
	if err := ps.checkRange(i, i+1); err != nil {
		return 0, err
	}
	if ps.truncateTerm() == ps.raftState.LastTerm ||
		i == ps.raftState.LastIndex {
		return ps.raftState.LastTerm, nil
	}
	key := meta.RaftLogEntryKey(i)
	val, err := ps.engines.ReadMeta(key)
	if err != nil {
		return 0, err
	}
	var entry raftpb.Entry
	if err = entry.Unmarshal(val); err != nil {
		return 0, err
	}
	return entry.Term, nil
}

func (ps *peerStorage) LastIndex() (uint64, error) {
	return ps.raftState.LastIndex, nil
}

func (ps *peerStorage) FirstIndex() (uint64, error) {
	return ps.truncateIndex() + 1, nil
}

func (ps *peerStorage) Snapshot() (raftpb.Snapshot, error) {
	logger.Infof("follower need snapshot, generating...")
	var snapshot raftpb.Snapshot
	if ps.snapshotState.StateType == snap.SnapshotGenerating {
		select {
		case s := <-ps.snapshotState.Receiver:
			if s != nil {
				snapshot = *s
			}
		default:
			return snapshot, raft.ErrSnapshotTemporarilyUnavailable
		}
		ps.snapshotState.StateType = snap.SnapshotRelaxed
		if !raft.IsEmptySnap(snapshot) {
			ps.snapshotTryCount = 0
			if ps.isValidateSnapshot(snapshot) {
				return snapshot, nil
			}
		}
	}
	if ps.snapshotTryCount >= config.GlobalConfig.RaftConfig.SnapshotTryCount {
		ps.snapshotTryCount = 0
		return snapshot, errors.Errorf("Failed to get snapshot after %d times", ps.snapshotTryCount)
	}

	ps.snapshotTryCount++
	receiver := make(chan *raftpb.Snapshot, 1)
	ps.snapshotState = &snap.SnapshotState{
		StateType: snap.SnapshotGenerating,
		Receiver:  receiver,
	}
	// Schedule a snapshot generating task
	go ps.doSnapshot()
	return snapshot, raft.ErrSnapshotTemporarilyUnavailable
}

func (ps *peerStorage) doSnapshot() {
	idx := ps.AppliedIndex()
	term, err := ps.Term(idx)
	if err != nil {
		return
	}
	snapshot := &raftpb.Snapshot{
		Metadata: raftpb.SnapshotMetadata{
			ConfState: *ps.confState,
			Index:     idx,
			Term:      term,
		},
	}

	snapshot.Data = ps.engines.KVSnapshot()
	logger.Infof("send snapshot")
	ps.snapshotState.Receiver <- snapshot
}

func (ps *peerStorage) AppliedIndex() uint64 {
	return ps.applyState.ApplyIndex
}

func (ps *peerStorage) truncateIndex() uint64 {
	return ps.applyState.TruncatedState.Index
}

func (ps *peerStorage) truncateTerm() uint64 {
	return ps.applyState.TruncatedState.Term
}

// append the entries to raft log and update raftState
func (ps *peerStorage) appendAndUpdate(entries []raftpb.Entry) bool {
	if len(entries) == 0 {
		return false
	}
	localFirst, _ := ps.FirstIndex()
	lastIndex := entries[len(entries)-1].Index
	if localFirst > lastIndex {
		return false
	}
	// some entries have been compacted already,
	// so we should truncate entries
	if localFirst > entries[0].Index {
		truncatedIndex := ps.truncateIndex()
		entries = entries[truncatedIndex-entries[0].Index+1:]
	}

	ps.appendRaftLogEntries(entries)
	localLastIndex, _ := ps.LastIndex()
	if localLastIndex > lastIndex {
		var shouldDeleteEntries []raftpb.Entry
		for i := lastIndex + 1; i <= localLastIndex; i++ {
			shouldDeleteEntries = append(shouldDeleteEntries, raftpb.Entry{
				Index: i,
			})
		}
		ps.deleteRaftLogEntries(shouldDeleteEntries)
	}

	needPersist := false
	if ps.raftState.LastTerm != entries[len(entries)-1].Term ||
		ps.raftState.LastIndex != entries[len(entries)-1].Index {
		ps.raftState.LastTerm = entries[len(entries)-1].Term
		ps.raftState.LastIndex = entries[len(entries)-1].Index
		needPersist = true
	}
	return needPersist
}

func (ps *peerStorage) applySnapshot(snapshot raftpb.Snapshot) bool {
	needPersist := false
	if ps.raftState.LastTerm != snapshot.Metadata.Term ||
		ps.raftState.LastIndex != snapshot.Metadata.Index {
		ps.raftState.LastIndex = snapshot.Metadata.Index
		ps.raftState.LastTerm = snapshot.Metadata.Term
		needPersist = true
	}

	ps.applyState.ApplyIndex = snapshot.Metadata.Index
	ps.applyState.TruncatedState.Index = snapshot.Metadata.Index
	ps.applyState.TruncatedState.Term = snapshot.Metadata.Term
	ps.raftApplyStateWriteToDB(ps.applyState)

	ps.confState = &snapshot.Metadata.ConfState
	ps.raftConfStateWriteToDB(ps.confState)

	go ps.applySnapToDB(snapshot.Data)
	return needPersist
}

func (ps *peerStorage) saveReadyState(rd raft.Ready) error {
	// make sure ready.Snapshot is not nil
	needPersist := false
	if !raft.IsEmptySnap(rd.Snapshot) {
		needPersist = ps.applySnapshot(rd.Snapshot)
	}
	needPersist = needPersist || ps.appendAndUpdate(rd.Entries)
	if !raft.IsEmptyHardState(rd.HardState) {
		ps.raftState.HardState = &rd.HardState
		needPersist = true
	}
	// persist raft local state once it is changed
	if needPersist {
		if err := ps.raftLocalStateWriteToDB(ps.raftState); err != nil {
			return err
		}
	}
	return nil
}

func (ps *peerStorage) applySnapToDB(data []byte) {
	logger.Infof("apply snapshot")
	ps.snapshotState.StateType = snap.SnapshotApplying
	pairs := storage.Decode(data)
	for _, pair := range pairs {
		ps.engines.WriteKV(storage.PutData(pair.Key, pair.Val, true))
	}
	ps.snapshotState.StateType = snap.SnapshotApplied
}

func (ps *peerStorage) deleteRaftLogEntries(entries []raftpb.Entry) error {
	for _, entry := range entries {
		key := meta.RaftLogEntryKey(entry.Index)
		if err := ps.deleteMeta(key, &entry); err != nil {
			return err
		}
	}
	return nil
}

func (ps *peerStorage) appendRaftLogEntries(entries []raftpb.Entry) error {
	for _, entry := range entries {
		key := meta.RaftLogEntryKey(entry.Index)
		if err := ps.putMeta(key, &entry); err != nil {
			return err
		}
	}
	return nil
}

func (ps *peerStorage) raftLocalStateWriteToDB(localState *raftstorepb.RaftLocalState) error {
	return ps.putMeta(meta.RaftLocalStateKey(), localState)
}

func (ps *peerStorage) raftApplyStateWriteToDB(applyState *raftstorepb.RaftApplyState) error {
	return ps.putMeta(meta.RaftApplyStateKey(), applyState)
}

func (ps *peerStorage) raftConfStateWriteToDB(confState *raftpb.ConfState) error {
	return ps.putMeta(meta.RaftConfStateKey(), confState)
}

func (ps *peerStorage) putMeta(key []byte, msg proto.Message) error {
	modify := storage.PutMeta(key, msg, true)
	return ps.engines.WriteMeta(modify)
}

func (ps *peerStorage) deleteMeta(key []byte, msg proto.Message) error {
	modify := storage.DeleteMeta(key, true)
	return ps.engines.WriteMeta(modify)
}

func (ps *peerStorage) checkRange(lo, hi uint64) error {
	if lo > hi {
		return errors.Errorf("Range error: low %d is greater than high %d", lo, hi)
	} else if hi > ps.raftState.LastIndex+1 {
		return errors.Errorf("Range error: high %d is out of bound", hi)
	} else if lo <= ps.truncateIndex() {
		return raft.ErrCompacted
	}
	return nil
}

func (ps *peerStorage) isValidateSnapshot(snapshot raftpb.Snapshot) bool {
	return snapshot.Metadata.Index >= ps.truncateIndex()
}
