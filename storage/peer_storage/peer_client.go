package peer_storage

import (
	"bullfrogkv/raftstore/raftstorepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"sync"
	"time"
)

type PeerClient struct {
	sync.RWMutex
	conns map[string]*peerConn
}

type peerConn struct {
	streamMu sync.Mutex
	stream   raftstorepb.Message_SendRaftMessageClient
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewPeerClient() *PeerClient {
	return &PeerClient{
		conns: make(map[string]*peerConn),
	}
}

func (p *PeerClient) GetPeerConn(addr string) (*peerConn, error) {
	// grpc connection already established
	p.RLock()
	conn, ok := p.conns[addr]
	if ok {
		p.RUnlock()
		return conn, nil
	}

	//establish grpc connection
	newConn, err := newPeerConn(addr)
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

func newPeerConn(addr string) (*peerConn, error) {
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
	stream, err := raftstorepb.NewMessageClient(conn).SendRaftMessage(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	return &peerConn{
		cancel: cancel,
		ctx:    ctx,
		stream: stream,
	}, nil
}

func (c *peerConn) Send(msg *raftstorepb.RaftMessage) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()
	return c.stream.Send(msg)
}
