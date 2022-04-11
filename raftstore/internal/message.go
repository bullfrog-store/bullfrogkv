package internal

import "bullfrogkv/raftstore/raftstorepb"

func NewGetCmdResponse(value []byte) *raftstorepb.RaftCmdResponse {
	header := &raftstorepb.RaftResponseHeader{}
	resp := &raftstorepb.Response{
		Get: &raftstorepb.GetResponse{Value: value},
	}
	return &raftstorepb.RaftCmdResponse{
		Header:   header,
		Response: resp,
	}
}

func NewPutCmdRequest(key, value []byte) *raftstorepb.RaftCmdRequest {
	header := &raftstorepb.RaftRequestHeader{}
	req := &raftstorepb.Request{
		CmdType: raftstorepb.CmdType_Put,
		Put: &raftstorepb.PutRequest{
			Key:   key,
			Value: value,
		},
	}
	return &raftstorepb.RaftCmdRequest{
		Header:  header,
		Request: req,
	}
}

func NewDeleteCmdRequest(key []byte) *raftstorepb.RaftCmdRequest {
	header := &raftstorepb.RaftRequestHeader{}
	req := &raftstorepb.Request{
		CmdType: raftstorepb.CmdType_Delete,
		Delete: &raftstorepb.DeleteRequest{
			Key: key,
		},
	}
	return &raftstorepb.RaftCmdRequest{
		Header:  header,
		Request: req,
	}
}

func NewCompactCmdRequest(index, term uint64) *raftstorepb.RaftCmdRequest {
	header := &raftstorepb.RaftRequestHeader{}
	request := &raftstorepb.AdminRequest{
		CmdType: raftstorepb.AdminCmdType_CompactLog,
		CompactLog: &raftstorepb.CompactLogRequest{
			CompactIndex: index,
			CompactTerm:  term,
		},
	}
	return &raftstorepb.RaftCmdRequest{
		Header:       header,
		AdminRequest: request,
	}
}
