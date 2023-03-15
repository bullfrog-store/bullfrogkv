package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDataPath = "./testdata"
)

func TestBasic(t *testing.T) {
	e, err := New(NewDefaultConfig(testDataPath))
	assert.NotNil(t, e)
	assert.NoError(t, err)

	err = e.WriteData(PutData([]byte("hello"), []byte("world"), true))
	assert.NoError(t, err)

	var value []byte
	value, err = e.ReadData([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("world"), value)

	value, err = e.ReadData([]byte("world"))
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, value)

	assert.NoError(t, e.Destroy())
}
