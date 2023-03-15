package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPapi(t *testing.T) {
	p, err := newPapi(testDataPath, nil)
	defer func() {
		assert.NoError(t, p.close())
	}()
	assert.NoError(t, err)
	assert.NotNil(t, p)

	assert.NoError(t, removeAll(testDataPath))
}

func TestPapiOps(t *testing.T) {
	p, err := newPapi(testDataPath, nil)
	defer func() {
		assert.NoError(t, p.close())
	}()
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// set some data
	err = p.set([]byte("ok"), []byte("okay"), true)
	assert.NoError(t, err)
	err = p.set([]byte("time"), []byte("ti"), true)
	assert.NoError(t, err)
	err = p.set([]byte("foo"), []byte("bar"), true)
	assert.NoError(t, err)
	err = p.set([]byte("raft"), []byte("paxos"), true)
	assert.NoError(t, err)
	err = p.set([]byte("bye"), []byte("bye"), true)
	assert.NoError(t, err)

	// get some data
	var value []byte
	value, err = p.get([]byte("time"))
	assert.NoError(t, err)
	assert.Equal(t, "ti", string(value))
	value, err = p.get([]byte("raft"))
	assert.NoError(t, err)
	assert.Equal(t, "paxos", string(value))
	value, err = p.get([]byte("foo"))
	assert.NoError(t, err)
	assert.Equal(t, "bar", string(value))

	// delete some data
	err = p.delete([]byte("foo"), true)
	assert.NoError(t, err)
	err = p.delete([]byte("raft"), true)
	assert.NoError(t, err)

	// verification
	value, err = p.get([]byte("raft"))
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, value)
	value, err = p.get([]byte("foo"))
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, value)

	assert.NoError(t, removeAll(testDataPath))
}

func TestPapiSnapshot(t *testing.T) {
	p, err := newPapi(testDataPath, nil)
	defer func() {
		assert.NoError(t, p.close())
	}()
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// set some data
	err = p.set([]byte("ok"), []byte("okay"), true)
	assert.NoError(t, err)
	err = p.set([]byte("time"), []byte("ti"), true)
	assert.NoError(t, err)
	err = p.set([]byte("foo"), []byte("bar"), true)
	assert.NoError(t, err)
	err = p.set([]byte("raft"), []byte("paxos"), true)
	assert.NoError(t, err)
	err = p.set([]byte("bye"), []byte("bye"), true)
	assert.NoError(t, err)

	// delete some data
	err = p.delete([]byte("foo"), true)
	assert.NoError(t, err)
	err = p.delete([]byte("ok"), true)
	assert.NoError(t, err)

	var buf []byte
	var entry Entry
	var mp = make(map[string]string)

	buf, err = p.snapshot()
	assert.NoError(t, err)
	for len(buf) > 0 {
		entry, buf, err = Deserialize(buf)
		assert.NoError(t, err)
		mp[string(entry.Key)] = string(entry.Value)
	}
	assert.Equal(t, 0, len(buf))
	assert.Equal(t, 3, len(mp))

	// verification
	assert.Equal(t, "ti", mp["time"])
	assert.Equal(t, "paxos", mp["raft"])
	assert.Equal(t, "bye", mp["bye"])

	assert.NoError(t, removeAll(testDataPath))
}
