package raftstore

import (
	"bullfrogkv/raftstore/raftstorepb"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"sync"
	"testing"
)

//func newTestPeerStorage(t *testing.T) *peerStorage {
//	storage := newPeerStorage("log_test")
//	return storage
//}

func TestNewPeer(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)
	go newTestPeer(1, 2, wg)
	go newTestPeer(2, 3, wg)
	go newTestPeer(3, 1, wg)
	wg.Wait()
}

func newTestPeer(id uint64, l uint64, wg sync.WaitGroup) {
	path := fmt.Sprintf("log_test/p%d", id)
	p := newPeer(id, path)
	lis, err := net.Listen("tcp", peerMap[l])
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d listen %v success\n", id, peerMap[l])
	g := grpc.NewServer()
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
