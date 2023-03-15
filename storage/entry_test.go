package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeAndDeserialize(t *testing.T) {
	entries := []Entry{
		{
			Key:   []byte("foo"),
			Value: []byte("bar"),
		},
		{
			Key:   []byte("a"),
			Value: []byte("b"),
		},
		{
			Key:   []byte("why"),
			Value: []byte("what"),
		},
		{
			Key:   []byte("sun"),
			Value: []byte("moon"),
		},
		{
			Key:   []byte("123"),
			Value: []byte("45"),
		},
	}

	buf := SerializeMulti(entries)
	es, err := DeserializeMulti(buf)
	assert.NoError(t, err)
	for i := 0; i < 5; i++ {
		assert.Equal(t, entries[i].Key, es[i].Key)
		assert.Equal(t, entries[i].Value, es[i].Value)
	}
}
