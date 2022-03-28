package main

import (
	"bullfrogkv/storage"
	"flag"
)

func main() {

	kvPath := flag.String("k", storage.KvPath, "comma kv storage path")
	raftPath := flag.String("r", storage.RaftPath, "comma raft storage path")
	addr := flag.String("a", ":8080", "comma port")
	flag.Parse()

	kve := newKVEngine(*kvPath, *raftPath)
	defer kve.kve.Close()
	ge := router(kve)
	ge.Run(*addr)
}
