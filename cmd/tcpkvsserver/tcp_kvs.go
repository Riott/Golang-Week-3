package main

import (
	"flag"
	"fmt"
	"tcpkvs/internal/server"
)

var portFlag = flag.Int("port", 27015, "port number the store will listen on")
var updateFlag = flag.Int("udp", 2306, "port number broadcast updates are listened for.")

func main() {
	flag.Parse()
	port := uint16(*portFlag)
	updatePort := uint16(*updateFlag)
	listenAddress := fmt.Sprintf(":%d", port)
	udpUpdateAddress := fmt.Sprintf(":%d", updatePort)
	server.Listen(listenAddress, udpUpdateAddress)
}
