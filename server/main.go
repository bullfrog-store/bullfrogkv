package main

import (
	"flag"
)

func main() {
	// Start command:
	//     ./main -id=1 -path=/tmp/bullfrog/node1 -addr=127.0.0.1:8080
	// Maybe we will use a file as a configuration later.

	storeId := flag.Uint64("id", 0, "storage ID for grpc communication")
	dataPath := flag.String("path", "/tmp/bullfrog", "based data path prefix")
	addr := flag.String("addr", ":8080", "the port of receiving request")
	flag.Parse()

	engine := newRaftEngine(*storeId, *dataPath)
	ge := router(engine)
	ge.Run(*addr)
}
