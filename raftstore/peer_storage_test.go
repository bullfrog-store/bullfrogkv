package raftstore

import (
	"bullfrogkv/config"
	"bullfrogkv/raftstore/meta"
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/raftstore/snap"
	"bullfrogkv/storage"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"testing"
	"time"
)

const (
	testPeerStoragePath = "../testdata"
)

func newTestPeerStorage() (*peerStorage, error) {
	return newPeerStorage(testPeerStoragePath)
}

func newTestPeerStorageWithPath(path string) (*peerStorage, error) {
	return newPeerStorage(path)
}

func newTestPeerStorageFromEntries(entries []raftpb.Entry) (*peerStorage, error) {
	ps, err := newTestPeerStorage()
	if err != nil {
		return nil, err
	}

	ps.appendAndUpdate(entries[1:])
	applyState := ps.applyState
	applyState.TruncatedState = &raftstorepb.RaftTruncatedState{
		Term:  entries[0].Term,
		Index: entries[0].Index,
	}
	applyState.ApplyIndex = entries[len(entries)-1].Index
	if err = ps.writeRaftLocalState(ps.raftState); err != nil {
		return nil, err
	}
	if err = ps.appendRaftLogEntries(entries); err != nil {
		return nil, err
	}
	if err = ps.writeRaftApplyState(ps.applyState); err != nil {
		return nil, err
	}
	return ps, nil
}

func cleanupData(ps *peerStorage) error {
	return ps.engine.Destroy()
}

func newTestEntry(term, index uint64) raftpb.Entry {
	return raftpb.Entry{
		Term:  term,
		Index: index,
		Data:  []byte("0"),
	}
}

func TestPeerStorageTerm(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	testDatas := []struct {
		term  uint64
		index uint64
		err   error
	}{
		{1, 2, raft.ErrCompacted},
		{3, 3, nil},
		{4, 4, nil},
		{5, 5, nil},
	}
	for _, testData := range testDatas {
		ps, err := newTestPeerStorageFromEntries(entries)
		assert.NoError(t, err)

		term, err := ps.Term(testData.index)
		if err != nil {
			assert.Equal(t, testData.err, err)
		} else {
			assert.Equal(t, testData.term, term)
		}
		assert.NoError(t, cleanupData(ps))
	}
}

func TestPeerStorageFirstIndex(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	ps, err := newTestPeerStorageFromEntries(entries)
	assert.NoError(t, err)

	firstIndex, err := ps.FirstIndex()
	assert.NoError(t, err)
	assert.Equal(t, uint64(4), firstIndex)

	assert.NoError(t, cleanupData(ps))
}

func TestPeerStorageLastIndex(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	ps, err := newTestPeerStorageFromEntries(entries)
	assert.NoError(t, err)

	lastIndex, err := ps.LastIndex()
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), lastIndex)

	assert.NoError(t, cleanupData(ps))
}

func TestPeerStorageEntries(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
		newTestEntry(6, 6),
		newTestEntry(7, 7),
	}
	testDatas := []struct {
		lo   uint64
		hi   uint64
		ents []raftpb.Entry
		err  error
	}{
		{3, 4, nil, raft.ErrCompacted},
		{4, 5, []raftpb.Entry{newTestEntry(4, 4)}, nil},
		{4, 8, []raftpb.Entry{
			newTestEntry(4, 4),
			newTestEntry(5, 5),
			newTestEntry(6, 6),
			newTestEntry(7, 7),
		}, nil},
	}

	for _, testData := range testDatas {
		ps, err := newTestPeerStorageFromEntries(entries)
		assert.NoError(t, err)

		// truncated.Index = 3, firstIndex = 4, lastIndex = 7,
		ents, err := ps.Entries(testData.lo, testData.hi, 0)
		if err != nil {
			assert.Equal(t, testData.err, err)
		} else {
			assert.Equal(t, testData.ents, ents)
		}

		assert.NoError(t, cleanupData(ps))
	}
}

func TestPeerStorageAppliedIndex(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	ps, err := newTestPeerStorageFromEntries(entries)
	assert.NoError(t, err)

	applyIndex := ps.appliedIndex()
	assert.Equal(t, uint64(5), applyIndex)

	assert.NoError(t, cleanupData(ps))
}

func TestPeerStorageAppendAndUpdate(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	testDatas := []struct {
		append []raftpb.Entry
		result []raftpb.Entry
	}{
		{
			[]raftpb.Entry{
				newTestEntry(3, 3),
				newTestEntry(4, 4),
			},
			[]raftpb.Entry{
				newTestEntry(4, 4),
			},
		},
		{
			[]raftpb.Entry{
				newTestEntry(3, 3),
				newTestEntry(6, 4),
				newTestEntry(6, 5),
			},
			[]raftpb.Entry{
				newTestEntry(6, 4),
				newTestEntry(6, 5),
			},
		},
		{
			[]raftpb.Entry{
				newTestEntry(3, 3),
				newTestEntry(4, 4),
				newTestEntry(5, 5),
				newTestEntry(5, 6),
			},
			[]raftpb.Entry{
				newTestEntry(4, 4),
				newTestEntry(5, 5),
				newTestEntry(5, 6),
			},
		},
		{
			[]raftpb.Entry{
				newTestEntry(3, 2),
				newTestEntry(3, 3),
				newTestEntry(5, 4),
			},

			[]raftpb.Entry{
				newTestEntry(5, 4),
			},
		},
		{
			[]raftpb.Entry{
				newTestEntry(5, 4),
			},
			[]raftpb.Entry{
				newTestEntry(5, 4),
			},
		},
		{
			[]raftpb.Entry{
				newTestEntry(5, 6),
			},
			[]raftpb.Entry{
				newTestEntry(4, 4),
				newTestEntry(5, 5),
				newTestEntry(5, 6),
			},
		},
	}
	for _, testData := range testDatas {
		ps, err := newTestPeerStorageFromEntries(entries)
		assert.NoError(t, err)

		ps.appendAndUpdate(testData.append)
		firstIndex, _ := ps.FirstIndex()
		lastIndex, _ := ps.LastIndex()
		result, err := ps.Entries(firstIndex, lastIndex+1, 0)
		assert.NoError(t, err)
		assert.Equal(t, testData.result, result)

		assert.NoError(t, cleanupData(ps))
	}
}

func TestPeerStorageRestart(t *testing.T) {
	var err error
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	// Start peer storage with given entries
	ps, err := newTestPeerStorageFromEntries(entries)
	assert.NoError(t, err)
	assert.NoError(t, ps.engine.Close())

	// Restart peer storage without given entries
	time.Sleep(1)
	ps, err = newTestPeerStorage()
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), ps.raftState.LastTerm)
	assert.Equal(t, uint64(5), ps.raftState.LastIndex)
	assert.Equal(t, uint64(5), ps.applyState.ApplyIndex)
	assert.Equal(t, uint64(3), ps.applyState.TruncatedState.Term)
	assert.Equal(t, uint64(3), ps.applyState.TruncatedState.Index)
	for index := 3; index <= 5; index++ {
		key := meta.RaftLogEntryKey(uint64(index))
		val, err := ps.engine.ReadMeta(key)
		assert.NoError(t, err)
		var entry raftpb.Entry
		assert.NoError(t, entry.Unmarshal(val))
		assert.Equal(t, newTestEntry(uint64(index), uint64(index)), entry)
	}

	assert.NoError(t, cleanupData(ps))
}

func setToPebble(t *testing.T) *peerStorage {
	config.GlobalConfig = &config.Config{
		RaftConfig: config.RaftConfig{
			SnapshotTryCount: 5,
		},
	}
	entries := []raftpb.Entry{
		newTestEntry(2, 3),
		newTestEntry(2, 4),
		newTestEntry(2, 5),
		newTestEntry(2, 6),
	}
	ps, err := newTestPeerStorageFromEntries(entries)
	assert.NoError(t, err)

	testDatas := []struct {
		Key []byte
		Val []byte
	}{
		{
			Key: []byte("1"),
			Val: []byte("1"),
		},
		{
			Key: []byte("2"),
			Val: []byte("2"),
		},
		{
			Key: []byte("3"),
			Val: []byte("3"),
		},
		{
			Key: []byte("4"),
			Val: []byte("4"),
		},
	}

	for _, data := range testDatas {
		assert.NoError(t, ps.engine.WriteData(storage.Modify{
			Data: storage.Put{
				Key:   data.Key,
				Value: data.Val,
				Sync:  true,
			},
		}))
	}
	return ps
}

func TestSetToPebble(t *testing.T) {
	ps := setToPebble(t)
	assert.NoError(t, cleanupData(ps))
}

func TestSnapshot(t *testing.T) {
	ps := setToPebble(t)
	snapshot, err := ps.Snapshot()
	if err != nil {
		assert.Equal(t, err, raft.ErrSnapshotTemporarilyUnavailable)
	}
	for {
		snapshot, err = ps.Snapshot()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	fmt.Println(snapshot.Metadata)
	ss, err := storage.DeserializeMulti(snapshot.Data)
	assert.NoError(t, err)
	for _, s := range ss {
		fmt.Println(string(s.Key), ":", string(s.Value))
	}
	assert.NoError(t, cleanupData(ps))
}

func TestApplySnap(t *testing.T) {
	ps := setToPebble(t)
	snapshot, err := ps.Snapshot()
	if err != nil {
		assert.Equal(t, err, raft.ErrSnapshotTemporarilyUnavailable)
	}

	for {
		snapshot, err = ps.Snapshot()
		if err == nil {
			break
		}
	}

	newps, err := newTestPeerStorageWithPath("../testdata2")
	assert.NoError(t, err)
	newps.applySnapshot(snapshot)
	for newps.snapshotState.StateType != snap.SnapshotApplied {

	}
	val, err := newps.engine.ReadData([]byte("1"))
	assert.NoError(t, err)
	fmt.Println(string(val))

	val, err = newps.engine.ReadData([]byte("2"))
	assert.NoError(t, err)
	fmt.Println(string(val))

	val, err = newps.engine.ReadData([]byte("3"))
	assert.NoError(t, err)
	fmt.Println(string(val))

	val, err = newps.engine.ReadData([]byte("4"))
	assert.NoError(t, err)
	fmt.Println(string(val))
	assert.NoError(t, cleanupData(ps))
	assert.NoError(t, cleanupData(newps))
}
