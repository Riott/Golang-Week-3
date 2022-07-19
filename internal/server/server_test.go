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
	serverPutHandler(buffer)

	response := make([]byte, 3)
	io.ReadFull(buffer, response)

	if string(response) != "ack" {
		t.Error("server put handler failed")
	}
	// break first arg
	buffer = bytes.NewBuffer([]byte("put93key212stored value"))
	readCommandHeader(buffer)
	serverPutHandler(buffer)
	response = make([]byte, 3)
	io.ReadFull(buffer, response)
	if string(response) != "err" {
		t.Error("should have got an error from reading args")
	}

	// break second arg
	buffer = bytes.NewBuffer([]byte("put13key312stored value"))
	readCommandHeader(buffer)
	serverPutHandler(buffer)
	response = make([]byte, 3)
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
	serverPutHandler(buffer)

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
	buffer = bytes.NewBuffer([]byte("put13key312stored value"))
	readCommandHeader(buffer)
	serverPutHandler(buffer)
	response = make([]byte, 3)
	io.ReadFull(buffer, response)
	if string(response) != "err" {
		t.Error("should have got an error from reading args")
	}
}

func TestServerDeleteHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte("get13key"))
	_, err := readCommandHeader(buffer)

	if err != nil {
		t.Error("something went wrong reading the command")
	}

	serverPutHandler(buffer)

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
	serverDeleteHandler(buffer)
	response = make([]byte, 3)
	io.ReadFull(buffer, response)

	if string(response) != "ack" {
		t.Error("server delete handler failed")
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
	handle(putBuffer)

	getBuffer := bytes.NewBuffer([]byte("get13key"))
	handle(getBuffer)

	deleteBuffer := bytes.NewBuffer([]byte("del13key"))
	handle(deleteBuffer)

	byeBuffer := bytes.NewBuffer([]byte("bye"))
	handle(byeBuffer)
}
