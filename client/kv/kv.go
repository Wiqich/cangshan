package kv

import (
	"errors"
	"time"
)

var (
	// ErrNotFound present
	ErrNotFound = errors.New("not found")
)

type Item struct {
	Key    string
	Value  []byte
	MaxAge time.Duration
}

// KV is general key-value storage client interface
type KV interface {
	Ping() error
	Get(key string) ([]byte, error)
	GetMulti(keys ...string) (map[string][]byte, error)
	Set(key string, value []byte, maxage time.Duration) error
	SetMulti(items []Item) error
	Remove(key string) error
}
