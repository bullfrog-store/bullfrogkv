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
