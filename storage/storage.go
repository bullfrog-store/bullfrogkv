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

func (s *storage) Set(key, val []byte, sync bool) error {
	return s.db.Set(key, val, toWriteOptions(sync))
}

func (s *storage) Get(key []byte) ([]byte, error) {
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

func (s *storage) Delete(key []byte, sync bool) error {
	return s.db.Delete(key, toWriteOptions(sync))
}

func (s *storage) Close() error {
	return s.db.Close()
}

func toWriteOptions(sync bool) *pebble.WriteOptions {
	if sync {
		return pebble.Sync
	}
	return pebble.NoSync
}
