package raftstore

import (
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/storage"
	"github.com/golang/protobuf/proto"
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
	pr.router = newRouter(peerMap[id], pr.raftMsgReceiver)

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
	go pr.handleRaftMsgs()
	return pr
}

func (pr *peer) propose(cmd *raftstorepb.RaftCmdRequest) error {
	data, err := proto.Marshal(cmd)
	if err != nil {
		return err
	}
	return pr.raftGroup.Propose(nil, data)
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

func (pr *peer) handleRaftMsgs() {
	for {
		msgs := make([]raftpb.Message, 0)
		select {
		case msg := <-pr.raftMsgReceiver:
			msgs = append(msgs, msg)
		}
		pending := len(pr.raftMsgReceiver)
		for i := 0; i < pending; i++ {
			msgs = append(msgs, <-pr.raftMsgReceiver)
		}
		for _, msg := range msgs {
			pr.raftGroup.Step(nil, msg)
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
		// TODO(qyl): try to compact log
		// propose admin request
	}
}

func (pr *peer) handleReady(rd raft.Ready) {
	pr.ps.saveReadyState(rd)
	pr.router.sendRaftMessage(rd.Messages)
	for _, ent := range rd.CommittedEntries {
		pr.process(ent)
	}
	pr.raftGroup.Advance()
}

func (pr *peer) process(ent raftpb.Entry) {
	cmd := &raftstorepb.RaftCmdRequest{}
	if err := proto.Unmarshal(ent.Data, cmd); err != nil {
		panic(err)
	}
	if cmd.Request != nil {
		// process common request
		pr.processRequest(cmd.Request)
	} else if cmd.AdminRequest != nil {
		// TODO(qyl): process admin request
	}
}

func (pr *peer) processRequest(cmd *raftstorepb.Request) {
	switch cmd.CmdType {
	case raftstorepb.CmdType_Put:
		modify := storage.PutData(cmd.Put.Key, cmd.Put.Value, true)
		pr.ps.engine.WriteKV(modify)
	case raftstorepb.CmdType_Delete:
		modify := storage.DeleteData(cmd.Delete.Key, true)
		pr.ps.engine.WriteKV(modify)
	case raftstorepb.CmdType_Get:
		// TODO(qyl): use read index
	}
}

func (pr *peer) term() uint64 {
	return pr.raftGroup.Status().Term
}

func (pr *peer) isLeader() bool {
	return pr.raftGroup.Status().Lead == pr.id
}
