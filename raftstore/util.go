package raftstore

func byteEqual(a, b []byte) bool {
	return string(a) == string(b)
}