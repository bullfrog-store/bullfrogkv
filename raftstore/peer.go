package raftstore

import (
	"bullfrogkv/logger"
	"bullfrogkv/raftstore/internal"
	"bullfrogkv/raftstore/meta"
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/storage"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/quorum"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"golang.org/x/net/context"
	"math"
	"time"
)

const (
	defaultLogGCCountLimit    = 10
	defaultCompactCheckPeriod = 100
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

	readRequestCh   chan *readRequest
	readStateMap    map[string]raft.ReadState
	readStateComing chan struct{}

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
		readStateMap:      make(map[string]raft.ReadState),
		readStateComing:   make(chan struct{}, 1),
		raftMsgReceiver:   make(chan raftpb.Message, 1024),
		compactionTimeout: defaultCompactCheckPeriod,
	}
	pr.router = newRouter(peerMap[id], pr.raftMsgReceiver)

	logger.Infof("new etcd raft, node: %d", id)
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
	if pr.isInitialBootstrap() {
		rpeers := make([]raft.Peer, len(peerMap))
		for i := range rpeers {
			rpeers[i] = raft.Peer{ID: uint64(i + 1)}
		}
		pr.raftGroup = raft.StartNode(c, rpeers)
		pr.ps.confState = pr.confState()
		pr.ps.raftConfStateWriteToDB(pr.ps.confState)
	} else {
		pr.raftGroup = raft.RestartNode(c)
	}
	logger.Infof("etcd raft is started")
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
	logger.Infof("peer drive run")
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
	logger.Infof("peer handle raft msgs")
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
	pr.tickLogGC()
}

func (pr *peer) tickLogGC() {
	pr.compactionElapse++
	if pr.compactionElapse >= pr.compactionTimeout {
		pr.compactionElapse = 0
		// try to compact log
		pr.onLogGCTask()
	}
}

func (pr *peer) onLogGCTask() {
	if !pr.isLeader() {
		return
	}
	appliedIdx := pr.ps.AppliedIndex()
	firstIdx, _ := pr.ps.FirstIndex()
	var compactIdx uint64
	if appliedIdx > firstIdx && appliedIdx-firstIdx >= defaultLogGCCountLimit {
		compactIdx = appliedIdx
	} else {
		return
	}

	// improve the success rate of log compaction
	compactIdx--
	term, err := pr.ps.Term(compactIdx)
	if err != nil {
		logger.Fatalf("appliedIdx: %d, firstIdx: %d, compactIdx: %d", appliedIdx, firstIdx, compactIdx)
		panic(err)
	}

	header := &raftstorepb.RaftRequestHeader{Term: term}
	request := &raftstorepb.AdminRequest{
		CmdType: raftstorepb.AdminCmdType_CompactLog,
		CompactLog: &raftstorepb.CompactLogRequest{
			CompactIndex: compactIdx,
			CompactTerm:  term,
		},
	}
	pr.propose(internal.NewRaftAdminCmdRequest(header, request))
}

func (pr *peer) handleReady(rd raft.Ready) {
	logger.Infof("peer handle ready: %+v", rd)
	pr.ps.saveReadyState(rd)
	for _, state := range rd.ReadStates {
		pr.readStateMap[string(state.RequestCtx)] = state
	}
	if len(rd.ReadStates) > 0 {
		pr.readStateComing <- struct{}{}
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
		// process admin request
		pr.processAdminRequest(cmd.AdminRequest)
	}
}

func (pr *peer) processRequest(request *raftstorepb.Request) {
	switch request.CmdType {
	case raftstorepb.CmdType_Put:
		logger.Infof("apply CmdType_Put request: %+v", request.Put)
		modify := storage.PutData(request.Put.Key, request.Put.Value, true)
		pr.ps.engine.WriteKV(modify)
	case raftstorepb.CmdType_Delete:
		logger.Infof("apply CmdType_Delete request: %+v", request.Delete)
		modify := storage.DeleteData(request.Delete.Key, true)
		pr.ps.engine.WriteKV(modify)
	}
}

func (pr *peer) processAdminRequest(request *raftstorepb.AdminRequest) {
	switch request.CmdType {
	case raftstorepb.AdminCmdType_CompactLog:
		compactLog := request.GetCompactLog()
		applySt := pr.ps.applyState
		if compactLog.CompactIndex >= applySt.TruncatedState.Index {
			applySt.TruncatedState.Index = compactLog.CompactIndex
			applySt.TruncatedState.Term = compactLog.CompactTerm
			pr.ps.raftApplyStateWriteToDB(applySt)
			go pr.gcRaftLog(pr.lastCompactedIdx, applySt.TruncatedState.Index+1)
			pr.lastCompactedIdx = applySt.TruncatedState.Index
		}
	}
}

func (pr *peer) gcRaftLog(start, end uint64) error {
	entries, err := pr.ps.Entries(start, end, math.MaxUint64)
	if err != nil {
		return err
	}
	return pr.ps.raftLogEntriesDeleteDB(entries)
}

func (pr *peer) linearizableRead(key []byte) *internal.Callback {
	readCtx := pr.buildReadCtx(key)
	cb := internal.NewCallback()
	rr := &readRequest{
		readCtx:  readCtx,
		callback: cb,
	}
	logger.Infof("receive a read request: %+v", rr)
	pr.readRequestCh <- rr
	if err := pr.raftGroup.ReadIndex(context.TODO(), readCtx); err != nil {
		panic(err)
	}
	return cb
}

func (pr *peer) buildReadCtx(key []byte) []byte {
	buf := make([]byte, 8)
	ts := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(buf, uint64(ts))
	return append(buf, key...)
}

func (pr *peer) handleReadState() {
	for {
		select {
		case <-pr.readStateComing:
			var rr *readRequest
			for len(pr.readRequestCh) > 0 {
				rr = <-pr.readRequestCh
				if _, ok := pr.readStateMap[string(rr.readCtx)]; ok {
					break
				}
			}
			state := pr.readStateMap[string(rr.readCtx)]
			delete(pr.readStateMap, string(rr.readCtx))
			logger.Infof("ReadState: %+v, ReadRequest: %+v", state, rr)
			go pr.readApplied(state, rr)
		}
	}
}

func (pr *peer) readApplied(state raft.ReadState, rr *readRequest) {
	logger.Infof("wait for applied index >= state.Index")
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
	logger.Infof("%s get response successfully: %+v", string(rr.key()), resp.Get)
	rr.callback.Done(internal.NewRaftCmdResponse(resp))
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
	// wait for applied index >= state.Index
	<-doneCh
	close(doneCh)
}

func (pr *peer) confState() *raftpb.ConfState {
	c := pr.raftGroup.Status().Config
	return &raftpb.ConfState{
		Voters:         c.Voters[0].Slice(),
		VotersOutgoing: c.Voters[1].Slice(),
		Learners:       quorum.MajorityConfig(c.Learners).Slice(),
		LearnersNext:   quorum.MajorityConfig(c.LearnersNext).Slice(),
		AutoLeave:      c.AutoLeave,
	}
}

func (pr *peer) term() uint64 {
	return pr.raftGroup.Status().Term
}

func (pr *peer) isLeader() bool {
	return pr.raftGroup.Status().Lead == pr.id
}

func (pr *peer) isInitialBootstrap() bool {
	cs, err := meta.GetRaftConfState(pr.ps.engine)
	if err != nil {
		panic(err)
	}
	return isEmptyConfState(*cs)
}
