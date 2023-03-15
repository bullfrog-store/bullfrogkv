package storage

import "github.com/golang/protobuf/proto"

type Modify struct {
	Data interface{}
}

type Put struct {
	Key   []byte
	Value []byte
	Sync  bool
}

type Delete struct {
	Key  []byte
	Sync bool
}

func PutMeta(key []byte, msg proto.Message, sync bool) Modify {
	value, err := proto.Marshal(msg)
	if err != nil {
		// FIXME: handle the error more elegantly.
		// However, it's an unexpected error.
		panic(err)
	}
	return Modify{
		Data: Put{key, value, sync},
	}
}

func DeleteMeta(key []byte, sync bool) Modify {
	return Modify{
		Data: Delete{key, sync},
	}
}

func PutData(key, value []byte, sync bool) Modify {
	return Modify{
		Data: Put{key, value, sync},
	}
}

func DeleteData(key []byte, sync bool) Modify {
	return Modify{
		Data: Delete{key, sync},
	}
}

func (m Modify) Key() []byte {
	switch m.Data.(type) {
	case Put:
		return m.Data.(Put).Key
	case Delete:
		return m.Data.(Delete).Key
	}
	return nil
}

func (m Modify) Value() []byte {
	if putData, ok := m.Data.(Put); ok {
		return putData.Value
	}
	return nil
}

func (m Modify) Sync() bool {
	switch m.Data.(type) {
	case Put:
		return m.Data.(Put).Sync
	case Delete:
		return m.Data.(Delete).Sync
	}
	return false
}
