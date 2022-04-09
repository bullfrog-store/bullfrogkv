package raftstore

import (
	"bullfrogkv/config"
	"bullfrogkv/logger"
	"bullfrogkv/raftstore/internal"
	"bullfrogkv/raftstore/raftstorepb"
	"bullfrogkv/storage"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/quorum"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"math"
	"net"
	"time"
)

var peerMap = map[uint64]string{}

func loadPeerMap(addrs []string) {
	for peer, addr := range addrs {
		peerMap[uint64(peer+1)] = addr
	}
}

func raftPeers() []raft.Peer {
	rpeers := make([]raft.Peer, len(peerMap))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	return rpeers
}

type reader struct {
	ctx []byte
	cb  *internal.Callback
}

func (r *reader) timestamp() int64 {
	return int64(binary.LittleEndian.Uint64(r.ctx))
}

func (r *reader) key() []byte {
	return r.ctx[8:]
}

type peer struct {
	id        uint64
	raftGroup raft.Node
	ps        *peerStorage
	router    *router

	waitReadChannel chan *reader
	readStateTable  map[string]raft.ReadState
	readStateComing chan struct{}

	raftMsgReceiver chan raftpb.Message

	compactionElapse  int
	compactionTimeout int

	lastCompactedIdx uint64
}

func newPeer() *peer {
	cfg := config.GlobalConfig
	loadPeerMap(cfg.RouteConfig.GrpcAddrs)

	bufferSize := cfg.RaftConfig.MaxSizePerMsg * 4
	pr := &peer{
		id:                cfg.StoreConfig.StoreId,
		ps:                newPeerStorage(cfg.StoreConfig.DataPath),
		waitReadChannel:   make(chan *reader, bufferSize),
		readStateTable:    make(map[string]raft.ReadState),
		readStateComing:   make(chan struct{}, 1),
		raftMsgReceiver:   make(chan raftpb.Message, bufferSize),
		compactionTimeout: cfg.RaftConfig.CompactCheckPeriod,
	}
	pr.router = newRouter(peerMap[pr.id], pr.raftMsgReceiver)
	pr.lastCompactedIdx = pr.ps.truncateIndex()

	c := &raft.Config{
		ID:              pr.id,
		ElectionTick:    cfg.RaftConfig.ElectionTick,
		HeartbeatTick:   cfg.RaftConfig.HeartbeatTick,
		Storage:         pr.ps,
		Applied:         pr.ps.AppliedIndex(),
		MaxSizePerMsg:   cfg.RaftConfig.MaxSizePerMsg,
		MaxInflightMsgs: cfg.RaftConfig.MaxInflightMsgs,
		PreVote:         true,
	}
	if pr.isInitial() {
		pr.raftGroup = raft.StartNode(c, raftPeers())
		pr.ps.confState = pr.confState()
		pr.ps.raftConfStateWriteToDB(pr.ps.confState)
	} else {
		pr.raftGroup = raft.RestartNode(c)
	}
	logger.Infof("etcd raft is started, node: %d", pr.id)

	pr.run()
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
	go pr.serveGrpc(pr.id)
	go pr.onTick()
	go pr.handleRaftMsgs()
	go pr.handleReadState()
}

func (pr *peer) serveGrpc(id uint64) {
	lis, err := net.Listen("tcp", peerMap[id])
	if err != nil {
		panic(err)
	}
	logger.Infof("%d listen %v success\n", id, peerMap[id])
	g := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 15 * time.Second,
			Time:              5 * time.Second,
			Timeout:           1 * time.Second,
		}),
	)
	raftstorepb.RegisterMessageServer(g, pr.router.raftServer)
	g.Serve(lis)
}

func (pr *peer) onTick() {
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
			if msg.To == pr.id {
				pr.raftGroup.Step(context.TODO(), msg)
			}
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
	if appliedIdx > firstIdx && appliedIdx-firstIdx >= config.GlobalConfig.RaftConfig.LogGCCountLimit {
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

	pr.propose(internal.NewCompactCmdRequest(compactIdx, term))
}

func (pr *peer) handleReady(rd raft.Ready) {
	pr.ps.saveReadyState(rd)
	for _, state := range rd.ReadStates {
		pr.readStateTable[string(state.RequestCtx)] = state
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
		logger.Debugf("apply CmdType_Put request: %+v", request.Put)
		modify := storage.PutData(request.Put.Key, request.Put.Value, true)
		pr.ps.engines.WriteKV(modify)
	case raftstorepb.CmdType_Delete:
		logger.Debugf("apply CmdType_Delete request: %+v", request.Delete)
		modify := storage.DeleteData(request.Delete.Key, true)
		pr.ps.engines.WriteKV(modify)
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
			go pr.gcRaftLog(pr.lastCompactedIdx+1, applySt.TruncatedState.Index+1)
			pr.lastCompactedIdx = applySt.TruncatedState.Index
		}
	}
}

func (pr *peer) gcRaftLog(start, end uint64) error {
	entries, err := pr.ps.Entries(start, end, math.MaxUint64)
	if err != nil {
		return err
	}
	return pr.ps.deleteRaftLogEntries(entries)
}

func (pr *peer) linearizableRead(key []byte) *internal.Callback {
	ctx := buildReadCtx(key)
	cb := internal.NewCallback()
	reader := &reader{
		ctx: ctx,
		cb:  cb,
	}
	logger.Debugf("receive a read request: %+v", reader)
	pr.waitReadChannel <- reader
	if err := pr.raftGroup.ReadIndex(context.TODO(), ctx); err != nil {
		panic(err)
	}
	return cb
}

func (pr *peer) handleReadState() {
	for {
		select {
		case <-pr.readStateComing:
			var r *reader
			size := len(pr.waitReadChannel)
			for size > 0 {
				r = <-pr.waitReadChannel
				if _, ok := pr.readStateTable[string(r.ctx)]; ok {
					break
				}
				if time.Duration(time.Now().UnixNano()-r.timestamp()) < 15*time.Second {
					pr.waitReadChannel <- r
				}
				size--
			}
			state := pr.readStateTable[string(r.ctx)]
			delete(pr.readStateTable, string(r.ctx))
			logger.Debugf("ReadState: %+v, ReadRequest: %+v", state, r)
			go pr.readApplied(state, r)
		}
	}
}

func (pr *peer) readApplied(state raft.ReadState, reader *reader) {
	pr.waitAppliedAdvance(state.Index)
	value, err := pr.ps.engines.ReadKV(reader.key())
	if err != nil {
		if err != storage.ErrNotFound {
			panic(err)
		}
	}
	resp := internal.NewGetCmdResponse(value)
	logger.Infof("%s get response successfully: %+v", string(reader.key()), resp.GetResponse())
	reader.cb.Done(resp)
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

func (pr *peer) isInitial() bool {
	return isEmptyConfState(*pr.ps.confState)
}

func buildReadCtx(key []byte) []byte {
	buf := make([]byte, 8)
	ts := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(buf, uint64(ts))
	return append(buf, key...)
}
