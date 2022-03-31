package meta

import "encoding/binary"

const (
	RaftLocalStatePrefix = 0x01
	RaftLogEntryPrefix   = 0x02
)

func buildRaftLogEntryKey(prefix byte, index uint64) []byte {
	key := make([]byte, 9)
	key[0] = prefix
	binary.BigEndian.PutUint64(key[1:], index)
	return key
}

func buildRaftLocalStateKey(prefix byte) []byte {
	key := make([]byte, 9)
	key[0] = prefix
	// To ensure that the key length is consistent, we set aside 8 bytes here.
	return key
}

func RaftLogEntryKey(index uint64) []byte {
	return buildRaftLogEntryKey(RaftLogEntryPrefix, index)
}

func RaftLocalStateKey() []byte {
	return buildRaftLocalStateKey(RaftLocalStatePrefix)
}
