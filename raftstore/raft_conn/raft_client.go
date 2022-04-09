package raft_conn

import (
	"bullfrogkv/raftstore/raftstorepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"sync"
	"time"
)

type RaftClient struct {
	sync.RWMutex
	conns map[string]*raftConn
}

type ToSendReq struct {
	Msg  *raftstorepb.RaftMsgReq
	Addr string
}

type raftConn struct {
	streamMu sync.Mutex
	stream   raftstorepb.Message_RaftMessageClient
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewRaftClient() *RaftClient {
	return &RaftClient{
		conns: make(map[string]*raftConn),
	}
}

func (rc *RaftClient) GetRaftConn(addr string) (*raftConn, bool) {
	// grpc connection already established
	rc.RLock()
	defer rc.RUnlock()
	conn, ok := rc.conns[addr]
	return conn, ok
}

func (rc *RaftClient) DialAndSend(addr string, msg *raftstorepb.RaftMsgReq) {
	conn, err := rc.newRaftConn(addr)
	if err != nil {
		return
	}

	err = conn.Send(msg)
	if err == nil {
		if oldConn, ok := rc.GetRaftConn(addr); ok {
			oldConn.Stop()
		}
		rc.Lock()
		rc.conns[addr] = conn
		rc.Unlock()
	}
}

func (rc *RaftClient) newRaftConn(addr string) (*raftConn, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		// reconnect after disconnection
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             1 * time.Second,
			PermitWithoutStream: true,
		}))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := raftstorepb.NewMessageClient(conn).RaftMessage(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	c := &raftConn{
		stream: stream,
		ctx:    ctx,
		cancel: cancel,
	}
	return c, nil
}

func (c *raftConn) Send(msg *raftstorepb.RaftMsgReq) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()
	return c.stream.Send(msg)
}

func (c *raftConn) Stop() {
	c.cancel()
}
