package raftstore

import (
	"bullfrogkv/raftstore/internal"
	"go.etcd.io/etcd/raft/v3"
	"time"
)

const (
	defaultCompactionTimeout = 200
)

var (
	// peerMap is temporary peer id to address map.
	peerMap = map[uint64]string{
		1: "127.0.0.1:8080",
		2: "127.0.0.1:8081",
		3: "127.0.0.1:8082",
	}
)

type proposal struct {
	index uint64
	term  uint64
	cb    *internal.Callback
}

type peer struct {
	id        uint64
	raftGroup raft.Node
	ps        *peerStorage
	router    *router

	proposals []*proposal

	compactionElapse  int
	compactionTimeout int
}

func newPeer(id uint64) *peer {
	pr := &peer{
		id:                id,
		ps:                newPeerStorage(),
		compactionTimeout: defaultCompactionTimeout,
	}
	// TODO: start raft
	c := &raft.Config{
		ID:                        id, // TODO: peer id
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   pr.ps,
		Applied:                   0, // TODO: in kv state
		MaxSizePerMsg:             1024 * 1024,
		MaxUncommittedEntriesSize: 1 << 30,
		MaxInflightMsgs:           256,
	}
	rpeers := make([]raft.Peer, len(peerMap))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	pr.raftGroup = raft.StartNode(c, rpeers)
	go pr.run()
	return pr
}

func (pr *peer) propose(cmd *internal.MsgRaftCmd) error {
	// TODO: handle propose
	return pr.raftGroup.Propose(nil, nil)
}

func (pr *peer) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			pr.tick()
		case rd := <-pr.raftGroup.Ready():
			pr.handleReady(rd)
		}
	}
}

func (pr *peer) tick() {
	pr.raftGroup.Tick()
	pr.compactionElapse++
	if pr.compactionElapse >= pr.compactionTimeout {
		// TODO: try to compact log
	}
}

func (pr *peer) handleReady(rd raft.Ready) {
	// TODO: handle ready
	pr.raftGroup.Advance()
}
