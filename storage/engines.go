package storage

import (
	"errors"
	"github.com/cockroachdb/pebble"
	"os"
)

var (
	ErrUnknownModify = errors.New("unknown modify type")
)

const (
	KvPath   = "/kv"
	RaftPath = "/raft"
)

type Engines struct {
	// data
	kv     *storage
	kvPath string
	// metadata used by raft
	raft     *storage
	raftPath string
}

func NewEngines(kvPath, raftPath string) *Engines {
	opts := (&pebble.Options{}).EnsureDefaults()
	return &Engines{
		kv:       newStorage(kvPath, opts),
		kvPath:   kvPath,
		raft:     newStorage(raftPath, opts),
		raftPath: raftPath,
	}
}

func (e *Engines) WriteKV(m Modify) error {
	switch m.Data.(type) {
	case Put:
		return e.kv.Set(m.Key(), m.Value(), m.Sync())
	case Delete:
		return e.kv.Delete(m.Key(), m.Sync())
	default:
		return ErrUnknownModify
	}
}

func (e *Engines) WriteRaft(m Modify) error {
	switch m.Data.(type) {
	case Put:
		return e.raft.Set(m.Key(), m.Value(), m.Sync())
	case Delete:
		return e.raft.Delete(m.Key(), m.Sync())
	default:
		return ErrUnknownModify
	}
}

func (e *Engines) Close() error {
	if err := e.kv.Close(); err != nil {
		return err
	}
	if err := e.raft.Close(); err != nil {
		return err
	}
	return nil
}

func (e *Engines) Destroy() error {
	if err := e.Close(); err != nil {
		return err
	}
	if err := os.RemoveAll(e.kvPath); err != nil {
		return err
	}
	if err := os.RemoveAll(e.raftPath); err != nil {
		return err
	}
	return nil
}
