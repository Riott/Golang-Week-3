package server

import (
	"bytes"
	"io"
	"testing"
)

func TestReadCommandHeader(t *testing.T) {
	buffer := bytes.NewBuffer([]byte("put13key212stored value"))
	cmd, err := readCommandHeader(buffer)
	if err != nil {
		t.Error("something went wrong reading the command")
	}
	if cmd != commandPut {
		t.Error("command parsing failed")
	}
}

func TestServerPutHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte("put13key212stored value"))
	_, err := readCommandHeader(buffer)

	if err != nil {
		t.Error("something went wrong reading the command")
	}
	serverPutHandler(buffer, true)

	response := make([]byte, 3)
	io.ReadFull(buffer, response)

	if string(response) != "ack" {
		t.Error("server put handler failed")
	}
	// break first arg
	buffer = bytes.NewBuffer([]byte("put93key212stored value"))
	readCommandHeader(buffer)
	serverPutHandler(buffer, true)
	response = make([]byte, 3)
	buffer.Next(10)
	io.ReadFull(buffer, response)

	if string(response) != "err" {
		t.Error("should have got an error from reading args")
	}

	// break second arg
	buffer = bytes.NewBuffer([]byte("put13key312stored value"))

	readCommandHeader(buffer)
	serverPutHandler(buffer, true)

	response = make([]byte, 3)
	buffer.Next(11)
	io.ReadFull(buffer, response)

	if string(response) != "err" {
		t.Error("should have got an error from reading args")
	}
}

func TestServerGetHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte("put13key212stored value"))
	_, err := readCommandHeader(buffer)

	if err != nil {
		t.Error("something went wrong reading the command")
	}
	serverPutHandler(buffer, true)

	response := make([]byte, 3)
	io.ReadFull(buffer, response)

	if string(response) != "ack" {
		t.Error("server put handler failed")
	}

	buffer = bytes.NewBuffer([]byte("get13key"))

	_, err = readCommandHeader(buffer)
	if err != nil {
		t.Error("something went wrong reading the command")
	}
	serverGetHandler(buffer)
	response = make([]byte, 18)
	io.ReadFull(buffer, response)

	if string(response) != "val212stored value" {
		t.Error("server get handler failed")
	}

	//break argument
	buffer = bytes.NewBuffer([]byte("get23key"))
	readCommandHeader(buffer)
	serverGetHandler(buffer)

	response = make([]byte, 3)
	buffer.Next(2)
	io.ReadFull(buffer, response)

	if string(response) != "err" {
		t.Error("should have got an error from reading args")
	}
	// try to get a non existing key
	buffer = bytes.NewBuffer([]byte("get14key1"))
	readCommandHeader(buffer)
	serverGetHandler(buffer)

	response = make([]byte, 3)
	io.ReadFull(buffer, response)

	if string(response) != "nil" {
		t.Error("should have received nil trying to query non existing key")
	}
}

func TestServerDeleteHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte("put13key212stored value"))
	_, err := readCommandHeader(buffer)

	if err != nil {
		t.Error("something went wrong reading the command")
	}

	serverPutHandler(buffer, true)

	response := make([]byte, 3)
	io.ReadFull(buffer, response)

	if string(response) != "ack" {
		t.Error("server put handler failed")
	}

	buffer = bytes.NewBuffer([]byte("del13key"))

	_, err = readCommandHeader(buffer)
	if err != nil {
		t.Error("something went wrong reading the command")
	}
	serverDeleteHandler(buffer, true)
	response = make([]byte, 3)
	io.ReadFull(buffer, response)

	if string(response) != "ack" {
		t.Error("server delete handler failed")
	}

	// break the argument
	buffer = bytes.NewBuffer([]byte("del23key"))
	readCommandHeader(buffer)
	serverDeleteHandler(buffer, true)

	response = make([]byte, 3)
	buffer.Next(2)
	io.ReadFull(buffer, response)

	if string(response) != "err" {
		t.Error("should have got an error from reading args")
	}
}

func TestReadArgument(t *testing.T) {
	buffer := bytes.NewBuffer([]byte("13key"))
	value, err := readArgument(buffer)

	if err != nil {
		t.Error("something went wrong reading the argument")
	}

	if value != "key" {
		t.Error("incorrect value receieved, expected `key`")
	}

	// lets break it
	buffer = bytes.NewBuffer([]byte("23key"))
	_, err = readArgument(buffer)

	if err == nil {
		t.Error("should have thrown an err as argument was invalid")
	}

	buffer = bytes.NewBuffer([]byte("!^123£key"))
	_, err = readArgument(buffer)

	if err == nil {
		t.Error("should have thrown an err as argument was invalid")
	}

}

func TestSendValue(t *testing.T) {
	value := "im a little value short and stout"

	buffer := bytes.NewBuffer([]byte(""))

	sendValue(value, buffer)

	response := make([]byte, 39)
	io.ReadFull(buffer, response)

	if string(response) != "val233im a little value short and stout" {
		t.Error("incorrect value receieved, expected `val233im a little value short and stout`")
	}

}

func TestHandleServerCommand(t *testing.T) {
	putBuffer := bytes.NewBuffer([]byte("put13key212stored value"))
	handle(putBuffer, true)

	getBuffer := bytes.NewBuffer([]byte("get13key"))
	handle(getBuffer, true)

	deleteBuffer := bytes.NewBuffer([]byte("del13key"))
	handle(deleteBuffer, true)

	byeBuffer := bytes.NewBuffer([]byte("bye"))
	handle(byeBuffer, true)
}

var value = ""

func BenchmarkReadArgument(b *testing.B) {
	v := ""
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer([]byte("212stored value"))
		v, _ = readArgument(buffer)
	}
	value = v
}

func BenchmarkSendValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer([]byte("212stored value"))
		sendValue("myvalue", buffer)
	}
}

var command Command

func BenchmarkReadCommandHeader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer([]byte("put13key212stored"))
		command, _ = readCommandHeader(buffer)
	}
}

func BenchmarkServerPutHandler(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer([]byte("put13key212stored"))
		serverPutHandler(buffer, true)
	}
}

func BenchmarkServerGetHandler(b *testing.B) {
	buffer := bytes.NewBuffer([]byte("put13key212stored"))
	serverPutHandler(buffer, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer([]byte("get13key"))
		serverGetHandler(buffer)
	}
}

func BenchmarkServerDeleteHandler(b *testing.B) {
	buffer := bytes.NewBuffer([]byte("put13key212stored"))
	serverPutHandler(buffer, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer([]byte("del13key"))
		serverDeleteHandler(buffer, true)
	}
}

func BenchmarkHandle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer([]byte("put13key212stored"))
		handle(buffer, true)
	}
}
