package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
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

func Listen(address string) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Server listening at %v\n", address)

	defer func() { _ = listener.Close() }()

	for {
		connection, err := listener.Accept()

		if err != nil {
			break
		}

		go func() {
			ServerLogger.Printf("%s: Client Connected", connection.RemoteAddr())
			handle(connection)
			connection.Close()
		}()
	}
}

func handle(connection io.ReadWriter) {
	for {
		cmd, err := readCommandHeader(connection)

		if err != nil {
			return
		}

		if finish := handleServerCommand(cmd, connection); finish {
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

func handleServerCommand(cmd Command, connection io.ReadWriter) bool {
	ServerLogger.Printf("Command Receieved: %s\n", cmd)

	switch cmd {

	case commandPut:
		serverPutHandler(connection)

	case commandGet:
		serverGetHandler(connection)

	case commandDelete:
		serverDeleteHandler(connection)

	case commandBye:
		return true
	}
	return false
}

func serverPutHandler(connection io.ReadWriter) {
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

func serverDeleteHandler(connection io.ReadWriter) {
	key, err := readArgument(connection)

	if err != nil {
		fmt.Fprintf(connection, "err")
		return
	}

	store.Delete(key)
	fmt.Fprintf(connection, "ack")
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
