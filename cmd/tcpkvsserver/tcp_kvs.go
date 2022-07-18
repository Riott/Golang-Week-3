package main

import (
	"flag"
	"fmt"
	"tcpkvs/internal/server"
)

var portFlag = flag.Int("port", 27015, "port number the store will listen on")

func main() {
	flag.Parse()
	port := uint16(*portFlag)
	listenAddress := fmt.Sprintf(":%d", port)
	server.Listen(listenAddress)
}
