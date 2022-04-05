package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {
	p := Pair{
		Key: []byte("foo"),
		Val: []byte("bar"),
	}
	enc := Encode(p)
	dec := Decode(enc)
	assert.Equal(t, p.Key, dec.Key)
	assert.Equal(t, p.Val, dec.Val)
}
