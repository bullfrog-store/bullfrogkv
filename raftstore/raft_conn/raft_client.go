package raft_conn

import (
	"bullfrogkv/logger"
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

func (rc *RaftClient) GetRaftConn(addr string) (*raftConn, error) {
	// grpc connection already established
	rc.RLock()
	conn, ok := rc.conns[addr]
	if ok {
		rc.RUnlock()
		return conn, nil
	}
	rc.RUnlock()
	//establish grpc connection
	newConn, err := newRaftConn(addr)
	if err != nil {
		return nil, err
	}
	rc.Lock()
	defer rc.Unlock()
	if oldConn, ok := rc.conns[addr]; ok {
		oldConn.cancel()
	}
	rc.conns[addr] = newConn
	return newConn, nil
}

func (rc *RaftClient) Send(addr string, msg *raftstorepb.RaftMsgReq) error {
	conn, err := rc.GetRaftConn(addr)
	if err != nil {
		return err
	}
	err = conn.Send(msg)
	if err == nil {
		return nil
	}
	logger.Warnf("meet error when message send: %+v", err)

	rc.Lock()
	conn.Stop()
	delete(rc.conns, addr)
	rc.Unlock()

	conn, err = rc.GetRaftConn(addr)
	if err != nil {
		return err
	}
	return conn.Send(msg)
}

func newRaftConn(addr string) (*raftConn, error) {
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
	return &raftConn{
		cancel: cancel,
		ctx:    ctx,
		stream: stream,
	}, nil
}

func (c *raftConn) Send(msg *raftstorepb.RaftMsgReq) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()
	return c.stream.Send(msg)
}

func (c *raftConn) Stop() {
	c.cancel()
}
