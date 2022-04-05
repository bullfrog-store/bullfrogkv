package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {
	pairs := []Pair{
		{
			Key: []byte("foo"),
			Val: []byte("bar"),
		},
		{
			Key: []byte("bull"),
			Val: []byte("frog"),
		},
		{
			Key: []byte("why"),
			Val: []byte("what"),
		},
		{
			Key: []byte("www"),
			Val: []byte("com"),
		},
		{
			Key: []byte("blue"),
			Val: []byte("white"),
		},
	}
	enc := Encode(pairs)
	dec := Decode(enc)
	for i := range dec {
		assert.Equal(t, pairs[i].Key, dec[i].Key)
		assert.Equal(t, pairs[i].Val, dec[i].Val)
	}
}
