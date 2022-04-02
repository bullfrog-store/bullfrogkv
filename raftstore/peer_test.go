package raftstore

import (
	"bullfrogkv/raftstore/raftstorepb"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

//func newTestPeerStorage(t *testing.T) *peerStorage {
//	storage := newPeerStorage("log_test")
//	return storage
//}

func TestNewPeer1(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	go newTestPeer(1, wg)
	wg.Wait()
}

func TestNewPeer2(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go newTestPeer(2, wg)
	wg.Wait()
}

func TestNewPeer3(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go newTestPeer(3, wg)
	wg.Wait()
}

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second,
	PermitWithoutStream: true,
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle: 15 * time.Second,
	Time:              5 * time.Second,
	Timeout:           1 * time.Second,
}

func newTestPeer(id uint64, wg sync.WaitGroup) {
	path := fmt.Sprintf("log_test/p%d", id)
	p := newPeer(id, path)
	lis, err := net.Listen("tcp", peerMap[id])
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d listen %v success\n", id, peerMap[id])
	g := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
	raftstorepb.RegisterMessageServer(g, p.router.raftServer)
	g.Serve(lis)
	wg.Done()
}

func TestCD(t *testing.T) {
	os.RemoveAll("log_test/p1")
	os.RemoveAll("log_test/p2")
	os.RemoveAll("log_test/p3")
	os.Mkdir("log_test/p1", 0777)
	os.Mkdir("log_test/p2", 0777)
	os.Mkdir("log_test/p3", 0777)
}
