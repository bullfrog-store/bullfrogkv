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
	engine           storage.Engine
	raftState        *raftstorepb.RaftLocalState
	applyState       *raftstorepb.RaftApplyState
	confState        *raftpb.ConfState
	snapshotState    *snap.SnapshotState
	snapshotTryCount int
}

func newPeerStorage(path string) (*peerStorage, error) {
	var err error
	ps := &peerStorage{
		snapshotState: &snap.SnapshotState{
			StateType: snap.SnapshotToGen,
		},
	}

	if ps.engine, err = storage.New(storage.NewDefaultConfig(path)); err != nil {
		return nil, err
	}
	if ps.raftState, err = meta.InitRaftLocalState(ps.engine); err != nil {
		return nil, err
	}
	if ps.applyState, err = meta.InitRaftApplyState(ps.engine); err != nil {
		return nil, err
	}
	if ps.confState, err = meta.InitConfState(ps.engine); err != nil {
		return nil, err
	}

	return ps, nil
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
		val, err := ps.engine.ReadMeta(key)
		if err != nil {
			return nil, err
		}
		var entry raftpb.Entry
		if err = entry.Unmarshal(val); err != nil {
			return nil, err
		}
		// Maybe here has been compacted
		if entry.Index != entryIndex {
			break
		}
		entryIndex++
		entries = append(entries, entry)
	}
	// Here is correct result.
	if len(entries) == entrySize {
		return entries, nil
	}
	// We can't fetch enough entries.
	return nil, raft.ErrUnavailable
}

func (ps *peerStorage) Term(i uint64) (uint64, error) {
	// If index is truncated log.index.
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
	val, err := ps.engine.ReadMeta(key)
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
			if ps.isSnapshotValid(snapshot) {
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
	go ps.generateSnapshot()
	return snapshot, raft.ErrSnapshotTemporarilyUnavailable
}

func (ps *peerStorage) generateSnapshot() {
	idx := ps.appliedIndex()
	term, err := ps.Term(idx)
	if err != nil {
		logger.Errorf("snapshot generation error: %s", err.Error())
		return
	}
	snapshot := &raftpb.Snapshot{
		Metadata: raftpb.SnapshotMetadata{
			ConfState: *ps.confState,
			Index:     idx,
			Term:      term,
		},
	}

	snapshot.Data, err = ps.engine.DataSnapshot()
	if err != nil {
		logger.Errorf("snapshot generation error: %s", err.Error())
		return
	}
	logger.Infof("send snapshot")
	ps.snapshotState.Receiver <- snapshot
}

func (ps *peerStorage) appliedIndex() uint64 {
	return ps.applyState.ApplyIndex
}

func (ps *peerStorage) truncateIndex() uint64 {
	return ps.applyState.TruncatedState.Index
}

func (ps *peerStorage) truncateTerm() uint64 {
	return ps.applyState.TruncatedState.Term
}

// Append the entries to raft log and update raftState
func (ps *peerStorage) appendAndUpdate(entries []raftpb.Entry) (bool, error) {
	if len(entries) == 0 {
		return false, nil
	}
	localFirst, _ := ps.FirstIndex()
	lastIndex := entries[len(entries)-1].Index
	if localFirst > lastIndex {
		return false, nil
	}

	// Some entries have been compacted already, so we should truncate entries
	if localFirst > entries[0].Index {
		truncatedIndex := ps.truncateIndex()
		entries = entries[truncatedIndex-entries[0].Index+1:]
	}

	if err := ps.appendRaftLogEntries(entries); err != nil {
		return false, err
	}
	localLastIndex, err := ps.LastIndex()
	if err != nil {
		return false, err
	}
	if localLastIndex > lastIndex {
		var shouldDeleteEntries []raftpb.Entry
		for i := lastIndex + 1; i <= localLastIndex; i++ {
			shouldDeleteEntries = append(shouldDeleteEntries, raftpb.Entry{
				Index: i,
			})
		}
		if err = ps.deleteRaftLogEntries(shouldDeleteEntries); err != nil {
			return false, err
		}
	}

	needSync := false
	if ps.raftState.LastTerm != entries[len(entries)-1].Term ||
		ps.raftState.LastIndex != entries[len(entries)-1].Index {
		ps.raftState.LastTerm = entries[len(entries)-1].Term
		ps.raftState.LastIndex = entries[len(entries)-1].Index
		needSync = true
	}
	return needSync, nil
}

func (ps *peerStorage) applySnapshot(snapshot raftpb.Snapshot) (bool, error) {
	needSync := false
	if ps.raftState.LastTerm != snapshot.Metadata.Term ||
		ps.raftState.LastIndex != snapshot.Metadata.Index {
		ps.raftState.LastIndex = snapshot.Metadata.Index
		ps.raftState.LastTerm = snapshot.Metadata.Term
		needSync = true
	}

	var err error
	ps.applyState.ApplyIndex = snapshot.Metadata.Index
	ps.applyState.TruncatedState.Index = snapshot.Metadata.Index
	ps.applyState.TruncatedState.Term = snapshot.Metadata.Term
	if err = ps.writeRaftApplyState(ps.applyState); err != nil {
		return false, err
	}

	ps.confState = &snapshot.Metadata.ConfState
	if err = ps.writeRaftConfState(ps.confState); err != nil {
		return false, err
	}

	go ps.doApplySnapshot(snapshot.Data)
	return needSync, nil
}

func (ps *peerStorage) saveReadyState(rd raft.Ready) error {
	var err error
	needSync := false
	if !raft.IsEmptySnap(rd.Snapshot) {
		if needSync, err = ps.applySnapshot(rd.Snapshot); err != nil {
			return err
		}
	}
	result, err := ps.appendAndUpdate(rd.Entries)
	if err != nil {
		return err
	}

	needSync = needSync || result
	if !raft.IsEmptyHardState(rd.HardState) {
		ps.raftState.HardState = &rd.HardState
		needSync = true
	}
	// Persist raft local state once it is changed.
	if needSync {
		err = ps.writeRaftLocalState(ps.raftState)
	}
	return err
}

func (ps *peerStorage) doApplySnapshot(data []byte) {
	logger.Infof("applying snapshot data...")
	ps.snapshotState.StateType = snap.SnapshotApplying
	entries, err := storage.DeserializeMulti(data)
	if err != nil {
		logger.Errorf("get error when applying snapshot data, error: %s", err.Error())
		return
	}
	for _, entry := range entries {
		if err = ps.engine.WriteData(storage.PutData(entry.Key, entry.Value, true)); err != nil {
			logger.Errorf("get error when applying snapshot data, error: %s", err.Error())
			return
		}
	}
	ps.snapshotState.StateType = snap.SnapshotApplied
}

func (ps *peerStorage) deleteRaftLogEntries(entries []raftpb.Entry) error {
	for _, entry := range entries {
		key := meta.RaftLogEntryKey(entry.Index)
		if err := ps.deleteMeta(key); err != nil {
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

func (ps *peerStorage) writeRaftLocalState(localState *raftstorepb.RaftLocalState) error {
	return ps.putMeta(meta.RaftLocalStateKey(), localState)
}

func (ps *peerStorage) writeRaftApplyState(applyState *raftstorepb.RaftApplyState) error {
	return ps.putMeta(meta.RaftApplyStateKey(), applyState)
}

func (ps *peerStorage) writeRaftConfState(confState *raftpb.ConfState) error {
	return ps.putMeta(meta.RaftConfStateKey(), confState)
}

func (ps *peerStorage) putMeta(key []byte, msg proto.Message) error {
	modify := storage.PutMeta(key, msg, true)
	return ps.engine.WriteMeta(modify)
}

func (ps *peerStorage) deleteMeta(key []byte) error {
	modify := storage.DeleteMeta(key, true)
	return ps.engine.WriteMeta(modify)
}

func (ps *peerStorage) checkRange(lo, hi uint64) error {
	if lo > hi {
		return errors.Errorf("range error: low %d is greater than high %d", lo, hi)
	} else if hi > ps.raftState.LastIndex+1 {
		return errors.Errorf("range error: high %d is out of bound", hi)
	} else if lo <= ps.truncateIndex() {
		return raft.ErrCompacted
	}
	return nil
}

func (ps *peerStorage) isSnapshotValid(snapshot raftpb.Snapshot) bool {
	return snapshot.Metadata.Index >= ps.truncateIndex()
}
