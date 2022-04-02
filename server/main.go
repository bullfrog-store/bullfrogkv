package main

import (
	"flag"
)

func main() {
	// Start command:
	//     c
	// Maybe we will use a file as a configuration later.

	storeId := flag.Uint64("i", 0, "storage ID for grpc communication")
	dataPath := flag.String("p", "/tmp/bullfrog", "based data path prefix")
	addr := flag.String("a", ":8080", "the port of receiving request")
	flag.Parse()

	engine := newRaftEngine(*storeId, *dataPath)
	ge := router(engine)
	ge.Run(*addr)
}
