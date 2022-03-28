package main

import (
	"flag"
)

func main() {
	kvPath := flag.String("k", "kv", "comma kv storage path")
	raftPath := flag.String("r", "raft", "comma raft storage path")
	addr := flag.String("a", ":8080", "comma port")
	flag.Parse()

	kve := newKVEngine(*kvPath, *raftPath)
	defer kve.kve.Close()
	ge := router(kve)
	ge.Run(*addr)
}
