package raftstore

import (
	"bullfrogkv/raftstore/internal"
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/storage"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"golang.org/x/net/context"
	"time"
)

type readRequest struct {
	readCtx  []byte
	callback *internal.Callback
}

func (rr *readRequest) key() []byte {
	return rr.readCtx[8:]
}

type peer struct {
	id        uint64
	raftGroup raft.Node
	ps        *peerStorage
	router    *router

	readRequestCh chan *readRequest
	readStateCh   chan raft.ReadState

	raftMsgReceiver chan raftpb.Message

	compactionElapse  int
	compactionTimeout int

	lastCompactedIdx uint64
}

func newPeer(id uint64, path string) *peer {
	pr := &peer{
		id:                id,
		ps:                newPeerStorage(path),
		readRequestCh:     make(chan *readRequest, 1024),
		readStateCh:       make(chan raft.ReadState, 1024),
		raftMsgReceiver:   make(chan raftpb.Message, 1024),
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
		PreVote:                   true,
	}
	rpeers := make([]raft.Peer, len(peerMap))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	pr.raftGroup = raft.StartNode(c, rpeers)
	go pr.run()
	go pr.handleRaftMsgs()
	go pr.handleReadState()
	return pr
}

func (pr *peer) propose(cmd *raftstorepb.RaftCmdRequest) error {
	data, err := proto.Marshal(cmd)
	if err != nil {
		return err
	}
	return pr.raftGroup.Propose(context.TODO(), data)
}

func (pr *peer) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			pr.tick()
		case rd := <-pr.raftGroup.Ready():
			//fmt.Println("msg:", rd.Messages, "entries :", rd.Entries)
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
			pr.raftGroup.Step(context.TODO(), msg)
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
	for _, state := range rd.ReadStates {
		pr.readStateCh <- state
	}
	pr.router.sendRaftMessage(rd.Messages)
	for _, ent := range rd.CommittedEntries {
		pr.process(ent)
	}
	pr.raftGroup.Advance()
}

func (pr *peer) process(ent raftpb.Entry) {
	pr.ps.applyState.ApplyIndex = ent.Index
	pr.ps.raftApplyStateWriteToDB(pr.ps.applyState)
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
	}
}

func (pr *peer) linearizableRead(key []byte) *internal.Callback {
	ts := time.Now().UnixNano()
	readCtx := pr.buildReadCtx(ts, key)
	cb := internal.NewCallback()
	rr := &readRequest{
		readCtx:  readCtx,
		callback: cb,
	}
	pr.readRequestCh <- rr
	if err := pr.raftGroup.ReadIndex(context.TODO(), readCtx); err != nil {
		panic(err)
	}
	return cb
}

func (pr *peer) buildReadCtx(ts int64, key []byte) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(ts))
	return append(buf, key...)
}

func (pr *peer) handleReadState() {
	for {
		select {
		case state := <-pr.readStateCh:
			rr := <-pr.readRequestCh
			if byteEqual(rr.readCtx, state.RequestCtx) {
				// wait for applied index >= state.Index
				pr.waitAppliedAdvance(state.Index)
				value, err := pr.ps.engine.ReadKV(rr.key())
				if err != nil {
					if err != storage.ErrNotFound {
						panic(err)
					}
				}
				resp := &raftstorepb.Response{
					Get: &raftstorepb.GetResponse{Value: value},
				}
				rr.callback.Done(internal.NewRaftCmdResponse(resp))
			}
		}
	}
}

func (pr *peer) waitAppliedAdvance(index uint64) {
	applied := pr.ps.AppliedIndex()
	if applied >= index {
		return
	}
	doneCh := make(chan struct{})
	go func() {
		for applied < index {
			time.Sleep(time.Millisecond)
			applied = pr.ps.AppliedIndex()
		}
		doneCh <- struct{}{}
	}()
	<-doneCh
	close(doneCh)
}

func (pr *peer) term() uint64 {
	return pr.raftGroup.Status().Term
}

func (pr *peer) isLeader() bool {
	return pr.raftGroup.Status().Lead == pr.id
}
