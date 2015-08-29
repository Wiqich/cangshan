package cache

import (
	"errors"
	"fmt"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/kv"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("KVCache", new(KVCache))
}

type KVCacheValueEncoding interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte) (interface{}, error)
}

type KVCache struct {
	KV             kv.KV
	Timeout        time.Duration
	IgnoreNilValue bool
	Encoding       KVCacheValueEncoding
}

func (cache *KVCache) Initialize() error {
	if cache.Encoding == nil {
		return errors.New("no Encoding")
	}
	if cache.KV == nil {
		return errors.New("no KV")
	}
	return nil
}

func (cache *KVCache) Set(key string, value interface{}) error {
	logging.Debug("KVCache.Set: %s -> %v", key, value)
	if value == nil && cache.IgnoreNilValue {
		return nil
	}
	if content, err := cache.Encoding.Encode(value); err != nil {
		return fmt.Errorf("encode value fail: %s", err)
	} else if err := cache.KV.Set(key, content, cache.Timeout); err != nil {
		return fmt.Errorf("set to kv fail: %s", err)
	}
	return nil
}

func (cache *KVCache) Get(key string) (interface{}, bool, error) {
	logging.Debug("KVCache.Get: %s", key)
	if content, err := cache.KV.Get(key); err != nil {
		return nil, false, fmt.Errorf("get from kv fail: %s", err)
	} else if value, err := cache.Encoding.Decode(content); err != nil {
		return nil, true, fmt.Errorf("decode value fail: %s", err)
	} else if value == "" && cache.IgnoreNilValue {
		return nil, false, nil
	} else {
		return value, true, nil
	}
}

func (cache *KVCache) Remove(key string) error {
	return cache.KV.Remove(key)
}
