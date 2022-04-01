package storage

import (
	"errors"
	"github.com/cockroachdb/pebble"
	"os"
)

var (
	ErrUnknownModify = errors.New("unknown modify type")
	ErrNotFound      = errors.New("key doesn't exist")
)

const (
	KvPath   = "/kv"
	MetaPath = "/meta"
)

type Engines struct {
	// data
	kv     *storage
	kvPath string
	// metadata used by meta
	meta     *storage
	metaPath string
}

func NewEngines(kvPath, metaPath string) *Engines {
	opts := (&pebble.Options{}).EnsureDefaults()
	return &Engines{
		kv:       newStorage(kvPath, opts),
		kvPath:   kvPath,
		meta:     newStorage(metaPath, opts),
		metaPath: metaPath,
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

func (e *Engines) WriteMeta(m Modify) error {
	switch m.Data.(type) {
	case Put:
		return e.meta.Set(m.Key(), m.Value(), m.Sync())
	case Delete:
		return e.meta.Delete(m.Key(), m.Sync())
	default:
		return ErrUnknownModify
	}
}

func (e *Engines) ReadKV(key []byte) ([]byte, error) {
	val, err := e.kv.Get(key)
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, ErrNotFound
	}
	return val, err
}

func (e *Engines) ReadMeta(key []byte) ([]byte, error) {
	val, err := e.meta.Get(key)
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, ErrNotFound
	}
	return val, err
}

func (e *Engines) Close() error {
	if err := e.kv.Close(); err != nil {
		return err
	}
	if err := e.meta.Close(); err != nil {
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
	if err := os.RemoveAll(e.metaPath); err != nil {
		return err
	}
	return nil
}
