package kvs

import (
	"errors"
)

type KVStore struct {
	store    map[string]string
	requests chan *readWriteOperation
}

type readWriteOperation struct {
	key       string
	value     string
	operation Operation
	response  chan Response
}

type Operation string

type Response struct {
	responseCode  ResponseCode
	responseValue string
}

type ResponseCode int

const (
	StorePut    = "PUT"
	StoreGet    = "GET"
	StoreDelete = "DELETE"
)

const (
	ResponseOk         = 0
	ResponseBadRequest = 418
	ResponseNotFound   = 404
)

var ErrKeyNotFound = errors.New("the key wasn't present in the store")
var ErrPutFailed = errors.New("something went wrong trying to put key into store")

func InitStore() *KVStore {
	keyValueStore := KVStore{store: map[string]string{}, requests: make(chan *readWriteOperation)}
	go func() {
		for {
			select {
			case request := <-keyValueStore.requests:

				switch request.operation {

				case StorePut:
					keyValueStore.store[request.key] = request.value
					request.response <- Response{responseCode: ResponseOk, responseValue: request.key}

				case StoreGet:
					value, found := keyValueStore.store[request.key]

					if found {
						request.response <- Response{responseCode: ResponseOk, responseValue: value}
						continue
					}

					request.response <- Response{responseCode: ResponseNotFound, responseValue: request.key}
				case StoreDelete:
					delete(keyValueStore.store, request.key)
					request.response <- Response{responseCode: ResponseOk, responseValue: request.key}

				}
			}
		}
	}()
	return &keyValueStore
}

func (store *KVStore) Put(key, value string) error {
	put := &readWriteOperation{key: key, value: value, operation: StorePut, response: make(chan Response)}
	store.requests <- put
	response := <-put.response

	if response.responseCode != ResponseOk {
		return ErrPutFailed
	}

	return nil
}

func (store KVStore) Get(key string) (string, error) {
	get := &readWriteOperation{key: key, operation: StoreGet, response: make(chan Response)}
	store.requests <- get

	response := <-get.response

	if response.responseCode != ResponseOk {
		return "", ErrKeyNotFound
	}

	return response.responseValue, nil
}

func (store *KVStore) Delete(key string) {
	delete := &readWriteOperation{key: key, operation: StoreDelete, response: make(chan Response)}
	store.requests <- delete
	<-delete.response
}
