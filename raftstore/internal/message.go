package internal

import "bullfrogkv/raftstore/raftstorepb"

func NewRaftCmdRequest(header *raftstorepb.RaftRequestHeader, request *raftstorepb.Request) *raftstorepb.RaftCmdRequest {
	return &raftstorepb.RaftCmdRequest{
		Header:  header,
		Request: request,
	}
}

func NewRaftAdminCmdRequest(header *raftstorepb.RaftRequestHeader, request *raftstorepb.AdminRequest) *raftstorepb.RaftCmdRequest {
	return &raftstorepb.RaftCmdRequest{
		Header:       header,
		AdminRequest: request,
	}
}

func NewRaftCmdResponse(response *raftstorepb.Response) *raftstorepb.RaftCmdResponse {
	return &raftstorepb.RaftCmdResponse{
		Response: response,
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
	return NewRaftAdminCmdRequest(header, request)
}
