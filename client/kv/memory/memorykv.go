package memorykv

import (
	"container/list"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/client/kv"
)

// A MemoryKV provide in-memory kv storage. A MemoryKV can limit total data size, and discard old
// data for saving new data.
type MemoryKV struct {
	Capacity uint64
	data     map[string][]byte
	keys     *list.List
	size     uint64
	keysLock sync.Mutex
}

// Initialize the MemoryKV module for application
func (k *MemoryKV) Initialize() error {
	k.data = make(map[string][]byte)
	k.keys = list.New()
	return nil
}

// Ping always success
func (k *MemoryKV) Ping() error {
	return nil
}

// Get value with specified key
func (k *MemoryKV) Get(key string) ([]byte, error) {
	value := k.data[key]
	if value == nil {
		return nil, kv.ErrNotFound
	}
	return value, nil
}

// Set value with specified key
func (k *MemoryKV) Set(key string, value []byte, maxAge time.Duration) error {
	// ignore empty value
	if value == nil {
		return nil
	}
	//
	if k.Capacity-k.size < uint64(len(value)) {
	}
	return nil
}
