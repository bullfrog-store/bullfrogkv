package storage

import (
	"encoding/binary"
	"errors"
)

const (
	MagicNumber uint64 = 0xDEADBEEF
)

var (
	ErrDeserialize error = errors.New("error encountered while deserialization")
)

// Entry design as below:
// |-------------|---------|-----|-----------|-------|
// | MagicNumber | KeySize | Key | ValueSize | Value |
// |-------------|---------|-----|-----------|-------|
// - MagicNumber: 8byte
// - KeySize, ValueSize: 4byte
type Entry struct {
	Key   []byte
	Value []byte
}

func MakeEntry(key, value []byte) Entry {
	return Entry{key, value}
}

func SerializeMulti(es []Entry) []byte {
	buf := make([]byte, 0)
	for i := range es {
		buf = append(buf, Serialize(es[i])...)
	}
	return buf
}

func DeserializeMulti(buf []byte) ([]Entry, error) {
	var es []Entry
	var e Entry
	var err error

	for len(buf) > 0 {
		e, buf, err = Deserialize(buf)
		if err != nil {
			return nil, err
		}
		es = append(es, e)
	}
	return es, nil
}

func Serialize(e Entry) []byte {
	ksize, vsize := len(e.Key), len(e.Value)
	buf := make([]byte, 8+4+ksize+4+vsize)
	// magic number
	binary.LittleEndian.PutUint64(buf, MagicNumber)
	// key size and key
	binary.LittleEndian.PutUint32(buf[8:], uint32(ksize))
	copy(buf[8+4:], e.Key)
	// value size and value
	binary.LittleEndian.PutUint32(buf[8+4+ksize:], uint32(vsize))
	copy(buf[8+4+ksize+4:], e.Value)
	return buf
}

func Deserialize(buf []byte) (Entry, []byte, error) {
	e := Entry{}
	if binary.LittleEndian.Uint64(buf) != MagicNumber {
		return e, nil, ErrDeserialize
	}

	ksize := binary.LittleEndian.Uint32(buf[8:])
	vsize := binary.LittleEndian.Uint32(buf[8+4+ksize:])
	if int(8+4+ksize+4+vsize) > len(buf) {
		return e, nil, ErrDeserialize
	}

	e.Key = make([]byte, ksize)
	copy(e.Key, buf[8+4:8+4+ksize])

	e.Value = make([]byte, vsize)
	copy(e.Value, buf[8+4+ksize+4:8+4+ksize+4+vsize])

	buf = buf[8+4+ksize+4+vsize:]
	return e, buf, nil
}
