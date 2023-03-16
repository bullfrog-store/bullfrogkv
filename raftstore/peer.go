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
	for pr, addr := range addrs {
		peerMap[uint64(pr+1)] = addr
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

func newPeer() (*peer, error) {
	var err error
	cfg := config.GlobalConfig
	loadPeerMap(cfg.RouteConfig.GrpcAddrs)

	bufferSize := cfg.RaftConfig.MaxSizePerMsg * 4
	pr := &peer{
		id:                cfg.StoreConfig.StoreId,
		waitReadChannel:   make(chan *reader, bufferSize),
		readStateTable:    make(map[string]raft.ReadState),
		readStateComing:   make(chan struct{}, 1),
		raftMsgReceiver:   make(chan raftpb.Message, bufferSize),
		compactionTimeout: cfg.RaftConfig.CompactCheckPeriod,
	}
	if pr.ps, err = newPeerStorage(cfg.StoreConfig.DataPath); err != nil {
		return nil, err
	}
	pr.router = newRouter(peerMap[pr.id], pr.raftMsgReceiver)
	pr.lastCompactedIdx = pr.ps.truncateIndex()

	c := &raft.Config{
		ID:              pr.id,
		ElectionTick:    cfg.RaftConfig.ElectionTick,
		HeartbeatTick:   cfg.RaftConfig.HeartbeatTick,
		Storage:         pr.ps,
		Applied:         pr.ps.appliedIndex(),
		MaxSizePerMsg:   cfg.RaftConfig.MaxSizePerMsg,
		MaxInflightMsgs: cfg.RaftConfig.MaxInflightMsgs,
		PreVote:         true,
		Logger:          logger.GlobalLogger(),
	}
	if pr.isInitial() {
		pr.raftGroup = raft.StartNode(c, raftPeers())
		pr.ps.confState = pr.confState()
		if err = pr.ps.writeRaftConfState(pr.ps.confState); err != nil {
			return nil, err
		}
	} else {
		pr.raftGroup = raft.RestartNode(c)
	}
	logger.Infof("etcd raft is running, node: %d", pr.id)

	pr.run()
	return pr, nil
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
		logger.Fatalf("grpc connect failed, error: %s", err)
	}

	logger.Infof("peer-%d listen at %v successfully", id, peerMap[id])
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
	if err = g.Serve(lis); err != nil {
		logger.Fatalf("grpc serve failed, error: %s", err)
	}
}

func (pr *peer) onTick() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		var err error
		select {
		case <-ticker.C:
			err = pr.tick()
		case rd := <-pr.raftGroup.Ready():
			err = pr.handleReady(rd)
		}
		if err != nil {
			logger.Errorf("tick error: %s", err.Error())
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
				if err := pr.raftGroup.Step(context.TODO(), msg); err != nil {
					logger.Errorf("get error when RaftGroup step message, error: %s", err.Error())
				}
			}
		}
	}
}

func (pr *peer) tick() error {
	pr.raftGroup.Tick()
	return pr.tickLogGC()
}

func (pr *peer) tickLogGC() error {
	var err error
	pr.compactionElapse++
	if pr.compactionElapse >= pr.compactionTimeout {
		pr.compactionElapse = 0
		// Try to compact log.
		err = pr.onLogGCTask()
	}
	return err
}

func (pr *peer) onLogGCTask() error {
	if !pr.isLeader() {
		return nil
	}
	appliedIdx := pr.ps.appliedIndex()
	firstIdx, _ := pr.ps.FirstIndex()
	var compactIdx uint64
	if appliedIdx > firstIdx && appliedIdx-firstIdx >= config.GlobalConfig.RaftConfig.LogGCCountLimit {
		compactIdx = appliedIdx
	} else {
		return nil
	}

	// Improve the success rate of log compaction.
	compactIdx--
	term, err := pr.ps.Term(compactIdx)
	if err != nil {
		logger.Errorf("compaction error: appliedIdx: %d, firstIdx: %d, compactIdx: %d", appliedIdx, firstIdx, compactIdx)
		return err
	}

	return pr.propose(internal.NewCompactCmdRequest(compactIdx, term))
}

func (pr *peer) handleReady(rd raft.Ready) error {
	var err error
	if err = pr.ps.saveReadyState(rd); err != nil {
		return err
	}
	for _, state := range rd.ReadStates {
		pr.readStateTable[string(state.RequestCtx)] = state
	}
	if len(rd.ReadStates) > 0 {
		pr.readStateComing <- struct{}{}
	}
	pr.router.sendRaftMessage(rd.Messages)
	for _, ent := range rd.CommittedEntries {
		if err = pr.process(ent); err != nil {
			return err
		}
	}
	pr.raftGroup.Advance()

	return nil
}

func (pr *peer) process(ent raftpb.Entry) error {
	var err error
	pr.ps.applyState.ApplyIndex = ent.Index
	if err = pr.ps.writeRaftApplyState(pr.ps.applyState); err != nil {
		return err
	}
	cmd := &raftstorepb.RaftCmdRequest{}
	if err = proto.Unmarshal(ent.Data, cmd); err != nil {
		return err
	}
	if cmd.Request != nil {
		// Process common request.
		err = pr.processRequest(cmd.Request)
	} else if cmd.AdminRequest != nil {
		// Process admin request.
		err = pr.processAdminRequest(cmd.AdminRequest)
	}
	return err
}

func (pr *peer) processRequest(request *raftstorepb.Request) error {
	switch request.CmdType {
	case raftstorepb.CmdType_Put:
		logger.Debugf("apply CmdType_Put request: %+v", request.Put)
		modify := storage.PutData(request.Put.Key, request.Put.Value, true)
		return pr.ps.engine.WriteData(modify)
	case raftstorepb.CmdType_Delete:
		logger.Debugf("apply CmdType_Delete request: %+v", request.Delete)
		modify := storage.DeleteData(request.Delete.Key, true)
		return pr.ps.engine.WriteData(modify)
	default:
		// Unreachable branch for now.
		return nil
	}
}

func (pr *peer) processAdminRequest(request *raftstorepb.AdminRequest) error {
	switch request.CmdType {
	case raftstorepb.AdminCmdType_CompactLog:
		compactLog := request.GetCompactLog()
		applySt := pr.ps.applyState
		if compactLog.CompactIndex >= applySt.TruncatedState.Index {
			applySt.TruncatedState.Index = compactLog.CompactIndex
			applySt.TruncatedState.Term = compactLog.CompactTerm
			if err := pr.ps.writeRaftApplyState(applySt); err != nil {
				return err
			}
			go pr.gcRaftLog(pr.lastCompactedIdx+1, applySt.TruncatedState.Index+1)
			pr.lastCompactedIdx = applySt.TruncatedState.Index
		}
	default:
		// Unreachable branch for now.
	}
	return nil
}

func (pr *peer) gcRaftLog(start, end uint64) {
	entries, err := pr.ps.Entries(start, end, math.MaxUint64)
	if err != nil {
		logger.Errorf("get error when doing raft log GC, error: %s", err.Error())
		return
	}
	if err = pr.ps.deleteRaftLogEntries(entries); err != nil {
		logger.Errorf("get error when doing raft log GC, error: %s", err.Error())
	}
}

func (pr *peer) linearizableRead(key []byte) (*internal.Callback, error) {
	ctx := buildReadCtx(key)
	cb := internal.NewCallback()
	reader := &reader{
		ctx: ctx,
		cb:  cb,
	}
	logger.Debugf("receive a read request: %+v", reader)
	pr.waitReadChannel <- reader
	if err := pr.raftGroup.ReadIndex(context.TODO(), ctx); err != nil {
		return nil, err
	}
	return cb, nil
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
	value, err := pr.ps.engine.ReadData(reader.key())
	if err != nil && err != storage.ErrNotFound {
		logger.Errorf("get error when reading applied data, error: %s", err.Error())
		return
	}
	resp := internal.NewGetCmdResponse(value)
	logger.Infof("get response successfully, value: %s, response: %+v", string(reader.key()), resp.GetResponse())
	reader.cb.Done(resp)
}

func (pr *peer) waitAppliedAdvance(index uint64) {
	applied := pr.ps.appliedIndex()
	if applied >= index {
		return
	}
	doneCh := make(chan struct{})
	go func() {
		for applied < index {
			time.Sleep(time.Millisecond)
			applied = pr.ps.appliedIndex()
		}
		doneCh <- struct{}{}
	}()
	// Wait for applied index >= state.Index.
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
