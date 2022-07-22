package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"tcpkvs/internal/logger"
	"tcpkvs/pkg/kvs"
)

const (
	commandPut    = "put"
	commandGet    = "get"
	commandDelete = "del"
	commandBye    = "bye"
	commandLength = 3
)

type Command string

var ServerLogger = logger.InitLogger()
var store = kvs.InitStore()

// notest
func Listen(listenAddress, udpUpdateAddress string) {
	listener, err := net.Listen("tcp", listenAddress)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Server listening at %v\n", listenAddress)
	go listenForBroadcast(udpUpdateAddress)
	defer func() { _ = listener.Close() }()

	for {
		connection, err := listener.Accept()

		if err != nil {
			break
		}

		go func() {
			ServerLogger.Printf("%s: Client Connected", connection.RemoteAddr())
			handle(connection, false)
			connection.Close()
		}()
	}
}

func listenForBroadcast(address string) {
	fmt.Printf("Server listening for updates at %v\n", address)

	packetConn, err := net.ListenPacket("udp4", address)

	if err != nil {
		return
	}

	defer packetConn.Close()

	for {
		buffer := make([]byte, 1024)
		n, addr, err := packetConn.ReadFrom(buffer)

		if err != nil {
			return
		}

		if err != nil {
			return
		}

		localIP := getLocalIP()
		updateIP := addr.String()

		if strings.Contains(updateIP, localIP) {
			fmt.Printf("%s ignoring update from self %s\n", addr, buffer[:n])
			ServerLogger.Printf("%s ignoring update from self %s\n", addr, buffer[:n])
			continue
		}

		fmt.Printf("%s server sent update %s\n", addr, buffer[:n])
		ServerLogger.Printf("%s server sent update %s\n", addr, buffer[:n])
		cmdBuffer := bytes.NewBuffer(buffer[:n])

		handle(cmdBuffer, true)
	}

}

func getLocalIP() string {
	list, err := net.Interfaces()

	if err != nil {
		return ""
	}

	for _, iface := range list {
		if iface.Name == "Valhalla" {
			addrs, _ := iface.Addrs()
			return addrs[1].String()[:len(addrs[1].String())-3]
		}
	}

	return ""
}

func handle(connection io.ReadWriter, udpUpdate bool) {
	for {
		cmd, err := readCommandHeader(connection)

		if err != nil {
			return
		}

		if finish := handleServerCommand(cmd, connection, udpUpdate); finish {
			return
		}
	}
}

func readCommandHeader(connection io.ReadWriter) (Command, error) {
	buffer := make([]byte, commandLength)
	_, err := io.ReadFull(connection, buffer)

	if err != nil {
		return "", err
	}

	return Command(buffer), nil
}

func handleServerCommand(cmd Command, connection io.ReadWriter, udpUpdate bool) bool {
	ServerLogger.Printf("Command Receieved: %s\n", cmd)

	switch cmd {

	case commandPut:
		serverPutHandler(connection, udpUpdate)

	case commandGet:
		serverGetHandler(connection)

	case commandDelete:
		serverDeleteHandler(connection, udpUpdate)

	case commandBye:
		return true
	}

	return false
}

func serverPutHandler(connection io.ReadWriter, udpUpdate bool) {
	key, err := readArgument(connection)

	if err != nil {
		fmt.Fprintf(connection, "err")
		return
	}

	value, err := readArgument(connection)

	if err != nil {
		fmt.Fprintf(connection, "err")
		return
	}

	store.Put(key, value)
	fmt.Fprintf(connection, "ack")

	if !udpUpdate {
		buffer := rebuildCommandForUpdate(commandPut, key, value)
		sendBroadcastCommand(buffer)
	}
}

func serverGetHandler(connection io.ReadWriter) {
	key, err := readArgument(connection)

	if err != nil {
		fmt.Fprintf(connection, "err")
		return
	}

	value, err := store.Get(key)

	if err == kvs.ErrKeyNotFound {
		fmt.Fprintf(connection, "nil")
		return
	}

	sendValue(value, connection)
}

func serverDeleteHandler(connection io.ReadWriter, udpUpdate bool) {
	key, err := readArgument(connection)

	if err != nil {
		fmt.Fprintf(connection, "err")
		return
	}

	store.Delete(key)
	fmt.Fprintf(connection, "ack")

	if !udpUpdate {
		buffer := rebuildCommandForUpdate(commandDelete, key, "")
		sendBroadcastCommand(buffer)
	}
}

func readArgument(connection io.ReadWriter) (string, error) {
	// first byte determines how many bytes to read next
	argLengthBuffer := make([]byte, 1)
	io.ReadFull(connection, argLengthBuffer)

	argLengthSize, err := strconv.Atoi(string(argLengthBuffer[0]))

	if err != nil {
		return "", err
	}

	// read the length from the remaining buffer to determine value size
	argValueLengthBuffer := make([]byte, argLengthSize)
	io.ReadFull(connection, argValueLengthBuffer)

	argValueLength, err := strconv.Atoi(string(argValueLengthBuffer))

	if err != nil {
		return "", err
	}

	// read the value and return
	argBuffer := make([]byte, argValueLength)
	io.ReadFull(connection, argBuffer)

	arg := string(argBuffer)
	ServerLogger.Println("received arg", arg)
	return arg, nil
}

func sendValue(value string, connection io.ReadWriter) {
	length := len(value)
	digitLength := len(fmt.Sprintf("%d", length))

	fmt.Fprintf(connection, "val%d%d%s", digitLength, length, value)
	ServerLogger.Printf("sent to client: val%d%d%s", digitLength, length, value)
}

func rebuildCommandForUpdate(cmd Command, arg1, arg2 string) io.ReadWriter {
	switch cmd {

	case commandPut:
		keyLength := len(arg1)
		keyDigitLength := len(fmt.Sprintf("%d", keyLength))

		valueLength := len(arg2)
		valueDigitLength := len(fmt.Sprintf("%d", valueLength))

		commandBuffer := bytes.NewBuffer([]byte(fmt.Sprintf("%s%d%d%s%d%d%s", cmd, keyDigitLength, keyLength, arg1, valueDigitLength, valueLength, arg2)))
		return commandBuffer

	case commandDelete:
		keyLength := len(arg1)
		keyDigitLength := len(fmt.Sprintf("%d", keyLength))

		commandBuffer := bytes.NewBuffer([]byte(fmt.Sprintf("%s%d%d%s", cmd, keyDigitLength, keyLength, arg1)))
		return commandBuffer

	}

	return nil
}

// notest
func sendBroadcastCommand(connection io.ReadWriter) {
	packetConn, err := net.ListenPacket("udp4", ":2307")

	if err != nil {
		return
	}

	defer packetConn.Close()

	addr, err := net.ResolveUDPAddr("udp4", "192.168.1.255:2306")
	if err != nil {
		return
	}

	data, err := io.ReadAll(connection)
	if err != nil {
		return
	}
	_, err = packetConn.WriteTo([]byte(data), addr)
	if err != nil {
		return
	}
}
