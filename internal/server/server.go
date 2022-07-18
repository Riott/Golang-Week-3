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
	defer func() { _ = listener.Close() }()
	for {
		connection, err := listener.Accept()
		if err != nil {
			break
		}
		go handle(connection)

	}
}

func handle(connection net.Conn) {
	ServerLogger.Printf("%s: Client Connected", connection.RemoteAddr())
	for {
		cmd, err := readCommandHeader(connection)
		if err == nil {
			handleServerCommand(cmd, connection)

		}
	}
}

func readCommandHeader(connection net.Conn) (Command, error) {
	buffer := make([]byte, commandLength)
	_, err := io.ReadFull(connection, buffer)

	if err != nil {
		return "", err
	}

	return Command(buffer), nil
}

func handleServerCommand(cmd Command, connection net.Conn) {
	ServerLogger.Printf("Command Receieved: %s\n", cmd)

	switch cmd {
	case commandPut:
		serverPutHandler(connection)

	case commandGet:
		serverGetHandler(connection)

	case commandDelete:
		serverDeleteHandler(connection)
	case commandBye:
		serverByeHandler(connection)

	}
}

func serverPutHandler(connection net.Conn) {
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

func serverGetHandler(connection net.Conn) {
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

func serverDeleteHandler(connection net.Conn) {
	key, err := readArgument(connection)

	if err != nil {
		fmt.Fprintf(connection, "err")
		return
	}

	err = store.Delete(key)

	if err != nil {
		fmt.Fprintf(connection, "err")
		return
	}

	fmt.Fprintf(connection, "ack")
}

func serverByeHandler(connection net.Conn) {
	connection.Close()
	ServerLogger.Printf("%s: Client Connection closed forcibly by host", connection.RemoteAddr())

}

func readArgument(connection net.Conn) (string, error) {
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

func sendValue(value string, connection net.Conn) {
	length := len(value)
	digitLength := countDigits(length)
	fmt.Fprintf(connection, "val%d%d%s", digitLength, length, value)
	ServerLogger.Printf("sent to client: val%d%d%s", digitLength, length, value)
}

// there has to be a better way
func countDigits(number int) int {
	if number < 10 {
		return 1
	} else {
		return 1 + countDigits(number/10)
	}
}
