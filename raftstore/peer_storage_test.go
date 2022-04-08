package raftstore

import (
	"bullfrogkv/raftstore/meta"
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/raftstore/snap"
	"bullfrogkv/storage"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"testing"
	"time"
)

const (
	enginePath = "../test_data"
)

func newTestPeerStorage() *peerStorage {
	return newPeerStorage(enginePath)
}

func newTestPsWithPath(path string) *peerStorage {
	return newPeerStorage(path)
}

func newTestPeerStorageFromEntries(t *testing.T, entries []raftpb.Entry) *peerStorage {
	ps := newTestPeerStorage()
	ps.appendAndUpdate(entries[1:])
	applyState := ps.applyState
	applyState.TruncatedState = &raftstorepb.RaftTruncatedState{
		Term:  entries[0].Term,
		Index: entries[0].Index,
	}
	applyState.ApplyIndex = entries[len(entries)-1].Index
	err := ps.raftLocalStateWriteToDB(ps.raftState)
	require.Nil(t, err)
	err = ps.raftLogEntriesWriteToDB(entries)
	require.Nil(t, err)
	err = ps.raftApplyStateWriteToDB(ps.applyState)
	require.Nil(t, err)
	return ps
}

func cleanUpData(ps *peerStorage) {
	if err := ps.engines.Destroy(); err != nil {
		panic(err)
	}
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
		ps := newTestPeerStorageFromEntries(t, entries)
		term, err := ps.Term(testData.index)
		if err != nil {
			assert.Equal(t, testData.err, err)
		} else {
			assert.Equal(t, testData.term, term)
		}
		cleanUpData(ps)
	}
}

func TestPeerStorageFirstIndex(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	ps := newTestPeerStorageFromEntries(t, entries)
	firstIndex, err := ps.FirstIndex()
	assert.Nil(t, err)
	assert.Equal(t, uint64(4), firstIndex)
	cleanUpData(ps)
}

func TestPeerStorageLastIndex(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	ps := newTestPeerStorageFromEntries(t, entries)
	lastIndex, err := ps.LastIndex()
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), lastIndex)
	cleanUpData(ps)
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
	const maxSize = 0
	for _, testData := range testDatas {
		ps := newTestPeerStorageFromEntries(t, entries)
		// truncated.Index = 3, firstIndex = 4, lastIndex = 7,
		ents, err := ps.Entries(testData.lo, testData.hi, maxSize)
		if err != nil {
			assert.Equal(t, testData.err, err)
		} else {
			assert.Equal(t, testData.ents, ents)
		}
		cleanUpData(ps)
	}
}

func TestPeerStorageAppliedIndex(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	ps := newTestPeerStorageFromEntries(t, entries)
	defer cleanUpData(ps)
	applyIndex := ps.AppliedIndex()
	assert.Equal(t, uint64(5), applyIndex)
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
		ps := newTestPeerStorageFromEntries(t, entries)
		ps.appendAndUpdate(testData.append)
		const maxSize = 0
		firstIndex, _ := ps.FirstIndex()
		lastIndex, _ := ps.LastIndex()
		result, err := ps.Entries(firstIndex, lastIndex+1, maxSize)
		assert.Nil(t, err)
		assert.Equal(t, testData.result, result)
		cleanUpData(ps)
	}
}

func TestPeerStorageRestart(t *testing.T) {
	entries := []raftpb.Entry{
		newTestEntry(3, 3),
		newTestEntry(4, 4),
		newTestEntry(5, 5),
	}
	// Start peer storage with given entries
	ps := newTestPeerStorageFromEntries(t, entries)
	assert.Nil(t, ps.engines.Close())
	// Restart peer storage without given entries
	time.Sleep(1)
	ps = newTestPeerStorage()
	assert.Equal(t, uint64(5), ps.raftState.LastTerm)
	assert.Equal(t, uint64(5), ps.raftState.LastIndex)
	assert.Equal(t, uint64(5), ps.applyState.ApplyIndex)
	assert.Equal(t, uint64(3), ps.applyState.TruncatedState.Term)
	assert.Equal(t, uint64(3), ps.applyState.TruncatedState.Index)
	for index := 3; index <= 5; index++ {
		key := meta.RaftLogEntryKey(uint64(index))
		val, err := ps.engines.ReadMeta(key)
		assert.Nil(t, err)
		var entry raftpb.Entry
		assert.Nil(t, entry.Unmarshal(val))
		assert.Equal(t, newTestEntry(uint64(index), uint64(index)), entry)
	}
}

func setToPebble(t *testing.T) *peerStorage {
	entries := []raftpb.Entry{
		newTestEntry(2, 3),
		newTestEntry(2, 4),
		newTestEntry(2, 5),
		newTestEntry(2, 6),
	}
	ps := newTestPeerStorageFromEntries(t, entries)

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
		ps.engines.WriteKV(storage.Modify{
			Data: storage.Put{
				Key:   data.Key,
				Value: data.Val,
				Sync:  true,
			},
		})
	}
	return ps
}

func TestSetToPebble(t *testing.T) {
	setToPebble(t)
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
	ss := storage.Decode(snapshot.Data)
	for _, s := range ss {
		fmt.Println(string(s.Key), ":", string(s.Val))
	}
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
	//ss := storage.Decode(snapshot.Data)
	//for _, s := range ss {
	//	fmt.Println(string(s.Key), ":", string(s.Val))
	//}

	newPs := newTestPsWithPath("../test_data_new")
	newPs.applySnapshot(snapshot)
	for newPs.snapshotState.StateType != snap.SnapshotApplied {

	}
	val, err := newPs.engines.ReadKV([]byte("1"))
	assert.Nil(t, err)
	fmt.Println(string(val))

	val, err = newPs.engines.ReadKV([]byte("2"))
	assert.Nil(t, err)
	fmt.Println(string(val))

	val, err = newPs.engines.ReadKV([]byte("3"))
	assert.Nil(t, err)
	fmt.Println(string(val))

	val, err = newPs.engines.ReadKV([]byte("4"))
	assert.Nil(t, err)
	fmt.Println(string(val))
	cleanUpData(ps)
	cleanUpData(newPs)
}
