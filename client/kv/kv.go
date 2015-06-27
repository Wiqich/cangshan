package kv

import (
	"time"
)

type KV interface {
	Ping() error
	Get(string) ([]byte, error)
	Set(string, []byte, time.Duration) error
}
