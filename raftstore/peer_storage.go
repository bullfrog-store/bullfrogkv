package raftstore

import (
	"bullfrogkv/raftstore/meta"
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/raftstore/snap"
	"bullfrogkv/storage"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"log"
)

type peerStorage struct {
	engine           *storage.Engines
	raftState        *raftstorepb.RaftLocalState
	applyState       *raftstorepb.RaftApplyState
	confState        *raftpb.ConfState
	snapshotState    *snap.SnapshotState
	snapshotTryCount int
}

func newPeerStorage(path string) *peerStorage {
	// snapshotState and snapshotTryCount need not to init
	ps := &peerStorage{
		engine: storage.NewEngines(path+storage.KvPath, path+storage.MetaPath),
	}
	ps.raftState = meta.InitRaftLocalState(ps.engine)
	ps.applyState = meta.InitRaftApplyState(ps.engine)
	ps.confState = meta.InitConfState(ps.engine)
	ps.snapshotState = &snap.SnapshotState{
		StateType: snap.SnapshotToGen,
	}
	return ps
}

func (ps *peerStorage) InitialState() (raftpb.HardState, raftpb.ConfState, error) {
	raftState := ps.raftState
	if ps.isEmptyHardState(*ps.raftState.HardState) {
		log.Printf("[Peer Storage]: local state %+v is empty", raftState)
		return raftpb.HardState{}, raftpb.ConfState{}, nil
	}
	if isEmptyConfState(*ps.confState) {
		return *raftState.HardState, raftpb.ConfState{}, nil
	}
	return *raftState.HardState, *ps.confState, nil
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
	if i == ps.applyState.TruncatedState.Index {
		return ps.applyState.TruncatedState.Term, nil
	}
	if err := ps.checkRange(i, i+1); err != nil {
		return 0, err
	}
	if ps.applyState.TruncatedState.Term == ps.raftState.LastTerm ||
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
	return ps.applyState.TruncatedState.Index + 1, nil
}

func (ps *peerStorage) Snapshot() (raftpb.Snapshot, error) {
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
	if ps.snapshotTryCount >= 5 {
		err := errors.Errorf("Failed to get snapshot after %d times", ps.snapshotTryCount)
		ps.snapshotTryCount = 0
		return snapshot, err
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
			Term:  term,
			Index: idx,
		},
	}

	data := ps.engine.KVSnapshot()
	snapshot.Data = data
	ps.snapshotState.Receiver <- snapshot
}

func (ps *peerStorage) AppliedIndex() uint64 {
	return ps.applyState.ApplyIndex
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
		truncatedIndex := ps.applyState.TruncatedState.Index
		entries = entries[truncatedIndex-entries[0].Index+1:]
	}
	var raftStateUpdated bool
	if ps.raftState.LastTerm == entries[len(entries)-1].Term &&
		ps.raftState.LastIndex == entries[len(entries)-1].Index {
		raftStateUpdated = false
	} else {
		raftStateUpdated = true
	}
	ps.raftLogEntriesWriteToDB(entries)
	localLastIndex, _ := ps.LastIndex()
	if localLastIndex > lastIndex {
		var shouldDeleteEntries []raftpb.Entry
		for i := lastIndex + 1; i <= localLastIndex; i++ {
			shouldDeleteEntries = append(shouldDeleteEntries, raftpb.Entry{
				Index: i,
			})
		}
		ps.raftLogEntriesDeleteDB(shouldDeleteEntries)
	}
	ps.raftState.LastTerm = entries[len(entries)-1].Term
	ps.raftState.LastIndex = entries[len(entries)-1].Index
	return raftStateUpdated
}

func (ps *peerStorage) applySnapshot(snapshot raftpb.Snapshot) bool {
	var raftStateUpdated bool
	if ps.raftState.LastTerm == snapshot.Metadata.Term &&
		ps.raftState.LastIndex == snapshot.Metadata.Index {
		raftStateUpdated = false
	} else {
		raftStateUpdated = true
	}
	ps.raftState.LastIndex = snapshot.Metadata.Index
	ps.raftState.LastTerm = snapshot.Metadata.Term
	ps.applyState.ApplyIndex = snapshot.Metadata.Index
	ps.applyState.TruncatedState.Index = snapshot.Metadata.Index
	ps.applyState.TruncatedState.Term = snapshot.Metadata.Term
	//ps.snapshotState.StateType = snap.SnapshotApplying
	// persist
	ps.raftApplyStateWriteToDB(ps.applyState)
	go ps.applySnapToDB(snapshot.Data)
	return raftStateUpdated
}

func (ps *peerStorage) saveReadyState(rd raft.Ready) error {
	// make sure ready.Snapshot is not nil
	var raftStateUpdatedAfterApply bool
	if !raft.IsEmptySnap(rd.Snapshot) {
		raftStateUpdatedAfterApply = ps.applySnapshot(rd.Snapshot)
	}
	raftStateUpdatedAfterAppend := ps.appendAndUpdate(rd.Entries)
	if !ps.isEmptyHardState(rd.HardState) {
		ps.raftState.HardState = &rd.HardState
		raftStateUpdatedAfterAppend = true
	}
	// persist raft local state once it is changed
	if raftStateUpdatedAfterApply || raftStateUpdatedAfterAppend {
		err := ps.raftLocalStateWriteToDB(ps.raftState)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *peerStorage) applySnapToDB(data []byte) {
	ps.snapshotState.StateType = snap.SnapshotApplying
	pairs := storage.Decode(data)
	for _, pair := range pairs {
		ps.engine.WriteKV(storage.PutData(pair.Key, pair.Val, true))
	}
	ps.snapshotState.StateType = snap.SnapshotApplied
}

func (ps *peerStorage) raftLogEntriesDeleteDB(entries []raftpb.Entry) error {
	for _, entry := range entries {
		key := meta.RaftLogEntryKey(entry.Index)
		modify := storage.DeleteMeta(key, true)
		if err := ps.engine.WriteMeta(modify); err != nil {
			return err
		}
	}
	return nil
}

func (ps *peerStorage) raftLogEntriesWriteToDB(entries []raftpb.Entry) error {
	for _, entry := range entries {
		key := meta.RaftLogEntryKey(entry.Index)
		if err := ps.doWriteToDB(key, &entry, true); err != nil {
			return err
		}
	}
	return nil
}

func (ps *peerStorage) raftLocalStateWriteToDB(localState *raftstorepb.RaftLocalState) error {
	key := meta.RaftLocalStateKey()
	if err := ps.doWriteToDB(key, localState, true); err != nil {
		return err
	}
	return nil
}

func (ps *peerStorage) raftApplyStateWriteToDB(applyState *raftstorepb.RaftApplyState) error {
	key := meta.RaftApplyStateKey()
	if err := ps.doWriteToDB(key, applyState, true); err != nil {
		return err
	}
	return nil
}

func (ps *peerStorage) raftConfStateWriteToDB(confState *raftpb.ConfState) error {
	return ps.doWriteToDB(meta.RaftConfStateKey(), confState, true)
}

func (ps *peerStorage) doWriteToDB(key []byte, msg proto.Message, sync bool) error {
	modify := storage.PutMeta(key, msg, sync)
	if err := ps.engine.WriteMeta(modify); err != nil {
		return err
	}
	return nil
}

func (ps *peerStorage) isHardStateChanged(recentRaftState raftpb.HardState) bool {
	if ps.raftState.HardState.Term == recentRaftState.Term && ps.raftState.HardState.Vote == recentRaftState.Vote &&
		ps.raftState.HardState.Commit == recentRaftState.Commit {
		return false
	}
	return true
}

func (ps *peerStorage) isEmptyHardState(hardState raftpb.HardState) bool {
	emptyHardState := raftpb.HardState{}
	if hardState.Term == emptyHardState.Term && hardState.Vote == emptyHardState.Vote &&
		hardState.Commit == emptyHardState.Commit {
		return true
	}
	return false
}

func (ps *peerStorage) checkRange(lo, hi uint64) error {
	if lo > hi {
		return errors.Errorf("Range error: low %d is greater than high %d", lo, hi)
	} else if hi > ps.raftState.LastIndex+1 {
		return errors.Errorf("Range error: high %d is out of bound", hi)
	} else if lo <= ps.applyState.TruncatedState.Index {
		return raft.ErrCompacted
	}
	return nil
}

func (ps *peerStorage) isValidateSnapshot(snapshot raftpb.Snapshot) bool {
	index := snapshot.Metadata.Index
	if index < ps.applyState.TruncatedState.Index {
		return false
	}
	return true
}
