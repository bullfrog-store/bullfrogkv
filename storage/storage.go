package storage

import (
	"github.com/cockroachdb/pebble"
	"os"
)

type storage struct {
	db *pebble.DB
}

func newStorage(dir string, opts *pebble.Options) *storage {
	if !exist(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(err)
		}
	}

	db, err := pebble.Open(dir, opts)
	if err != nil {
		panic(err)
	}

	return &storage{db}
}

func (s *storage) set(key, val []byte, sync bool) error {
	return s.db.Set(key, val, toWriteOptions(sync))
}

func (s *storage) get(key []byte) ([]byte, error) {
	val, closer, err := s.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	if len(val) == 0 {
		return nil, nil
	}
	v := make([]byte, len(val))
	copy(v, val)
	return v, nil
}

func (s *storage) delete(key []byte, sync bool) error {
	return s.db.Delete(key, toWriteOptions(sync))
}

func (s *storage) snapshot() []byte {
	snap := s.db.NewSnapshot()
	iter := snap.NewIter(nil)
	pairs := make([]Pair, 0)
	for iter.First(); iter.Valid(); iter.Next() {
		pairs = append(pairs, Pair{Key: iter.Key(), Val: iter.Value()})
	}
	return Encode(pairs)
}

func (s *storage) close() error {
	return s.db.Close()
}

func toWriteOptions(sync bool) *pebble.WriteOptions {
	if sync {
		return pebble.Sync
	}
	return pebble.NoSync
}
