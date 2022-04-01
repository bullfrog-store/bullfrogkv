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
func (p *RaftClient) GetClientConn(addr string) (*raftConn, error) {
	// grpc connection already established
	p.RLock()
	conn, ok := p.conns[addr]
	if ok {
		p.RUnlock()
		return conn, nil
	}
	p.RUnlock()
	//establish grpc connection
	newConn, err := newClientConn(addr)
	if err != nil {
		return nil, err
	}
	p.Lock()
	defer p.Unlock()
	if oldConn, ok := p.conns[addr]; ok {
		newConn.cancel()
		return oldConn, err
	}
	p.conns[addr] = newConn
	return newConn, nil
}

func newClientConn(addr string) (*raftConn, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		// reconnect after disconnection
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                3 * time.Second,
			Timeout:             60 * time.Second,
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
