package storage

import (
	"encoding/binary"
)

type Pair struct {
	Key []byte
	Val []byte
}

func Encode(pairs []Pair) []byte {
	buf := make([]byte, 0)
	for _, p := range pairs {
		buf = append(buf, encode(p)...)
	}
	return buf
}

func Decode(encs []byte) []Pair {
	pairs := make([]Pair, 0)
	offset := 0
	for offset < len(encs) {
		i := binary.LittleEndian.Uint32(encs[offset:])
		p := decode(append([]byte{}, encs[offset:offset+int(i)]...))
		pairs = append(pairs, p)
		offset += int(i)
	}
	return pairs
}

func encode(p Pair) []byte {
	ksize, vsize := len(p.Key), len(p.Val)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf, uint32(ksize+vsize+8))
	binary.LittleEndian.PutUint32(buf[4:], uint32(ksize))
	buf = append(buf, p.Key...)
	buf = append(buf, p.Val...)
	return buf
}

func decode(enc []byte) Pair {
	p := Pair{}
	ksize := binary.LittleEndian.Uint32(enc[4:])
	p.Key = append([]byte{}, enc[8:8+ksize]...)
	p.Val = append([]byte{}, enc[8+ksize:]...)
	return p
}
