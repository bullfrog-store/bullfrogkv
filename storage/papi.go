package storage

import (
	"errors"

	"github.com/cockroachdb/pebble"
)

func NewDefaultPebbleOpts() *pebble.Options {
	opts := &pebble.Options{}
	return opts.EnsureDefaults()
}

// papi => Pebble API wrapper.
type papi struct {
	db *pebble.DB
}

func newPapi(path string, opts *pebble.Options) (*papi, error) {
	if !exist(path) {
		if err := mkdir(path, true, papiDefaultPerm); err != nil {
			return nil, err
		}
	}

	if opts == nil {
		opts = NewDefaultPebbleOpts()
	}
	db, err := pebble.Open(path, opts)
	if err != nil {
		return nil, err
	}

	return &papi{db}, nil
}

func (p *papi) set(key, value []byte, sync bool) error {
	return p.db.Set(key, value, toWriteOptions(sync))
}

func (p *papi) get(key []byte) ([]byte, error) {
	v, closer, err := p.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	value := make([]byte, len(v))
	copy(value, v)
	return value, nil
}

func (p *papi) delete(key []byte, sync bool) error {
	return p.db.Delete(key, toWriteOptions(sync))
}

// FIXME: need more research.
func (p *papi) snapshot() ([]byte, error) {
	snap := p.db.NewSnapshot()
	defer snap.Close()

	iter := snap.NewIter(nil)
	defer iter.Close()

	entries := make([]Entry, 0)
	for iter.First(); iter.Valid(); iter.Next() {
		var key, value []byte

		k, v := iter.Key(), iter.Value()
		key = make([]byte, len(k))
		copy(key, k)
		value = make([]byte, len(v))
		copy(value, v)

		entries = append(entries, MakeEntry(key, value))
	}

	return SerializeMulti(entries), nil
}

func (p *papi) close() error {
	return p.db.Close()
}

func toWriteOptions(sync bool) *pebble.WriteOptions {
	if sync {
		return pebble.Sync
	}
	return pebble.NoSync
}
