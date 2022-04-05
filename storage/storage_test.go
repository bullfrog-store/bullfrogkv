package storage

import (
	"github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStorage(t *testing.T) {
	opts := &pebble.Options{}
	s := newStorage("/tmp/data", opts.EnsureDefaults())
	assert.NotNil(t, s)
}

func TestStorageOps(t *testing.T) {
	var err error

	opts := &pebble.Options{}
	s := newStorage("/tmp/data", opts.EnsureDefaults())
	assert.NotNil(t, s)

	err = s.set([]byte("ok"), []byte("okay"), true)
	assert.NoError(t, err)
	err = s.set([]byte("time"), []byte("ti"), true)
	assert.NoError(t, err)
	err = s.set([]byte("foo"), []byte("bar"), true)
	assert.NoError(t, err)
	err = s.set([]byte("raft"), []byte("paxos"), true)
	assert.NoError(t, err)
	err = s.set([]byte("bye"), []byte("bye"), true)
	assert.NoError(t, err)

	var value []byte

	value, err = s.get([]byte("time"))
	assert.NoError(t, err)
	assert.Equal(t, "ti", string(value))
	value, err = s.get([]byte("raft"))
	assert.NoError(t, err)
	assert.Equal(t, "paxos", string(value))
	value, err = s.get([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, "bar", string(value))

	err = s.delete([]byte("foo"), true)
	assert.NoError(t, err)
	err = s.delete([]byte("raft"), true)
	assert.NoError(t, err)
	value, err = s.get([]byte("raft"))
	assert.NoError(t, err)
	assert.Nil(t, value)
	value, err = s.get([]byte("foo"))
	assert.NoError(t, err)
	assert.Nil(t, value)
}