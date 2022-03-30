package internal

import (
	"bullfrogkv/raftstore/raftstorepb"
	"time"
)

type Callback struct {
	Resp *raftstorepb.RaftCmdResponse
	done chan struct{}
}

func (cb *Callback) Done(resp *raftstorepb.RaftCmdResponse) {
	if cb == nil {
		return
	}
	if resp != nil {
		cb.Resp = resp
	}
	cb.done <- struct{}{}
}

func (cb *Callback) WaitResp() *raftstorepb.RaftCmdResponse {
	select {
	case <-cb.done:
		return cb.Resp
	}
}

func (cb *Callback) WaitRespWithTimeout(timeout time.Duration) *raftstorepb.RaftCmdResponse {
	select {
	case <-cb.done:
		return cb.Resp
	case <-time.After(timeout):
		return cb.Resp
	}
}

func NewCallback() *Callback {
	done := make(chan struct{}, 1)
	cb := &Callback{done: done}
	return cb
}

