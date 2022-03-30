package raftstore

import (
	"bullfrogkv/raftstore/raftstorepb"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"time"
)

type peer struct {
	id        uint64
	raftGroup raft.Node
	ps        *peerStorage
	router    *router

	raftMsgReceiver chan raftpb.Message

	compactionElapse  int
	compactionTimeout int

	lastCompactedIdx uint64
}

func newPeer(id uint64, path string) *peer {
	// TODO: router init
	pr := &peer{
		id:                id,
		ps:                newPeerStorage(path),
		raftMsgReceiver:   make(chan raftpb.Message, 256),
		compactionTimeout: 100,
	}
	pr.router = newRouter(pr.raftMsgReceiver)

	c := &raft.Config{
		ID:                        id,
		ElectionTick:              10,
		HeartbeatTick:             1,
		Storage:                   pr.ps,
		Applied:                   pr.ps.AppliedIndex(),
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

func (pr *peer) propose(cmd *raftstorepb.RaftCmdRequest) error {
	// TODO: handle propose
	return pr.raftGroup.Propose(nil, nil)
}

func (pr *peer) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			pr.tick()
		case msg := <-pr.raftMsgReceiver:
			pr.raftGroup.Step(nil, msg)
		case rd := <-pr.raftGroup.Ready():
			pr.handleReady(rd)
		}
	}
}

func (pr *peer) tick() {
	pr.raftGroup.Tick()
	pr.tickCompact()
}

func (pr *peer) tickCompact() {
	pr.compactionElapse++
	if pr.compactionElapse >= pr.compactionTimeout {
		pr.compactionElapse = 0
		// TODO: try to compact log
		// propose admin request
	}
}

func (pr *peer) handleReady(rd raft.Ready) {
	// TODO: handle ready
	pr.raftGroup.Advance()
}

func (pr *peer) term() uint64 {
	return pr.raftGroup.Status().Term
}

func (pr *peer) isLeader() bool {
	return pr.raftGroup.Status().Lead == pr.id
}
