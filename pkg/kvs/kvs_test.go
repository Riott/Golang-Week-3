package kvs_test

import (
	"tcpkvs/pkg/kvs"
	"testing"
)

var store *kvs.KVStore

func TestPut(t *testing.T) {
	err := store.Put("key1", "1234")

	if err != nil {
		t.Errorf("An error occured trying to put a key into the store")
	}

	_, err = store.Get("key1")

	if err != nil {
		t.Errorf("Key: key1 should be present in the kvs")

	}
}

func TestGet(t *testing.T) {
	store.Put("key1", "1234")

	_, err := store.Get("key1")

	if err != nil {
		t.Errorf("Something went wrong trying to retrieve the key")
	}
}

func TestDelete(t *testing.T) {
	store.Put("key1", "1234")
	store.Delete("key1")
	_, err := store.Get("key1")
	if err != nil {
		return
	}
	t.Error("should have returned key not found")

}

func BenchmarkPut(b *testing.B) {
	for i := 0; i < b.N; i++ {
		store.Put("key1", "1234")
	}
}

var value string

func BenchmarkGet(b *testing.B) {
	store.Put("key1", "1234")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value, _ = store.Get("key1")

	}
}

func BenchmarkDelete(b *testing.B) {
	store.Put("key1", "1234")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Delete("key1")
	}
}

func TestMain(m *testing.M) {
	go func() {
		store = kvs.InitStore()
	}()
	m.Run()
}
