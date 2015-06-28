package kv

import (
	"errors"
	"time"
)

var (
	// ErrNotFound present
	ErrNotFound = errors.New("not found")
)

// KV is general key-value storage client interface
type KV interface {
	Ping() error
	Get(string) ([]byte, error)
	Set(string, []byte, time.Duration) error
}
