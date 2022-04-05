package storage

import (
	"encoding/binary"
)

type Pair struct {
	Key []byte
	Val []byte
}

func Encode(p Pair) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(p.Key)))
	buf = append(buf, p.Key...)
	buf = append(buf, p.Val...)
	return buf
}

func Decode(enc []byte) Pair {
	p := Pair{}
	size := binary.LittleEndian.Uint32(enc)
	p.Key = append([]byte{}, enc[4:4+size]...)
	p.Val = append([]byte{}, enc[4+size:]...)
	return p
}
